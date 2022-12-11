package conf

import (
	"fmt"
	"os"
	"strings"

	"gitlab.xfq.com/tech-lab/dionysus/pkg/conf"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/grpool"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/hystrix"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/logger"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/metrics"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/middle"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/orm"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/rabbitmq"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/redis"
)

func Setup() error {
	// init file config
	if conff := os.Getenv("GAPI_CONFIG"); conff != "" {
		if err := conf.InitConfFile(conff); err != nil {
			return fmt.Errorf("initConfFile fail, endpoints:%v, error: %v", conff, err)
		}
	}

	// init etcd config
	if confe := os.Getenv("GAPI_CONFIG_ETCD"); confe != "" {
		// verify project name
		pn := os.Getenv("GAPI_PROJECT_NAME")
		if pn == "" {
			return fmt.Errorf(" Project name should not be empty")
		}

		endpoints := strings.Split(confe, ",")
		if err := conf.InitEtcdClient(endpoints, pn); err != nil {
			return fmt.Errorf("initEtcdClient fail, endpoints:%v, error: %v", endpoints, err)
		}
	}

	conf.SetLog(logger.Getlogger())
	grpool.SetLog(logger.Getlogger())
	middle.SetLog(logger.Getlogger())
	redis.SetLog(logger.Getlogger())
	metrics.SetLog(logger.Getlogger())
	hystrix.SetLog(logger.Getlogger())
	rabbitmq.SetLog(logger.Getlogger())
	orm.SetLog(logger.Getlogger())
	return nil
}
