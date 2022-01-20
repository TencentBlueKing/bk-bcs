package web

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"

	"github.com/gin-gonic/gin"
)

type service struct {
	opts *route.Options
}

func NewRouteRegistrar(opts *route.Options) route.Registrar {
	return service{opts: opts}
}

func (e service) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/web_console/projects/:projectId/clusters/:clusterId/", e.IndexPageHandler)
}

func (s *service) IndexPageHandler(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	routePrefix := s.opts.Config.Get("web", "route_prefix").String("")
	sessionUrl := fmt.Sprintf("/web_console/api/projects/%s/clusters/%s/web_console/session/", projectId, clusterId)
	settings := map[string]string{
		"SITE_STATIC_URL":      routePrefix,
		"COMMON_EXCEPTION_MSG": "",
	}

	data := gin.H{
		"title":       clusterId,
		"session_url": sessionUrl,
		"settings":    settings,
	}

	c.HTML(http.StatusOK, "index.html", data)
}
