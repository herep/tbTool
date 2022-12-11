package grpool

import (
	logger "gitlab.xfq.com/tech-lab/ngkit"
)

var (
	defaultLogFields = map[string]interface{}{"pkg": "grpool", "type": "internal"}
	log              = initLog()
)

func initLog() logger.Logger {
	l, _ := logger.New(logger.ZapLogger)
	return l.WithFields(defaultLogFields)
}

func SetLog(grpoolLog logger.Logger) {
	log = grpoolLog.WithFields(defaultLogFields)
	log.Infof("grpool log is set")
}
