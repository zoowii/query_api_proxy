package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var cacheInstance = cache.New(5*time.Minute, 10*time.Minute)

func Set(key string, value interface{}, expireTimeDuration time.Duration) {
	cacheInstance.Set(key, value, expireTimeDuration)
}

func SetWithDefaultExpire(key string, value interface{}) {
	cacheInstance.Set(key, value, cache.DefaultExpiration)
}
func SetWithNoExpire(key string, value interface{}) {
	cacheInstance.Set(key, value, cache.NoExpiration)
}

func Get(key string) (value interface{}, found bool) {
	value, found = cacheInstance.Get(key)
	return
}
