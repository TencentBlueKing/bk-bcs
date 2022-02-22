package storage

// once synchronization
import (
	"sync"
)

// GLobals
var (
	GlobalRedisSession *RedisSession
)

var redisOnce sync.Once

// GetDefaultRedisSession : get default redis session
func GetDefaultRedisSession() *RedisSession {
	if GlobalRedisSession == nil {
		redisOnce.Do(func() {
			GlobalRedisSession = &RedisSession{}
			err := GlobalRedisSession.Init()
			if err != nil {
				panic(err)
			}
		})
	}
	return GlobalRedisSession
}
