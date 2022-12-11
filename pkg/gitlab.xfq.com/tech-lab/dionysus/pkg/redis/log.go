package redis

import (
	logger "gitlab.xfq.com/tech-lab/ngkit"
)

var (
	defaultLogFields = map[string]interface{}{"pkg": "redis", "type": "internal"}
	log              = initLog()
)

func initLog() logger.Logger {
	l, _ := logger.New(logger.ZapLogger)
	return l.WithFields(defaultLogFields)
}

func SetLog(redisLog logger.Logger) {
	log = redisLog.WithFields(defaultLogFields)
	log.Infof("redis log is set")
}
