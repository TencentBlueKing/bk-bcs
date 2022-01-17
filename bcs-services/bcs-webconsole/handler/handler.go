package handler

import (
	"fmt"
	"html/template"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/web"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"github.com/gin-gonic/gin"
	"go-micro.dev/v4/client"
)

const (
	IndexPageTpl = "web/templates/index.html"
)

func NewRouteRegistrar() route.Registrar {
	return BcsWebconsole{}
}

func (e BcsWebconsole) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/api/ping", e.Ping).
		GET("/web_console/projects/:projectId/clusters/:clusterId/", e.IndexPageHandler).
		GET("/web_console/", e.IndexPageHandler)
}

func (e *BcsWebconsole) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func (e *BcsWebconsole) StaticHandler(c *gin.Context) {

}

func (e *BcsWebconsole) IndexPageHandler(c *gin.Context) {
	t, err := template.ParseFS(web.FS, IndexPageTpl)
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	sessionUrl := fmt.Sprintf("/web_console/api/projects/%s/clusters/%s/web_console/session/", projectId, clusterId)
	settings := map[string]string{
		"SITE_STATIC_URL":      "/web_console",
		"COMMON_EXCEPTION_MSG": "",
	}

	data := map[string]interface{}{
		"title":       clusterId,
		"session_url": sessionUrl,
		"settings":    settings,
	}
	fmt.Println("leijiaomin", projectId, clusterId)
	if err != nil {
		blog.Error("index page templates not found, err : %v", err)
		return
	}

	t.Execute(c.Writer, data)
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
