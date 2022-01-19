package handler

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/web"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"github.com/gin-gonic/gin"
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

func Register(opts *route.Options) error {
	router := opts.Router
	h := NewRouteRegistrar()
	h.RegisterRoute(router.Group(""))
	for _, r := range []route.Registrar{
		web.NewRouteRegistrar(opts),
		api.NewRouteRegistrar(opts),
	} {
		r.RegisterRoute(router.Group(""))
	}
	return nil
}
