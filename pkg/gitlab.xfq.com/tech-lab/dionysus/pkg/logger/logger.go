package logger

import (
	"io"
	"os"

	"gitlab.xfq.com/tech-lab/dionysus/pkg/color"
	"gitlab.xfq.com/tech-lab/ngkit"
	"gitlab.xfq.com/tech-lab/utils/net"
)

var (
	logger, _ = log.New(log.ZapLogger)
	devMode   bool

	Debug      = logger.Debug
	Debugf     = logger.Debugf
	Info       = logger.Info
	Infof      = logger.Infof
	Warn       = logger.Warn
	Warnf      = logger.Warnf
	Error      = logger.Error
	Errorf     = logger.Errorf
	Fatal      = logger.Fatal
	Fatalf     = logger.Fatalf
	WithField  = logger.WithField
	WithFields = logger.WithFields
)

type Lwriter struct {
}

func (writer Lwriter) Write(p []byte) (n int, err error) {
	logger.Error(string(p))
	return len(p), nil
}

func Setup() error {
	logger = newLogger()

	Debug = logger.Debug
	Debugf = logger.Debugf
	Info = logger.Info
	Infof = logger.Infof
	Warn = logger.Warn
	Warnf = logger.Warnf
	Error = logger.Error
	Errorf = logger.Errorf
	Fatal = logger.Fatal
	Fatalf = logger.Fatalf
	WithField = logger.WithField
	WithFields = logger.WithFields
	return nil
}

func SetDevMode(b bool) {
	devMode = b
}

func NewTracingLogger() (log.Logger, error) {
	opts := []log.Option{
		log.WithEncoderCfg(log.EncoderConfig{MessageKey: "msg"}),
		log.WithEncoder(log.ConsoleEncoder),
	}

	// writer
	if path := os.Getenv("GAPI_TRACING_LOG"); path != "" {
		w, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			return nil, err
		}
		opts = append(opts, log.WithWriter(w))
	} else {
		opts = append(opts, log.WithWriter(os.Stdout))
	}

	return log.New(log.ZapLogger, opts...)
}

func newLogger() log.Logger {

	// 保持一致
	EncoderConfig := log.EncoderConfig{
		MessageKey:   "msg",
		LevelKey:     "level",
		EncodeLevel:  log.LowercaseLevelEncoder,
		TimeKey:      "@timestamp",
		EncodeTime:   log.RFC3339MilliTimeEncoder,
		CallerKey:    "caller",
		EncodeCaller: log.ShortCallerEncoder,
	}

	opts := []log.Option{
		log.WithEncoderCfg(EncoderConfig),
		log.WithLevelEnabler(log.DebugLevel),
		log.AddCaller(),
		log.AddCallerSkip(1),
	}

	// writer
	var ew, w io.Writer = os.Stdout, os.Stdout

	if devMode {
		ew = color.Writer(ew, color.Red)
	}

	if path := os.Getenv("GAPI_LOG"); path != "" {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			panic(err) // 保持一致
		}

		if devMode {
			// f+std.out
			ew, w = io.MultiWriter(f, ew), io.MultiWriter(f, w)
		} else {
			ew, w = f, f // only file
		}
	}

	opts = append(opts, log.ErrorOutput(ew), log.WithWriter(w))

	// 保持一致
	hostname, _ := os.Hostname()
	// todo entry.Data["project"] = fh.filename
	ip, err := net.GetIP()
	if err != nil {
		ip = hostname
	}

	initialFields := map[string]interface{}{
		"host":      ip,
		"host_name": hostname,
	}

	opts = append(opts, log.Fields(initialFields))

	logger, err := log.New(log.ZapLogger, opts...)
	if err != nil {
		panic(err) // 保持一致
	}

	return logger
}

// Deprecated: Use WithField or WithFields instead.
// func With(args ...interface{}) *zap.SugaredLogger {
//	return logger.With(args...)
// }

func Getlogger() log.Logger {
	return logger
}
