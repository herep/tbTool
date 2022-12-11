package hystrix

import (
	logger "gitlab.xfq.com/tech-lab/ngkit"
)

var (
	defaultLogFields = map[string]interface{}{"pkg": "hystrix", "type": "internal"}
	log              = initLog()
)

func initLog() logger.Logger {
	l, _ := logger.New(logger.ZapLogger)
	return l.WithFields(defaultLogFields)
}

func SetLog(hyxLog logger.Logger) {
	log = hyxLog.WithFields(defaultLogFields)
	log.Infof("hystrix log is set")
}
