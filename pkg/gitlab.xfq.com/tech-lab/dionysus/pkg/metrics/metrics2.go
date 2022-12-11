package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterMetrics register ginprometheus and pprof to gin
func InitPrometheus(metricsAddr string, subsystem string, metricsList []MetricType) (*Prometheus, error) {
	metricsRouter := gin.New()
	p, err := NewPrometheus(subsystem, metricsList)
	if err != nil {
		return nil, err
	}
	p.SetMetricsPath(metricsRouter)
	RegisterPProf(metricsRouter, pprofPath)
	metricsRouter.Use(gin.RecoveryWithWriter(Lwriter{}))
	metricsServer = &http.Server{Addr: metricsAddr, Handler: metricsRouter}
	return p, nil
}

//  return each metricCollector's method as middleware for gin
func RecordMetrics(r *http.Request, start time.Time, statusCode int, respSize int, metrics []MetricType) {
	record(r, start, statusCode, respSize, metrics)
}
