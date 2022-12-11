package strategy

import (
	"strings"
	"sync"
	"sync/atomic"

	"github.com/go-redis/redis/v7"
)

var countMap = sync.Map{}

// RoundRobin is a roundrobin strategy algorithm for node selection
func RoundRobin(redisCLis []*redis.Client, redisName string) *redis.Client {
	count := getCount(redisName)
	i := (*count) % uint64(len(redisCLis))
	atomic.AddUint64(count, 1)
	return redisCLis[i]
}

func initCountMap(redisName string) *uint64 {
	var count uint64
	lowerName := strings.ToLower(redisName)
	countMap.Store(lowerName, &count)
	return &count
}

func getCount(redisName string) *uint64 {
	lowerName := strings.ToLower(redisName)
	value, ok := countMap.Load(lowerName)
	if !ok {
		return initCountMap(redisName)
	}
	count := value.(*uint64)
	return count
}
