package rabbitmq

import (
	logger "gitlab.xfq.com/tech-lab/ngkit"
)

var (
	defaultLogFields = map[string]interface{}{"pkg": "rabbitmq", "type": "internal"}
	log              = initLog()
)

func initLog() logger.Logger {
	l, _ := logger.New(logger.ZapLogger)
	return l.WithFields(defaultLogFields)
}

func SetLog(rabbitLog logger.Logger) {
	log = rabbitLog.WithFields(defaultLogFields)
	log.Infof("log is set")
}
