package hystrix

type Option func(*HyxFunc)
type ConfigOption func(*CommandConfig)

// WithFallbackFunc sets the fallback function
func WithFallbackFunc(fn fallbackFunc) Option {
	return func(hf *HyxFunc) {
		hf.FallbackFunc = fn
	}
}

// WithTimeout sets hystrix timeout
// Timeout is how long to wait for command to complete, in milliseconds
func WithTimeout(timeout int) ConfigOption {
	return func(cc *CommandConfig) {
		cc.Timeout = timeout
	}
}

// WithMaxConcurrentRequests sets hystrix max concurrent requests
// MaxConcurrentRequests is how many commands of the same type can run at the same time
func WithMaxConcurrentRequests(maxConcurrentRequests int) ConfigOption {
	return func(cc *CommandConfig) {
		cc.MaxConcurrentRequests = maxConcurrentRequests
	}
}

// WithRequestVolumeThreshold sets hystrix request volume threshold
// RequestVolumeThreshold is the minimum number of requests needed before a circuit can be tripped due to health
func WithRequestVolumeThreshold(requestVolumeThreshold int) ConfigOption {
	return func(cc *CommandConfig) {
		cc.RequestVolumeThreshold = requestVolumeThreshold
	}
}

// WithSleepWindow sets hystrix sleep window
// SleepWindow is how long, in milliseconds, to wait after a circuit opens before testing for recovery
func WithSleepWindow(sleepWindow int) ConfigOption {
	return func(cc *CommandConfig) {
		cc.SleepWindow = sleepWindow
	}
}

// WithErrorPercentThreshold sets hystrix error percent threshold
// ErrorPercentThreshold causes circuits to open once the rolling measure of errors exceeds this percent of requests
func WithErrorPercentThreshold(errorPercentThreshold int) ConfigOption {
	return func(cc *CommandConfig) {
		cc.ErrorPercentThreshold = errorPercentThreshold
	}
}
