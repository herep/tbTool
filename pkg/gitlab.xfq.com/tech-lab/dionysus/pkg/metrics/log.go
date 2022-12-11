package metrics

import (
	logger "gitlab.xfq.com/tech-lab/ngkit"
)

var (
	defaultLogFields = map[string]interface{}{"pkg": "metrics", "type": "internal"}
	mLog             = initLog()
)

func initLog() logger.Logger {
	l, _ := logger.New(logger.ZapLogger)
	return l.WithFields(defaultLogFields)
}

type Lwriter struct {
}

func (writer Lwriter) Write(p []byte) (n int, err error) {
	mLog.Errorf(string(p))
	return len(p), nil
}

func SetLog(log logger.Logger) {
	mLog = log.WithFields(defaultLogFields)
	mLog.Infof("metrics log is set")
}
