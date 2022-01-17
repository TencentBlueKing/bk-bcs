package handler

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"github.com/gin-gonic/gin"
	"go-micro.dev/v4/client"
)

func NewRouteRegistrar() route.Registrar {
	return BcsWebconsole{}
}

func (e BcsWebconsole) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/api/ping", e.Ping).
		GET("/web_console/projects/:projectId/clusters/:clusterId/", e.IndexPageHandler).
		GET("/web_console/", e.StaticHandler)
}

func (e *BcsWebconsole) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func (e *BcsWebconsole) StaticHandler(c *gin.Context) {

}

func (e *BcsWebconsole) IndexPageHandler(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	sessionUrl := fmt.Sprintf("/web_console/api/projects/%s/clusters/%s/web_console/session/", projectId, clusterId)
	settings := map[string]string{
		"SITE_STATIC_URL":      "/web_console",
		"COMMON_EXCEPTION_MSG": "",
	}

	data := gin.H{
		"title":       clusterId,
		"session_url": sessionUrl,
		"settings":    settings,
	}
	fmt.Println("leijiaomin", projectId, clusterId)

	c.HTML(http.StatusOK, "index.html", data)
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
