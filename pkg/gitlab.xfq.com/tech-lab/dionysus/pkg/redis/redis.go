package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"

	utilerrors "gitlab.xfq.com/tech-lab/dionysus/pkg/errors"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/grpool"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/redis/strategy"
)

var initClientsMap sync.Map

const DefaultRWTimeout = time.Second

type Option func(options *redis.Options)

type Rdconfig struct {
	Addr        string
	DB          int
	Password    string
	PoolSize    int
	IdleTimeout int
}

// WithReadTimeout 设置读超时
func WithReadTimeout(d time.Duration) Option {
	return func(o *redis.Options) {
		o.ReadTimeout = d
	}
}

// WithWriteTimeout 设置写超时
func WithWriteTimeout(d time.Duration) Option {
	return func(o *redis.Options) {
		o.WriteTimeout = d
	}
}

func WithDB(db int) Option {
	return func(o *redis.Options) {
		o.DB = db
	}
}

func WithPoolSize(poolSize int) Option {
	return func(o *redis.Options) {
		o.PoolSize = poolSize
	}
}

func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(o *redis.Options) {
		o.IdleTimeout = idleTimeout
	}
}

func pickupClients(redisName string) ([]*redis.Client, error) {
	lowerName := strings.ToLower(redisName)
	result, ok := initClientsMap.Load(lowerName)
	if ok {
		return result.([]*redis.Client), nil
	}

	log.Errorf("can not get redis client ,prefix=" + redisName)
	return nil, errors.New("can not get redis client ,prefix=" + redisName)
}

func getClients(redisName string) []*redis.Client {
	lowerName := strings.ToLower(redisName)
	result, ok := initClientsMap.Load(lowerName)
	if ok {
		return result.([]*redis.Client)
	}
	return nil
}

// 组装转换数据
func InitClient(redisName string, redisConfig *Rdconfig) error {
	redisClients := newClients(redisConfig)
	if len(redisClients) == 0 {
		log.Errorf("get redis clients is 0, will not update RedisMap of redisConfig: %v", redisConfig)
		return fmt.Errorf("get redis clients is 0, will not update RedisMap of redisConfig: %v", redisConfig)
	}

	for _, redisClient := range redisClients {
		sc := redisClient.Ping()
		if sc.Val() == "" {
			_ = grpool.Submit(func() {
				err := closeClients(redisClients)
				if err != nil {
					log.Errorf("redis name: %v close err: %v", redisName, err)
				}
			})
			log.Errorf("redis连接池连接失败,error=%v, redisConfig: %v", sc.Err(), redisClient.Options().Addr)
			return fmt.Errorf("redis连接池连接失败,error=%v", sc.Err())
		}
	}

	lowerName := strings.ToLower(redisName)
	rdclis := getClients(lowerName)

	initClientsMap.Store(lowerName, redisClients)
	deferCloseClients(rdclis, lowerName)
	log.Infof("rebuild redis pool done - %v", redisName)

	return nil
}

// new redis pool from config center by watch
func newClients(redisConfig *Rdconfig) []*redis.Client {
	redisClients := []*redis.Client{}
	if redisConfig.Addr == "" {
		return redisClients
	}
	addrs := strings.Split(redisConfig.Addr, ",")
	for _, addr := range addrs {
		if addr == "" {
			continue
		}
		redisCli := redis.NewClient(&redis.Options{
			Addr:        addr,
			DB:          redisConfig.DB,
			PoolSize:    redisConfig.PoolSize,
			IdleTimeout: time.Duration(redisConfig.IdleTimeout) * time.Second,
			Password:    redisConfig.Password,
		})
		redisClients = append(redisClients, redisCli)
	}
	return redisClients
}

// Deprecated: use GetClient() instead
func GetRedis(ctx context.Context, redisName string, options ...Option) (*redis.Client, error) {
	return GetClientWithStrategy(ctx, redisName, strategy.RoundRobin, options...)
}

// GetClient get a redis client with open tracing context
func GetClient(ctx context.Context, redisName string, options ...Option) (*redis.Client, error) {
	return GetClientWithStrategy(ctx, redisName, strategy.RoundRobin, options...)
}

// GetClient get a redis client with open tracing context and strategy
func GetClientWithStrategy(ctx context.Context, redisName string, strategy strategy.Strategy, options ...Option) (*redis.Client, error) {
	clients, err := pickupClients(redisName)
	if err != nil {
		return nil, err
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("get redis clients is 0 of redisName: %v", redisName)
	}

	var client *redis.Client

	if len(clients) == 1 {
		client = clients[0]
	} else {
		client = strategy(clients, redisName)
	}

	if client == nil {
		return nil, fmt.Errorf("get redis client is nil redisName: %v", redisName)
	}

	cc := client.WithTimeout(DefaultRWTimeout)
	opts := cc.Options()

	// 处理外部进来的参数配置
	for _, o := range options {
		o(opts)
	}

	return cc.WithContext(ctx), nil
}

func deleteClientsData(redisName string) {
	lowerName := strings.ToLower(redisName)
	rdclis := getClients(lowerName)
	initClientsMap.Delete(lowerName)
	log.Infof("delete redis pool done - %v", redisName)
	deferCloseClients(rdclis, lowerName)
}

func deferCloseClients(rdclis []*redis.Client, lowerName string) {
	if rdclis == nil {
		return
	}
	time.AfterFunc(time.Second*30, func() {
		err := closeClients(rdclis)
		if err != nil {
			log.Errorf("redis: %v close error: %v", lowerName, err)
		}
	})
}

func closeClients(rdclis []*redis.Client) error {
	var errlist []error
	for _, rdcli := range rdclis {
		if rdcli != nil {
			err := rdcli.Close()
			if err != nil {
				errlist = append(errlist, err)
			}
		}
	}
	return utilerrors.NewAggregate(errlist)
}
