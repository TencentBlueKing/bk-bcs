package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"go-micro.dev/v4/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type service struct {
	Config config.Config
}

func NewRouteRegistrar(conf config.Config) route.Registrar {
	return service{Config: conf}
}

func (e service) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/web_console/api/projects/:projectId/clusters/:clusterId/web_console/session/", e.CreateWebConsoleSession)
}

func (s *service) CreateWebConsoleSession(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", s.Config.Get("redis", "host").String("127.0.0.1"), s.Config.Get("redis", "port").Int(6379)),
		Password: "",
		DB:       s.Config.Get("redis", "db").Int(0),
	})

	host := fmt.Sprintf("%s/clusters/%s", s.Config.Get("bcs_conf", "host").String(""), clusterId)
	token := s.Config.Get("bcs_conf", "token").String("")

	fmt.Println(host, token)

	config := &rest.Config{
		Host:        host,
		BearerToken: token,
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	fmt.Println(err)

	backend := manager.NewManager(nil, k8sClient, config, redisClient, s.Config)

	data := types.APIResponse{
		Result: true,
		Code:   1, // TODO code待确认
		Data:   map[string]string{},
	}

	session, err := store.Get(c.Request, "sessionID")
	if err != nil {
		data.Result = false
		data.Message = "获取session失败！"
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

	// TODO 封装获取wsURL方法
	wsUrl := "ws://127.0.0.1:8080/web_console/projects/clusters/ws?projectsID=%s&clustersID=%s&session_id=%s"
	wsUrl = fmt.Sprintf(wsUrl, projectId, clusterId, session.ID)
	data.Code = 0
	data.Message = "获取session成功"
	data.Data = map[string]string{
		"session_id": session.ID,
		"ws_url":     wsUrl,
	}

	manager.ResponseJSON(c.Writer, http.StatusOK, data)
}
