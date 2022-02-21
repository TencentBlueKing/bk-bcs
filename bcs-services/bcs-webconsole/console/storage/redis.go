package storage

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"

	redis "github.com/go-redis/redis/v8"
)

// RedisSession :
type RedisSession struct {
	Client *redis.Client
}

// Init : init the redis client
func (r *RedisSession) Init() error {
	redisConf := config.G.Redis

	r.Client = redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%v:%v", redisConf.Host, redisConf.Port),
		Password:    redisConf.Password,
		DB:          redisConf.DB,
		DialTimeout: time.Duration(redisConf.MaxConnTimeout) * time.Second,
		ReadTimeout: time.Duration(redisConf.ReadTimeout) * time.Second,
		PoolSize:    redisConf.MaxPoolSize,
		IdleTimeout: time.Duration(redisConf.IdleTimeout) * time.Second,
	})
	return nil
}

// Close : close redis session
func (r *RedisSession) Close() {
	r.Client.Close()
}
