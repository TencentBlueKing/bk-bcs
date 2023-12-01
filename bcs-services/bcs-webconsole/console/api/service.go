/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package api xxx
package api

import (
	"encoding/json"
	"net/url"
	"path"
	"strings"
	"time"

	gintrace "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/tracing"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
)

const (
	// FileSizeUnitMb xxx
	FileSizeUnitMb = 1024 * 1024
	// FileSizeLimits xxx
	FileSizeLimits = 30
)

type service struct {
	opts *route.Options
}

// NewRouteRegistrar xxx
func NewRouteRegistrar(opts *route.Options) route.Registrar {
	return service{opts: opts}
}

// RegisterRoute xxx
func (s service) RegisterRoute(router gin.IRoutes) {
	api := router.Use(route.APIAuthRequired(), gintrace.Middleware(tracing.ServiceName))

	// 用户登入态鉴权, session鉴权
	api.GET("/api/projects/:projectId/clusters/:clusterId/session/",
		metrics.RequestCollect("CreateWebConsoleSession"), route.PermissionRequired(),
		route.AuditHandler(), s.CreateWebConsoleSession)
	api.GET("/api/projects/:projectId/clusters/",
		metrics.RequestCollect("ListClusters"), s.ListClusters)

	// 蓝鲸API网关鉴权 & App鉴权
	api.GET("/api/portal/sessions/:sessionId/",
		metrics.RequestCollect("CreatePortalSession"), s.CreatePortalSession)
	api.POST("/api/portal/projects/:projectId/clusters/:clusterId/container/",
		metrics.RequestCollect("CreateContainerPortalSession"), route.CredentialRequired(), s.CreateContainerPortalSession)
	api.POST("/api/portal/projects/:projectId/clusters/:clusterId/cluster/",
		metrics.RequestCollect("CreateClusterPortalSession"), route.CredentialRequired(), s.CreateClusterPortalSession)

	// websocket协议, session鉴权
	api.GET("/ws/sessions/:sessionId/", metrics.RequestCollect("BCSWebSocket"), s.BCSWebSocketHandler)

	// 文件上传下载
	api.POST("/api/sessions/:sessionId/upload/", metrics.RequestCollect("Upload"), s.UploadHandler)
	api.GET("/api/sessions/:sessionId/download/", metrics.RequestCollect("Download"), s.DownloadHandler)
	api.GET("/api/sessions/:sessionId/download/check/", metrics.RequestCollect("CheckDownload"), s.CheckDownloadHandler)

	// 用户命令延时统计api
	api.PUT("/api/command/delay/:username", metrics.RequestCollect("SetUserDelaySwitch"),
		route.ManagersRequired(), s.SetUserDelaySwitch)
	api.GET("/api/command/delay/:username", metrics.RequestCollect("GetUserDelaySwitch"),
		route.ManagersRequired(), s.GetUserDelaySwitch)
	api.GET("/api/command/delay/:username/meter", metrics.RequestCollect("GetUserDelayMeter"),
		route.ManagersRequired(), s.GetUserDelayMeter)
	api.GET("/api/command/delay", metrics.RequestCollect("GetDelayUsers"),
		route.ManagersRequired(), s.GetDelayUsers)
}

// ListClusters 集群列表
func (s *service) ListClusters(c *gin.Context) {
	projectId := c.Param("projectId")
	project, err := bcs.GetProject(c.Request.Context(), config.G.BCS, projectId)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "项目不正确"))
		return
	}

	clusters, err := bcs.ListClusters(c.Request.Context(), project.ProjectId)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, err.Error()))
		return
	}
	rest.APIOK(c, i18n.GetMessage(c, "获取集群成功"), clusters)
}

// CreateWebConsoleSession 创建websocket session
func (s *service) CreateWebConsoleSession(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)

	consoleQuery := new(podmanager.ConsoleQuery)
	_ = c.BindQuery(consoleQuery)

	// 封装一个独立函数, 统计耗时
	podCtx, err := func() (podCtx *types.PodContext, err error) {
		start := time.Now()
		defer func() {
			if consoleQuery.IsContainerDirectMode() {
				return
			}

			// 单独统计 pod metrics
			podReadyDuration := time.Since(start)
			metrics.SetRequestIgnoreDuration(c, podReadyDuration)

			metrics.CollectPodReady(
				podmanager.GetAdminClusterId(authCtx.ClusterId),
				podmanager.GetNamespace(),
				podmanager.GetPodName(authCtx.ClusterId, authCtx.Username),
				err,
				podReadyDuration,
			)
		}()

		podCtx, err = podmanager.QueryAuthPodCtx(c.Request.Context(), authCtx.ClusterId, authCtx.Username, consoleQuery)
		return
	}()
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, err.Error()))
		return
	}

	podCtx.ProjectId = authCtx.ProjectId
	podCtx.Username = authCtx.Username
	podCtx.Source = consoleQuery.Source
	// 二次换取的 session 时间有效期24小时
	podCtx.SessionTimeout = types.MaxSessionTimeout

	sessionId, err := sessions.NewStore().WebSocketScope().Set(c.Request.Context(), podCtx)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "获取session失败{}", err))
		return
	}

	data := map[string]string{
		"session_id": sessionId,
		"ws_url":     makeWebSocketURL(sessionId, consoleQuery.Lang, false),
	}
	rest.APIOK(c, i18n.GetMessage(c, "获取session成功"), data)
}

// CreatePortalSession xxx
func (s *service) CreatePortalSession(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)
	if authCtx.BindSession == nil {
		rest.APIError(c, i18n.GetMessage(c, "session_id不合法或已经过期"))
		return
	}

	podCtx := authCtx.BindSession
	// 二次换取的 session 时间有效期24小时
	podCtx.SessionTimeout = types.MaxSessionTimeout

	sessionId, err := sessions.NewStore().WebSocketScope().Set(c.Request.Context(), podCtx)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "获取session失败{}", err))
		return
	}

	lang := c.Query("lang")
	lang = strings.TrimSuffix(lang, "/")
	data := map[string]string{
		"session_id": sessionId,
		"ws_url":     makeWebSocketURL(sessionId, lang, false),
	}
	rest.APIOK(c, i18n.GetMessage(c, "获取session成功"), data)
}

// CreateContainerPortalSession 创建 webconsole url api
func (s *service) CreateContainerPortalSession(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)

	consoleQuery := new(podmanager.OpenQuery)

	err := c.BindJSON(consoleQuery)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "请求参数错误{}", err))
		return
	}

	if e := consoleQuery.Validate(); e != nil {
		rest.APIError(c, i18n.GetMessage(c, "请求参数错误{}", e))
		return
	}

	// 自定义命令行
	commands, err := consoleQuery.SplitCommand()
	if err != nil {
		rest.APIError(
			c, i18n.GetMessage(c, "请求参数错误, command not valid{}", err))
		return
	}

	podCtx, err := podmanager.QueryOpenPodCtx(c.Request.Context(), authCtx.ClusterId, consoleQuery)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "请求参数错误{}", err))
		return
	}

	podCtx.ProjectId = authCtx.ProjectId
	// bkapigw 校验, 使用 Operator 做用户标识
	podCtx.Username = consoleQuery.Operator
	// 设置临时 session 过期时间
	podCtx.ConnIdleTimeout = consoleQuery.ConnIdleTimeout
	podCtx.SessionTimeout = consoleQuery.SessionTimeout
	podCtx.Viewers = consoleQuery.Viewers

	if len(commands) > 0 {
		podCtx.Commands = commands
	}

	sessionId, err := sessions.NewStore().OpenAPIScope().Set(c.Request.Context(), podCtx)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "获取session失败{}", err))
		return
	}

	data := map[string]string{
		"session_id":      sessionId,
		"web_console_url": makeWebConsoleURL(sessionId, podCtx),
	}

	// 这里直接置换新的session_id
	if consoleQuery.WSAcquire {
		wsSessionId, err := sessions.NewStore().WebSocketScope().Set(c.Request.Context(), podCtx)
		if err != nil {
			rest.APIError(c, i18n.GetMessage(c, "获取session失败{}", err))
			return
		}

		data["ws_url"] = makeWebSocketURL(wsSessionId, "", true)
	}

	rest.APIOK(c, i18n.GetMessage(c, "获取session成功"), data)
}

// makeWebConsoleURL webconsole 页面访问地址
func makeWebConsoleURL(sessionId string, podCtx *types.PodContext) string {
	u := *config.G.Web.BaseURL
	u.Path = path.Join(u.Path, "/portal/container/") + "/"

	query := url.Values{}
	query.Set("session_id", sessionId)
	query.Set("container_name", podCtx.ContainerName)

	u.RawQuery = query.Encode()

	return u.String()
}

// makeWebSocketURL http 转换为 ws 协议链接
func makeWebSocketURL(sessionId, lang string, withScheme bool) string {
	u := *config.G.Web.BaseURL
	u.Path = path.Join(u.Path, "/ws/sessions/", sessionId) + "/"

	query := url.Values{}
	if lang != "" {
		query.Set("lang", lang)
	}

	u.RawQuery = query.Encode()

	// https 协议 转换为 wss
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	// 去掉前缀, web 使用
	if !withScheme {
		u.Scheme = ""
		u.Host = ""
	}

	return u.String()
}

// CreateClusterPortalSession 集群级别的 webconsole openapi
func (s *service) CreateClusterPortalSession(c *gin.Context) {
	rest.APIError(c, "Not implemented")
}

// SetUserDelaySwitch 开启/关闭某个用户命令延时统计API
func (s *service) SetUserDelaySwitch(c *gin.Context) {
	// 参数解析
	username := c.Param("username")
	var commandDelays types.CommandDelay
	err := c.BindJSON(&commandDelays)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "请求参数错误{}", err))
		return
	}

	if len(commandDelays.ConsoleKey) != 1 {
		rest.APIError(c, i18n.GetMessage(c, "请求参数错误{}", errors.New("invalid console_key")))
		return
	}

	// 将用户设置的延时命令开关放到 redis 上保存
	err = storage.GetDefaultRedisSession().Client.HSet(c, types.GetMeterKey(), username, commandDelays.HashValue()).Err()
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "服务请求失败", err))
		return
	}
	rest.APIOK(c, i18n.GetMessage(c, "服务请求成功"), nil)
}

// GetUserDelaySwitch 获取某个用户命令延时统计API
func (s *service) GetUserDelaySwitch(c *gin.Context) {
	username := c.Param("username")
	result, err := storage.GetDefaultRedisSession().Client.HGet(c, types.GetMeterKey(), username).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			rest.APIError(c, i18n.GetMessage(c, "用户没有设置命令延时", err))
			return
		}
		rest.APIError(c, i18n.GetMessage(c, "服务请求失败", err))
		return
	}

	// 将用户设置的延时命令转成结构体数组输出
	commandDelay, err := types.MakeCommandDelay(result)
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "服务请求失败", err))
		return
	}
	rest.APIOK(c, i18n.GetMessage(c, "服务请求成功"), commandDelay)
}

// GetUserDelayMeter 查看用户+集群(选填)命令延时情况 API
func (s *service) GetUserDelayMeter(c *gin.Context) {
	username := c.Param("username")
	clusterId := c.Query("clusterId")
	key := types.GetMeterDataKey(username)
	result, err := storage.GetDefaultRedisSession().Client.LRange(c, key, 0, -1).Result()
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "服务请求失败", err))
		return
	}

	// 去重处理，键值为cluster_id
	userMeterMap := make(map[string]types.UserMeters)
	// 根据clusterId筛选
	for i := range result {
		// clusterId不为空的情况并且用户没有设置该clusterId的情况不做数据展示
		if clusterId != "" && !strings.Contains(result[i], clusterId) {
			continue
		}

		// 解析Redis中数据， string -> delayData
		var delayData types.DelayData
		err = json.Unmarshal([]byte(result[i]), &delayData)
		if err != nil {
			rest.APIError(c, i18n.GetMessage(c, "服务请求失败", err))
			return
		}

		// 解析消耗时间
		timeConsume, err := time.ParseDuration(delayData.TimeConsume)
		if err != nil {
			// 有错误则不追加列表
			continue
		}
		// 能取出数据，就追加
		if value, ok := userMeterMap[delayData.ClusterId]; ok {
			userMeter := types.UserConsume{
				TimeConsume: delayData.TimeConsume,
				CreateTime:  delayData.CreateTime,
				SessionId:   delayData.SessionId,
				PodName:     delayData.PodName,
				CommandKey:  delayData.CommandKey,
			}
			value.UserConsumes = append(value.UserConsumes, userMeter)
			// 统计数据
			value.AverageTimeConsume += timeConsume
			if value.MaxTimeConsume < timeConsume {
				value.MaxTimeConsume = timeConsume
			}
			if value.MinTimeConsume > timeConsume {
				value.MinTimeConsume = timeConsume
			}
			userMeterMap[delayData.ClusterId] = value
		} else {
			// 无法取出的情况则初始化值
			userMeterMap[delayData.ClusterId] = types.UserMeters{
				ClusterId:          delayData.ClusterId,
				AverageTimeConsume: timeConsume,
				MaxTimeConsume:     timeConsume,
				MinTimeConsume:     timeConsume,
				UserConsumes: []types.UserConsume{
					{
						TimeConsume: delayData.TimeConsume,
						CreateTime:  delayData.CreateTime,
						SessionId:   delayData.SessionId,
						PodName:     delayData.PodName,
						CommandKey:  delayData.CommandKey,
					},
				},
			}
		}
	}

	// 返回筛选的结果
	var rsp []types.UserMeterRsp
	for _, value := range userMeterMap {
		// 求平均值
		value.AverageTimeConsume /= time.Duration(len(value.UserConsumes))
		userMeterRsp := types.UserMeterRsp{
			ClusterId:          value.ClusterId,
			AverageTimeConsume: value.AverageTimeConsume.String(),
			MaxTimeConsume:     value.MaxTimeConsume.String(),
			MinTimeConsume:     value.MinTimeConsume.String(),
			UserConsumes:       value.UserConsumes,
		}
		rsp = append(rsp, userMeterRsp)
	}
	rest.APIOK(c, i18n.GetMessage(c, "服务请求成功"), rsp)
}

// GetDelayUsers 查看哪些用户开启命令延时情况 API
func (s *service) GetDelayUsers(c *gin.Context) {
	result, err := storage.GetDefaultRedisSession().Client.HGetAll(c, types.GetMeterKey()).Result()
	if err != nil {
		rest.APIError(c, i18n.GetMessage(c, "服务请求失败{}", err))
		return
	}

	// 以结构体数组的形式返回
	rsp := map[string]*types.CommandDelay{}
	for k, v := range result {
		commandDelay, err := types.MakeCommandDelay(v)
		if err != nil {
			rest.APIError(c, i18n.GetMessage(c, "服务请求失败{}", err))
			return
		}
		rsp[k] = commandDelay
	}

	rest.APIOK(c, i18n.GetMessage(c, "服务请求成功"), rsp)
}
