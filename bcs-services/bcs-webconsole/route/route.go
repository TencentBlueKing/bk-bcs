package route

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/config"
)

type Registrar interface {
	RegisterRoute(gin.IRoutes)
}

type Options struct {
	RoutePrefix string
	Config      config.Config
	Client      client.Client
	Router      *gin.Engine
	RedisClient *redis.Client
}
