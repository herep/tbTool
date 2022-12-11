package orm

import (
	logger "gitlab.xfq.com/tech-lab/ngkit"
)

var (
	defaultLogFields = map[string]interface{}{"pkg": "orm", "type": "internal"}
	log              = initLog()
)

func initLog() logger.Logger {
	l, _ := logger.New(logger.ZapLogger)
	return l.WithFields(defaultLogFields)
}

func SetLog(ormLog logger.Logger) {
	log = ormLog.WithFields(defaultLogFields)
	log.Infof("orm log is set")
}
