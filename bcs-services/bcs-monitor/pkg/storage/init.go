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

// GetDefaultRedisSession : get default redis session for default database
func GetDefaultRedisSession() *RedisSession {
	if GlobalRedisSession == nil {
		redisOnce.Do(func() {
			session := &RedisSession{}
			err := session.Init()
			if err != nil {
				panic(err)
			}
			GlobalRedisSession = session
		})
	}
	return GlobalRedisSession
}
