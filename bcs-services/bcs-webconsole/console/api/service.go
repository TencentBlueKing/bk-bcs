package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type service struct {
	opts *route.Options
}

func NewRouteRegistrar(opts *route.Options) route.Registrar {
	return service{opts: opts}
}

func (e service) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/api/projects/:projectId/clusters/:clusterId/session/", e.CreateWebConsoleSession).
		GET("/ws/projects/:projectId/clusters/:clusterId/", e.BCSWebSocketHandler).
		GET(filepath.Join(e.opts.RoutePrefix, "/api/projects/:projectId/clusters/:clusterId/session")+"/", e.CreateWebConsoleSession).
		GET(filepath.Join(e.opts.RoutePrefix, "/ws/projects/:projectId/clusters/:clusterId")+"/", e.BCSWebSocketHandler)
}

func (s *service) CreateWebConsoleSession(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")

	host := fmt.Sprintf("%s/clusters/%s", s.opts.Config.Get("bcs_conf", "host").String(""), clusterId)
	token := s.opts.Config.Get("bcs_conf", "token").String("")

	config := &rest.Config{
		Host:        host,
		BearerToken: token,
	}

	data := types.APIResponse{
		Result: true,
		Code:   1, // TODO code待确认
		Data:   map[string]string{},
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		data.Result = false
		data.Message = fmt.Sprintf("获取session失败, %s", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, data)
		return
	}

	backend := manager.NewManager(nil, k8sClient, config, s.opts.RedisClient, s.opts.Config)

	store := sessions.NewRedisStore(s.opts.RedisClient, projectId, clusterId)
	session, err := store.New(c.Request, "")
	if err != nil {
		data.Result = false
		data.Message = "获取session失败"
		manager.ResponseJSON(c.Writer, http.StatusBadRequest, data)
		return
	}

	podName, err := backend.GetK8sContext(c.Writer, c.Request, c.Request.Context(), projectId, clusterId)
	if err != nil {
		data.Result = false
		data.Message = fmt.Sprintf("获取session失败, %s", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, data)
		return
	}
	// 把创建好的pod信息保存到用户数据
	userPodData := &types.UserPodData{
		ProjectID:  projectId,
		ClustersID: clusterId,
		PodName:    podName,
		SessionID:  session.ID,
		CrateTime:  time.Now(),
	}
	backend.WritePodData(userPodData)

	wsUrl := filepath.Join(s.opts.RoutePrefix, fmt.Sprintf("/ws/projects/%s/clusters/%s/?session_id=%s", projectId, clusterId, session.ID))
	data.Code = 0
	data.Message = "获取session成功"
	data.Data = map[string]string{
		"session_id": session.ID,
		"ws_url":     wsUrl,
	}

	manager.ResponseJSON(c.Writer, http.StatusOK, data)
}

func (s *service) BCSWebSocketHandler(c *gin.Context) {

	data := types.APIResponse{
		Result: true,
		Code:   1, // TODO code待确认
		Data:   map[string]string{},
	}

	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	sessionId := c.Query("session_id")
	store := sessions.NewRedisStore(s.opts.RedisClient, projectId, clusterId)
	values, err := store.GetValues(c.Request, sessionId)
	if err != nil {
		data.Result = false
		data.Message = "获取session失败"
		manager.ResponseJSON(c.Writer, http.StatusBadRequest, data)
		return
	}
	username := values["username"]

	host := fmt.Sprintf("%s/clusters/%s", s.opts.Config.Get("bcs_conf", "host").String(""), clusterId)
	token := s.opts.Config.Get("bcs_conf", "token").String("")

	config := &rest.Config{
		Host:        host,
		BearerToken: token,
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		data.Result = false
		data.Message = fmt.Sprintf("获取session失败, %s", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, data)
		return
	}

	backend := manager.NewManager(nil, k8sClient, config, s.opts.RedisClient, s.opts.Config)

	podName := fmt.Sprintf("kubectld-%s-u%s", strings.ToLower(clusterId), projectId)

	webConsole := &types.WebSocketConfig{
		PodName:    podName,
		User:       username,
		ClusterID:  clusterId,
		ProjectsID: projectId,
	}

	// handler container web console
	backend.StartExec(c.Writer, c.Request, webConsole)
}
