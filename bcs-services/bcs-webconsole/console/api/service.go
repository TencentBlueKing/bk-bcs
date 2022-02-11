package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type service struct {
	opts *route.Options
}

func NewRouteRegistrar(opts *route.Options) route.Registrar {
	return service{opts: opts}
}

// 	router.Use(route.Localize())
func (e service) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/api/projects/:projectId/clusters/:clusterId/session/", e.CreateWebConsoleSession).
		GET("/ws/projects/:projectId/clusters/:clusterId/", e.BCSWebSocketHandler).
		POST("/web_console", e.CreateOpenWebConsoleSession).
		GET(filepath.Join(e.opts.RoutePrefix, "/api/projects/:projectId/clusters/:clusterId/session")+"/",
			e.CreateWebConsoleSession).
		GET(filepath.Join(e.opts.RoutePrefix, "/ws/projects/:projectId/clusters/:clusterId")+"/",
			e.BCSWebSocketHandler).
		POST(filepath.Join(e.opts.RoutePrefix, "/web_console/"), e.CreateOpenWebConsoleSession)
}

func (s *service) CreateWebConsoleSession(c *gin.Context) {
	s.opts.Config.Get("").String("")
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")

	host := fmt.Sprintf("%s/clusters/%s", s.opts.Config.Get("bcs_conf", "host").String(""), clusterId)
	token := s.opts.Config.Get("bcs_conf", "token").String("")

	config := &rest.Config{
		Host:        host,
		BearerToken: token,
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		msg := i18n.GetMessage("k8s客户端初始化失败{}", map[string]string{"err": err.Error()})
		utils.APIError(c, msg)
		return
	}

	backend := manager.NewManager(nil, k8sClient, config, s.opts.RedisClient, s.opts.Config)

	store := sessions.NewRedisStore(s.opts.RedisClient, projectId, clusterId)
	session, err := store.New(c.Request, "")
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", map[string]string{"err": err.Error()})
		utils.APIError(c, msg)
		return
	}

	podName, err := backend.GetK8sContext(c.Request.Context(), projectId, clusterId)
	if err != nil {
		msg := i18n.GetMessage("申请pod资源失败{}", map[string]string{"err": err.Error()})
		utils.APIError(c, msg)
		return
	}
	// TODO 把创建好的pod信息保存到用户数据session
	userPodData := &types.UserPodData{
		ProjectID:  projectId,
		ClustersID: clusterId,
		PodName:    podName,
		SessionID:  session.ID,
		CrateTime:  time.Now(),
	}
	backend.WritePodData(userPodData)

	wsUrl := filepath.Join(s.opts.RoutePrefix, fmt.Sprintf("/ws/projects/%s/clusters/%s/?session_id=%s",
		projectId, clusterId, session.ID))

	data := types.APIResponse{
		Data: map[string]string{
			"session_id": session.ID,
			"ws_url":     wsUrl,
		},
		Code:      types.NoError,
		Message:   i18n.GetMessage("获取session成功"),
		RequestID: uuid.New().String(),
	}
	c.JSON(http.StatusOK, data)
}

func (s *service) BCSWebSocketHandler(c *gin.Context) {

	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	sessionId := c.Query("session_id")
	store := sessions.NewRedisStore(s.opts.RedisClient, projectId, clusterId)
	values, err := store.GetValues(c.Request, sessionId)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", map[string]string{"err": err.Error()})
		utils.APIError(c, msg)
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
		msg := i18n.GetMessage("初始化k8s客户端失败{}", map[string]string{"err": err.Error()})
		utils.APIError(c, msg)
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
	backend.StartExec(c, webConsole)
}

func (s *service) CreateOpenWebConsoleSession(c *gin.Context) {

	projectId := c.Query("project_id")
	clusterId := c.Query("cluster_id")

	var containerName string

	// 优先使用containerID
	containerID, ok := c.GetPostForm("container_id")
	if ok {
		//	有containerID才检查
		host := fmt.Sprintf("%s/clusters/%s", s.opts.Config.Get("bcs_conf", "host").String(""), clusterId)
		token := s.opts.Config.Get("bcs_conf", "token").String("")
		config := &rest.Config{
			Host:        host,
			BearerToken: token,
		}

		k8sClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			msg := i18n.GetMessage("初始化k8s客户端失败{}", map[string]string{"err": err.Error()})
			utils.APIError(c, msg)
			return
		}

		backend := manager.NewManager(nil, k8sClient, config, s.opts.RedisClient, s.opts.Config)
		container, err := backend.GetK8sContextByContainerID(containerID)
		if err != nil {
			blog.Info("container_id is incorrect, err : %v", err)
			msg := i18n.GetMessage("container_id不正确，请检查参数")
			utils.APIError(c, msg)
			return
		}

		containerName = container.ContainerName

	} else {

		podName, _ := c.GetPostForm("pod_name")
		containerName, _ := c.GetPostForm("container_name")
		namespace, _ := c.GetPostForm("namespace")

		// 其他使用namespace, pod, container
		if namespace == "" || podName == "" || containerName == "" {
			msg := i18n.GetMessage("container_id或namespace/pod_name/container_name不能同时为空")
			utils.APIError(c, msg)
			return
		}
	}

	store := sessions.NewRedisStore(s.opts.RedisClient, projectId, clusterId)
	session, err := store.New(c.Request, "")
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", map[string]string{"err": err.Error()})
		utils.APIError(c, msg)
		return
	}

	wsUrl := filepath.Join(s.opts.RoutePrefix, fmt.Sprintf("/web_console/?session_id=%s&container_name=%s",
		session.ID, containerName))

	respData := types.APIResponse{
		Data: map[string]string{
			"session_id": session.ID,
			"ws_url":     wsUrl,
		},
		Code:      types.NoError,
		Message:   i18n.GetMessage("获取session成功"),
		RequestID: uuid.New().String(),
	}

	c.JSON(http.StatusOK, respData)
}
