package dionysus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gitlab.xfq.com/tech-lab/dionysus/cmd"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/algs"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/env"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/logger"
	"gitlab.xfq.com/tech-lab/dionysus/step"
)

type Dio struct {
	cmd   *cobra.Command
	steps *step.Steps // 全局启动依赖项,暂时注释掉入口。

	// e             *gin.Engine
	// groups        []*CustomGroup
	// globalMiddles []func() gin.HandlerFunc
	// server        *http.Server
}

var dio *Dio

var (
	BuildTime = ""
	CommitID  = ""
	GoVersion = ""
)

func init() {
	dio = NewDio()

	// Caution :: Inorder to compatible with micro, we should move global steps into sub command.
	//
	// dio.regActionSteps("logger", 1, logger.Setup)
	// dio.regActionSteps("conf", 2, conf.Setup)
	// regActionSteps("flag", 0, flag.Setup)
}

func NewDio() *Dio {
	d := &Dio{
		// Steps should be finished before run
		steps: step.New(),
		// root cmd
		cmd: &cobra.Command{},
		// gin cmd
	}

	return d
}

// name overwrite:: default bin name -> func param -> env -> flag
func Start(name string, cmds ...cmd.Commander) {
	// 0 set cmd use
	dio.cmd.Use = name
	if ex, err := os.Executable(); dio.cmd.Use == "" && err == nil {
		dio.cmd.Use = filepath.Base(ex)
	}

	// 1. global flags
	var config, log, endpoints, environment string
	var version bool
	dio.cmd.PersistentFlags().StringVarP(&config, "config", "c", os.Getenv("GAPI_CONFIG"), "config path")
	dio.cmd.PersistentFlags().StringVarP(&log, "log", "l", os.Getenv("GAPI_LOG"), "log file path; default console output")
	dio.cmd.PersistentFlags().StringVarP(&endpoints, "endpoints", "e", os.Getenv("GAPI_CONFIG_ETCD"), "the etcd endpoints")
	dio.cmd.PersistentFlags().StringVarP(&name, "name", "n", os.Getenv("GAPI_PROJECT_NAME"), "the project name")
	dio.cmd.PersistentFlags().StringVar(&environment, "env", algs.FirstNotEmpty(os.Getenv(env.SysEnvKey), "product"), "the project run mod available in [ product | gray| test | develop ]")
	dio.cmd.PersistentFlags().BoolVar(&version, "version", false, "Print build version")

	// 2. global pre run function
	dio.cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if config != "" {
			err := os.Setenv("GAPI_CONFIG", config)
			if err != nil {
				return err
			}
		}

		if log != "" {
			err := os.Setenv("GAPI_LOG", log)
			if err != nil {
				return err
			}
		}

		if endpoints != "" {
			err := os.Setenv("GAPI_CONFIG_ETCD", endpoints)
			if err != nil {
				return err
			}
		}

		if name != "" {
			err := os.Setenv("GAPI_PROJECT_NAME", name)
			if err != nil {
				return err
			}
		}

		if environment != "" {
			err := env.Set(env.Environment(environment))
			if err != nil {
				return err
			}

			if env.IsDevelop() {
				logger.SetDevMode(true)
			}
		}

		if version {
			fmt.Println("Build Time:", BuildTime)
			fmt.Println("Go Version:", GoVersion)
			fmt.Println("Commit  ID:", CommitID)
		}

		return nil
	}

	dio.cmd.Run = func(cmd *cobra.Command, args []string) {} // empty run, so we can trigger -v flag

	// 3. append other cmds
	for _, c := range cmds {
		dio.cmd.AddCommand(c.GetCmd())
	}

	// 4. start
	if err := dio.cmd.Execute(); err != nil {
		panic(err)
	}
}

// 注册Steps 0-10 内部保留使用,暂时不开放全局 steps
// nolint
func regActionSteps(value string, priority int, fn func() error) {
	dio.steps.RegActionSteps(value, priority+100, fn)
}
