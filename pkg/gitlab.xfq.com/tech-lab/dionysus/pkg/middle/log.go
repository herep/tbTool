package middle

import (
	logger "gitlab.xfq.com/tech-lab/ngkit"
)

var (
	defaultLogFields = map[string]interface{}{"pkg": "middle", "type": "internal", "notice": "true"}
	log              = initLog()
)

func initLog() logger.Logger {
	l, _ := logger.New(logger.ZapLogger)
	return l.WithFields(defaultLogFields)
}

func SetLog(middleLog logger.Logger) {
	log = middleLog.WithFields(defaultLogFields)
	log.Infof("middle ware log is set")
}
