package metrics

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var defaultMetricPath = "/metrics"
var httpClientTimeout = time.Second * 60

/*
RequestCounterURLLabelMappingFn is a function which can be supplied to the middleware to control
the cardinality of the request counter's "url" label, which might be required in some contexts.
For instance, if for a "/customer/:name" route you don't want to generate a time series for every
possible customer name, you could use this function:

func(c *gin.Context) string {
	url := c.Request.URL.Path
	for _, p := range c.Params {
		if p.Key == "name" {
			url = strings.Replace(url, p.Value, ":name", 1)
			break
		}
	}
	return url
}

which would map "/customer/alice" and "/customer/bob" to their template "/customer/:name".
*/
type RequestCounterURLLabelMappingFn func(c *gin.Context) string

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	Ppg PrometheusPushGateway

	//MetricsList []*Metric
	MetricsList []MetricType
	MetricsPath string

	ReqCntURLLabelMappingFn RequestCounterURLLabelMappingFn

	// gin.Context string to use as a prometheus URL label
	URLLabelFromContext string
}

// PrometheusPushGateway contains the configuration for pushing to a Prometheus pushgateway (optional)
type PrometheusPushGateway struct {

	// Push interval in seconds
	PushIntervalDuration time.Duration

	// Push Gateway URL in format http://domain:port
	// where JOBNAME can be any string of your choice
	PushGatewayURL string

	// Local metrics URL where metrics are fetched from, this could be omitted in the future
	// if implemented using prometheus common/expfmt instead
	MetricsURL string

	// pushgateway job name, defaults to "gin"
	Job string
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus(subsystem string, metricsList []MetricType) (*Prometheus, error) {

	var mList []MetricType
	if len(metricsList) == 0 {
		mList = defaultMetrics
	} else {
		for _, m := range metricsList {
			if !MetricTypeValidated(m) {
				mLog.Errorf("Invalid metric type : %v", m)
				err := fmt.Errorf("Push metrics IntervalDuration must be greater than 0!")
				return nil, err
			}
		}
		mList = metricsList
	}
	p := &Prometheus{
		MetricsList: mList,
		MetricsPath: defaultMetricPath,
		ReqCntURLLabelMappingFn: func(c *gin.Context) string {
			return c.Request.URL.Path // i.e. by default do nothing, i.e. return URL as is
		},
	}

	p.registerMetrics(subsystem)

	return p, nil
}

// SetPushGateway sends metrics to a remote pushgateway exposed on pushGatewayURL
// every pushIntervalDur. Metrics are fetched from metricsURL
func (p *Prometheus) SetPushGateway(pushGatewayURL, metricsURL string, pushIntervalDur time.Duration) error {
	p.Ppg.PushGatewayURL = pushGatewayURL
	p.Ppg.MetricsURL = metricsURL
	p.Ppg.PushIntervalDuration = pushIntervalDur
	return p.startPushTicker()
}

// SetPushGatewayJob job name, defaults to "gin"
func (p *Prometheus) SetPushGatewayJob(j string) {
	p.Ppg.Job = j
}

// SetMetricsPath set metrics paths
func (p *Prometheus) SetMetricsPath(e *gin.Engine) {
	e.GET(p.MetricsPath, prometheusHandler())
}

// SetMetricsPathWithAuth set metrics paths with authentication
func (p *Prometheus) SetMetricsPathWithAuth(e *gin.Engine, accounts gin.Accounts) {
	e.GET(p.MetricsPath, gin.BasicAuth(accounts), prometheusHandler())
}

func (p *Prometheus) getMetrics() []byte {
	client := &http.Client{
		Timeout: httpClientTimeout,
	}
	response, err := client.Get(p.Ppg.MetricsURL)
	if err != nil {
		mLog.Errorf("getMetrics Get failed, err: %v", err)
		return nil
	}

	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		mLog.Errorf("getMetrics ReadAll failed, err: %v", err)
	}
	return body
}

func (p *Prometheus) getPushGatewayURL() string {
	h, _ := os.Hostname()
	if p.Ppg.Job == "" {
		p.Ppg.Job = "gapi"
	}
	return p.Ppg.PushGatewayURL + defaultMetricPath + "/job/" + p.Ppg.Job + "/instance/" + h
}

func (p *Prometheus) sendMetricsToPushGateway(metrics []byte) error {
	req, err := http.NewRequest("POST", p.getPushGatewayURL(), bytes.NewBuffer(metrics))
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: httpClientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()
	return nil
}

func (p *Prometheus) startPushTicker() error {
	// If PushIntervalDuration is 0 or less than 0, time.NewTicker will panic.
	// In this case, we return directly to prevent panic.
	if p.Ppg.PushIntervalDuration <= 0 {
		mLog.Errorf("Push metrics IntervalDuration must be greater than 0!")
		err := fmt.Errorf("Push metrics IntervalDuration must be greater than 0!")
		return err
	}
	ticker := time.NewTicker(time.Second * p.Ppg.PushIntervalDuration)
	go func() {
		for range ticker.C {
			err := p.sendMetricsToPushGateway(p.getMetrics())
			if err != nil {
				mLog.Errorf("Push metrics failed, err: %v", err)
			}
		}
	}()
	return nil
}

func (p *Prometheus) registerMetrics(subsystem string) {
	for _, metricId := range p.MetricsList {
		metricDef := metricsMap[metricId]
		metric := NewMetricCollector(metricDef, subsystem)
		if err := prometheus.Register(metric); err != nil {
			mLog.Errorf("Register metric %s failed. err: %v", metricDef.Name, err)
		}
		metricDef.MetricCollector = metric
	}
}

// Use adds the middleware to a gin engine.
func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.HandlerFunc())
	p.SetMetricsPath(e)
}

// UseWithAuth adds the middleware to a gin engine with BasicAuth.
func (p *Prometheus) UseWithAuth(e *gin.Engine, accounts gin.Accounts) {
	e.Use(p.HandlerFunc())
	p.SetMetricsPathWithAuth(e, accounts)
}

// HandlerFunc defines handler function for middleware
func (p *Prometheus) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == p.MetricsPath {
			c.Next()
			return
		}

		GetMetricHandler(c, p.URLLabelFromContext, p.ReqCntURLLabelMappingFn, p.MetricsList)
	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
