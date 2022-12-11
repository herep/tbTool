package strategy

import (
	"github.com/go-redis/redis/v7"
)

type Strategy func(redisCLis []*redis.Client, redisName string) *redis.Client
