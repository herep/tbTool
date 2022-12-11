package gincmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gitlab.xfq.com/tech-lab/dionysus/cmd"
	"gitlab.xfq.com/tech-lab/dionysus/conf"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/algs"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/env"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/logger"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/metrics"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/middle"
	"gitlab.xfq.com/tech-lab/dionysus/shutdown"
	"gitlab.xfq.com/tech-lab/dionysus/step"
)

const (
	WebServerAddr     = "GAPI_ADDR"
	MetricsServerAddr = "GAPI_METRICS_ADDR"

	defaultMetricsServerAddr = ":9120"
	defaultWebServerAddr     = ":8080"

	addrFlagName    = "addr"
	metricsFlagName = "metricsAddr"
)

var sysExit = os.Exit

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type ginCmd struct {
	*gin.Engine
	cmd *cobra.Command

	server *http.Server

	preRunFuncs, postRunFuncs *step.Steps

	once sync.Once
}

func New() *ginCmd {
	g := &ginCmd{
		Engine:       gin.New(),
		server:       &http.Server{},
		cmd:          &cobra.Command{Use: "gin", Short: "Run as gin web server"},
		preRunFuncs:  step.New(),
		postRunFuncs: step.New(),
	}

	// default middle wares
	g.Use(gin.RecoveryWithWriter(logger.Lwriter{}))
	return g
}

func (g *ginCmd) Flags() *pflag.FlagSet {
	return g.cmd.Flags()
}

func (g *ginCmd) RegFlagSet(set *pflag.FlagSet) {
	g.cmd.Flags().AddFlagSet(set)
}

func (g *ginCmd) GetCmd() *cobra.Command {

	finishChan := make(chan struct{})

	// flags can register only once
	g.once.Do(func() {
		g.cmd.Flags().StringVarP(&g.server.Addr, addrFlagName, "a", algs.FirstNotEmpty(os.Getenv(WebServerAddr), defaultWebServerAddr), "the http server address")
		g.cmd.Flags().StringP(metricsFlagName, "m", algs.FirstNotEmpty(os.Getenv(MetricsServerAddr), defaultMetricsServerAddr), "the metrics http server address")
	})

	g.cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		switch env.Get() {
		case env.Develop:
			gin.SetMode(gin.DebugMode)
		// case env.Test:
		// 	gin.SetMode(gin.TestMode) // gin.TestMode only using in gin's unit test.
		default:
			gin.SetMode(gin.ReleaseMode)
		}

		g.preRunFuncs.RegActionSteps("logger", 1, logger.Setup)
		g.preRunFuncs.RegActionSteps("conf", 2, conf.Setup)
		if err := g.preRunFuncs.Run(); err != nil {
			return err
		}

		// set timeout and metrics
		// default < env < request_header (ps: request_header timeout args must less than env args)
		var defaultTimeout = 10
		if t := os.Getenv("GAPI_REQUEST_TIMEOUT"); t != "" {
			if s, err := strconv.Atoi(t); err == nil && s > 0 {
				defaultTimeout = s
			}
		}

		prom := g.regMetrics2("Gapi", nil)
		g.server.Handler = middle.TimeoutHandler(g, time.Second*time.Duration(defaultTimeout), "", prom)
		// g.regMetrics("Gapi", nil)
		g.regCheckHealth()

		return nil
	}

	g.cmd.Run = func(cmd *cobra.Command, args []string) {
		shutdown.NotifyAfterFinish(finishChan, g.startGin)
	}

	g.cmd.PostRunE = func(cmd *cobra.Command, args []string) error {
		shutdown.WaitingForNotifies(finishChan, g.shutdown)
		return g.postRunFuncs.Run()
	}

	return g.cmd
}

func (g *ginCmd) RegPreRunFunc(value string, priority cmd.Priority, f func() error) error {
	return g.preRunFuncs.RegActionStepsE(value, int(priority)+100, f)
}

func (g *ginCmd) RegPostRunFunc(value string, priority cmd.Priority, f func() error) error {
	return g.postRunFuncs.RegActionStepsE(value, int(priority)+100, f)
}

// for unit test
func (g *ginCmd) ServeHTTPForTest(w http.ResponseWriter, req *http.Request) {
	g.ServeHTTP(w, req)
}

func (g *ginCmd) startGin() {
	if metricsServer := metrics.GetMetricsServer(); metricsServer != nil {
		if metricsServer.Addr != g.server.Addr {
			go func() {
				if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("listen metrics server failed : %s\n", err)
				}
			}()
		} else { // handled at regMetrics()
			log.Printf(" We should never got this err!!! metrics server using same addr with gin : %s\n", metricsServer.Addr)
		}
	}

	log.Printf("[Dio] Engine setting with address %v", g.server.Addr)

	if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("listen: %s\n", err)
		sysExit(1)
	}
}

func (g *ginCmd) shutdown() {
	metricsServer := metrics.GetMetricsServer()
	if metricsServer != nil && metricsServer.Addr != g.server.Addr {
		if err := metrics.CloseMetricsServer(); err != nil {
			log.Println("[error] Metrics server forced to shutdown:", err)
		}
	}

	log.Println("[info] Server exiting")
	if err := g.server.Shutdown(context.TODO()); err != nil {
		log.Println("[error] Server forced to shutdown:", err)
		sysExit(1)
	}
}

// Deprecated :: this func will be removed after v0.5.0
func (g *ginCmd) regMetrics(subsystem string, metricsList []metrics.MetricType) { // nolint
	var metricsAddr string

	if str, err := g.cmd.Flags().GetString(metricsFlagName); err != nil {
		metricsAddr = defaultMetricsServerAddr
	} else {
		metricsAddr = str
	}

	// if MetricsAddr == gapiAddr, we can't setup metrics server for security consideration
	// let's just return then.
	if metricsAddr == g.server.Addr {
		log.Printf("Error: Can not use gin address as metrics address!!")
		return
	}

	_, err := metrics.RegisterMetrics(g.Engine, metricsAddr, subsystem, metricsList)
	if err == nil {
		log.Printf("Register metrics addr %s.", metricsAddr)
	} else {
		log.Printf("Register metrics addr %s failed!", metricsAddr)
	}
}

func (g *ginCmd) regMetrics2(subsystem string, metricsList []metrics.MetricType) *metrics.Prometheus {
	var metricsAddr string

	if str, err := g.cmd.Flags().GetString(metricsFlagName); err != nil {
		metricsAddr = defaultMetricsServerAddr
	} else {
		metricsAddr = str
	}

	// if MetricsAddr == gapiAddr, we can't setup metrics server for security consideration
	// let's just return then.
	if metricsAddr == g.server.Addr {
		err := "Can not use gin address as metrics address!! "
		log.Print(err)
		return nil
	}

	p, err := metrics.InitPrometheus(metricsAddr, subsystem, metricsList)
	if err == nil {
		log.Printf("Register metrics addr %s.", metricsAddr)
		return p
	} else {
		log.Printf("Register metrics addr %s failed!", metricsAddr)
		return nil
	}
}

func (g *ginCmd) regCheckHealth() {
	g.GET("/", func(c *gin.Context) {
		c.String(200, "%s", "success")
	})
	g.GET("/health", func(c *gin.Context) {
		c.String(200, "%s", "success")
	})
}

// Deprecated:: use RegPreRunFunc instead
func (g *ginCmd) RegActionSteps(value string, priority int, fn func() error) {
	panic("This func has been deprecated. Please use g.RegPreRunFunc() .")
}

// 注册全局中间件
// Deprecated:: This func has been deprecated. Please use g.Use() .
func (g *ginCmd) RegGlobalMiddle(fn func() gin.HandlerFunc) {
	panic("This func has been deprecated. Please use g.Use() .")
}
