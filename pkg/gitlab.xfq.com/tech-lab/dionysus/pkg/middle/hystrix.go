package middle

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/gin-gonic/gin"

	hyx "gitlab.xfq.com/tech-lab/dionysus/pkg/hystrix"
)

// RateLimitMiddle 用于服务端接口限流
func RateLimitMiddle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		commandName := hyx.JoinCommandName(strings.ToLower(ctx.Request.Method) + ctx.Request.URL.Path)
		hyx.ServerConfigureCommand(commandName)

		err := hystrix.Do(commandName, func() (err error) {
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("hystrix do run panic: %v", e)
				}
			}()

			ctx.Next()

			statusCode := ctx.Writer.Status()
			if statusCode >= http.StatusInternalServerError {
				return fmt.Errorf("status_code: %d", statusCode)
			}
			return nil
		}, func(err error) (rErr error) {
			log.Errorf("hystrix.Do run failed, error: %v\n", err)
			defer func() {
				if e := recover(); e != nil {
					rErr = fmt.Errorf("panic: %v", e)
				}
			}()

			ctx.Abort()

			if err == hystrix.ErrMaxConcurrency {
				ctx.Writer.WriteHeader(http.StatusTooManyRequests)
			} else if err == hystrix.ErrCircuitOpen {
				ctx.Writer.WriteHeader(http.StatusServiceUnavailable)
			} else if err == hystrix.ErrTimeout || err == context.Canceled || err == context.DeadlineExceeded {
				// 该逻辑分支目前不会触发
				// todo 之后如果需要添加链路追踪, context需要做处理
				log.Errorf("should never run, error: %v", err)
			}
			return nil
		})
		if err != nil {
			log.Errorf("hystrix.Do fallback failed, error: %v", err)
		}
	}
}
