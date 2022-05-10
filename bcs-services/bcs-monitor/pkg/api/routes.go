package api

import (
	"context"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/pod"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// APIServer
type APIServer struct {
	Engine *gin.Engine
	srv    *http.Server
}

// NewAPIServer
func NewAPIServer(ctx context.Context) (*APIServer, error) {
	gin.SetMode(gin.ReleaseMode)

	s := &APIServer{Engine: gin.Default()}
	registerRoutes(s.Engine)

	return s, nil
}

// Run
func (a *APIServer) Run(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: a.Engine,
	}

	return srv.ListenAndServe()

}

func (a *APIServer) Close(ctx context.Context) error {
	return a.srv.Shutdown(ctx)
}

func registerRoutes(engine *gin.Engine) {
	// 添加X-Request-Id 头部
	requestIdMiddleware := requestid.New(
		requestid.WithGenerator(func() string {
			return rest.RequestIdGenerator()
		}),
	)

	// route := engine.Group("/projects/{projectId}/clusters/{clusterId}")
	engine.Use(requestIdMiddleware)

	// 日志相关接口
	engine.GET("/projects/:projectId/clusters/:clusterId/namespaces/:namespace/pods/:pod/containers", rest.RestHandlerFunc(pod.GetContainerList))
	// route.GET("/namespace/{namespace}/pods/{pod}/log/download", pod.GetContainerList)
	// route.GET("/namespace/{namespace}/pods/{pod}/log", pod.GetContainerList)
}
