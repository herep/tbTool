package middle

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"

	"gitlab.xfq.com/tech-lab/dionysus/pkg"
)

// timeout middleware wraps the request context with a timeout
func TimeoutMiddleware() gin.HandlerFunc {
	//default < env < request_header (ps: request_header timeout args must less than env args)
	var defaultTimeout = 10
	if t := os.Getenv("GAPI_REQUEST_TIMEOUT"); t != "" {
		if s, err := strconv.Atoi(t); err == nil && s > 0 {
			defaultTimeout = s
		}
	}
	return func(c *gin.Context) {
		timeout := defaultTimeout
		if t := c.GetHeader("Request_Timeout"); t != "" {
			if s, err := strconv.Atoi(t); err == nil && s < defaultTimeout && s > 0 {
				timeout = s
			}
		}
		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*time.Duration(timeout))

		defer func() {
			// check if context timeout was reached
			if ctx.Err() == context.DeadlineExceeded {

				// write response and abort the request
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				c.Abort()
			}

			//cancel to clear resources after finished
			cancel()
		}()

		// replace request with context wrapped request
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

type Handler = func(*gin.Context) pkg.Render

func TimedHandler(handler Handler) func(c *gin.Context) {
	return func(c *gin.Context) {

		// get the underlying request context
		ctx := c.Request.Context()

		// create a done channel to tell the request it's done
		doneChan := make(chan render.Render, 1)

		// here you put the actual work needed for the request
		// and then send the doneChan with the status and body
		// to finish the request by writing the response
		go func() {
			defer func() {
				if c := recover(); c != nil {
					log.Errorf("response request panic: %v", c)
				}
			}()
			doneChan <- handler(c)
		}()

		// non-blocking select on two channels see if the request
		// times out or finishes
		select {

		// if the context is done it timed out or was canceled
		// so don't return anything
		case <-ctx.Done():
			return

		// if the request finished then finish the request by
		// writing the response
		case res := <-doneChan:
			if res == nil {
				if c.Request.Method != "HEAD" {
					c.JSON(500, nil)
				}
			} else {
				c.Render(c.Writer.Status(), res)
			}

		}
	}
}
