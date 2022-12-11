package request

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"time"
)

type Options struct {
	Headers   map[string]string
	WithTrace bool
}

type Option func(options *Options)

func Headers(headers map[string]string) Option {
	return func(options *Options) {
		options.Headers = headers
	}
}

func WithTrace(trace bool) Option {
	return func(options *Options) {
		options.WithTrace = trace
	}
}

func Post(url string, body []byte, timeout time.Duration, retries int, setters ...Option) (*http.Response, []byte, error) {

	args := &Options{}

	for _, setter := range setters {
		setter(args)
	}

	req, err := http.NewRequest(http.MethodPost, url, ioutil.NopCloser(bytes.NewBuffer(body)))

	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gapi-request")

	if args.Headers != nil {
		for k, v := range args.Headers {
			req.Header.Set(k, v)
		}
	}

	if args.WithTrace {
		return toRequestWithTrace(req, 5*time.Second, 3)
	}

	return toRequest(req, timeout, retries)
}

func Get(url string, timeout time.Duration, retries int, setters ...Option) (*http.Response, []byte, error) {
	args := &Options{}

	for _, setter := range setters {
		setter(args)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("User-Agent", "massage-request")

	if args.Headers != nil {
		for k, v := range args.Headers {
			req.Header.Set(k, v)
		}
	}

	if args.WithTrace {
		return toRequestWithTrace(req, 5*time.Second, 3)
	}

	return toRequest(req, timeout, retries)
}

func toRequest(req *http.Request, timeout time.Duration, retries int) (*http.Response, []byte, error) {
	cli := http.Client{
		Timeout: timeout,
	}
	var resp *http.Response
	var reqErr error

	for retries > 0 {
		resp, reqErr = cli.Do(req)
		if reqErr != nil {

			retries--
		} else {
			break
		}
	}

	if reqErr != nil {
		return nil, nil, reqErr
	}
	defer resp.Body.Close()

	rs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, rs, nil
}

func toRequestWithTrace(req *http.Request, _ time.Duration, _ int) (*http.Response, []byte, error) {
	var connect, dns time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) {
			dns = time.Now()
		},
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			//logger.WithField("uniqueId", requestId).
			//	WithField("module", "server_request").
			//	WithField("useTime", time.Since(dns).Seconds()).
			//	Info("dns end:", time.Now())
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			//logger.WithField("uniqueId", requestId).WithField("module", "server_request").
			//	WithField("useTime", time.Since(connect).Seconds()).
			//	Info("connect end:", time.Now())
		},

		GotFirstResponseByte: func() {
			//logger.WithField("uniqueId", requestId).WithField("module", "server_request").
			//	WithField("useTime", time.Since(start).Seconds()).
			//	Info("first response:", time.Now())
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, reqErr := http.DefaultTransport.RoundTrip(req)

	if reqErr != nil {
		return nil, nil, reqErr
	}
	defer resp.Body.Close()
	rs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, rs, nil
}