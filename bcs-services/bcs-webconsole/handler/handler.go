package handler

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"github.com/gin-gonic/gin"
	"go-micro.dev/v4/client"
)

func NewRouteRegistrar() route.Registrar {
	return BcsWebconsole{}
}

func (e BcsWebconsole) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/api/ping", e.Ping)
}

func (e *BcsWebconsole) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

type Options struct {
	Client client.Client
	Router *gin.Engine
}

func Register(opts Options) error {
	router := opts.Router
	h := NewRouteRegistrar()
	h.RegisterRoute(router.Group(""))
	return nil
}
