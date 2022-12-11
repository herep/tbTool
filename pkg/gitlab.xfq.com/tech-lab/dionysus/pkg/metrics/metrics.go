package metrics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Standard default metrics
//	counter, counter_vec, gauge, gauge_vec,
//	histogram, histogram_vec, summary, summary_vec
var reqCnt = &Metric{
	ID:          "reqCnt",
	Name:        "requests_total",
	Description: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	Type:        "counter_vec",
	Args:        []string{"code", "method", "host"},
}

var reqDur = &Metric{
	ID:          "reqDur",
	Name:        "request_duration_seconds",
	Description: "The HTTP request latencies in seconds.",
	Type:        "histogram_vec",
	Args:        []string{"code", "method"},
}

var resSz = &Metric{
	ID:          "resSz",
	Name:        "response_size_bytes",
	Description: "The HTTP response sizes in bytes.",
	Type:        "summary",
}

var reqSz = &Metric{
	ID:          "reqSz",
	Name:        "request_size_bytes",
	Description: "The HTTP request sizes in bytes.",
	Type:        "summary",
}

var reqEr = &Metric{
	ID:          "reqErr",
	Name:        "request_error_count",
	Description: "Numbers of error requests",
	Type:        "counter",
}

var memStk = &Metric{
	ID:          "memStk",
	Name:        "mem_heap_in_use_bytes",
	Description: "Top heap inuse bytes by method",
	Type:        "gauge_vec",
	Args:        []string{"InUseObjects", "allocateObjects", "acllocateBytes", "stack"},
}

type MetricType int

const (
	ReqCNT MetricType = iota
	ReqDUR
	ReqSZ
	ResSZ
	ReqER
	MemSTK
)

var metricsMap = map[MetricType]*Metric{
	ReqCNT: reqCnt,
	ReqDUR: reqDur,
	ReqSZ:  reqSz,
	ResSZ:  resSz,
	ReqER:  reqEr,
	MemSTK: memStk,
}

var defaultMetrics = []MetricType{
	ReqCNT,
	ReqDUR,
	ReqSZ,
	ResSZ,
	ReqER,
}

var (
	metricsServer *http.Server = nil
)

// Metric is a definition for the name, description, type, ID, and
// prometheus.Collector type (i.e. CounterVec, Summary, etc) of each metric
type Metric struct {
	MetricCollector prometheus.Collector
	ID              string
	Name            string
	Description     string
	Type            string
	Args            []string
}

var pprofPath = "dio/pprof"

func MetricTypeValidated(metric MetricType) bool {
	switch metric {
	case ReqCNT, ReqDUR, ReqSZ, ResSZ, ReqER, MemSTK:
		return true
	}
	return false
}

// RegisterMetrics register ginprometheus and pprof to gin
func RegisterMetrics(e *gin.Engine, metricsAddr string, subsystem string, metricsList []MetricType) (*http.Server, error) {
	metricsRouter := gin.New()
	p, err := NewPrometheus(subsystem, metricsList)
	if err != nil {
		return nil, err
	}
	e.Use(p.HandlerFunc())
	p.SetMetricsPath(metricsRouter)
	RegisterPProf(metricsRouter, pprofPath)
	metricsRouter.Use(gin.RecoveryWithWriter(Lwriter{}))
	metricsServer = &http.Server{Addr: metricsAddr, Handler: metricsRouter}
	return metricsServer, nil
}

func GetMetricsServer() *http.Server {
	return metricsServer
}

func CloseMetricsServer() error {
	if metricsServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := metricsServer.Shutdown(ctx); err != nil {
		mLog.Errorf("Metric server shutdown failed.")
		return err
	}
	return nil
}

// NewMetricCollector associates prometheus.Collector based on Metric.Type
func NewMetricCollector(m *Metric, subsystem string) prometheus.Collector {
	var metric prometheus.Collector
	switch m.Type {
	case "counter_vec":
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "counter":
		metric = prometheus.NewCounter(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "gauge_vec":
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "gauge":
		metric = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "histogram_vec":
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "histogram":
		metric = prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	case "summary_vec":
		metric = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
			m.Args,
		)
	case "summary":
		metric = prometheus.NewSummary(
			prometheus.SummaryOpts{
				Subsystem: subsystem,
				Name:      m.Name,
				Help:      m.Description,
			},
		)
	}
	return metric
}

// GetMetricHandler return each metricCollector's method as middleware for gin
func GetMetricHandler(c *gin.Context, urlLabelFromContext string, reqCntUrlLabelFn RequestCounterURLLabelMappingFn, metrics []MetricType) {
	start := time.Now()

	c.Next()
	nstatus := c.Writer.Status()

	record(c.Request, start, nstatus, c.Writer.Size(), metrics)

	// url := reqCntUrlLabelFn(c)
	// // jlambert Oct 2018 - sidecar specific mod
	// if len(urlLabelFromContext) > 0 {
	// 	u, found := c.Get(urlLabelFromContext)
	// 	if !found {
	// 		u = "unknown"
	// 	}
	// 	url = u.(string)
	// }

	// for _, metric := range metrics {
	// 	if metricsMap[metric] == nil || metricsMap[metric].MetricCollector == nil {
	// 		mLog.Errorf("Invalid metric type : %v", metric)
	// 		continue
	// 	}
	// 	mc := metricsMap[metric].MetricCollector
	// 	switch metric {
	// 	case ReqCNT:
	// 		reqCntMC := mc.(*prometheus.CounterVec)
	// 		reqCntMC.WithLabelValues(status, c.Request.Method, c.HandlerName(), c.Request.Host).Inc()
	// 	case ReqDUR:
	// 		reqDurMC := mc.(*prometheus.HistogramVec)
	// 		reqDurMC.WithLabelValues(status, c.Request.Method).Observe(elapsed)
	// 	case ResSZ:
	// 		resSzMC := mc.(prometheus.Summary)
	// 		resSzMC.Observe(resSz)
	// 	case ReqSZ:
	// 		reqSzMC := mc.(prometheus.Summary)
	// 		reqSzMC.Observe(float64(reqSz))
	// 	case ReqER:
	// 		if nstatus >= 500 {
	// 			reqErMC := mc.(prometheus.Counter)
	// 			reqErMC.Inc()
	// 		}
	// 	}
	// }

}

func record(r *http.Request, start time.Time, statusCode int, respSize int, metrics []MetricType) {
	reqSz := computeApproximateRequestSize(r)

	status := strconv.Itoa(statusCode)
	elapsed := float64(time.Since(start)) / float64(time.Second)

	for _, metric := range metrics {
		if metricsMap[metric] == nil || metricsMap[metric].MetricCollector == nil {
			mLog.Errorf("Invalid metric type : %v", metric)
			continue
		}
		mc := metricsMap[metric].MetricCollector
		switch metric {
		case ReqCNT:
			reqCntMC := mc.(*prometheus.CounterVec)
			reqCntMC.WithLabelValues(status, r.Method, r.Host).Inc()
		case ReqDUR:
			reqDurMC := mc.(*prometheus.HistogramVec)
			reqDurMC.WithLabelValues(status, r.Method).Observe(elapsed)
		case ResSZ:
			resSzMC := mc.(prometheus.Summary)
			resSzMC.Observe(float64(respSize))
		case ReqSZ:
			reqSzMC := mc.(prometheus.Summary)
			reqSzMC.Observe(float64(reqSz))
		case ReqER:
			if statusCode >= 500 {
				reqErMC := mc.(prometheus.Counter)
				reqErMC.Inc()
			}
		}
	}

}
