package strategy

import (
	"math/rand"
	"time"

	"github.com/go-redis/redis/v7"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Random is a random strategy algorithm for node selection
func Random(redisCLis []*redis.Client, redisName string) *redis.Client {
	i := rand.Int() % len(redisCLis) // nolint:gosec
	return redisCLis[i]
}
