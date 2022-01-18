package web

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"go-micro.dev/v4/config"

	"github.com/gin-gonic/gin"
)

type service struct {
	Config config.Config
}

func NewRouteRegistrar(conf config.Config) route.Registrar {
	return service{Config: conf}
}

func (e service) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/web_console/projects/:projectId/clusters/:clusterId/", e.IndexPageHandler)
}

func (s *service) IndexPageHandler(c *gin.Context) {
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

	c.HTML(http.StatusOK, "index.html", data)
}
