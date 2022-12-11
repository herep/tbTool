package conf

import (
	logger "gitlab.xfq.com/tech-lab/ngkit"
)

var (
	defaultLogFields = map[string]interface{}{"pkg": "etcd", "type": "internal"}
	log              = initLog()
)

func initLog() logger.Logger {
	l, _ := logger.New(logger.ZapLogger)
	return l.WithFields(defaultLogFields)
}

func SetLog(watchLog logger.Logger) {
	log = watchLog.WithFields(defaultLogFields)
	log.Infof("conf log is set")
}
