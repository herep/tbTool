package hystrix

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/afex/hystrix-go/hystrix"
)

const (
	defaultTimeout                = 2000
	defaultMaxConcurrentRequests  = 1000
	defaultRequestVolumeThreshold = 5000
	defaultSleepWindow            = 2000
	defaultErrorPercentThreshold  = 50
)

const (
	serverDefaultMaxConcurrentRequests  = 50000
	serverDefaultRequestVolumeThreshold = 50000
	serverDefaultSleepWindow            = 2000
	serverDefaultErrorPercentThreshold  = 50
)

const (
	DotSep   = "."
	SlashSep = "/"
	Hystrix  = "hystrix"
	Server   = "server"
	Client   = "client"
)

var (
	hm                   sync.Map
	emptyStruct          struct{}
	projectName          string
	serverDefaultTimeout = 20000
)

func init() {
	if t := os.Getenv("GAPI_REQUEST_TIMEOUT"); t != "" {
		if s, err := strconv.Atoi(t); err == nil && s > 0 {
			serverDefaultTimeout = s*1000 + 10*1000
		}
	}
}

type runFunc func() error
type fallbackFunc func(error) error

// CommandConfig is used to tune circuit settings at runtime
type CommandConfig struct {
	Type                   string `json:"type"`                     // 类型, server: 服务端, client: 客户端
	Timeout                int    `json:"timeout"`                  // 超时时间, 单位毫秒
	MaxConcurrentRequests  int    `json:"max_concurrent_requests"`  // command最大并发量
	RequestVolumeThreshold int    `json:"request_volume_threshold"` // 一个统计窗口10秒内请求数量。达到这个请求数量后才去判断是否要开启熔断
	SleepWindow            int    `json:"sleep_window"`             // 当熔断器被打开后，SleepWindow的时间就是控制过多久后去尝试服务是否可用了。单位毫秒
	ErrorPercentThreshold  int    `json:"error_percent_threshold"`  // 错误百分比，请求数量大于等于RequestVolumeThreshold并且错误率到达这个百分比后就会启动
}

type HyxFunc struct {
	CommandName  string
	RunFunc      func() error      // 业务逻辑处理函数
	FallbackFunc func(error) error // 降级处理函数
}

func ConfigureCommand(commandName string, opts ...ConfigOption) {
	if _, loaded := hm.LoadOrStore(commandName, emptyStruct); loaded {
		return
	}

	var commandConfig CommandConfig
	notExistSetDefaultCommandConfig(&commandConfig)

	for _, o := range opts {
		o(&commandConfig)
	}

	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		Timeout:                commandConfig.Timeout,
		MaxConcurrentRequests:  commandConfig.MaxConcurrentRequests,
		RequestVolumeThreshold: commandConfig.RequestVolumeThreshold,
		SleepWindow:            commandConfig.SleepWindow,
		ErrorPercentThreshold:  commandConfig.ErrorPercentThreshold,
	})
}

// NewHyxFunc returns a new HyxFunc
func NewHyxFunc(commandName string, run runFunc, opts ...Option) (*HyxFunc, error) {
	ConfigureCommand(commandName)

	fn := HyxFunc{
		CommandName: commandName,
		RunFunc:     run,
	}
	for _, o := range opts {
		o(&fn)
	}

	return &fn, nil
}

// Do runs your function in a synchronous manner, blocking until either your function succeeds
// or an error is returned, including hystrix circuit errors
func (hf *HyxFunc) Do() error {
	return hystrix.Do(hf.CommandName, func() (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("hystrix do run panic: %v", e)
			}
		}()
		return hf.RunFunc()
	}, func(err error) (rErr error) {
		defer func() {
			if e := recover(); e != nil {
				rErr = fmt.Errorf("hystrix do fallback panic: %v", e)
			}
		}()
		return hf.FallbackFunc(err)
	})
}

// GetCircuit returns the circuit for the given command and whether this call created it.
func GetCircuit(name string) (*hystrix.CircuitBreaker, bool, error) {
	return hystrix.GetCircuit(name)
}

// IsOpen is called before any Command execution to check whether or
// not it should be attempted. An "open" circuit means it is disabled.
func IsOpen(name string) (bool, error) {
	cb, _, err := GetCircuit(name)
	if err != nil {
		return false, err
	}
	if cb == nil {
		return false, errors.New("circuit not exist")
	}
	return cb.IsOpen(), nil
}

// AllowRequest is checked before a command executes, ensuring that circuit state and metric health allow it.
// When the circuit is open, this call will occasionally return true to measure whether the external service
// has recovered.
func AllowRequest(name string) (bool, error) {
	cb, _, err := GetCircuit(name)
	if err != nil {
		return false, err
	}
	if cb == nil {
		return false, errors.New("circuit not exist")
	}
	return cb.AllowRequest(), nil
}

// GetCircuitSettings get all circuit settings
func GetCircuitSettings() map[string]*hystrix.Settings {
	return hystrix.GetCircuitSettings()
}

// ServerConfigureCommand applies settings for a server-side circuit
func ServerConfigureCommand(commandName string, opts ...ConfigOption) {
	if _, loaded := hm.LoadOrStore(commandName, emptyStruct); loaded {
		return
	}

	var commandConfig CommandConfig
	notExistSetServerDefaultCommandConfig(&commandConfig)

	for _, o := range opts {
		o(&commandConfig)
	}

	// 此参数手动配置无效, 统一使用默认值
	commandConfig.Timeout = serverDefaultTimeout
	hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
		Timeout:                commandConfig.Timeout,
		MaxConcurrentRequests:  commandConfig.MaxConcurrentRequests,
		RequestVolumeThreshold: commandConfig.RequestVolumeThreshold,
		SleepWindow:            commandConfig.SleepWindow,
		ErrorPercentThreshold:  commandConfig.ErrorPercentThreshold,
	})
}

// notExistSetDefaultCommandConfig 如果未指定某个参数将其设置为默认值
func notExistSetDefaultCommandConfig(config *CommandConfig) {
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.MaxConcurrentRequests == 0 {
		config.MaxConcurrentRequests = defaultMaxConcurrentRequests
	}
	if config.RequestVolumeThreshold == 0 {
		config.RequestVolumeThreshold = defaultRequestVolumeThreshold
	}
	if config.SleepWindow == 0 {
		config.SleepWindow = defaultSleepWindow
	}
	if config.ErrorPercentThreshold == 0 {
		config.ErrorPercentThreshold = defaultErrorPercentThreshold
	}
}

// notExistSetServerDefaultCommandConfig 如果未指定某个参数将其设置为默认值
func notExistSetServerDefaultCommandConfig(config *CommandConfig) {
	// 将超时时间设置为较大的值, 使hystrix的超时时间失效(TimeoutMiddleware默认10秒超时)
	config.Timeout = serverDefaultTimeout

	if config.MaxConcurrentRequests == 0 {
		config.MaxConcurrentRequests = serverDefaultMaxConcurrentRequests
	}
	if config.RequestVolumeThreshold == 0 {
		config.RequestVolumeThreshold = serverDefaultRequestVolumeThreshold
	}
	if config.SleepWindow == 0 {
		config.SleepWindow = serverDefaultSleepWindow
	}
	if config.ErrorPercentThreshold == 0 {
		config.ErrorPercentThreshold = serverDefaultErrorPercentThreshold
	}
}

func JoinCommandName(name string) string {
	if projectName == "" {
		projectName = os.Getenv("GAPI_PROJECT_NAME")
	}
	return strings.Join([]string{"", projectName, Hystrix, name}, SlashSep)
}
