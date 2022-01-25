package api

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/utils"
	"github.com/google/uuid"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/i18n"
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

func (s service) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		Use().
		GET("/api/projects/:projectId/clusters/:clusterId/session/", s.CreateWebConsoleSession).
		GET("/ws/projects/:projectId/clusters/:clusterId/", s.BCSWebSocketHandler).
		POST("/web_console", s.CreateOpenWebConsoleSession).
		GET(filepath.Join(s.opts.RoutePrefix, "/api/projects/:projectId/clusters/:clusterId/session")+"/",
			s.CreateWebConsoleSession).
		GET(filepath.Join(s.opts.RoutePrefix, "/ws/projects/:projectId/clusters/:clusterId")+"/",
			s.BCSWebSocketHandler).
		POST(filepath.Join(s.opts.RoutePrefix, "/web_console/"), s.CreateOpenWebConsoleSession)
}

// f"{settings.DEVOPS_BCS_API_URL}/web_console/?session_id={session_id}&container_name={container_name}

func (s *service) CreateWebConsoleSession(c *gin.Context) {
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
		// TODO 初始化k8s失败
		msg := i18n.MustGetMessage(i18n.NewLocalizeConfig("初始化k8s客户端失败{}", map[string]string{
			"err": err.Error()}))
		utils.APIError(c, msg)
		return
	}

	backend := manager.NewManager(nil, k8sClient, config, s.opts.RedisClient, s.opts.Config)

	store := sessions.NewRedisStore(s.opts.RedisClient, projectId, clusterId)
	session, err := store.New(c.Request, "")
	if err != nil {
		// TODO 获取session失败
		msg := i18n.MustGetMessage(i18n.NewLocalizeConfig("获取session失败{}", map[string]string{"err": err.Error()}))
		utils.APIError(c, msg)
		return
	}

	podName, err := backend.GetK8sContext(c.Request.Context(), projectId, clusterId)
	if err != nil {
		msg := i18n.MustGetMessage(i18n.NewLocalizeConfig("申请pod资源失败{}", nil))
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
		Message:   i18n.MustGetMessage("获取session成功"),
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
		msg := i18n.MustGetMessage(i18n.NewLocalizeConfig("获取session失败{}", map[string]string{"err": err.Error()}))
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
		// TODO
		msg := i18n.MustGetMessage(i18n.NewLocalizeConfig("初始化k8s客户端失败{}", map[string]string{"err": err.Error()}))
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

	//?session_id={session_id}&container_name={container_name}

	projectId := c.Query("project_id")
	clusterId := c.Query("cluster_id")

	appCode, ok := c.GetPostForm("app_code")
	if !ok {
		utils.APIError(c, i18n.MustGetMessage("app_code 不能为空"))
		return
	}
	operator, ok := c.GetPostForm("operator")
	if !ok {
		utils.APIError(c, i18n.MustGetMessage("operator 不能为空"))
		return
	}

	command := c.DefaultPostForm("command", "sh") // 默认是sh命令
	appSecret, _ := c.GetPostForm("app_secret")

	data := &CreateOpenWebConsoleSessionReqData{
		AppCode:   appCode,
		AppSecret: appSecret,
		Operator:  operator,
		Command:   command,
		ClusterID: clusterId,
		ProjectID: projectId,
	}

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
			// TODO 初始化k8s失败
			msg := i18n.MustGetMessage(i18n.NewLocalizeConfig("初始化k8s客户端失败{}", map[string]string{
				"err": err.Error()}))
			utils.APIError(c, msg)
			return
		}

		backend := manager.NewManager(nil, k8sClient, config, s.opts.RedisClient, s.opts.Config)
		k8sContext, err := backend.GetK8sContextByContainerID(containerID)
		if err != nil {
			msg := i18n.MustGetMessage(i18n.NewLocalizeConfig("申请pod资源失败{}", map[string]string{
				"err": err.Error()}))
			utils.APIError(c, msg)
			return
		}

		data.Namespace = k8sContext.Namespace
		data.PodName = k8sContext.PodName
		data.ContainerName = k8sContext.ContainerName
		data.ContainerID = containerID
	}

	data.PodName, _ = c.GetPostForm("pod_name")
	data.ContainerName, _ = c.GetPostForm("container_name")
	data.Namespace, _ = c.GetPostForm("namespace")

	// 其他使用namespace, pod, container
	if data.Namespace == "" || data.PodName == "" || data.ContainerName == "" {
		msg := i18n.MustGetMessage("container_id或namespace/pod_name/container_name不能同时为空")
		utils.APIError(c, msg)
		return
	}

	// 保存session
	store := sessions.NewRedisStore(s.opts.RedisClient, projectId, clusterId)
	session, err := store.New(c.Request, "")
	if err != nil {
		msg := i18n.MustGetMessage(i18n.NewLocalizeConfig("获取session失败{}", map[string]string{"err": err.Error()}))
		utils.APIError(c, msg)
		return
	}

	wsUrl := filepath.Join(s.opts.RoutePrefix, fmt.Sprintf("/web_console/?session_id=%s&container_name=%s", session.ID, data.ContainerName))

	respData := types.APIResponse{
		Data: map[string]string{
			"session_id": session.ID,
			"ws_url":     wsUrl,
		},
		Code:      types.NoError,
		Message:   i18n.MustGetMessage("获取session成功"),
		RequestID: uuid.New().String(),
	}

	c.JSON(http.StatusOK, respData)

}

type CreateOpenWebConsoleSessionReqData struct {
	AppCode         string `json:"app_code"`       // 应用ID
	AppSecret       string `json:"app_secret"`     // 安全密钥
	ContainerID     string `json:"container_id"`   //
	Operator        string `json:"operator"`       // 当前使用者
	Command         string `json:"command"`        // 自定义web-console启动命令
	Namespace       string `json:"namespace"`      // 通过命名空间, pod, container名字进入
	PodName         string `json:"pod_name"`       //
	ContainerName   string `json:"container_name"` //
	ProjectIDOrCode string `json:"project_id_or_code"`
	ClusterID       string `json:"cluster_id"`
	ProjectID       string `json:"project_id"`
}
