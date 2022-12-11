package middle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.xfq.com/tech-lab/dionysus/pkg/metrics"
)

// Copy form http.TimeoutHandler, enhanced timeout control ability at single request.
// https://tools.ietf.org/html/rfc2616#section-4.2
const TimeoutInContext = "request-timeout"

// TimeoutHandler returns a Handler that runs h with the given time limit.
//
// The new Handler calls h.ServeHTTP to handle each request, but if a
// call runs for longer than its time limit, the handler responds with
// a 503 Service Unavailable error and the given message in its body.
// (If msg is empty, a suitable default message will be sent.)
// After such a timeout, writes by h to its ResponseWriter will return
// ErrHandlerTimeout.
//
// TimeoutHandler supports the Pusher interface but does not support
// the Hijacker or Flusher interfaces.
func TimeoutHandler(h http.Handler, dt time.Duration, msg string, prom *metrics.Prometheus) http.Handler {
	return &timeoutHandler{
		handler: h,
		body:    msg,
		dt:      dt,

		prom: prom,
	}
}

// ErrHandlerTimeout is returned on ResponseWriter Write calls
// in handlers which have timed out.
var ErrHandlerTimeout = errors.New("http: Handler timeout")

type timeoutHandler struct {
	handler http.Handler
	body    string
	dt      time.Duration

	// When set, no context will be created and this context will
	// be used instead.
	testContext context.Context

	// ///////////////////////////////////////////////////////////////////////////////
	// Metrics
	prom *metrics.Prometheus
	// ///////////////////////////////////////////////////////////////////////////////
}

func (h *timeoutHandler) errorBody() string {
	if h.body != "" {
		return h.body
	}
	return "<html><head><title>Timeout</title></head><body><h1>Timeout</h1></body></html>"
}

func (h *timeoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := h.testContext
	if ctx == nil {
		var cancelCtx context.CancelFunc

		// ///////////////////////////////////////////////////////////////////////////////
		//  The timeout logic has been changed.
		//  ctx, cancelCtx = context.WithTimeout(r.Context(), h.dt)
		timeout := h.dt
		if i, err := timeoutDecode(r.Header.Get(TimeoutInContext)); err == nil {
			timeout = i
		}
		ctx, cancelCtx = context.WithTimeout(r.Context(), timeout)
		// ///////////////////////////////////////////////////////////////////////////////
		defer cancelCtx()
	}

	r = r.WithContext(ctx)
	tw := &timeoutWriter{
		w:   w,
		h:   make(http.Header),
		req: r,
	}
	done := make(chan struct{})

	// ///////////////////////////////////////////////////////////////////////////////
	// Metrics might be panic
	defer func() {
		if p := recover(); p != nil {
			log.Error("Panic at timeout: ", p)
		}
	}()
	// ///////////////////////////////////////////////////////////////////////////////

	start := time.Now()
	panicChan := make(chan interface{}, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				panicChan <- p
			}
		}()
		h.handler.ServeHTTP(tw, r)
		close(done)
	}()
	select {
	case p := <-panicChan:
		panic(p)

	case <-done:
		tw.mu.Lock()
		defer tw.mu.Unlock()
		dst := w.Header()
		for k, vv := range tw.h {
			dst[k] = vv
		}
		if !tw.wroteHeader {
			tw.code = http.StatusOK
		}
		w.WriteHeader(tw.code)
		w.Write(tw.wbuf.Bytes()) // nolint

		// ///////////////////////////////////////////////////////////////////////////////
		// Add metrics
		if h.prom != nil {
			metrics.RecordMetrics(r, start, tw.code, tw.wbuf.Len(), h.prom.MetricsList)
		}
		// ///////////////////////////////////////////////////////////////////////////////
	case <-ctx.Done():
		tw.mu.Lock()
		defer tw.mu.Unlock()
		// ///////////////////////////////////////////////////////////////////////////////
		// change code from 503 to 504, delete the error body and add metrics
		w.WriteHeader(http.StatusGatewayTimeout)

		if h.prom != nil {
			metrics.RecordMetrics(r, start, http.StatusGatewayTimeout, 0, h.prom.MetricsList)
		}

		// io.WriteString(w, h.errorBody())
		// ///////////////////////////////////////////////////////////////////////////////
		tw.timedOut = true
	}
}

type timeoutWriter struct {
	w    http.ResponseWriter
	h    http.Header
	wbuf bytes.Buffer
	req  *http.Request

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
	code        int
}

var _ http.Pusher = (*timeoutWriter)(nil)

// Push implements the Pusher interface.
func (tw *timeoutWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := tw.w.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (tw *timeoutWriter) Header() http.Header { return tw.h }

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		// ///////////////////////////////////////////////////////////////////////////////
		// don't return error, so gin will not panic
		// return 0, ErrHandlerTimeout
		return 0, nil
		// ///////////////////////////////////////////////////////////////////////////////
	}
	if !tw.wroteHeader {
		tw.writeHeaderLocked(http.StatusOK)
	}
	return tw.wbuf.Write(p)
}

func (tw *timeoutWriter) writeHeaderLocked(code int) {
	checkWriteHeaderCode(code)

	switch {
	case tw.timedOut:
		return
	case tw.wroteHeader:
		if tw.req != nil {
			caller := relevantCaller()
			logf(tw.req, "http: superfluous response.WriteHeader call from %s (%s:%d)", caller.Function, path.Base(caller.File), caller.Line)
		}
	default:
		tw.wroteHeader = true
		tw.code = code
	}
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.writeHeaderLocked(code)
}

func checkWriteHeaderCode(code int) {
	// Issue 22880: require valid WriteHeader status codes.
	// For now we only enforce that it's three digits.
	// In the future we might block things over 599 (600 and above aren't defined
	// at https://httpwg.org/specs/rfc7231.html#status.codes)
	// and we might block under 200 (once we have more mature 1xx support).
	// But for now any three digits.
	//
	// We used to send "HTTP/1.1 000 0" on the wire in responses but there's
	// no equivalent bogus thing we can realistically send in HTTP/2,
	// so we'll consistently panic instead and help people find their bugs
	// early. (We can't return an error from WriteHeader even if we wanted to.)
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
}

// relevantCaller searches the call stack for the first function outside of net/http.
// The purpose of this function is to provide more helpful error messages.
func relevantCaller() runtime.Frame {
	pc := make([]uintptr, 16)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
	for {
		frame, more := frames.Next()
		if !strings.HasPrefix(frame.Function, "net/http.") {
			return frame
		}
		if !more {
			break
		}
	}
	return frame
}

// logf prints to the ErrorLog of the *Server associated with request r
// via ServerContextKey. If there's no associated server, or if ErrorLog
// is nil, logging is done via the log package's standard logger.
func logf(r *http.Request, format string, args ...interface{}) {
	s, _ := r.Context().Value(http.ServerContextKey).(*http.Server)
	if s != nil && s.ErrorLog != nil {
		s.ErrorLog.Printf(format, args...)
	} else {
		log.Infof(format, args...)
	}
}

// copy from micro, and add default unit logic
func timeoutDecode(s string) (time.Duration, error) {
	size := len(s)
	if len(s) == 0 {
		return 0, fmt.Errorf("timeout str is empty")
	}

	d, ok := timeoutUnitToDuration(s[size-1])
	if !ok {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("timeout is not recognized: %q", s)
		}
		// default unit: second
		return time.Duration(i) * time.Second, nil
	}

	t, err := strconv.ParseInt(s[:size-1], 10, 64)
	if err != nil {
		return 0, err
	}
	return d * time.Duration(t), nil
}

// copy from micro
func timeoutUnitToDuration(u uint8) (d time.Duration, ok bool) {
	switch u {
	case 'H':
		return time.Hour, true
	case 'M':
		return time.Minute, true
	case 'S':
		return time.Second, true
	case 'm':
		return time.Millisecond, true
	case 'u':
		return time.Microsecond, true
	case 'n':
		return time.Nanosecond, true
	default:
	}
	return
}
