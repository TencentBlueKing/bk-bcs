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
 *
 */

package api

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	gintrace "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/gin"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/tracing"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
)

const (
	// FileSizeUnitMb
	FileSizeUnitMb = 1024 * 1024
	// FileSizeLimits
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
		metrics.RequestCollect("CreateWebConsoleSession"), route.PermissionRequired(), s.CreateWebConsoleSession)
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
}

// ListClusters 集群列表
func (s *service) ListClusters(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)

	projectId := c.Param("projectId")
	project, err := bcs.GetProject(c.Request.Context(), config.G.BCS, projectId)
	if err != nil {
		APIError(c, i18n.GetMessage(c, "项目不正确"))
		return
	}

	clusters, err := bcs.ListClusters(c.Request.Context(), project.ProjectId)
	if err != nil {
		APIError(c, i18n.GetMessage(c, err.Error()))
		return
	}
	data := types.APIResponse{
		Data:      clusters,
		Code:      types.NoError,
		Message:   i18n.GetMessage(c, "获取集群成功"),
		RequestID: authCtx.RequestId,
	}
	c.JSON(http.StatusOK, data)
}

// CreateWebConsoleSession 创建websocket session
func (s *service) CreateWebConsoleSession(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)

	consoleQuery := new(podmanager.ConsoleQuery)
	c.BindQuery(consoleQuery)

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
		APIError(c, i18n.GetMessage(c, err.Error()))
		return
	}

	podCtx.ProjectId = authCtx.ProjectId
	podCtx.Username = authCtx.Username
	podCtx.Source = consoleQuery.Source

	sessionId, err := sessions.NewStore().WebSocketScope().Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage(c, "获取session失败{}", err)
		APIError(c, msg)
		return
	}

	data := types.APIResponse{
		Data: map[string]string{
			"session_id": sessionId,
			"ws_url":     makeWebSocketURL(sessionId, consoleQuery.Lang, false),
		},
		Code:      types.NoError,
		Message:   i18n.GetMessage(c, "获取session成功"),
		RequestID: authCtx.RequestId,
	}
	c.JSON(http.StatusOK, data)
}

// UploadHandler 上传文件
// NOCC:golint/fnsize(设计如此:)
func (s *service) UploadHandler(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)
	uploadPath := c.PostForm("upload_path")
	sessionId := c.Param("sessionId")
	data := types.APIResponse{RequestID: authCtx.RequestId}
	if uploadPath == "" {
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "请先输入上传路径")
		c.JSON(http.StatusOK, data)
		return
	}
	err := checkFileExists(uploadPath, sessionId)
	if err != nil {
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "目标路径不存在")
		c.JSON(http.StatusOK, data)
		return
	}
	err = checkPathIsDir(uploadPath, sessionId)
	if err != nil {
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "目标路径不存在")
		c.JSON(http.StatusOK, data)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		logger.Errorf("get file from request failed, err: %s", err.Error())
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "解析上传文件失败")
		c.JSON(http.StatusOK, data)
		return
	}

	opened, err := file.Open()
	if err != nil {
		logger.Errorf("open file from request failed, err: %s", err.Error())
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "解析上传文件失败")
		c.JSON(http.StatusOK, data)
		return
	}
	defer opened.Close()

	podCtx, err := sessions.NewStore().WebSocketScope().Get(c.Request.Context(), sessionId)
	if err != nil {
		logger.Errorf("get pod context by session %s failed, err: %s", sessionId, err.Error())
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "获取pod信息失败")
		c.JSON(http.StatusOK, data)
		return
	}
	reader, writer := io.Pipe()
	pe, err := podCtx.NewPodExec()
	if err != nil {
		logger.Errorf("new pod exec failed, err: %s", err.Error())
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "执行上传命令失败")
		c.JSON(http.StatusOK, data)
		return
	}
	errChan := make(chan error, 1)
	go func(r io.Reader, pw *io.PipeWriter) {
		tarWriter := tar.NewWriter(writer)
		defer func() {
			tarWriter.Close()
			writer.Close()
			close(errChan)
		}()
		e := tarWriter.WriteHeader(&tar.Header{
			Name: file.Filename,
			Size: file.Size,
			Mode: 0644,
		})
		if e != nil {
			logger.Errorf("writer tar header failed, err: %s", e.Error())
			errChan <- e
			return
		}
		_, e = io.Copy(tarWriter, opened)
		if e != nil {
			logger.Errorf("writer tar from opened file failed, err: %s", e.Error())
			errChan <- e
			return
		}
		errChan <- nil
	}(opened, writer)

	pe.Stdin = reader
	// 需要同时读取 stdout/stderr, 否则可能会 block 住
	pe.Stdout = &bytes.Buffer{}
	pe.Stderr = &bytes.Buffer{}

	pe.Command = []string{"tar", "-xmf", "-", "-C", uploadPath}
	pe.Tty = false

	if err := pe.Exec(); err != nil {
		logger.Errorf("pod exec failed, err: %s", err.Error())
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "执行上传命令失败")
		c.JSON(http.StatusOK, data)
		return
	}

	err, ok := <-errChan
	if ok && err != nil {
		logger.Errorf("writer to tar failed, err: %s", err.Error())
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "文件上传失败")
		c.JSON(http.StatusOK, data)
		return
	}
	data.Code = types.NoError
	data.Message = i18n.GetMessage(c, "文件上传成功")
	c.JSON(http.StatusOK, data)
}

// DownloadHandler 下载文件
func (s *service) DownloadHandler(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)
	data := types.APIResponse{RequestID: authCtx.RequestId}
	downloadPath := c.Query("download_path")
	sessionId := c.Param("sessionId")
	reader, writer := io.Pipe()
	defer reader.Close()
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			writer.Close()
			close(errChan)
		}()
		podCtx, err := sessions.NewStore().WebSocketScope().Get(c.Request.Context(), sessionId)
		if err != nil {
			errChan <- err
			return
		}

		pe, err := podCtx.NewPodExec()
		if err != nil {
			errChan <- err
			return
		}
		pe.Stdout = writer

		pe.Command = append([]string{"tar", "cf", "-"}, downloadPath)
		pe.Stderr = &bytes.Buffer{}
		pe.Tty = false
		err = pe.Exec()
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()
	tarReader := tar.NewReader(reader)
	_, err := tarReader.Next()
	if err != nil {
		data.Code = types.ApiErrorCode
		data.Message = i18n.GetMessage(c, "复制文件流失败")
		c.JSON(http.StatusOK, data)
		return
	}
	fileName := downloadPath[strings.LastIndex(downloadPath, "/")+1:]
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("X-File-Name", fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")
	io.Copy(c.Writer, tarReader)
}

// CheckDownloadHandler 下载文件预检查
func (s *service) CheckDownloadHandler(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)
	data := types.APIResponse{RequestID: authCtx.RequestId, Code: types.ApiErrorCode}
	downloadPath := c.Query("download_path")
	sessionId := c.Param("sessionId")

	if err := checkFileExists(downloadPath, sessionId); err != nil {
		data.Message = i18n.GetMessage(c, "目标文件不存在")
		c.JSON(http.StatusOK, data)
		return
	}

	if err := checkPathIsDir(downloadPath, sessionId); err == nil {
		data.Message = i18n.GetMessage(c, "暂不支持文件夹下载")
		c.JSON(http.StatusOK, data)
		return
	}

	if err := checkFileSize(downloadPath, sessionId, FileSizeLimits*FileSizeUnitMb); err != nil {
		data.Message = i18n.GetMessage(c, "文件不能超过{}MB", map[string]int{"fileLimit": FileSizeLimits})
		c.JSON(http.StatusOK, data)
		return
	}

	data.Code = types.NoError
	data.Message = i18n.GetMessage(c, "文件可以下载")
	c.JSON(http.StatusOK, data)
}

func checkPathIsDir(path, sessionID string) error {
	podCtx, err := sessions.NewStore().WebSocketScope().Get(context.Background(), sessionID)
	if err != nil {
		return err
	}

	pe, err := podCtx.NewPodExec()
	if err != nil {
		return err
	}
	pe.Command = append([]string{"test", "-d"}, path)
	pe.Stdout = &bytes.Buffer{}
	pe.Stderr = &bytes.Buffer{}
	pe.Tty = false
	err = pe.Exec()
	if err != nil {
		return err
	}
	return nil
}

func checkFileExists(path, sessionID string) error {
	podCtx, err := sessions.NewStore().WebSocketScope().Get(context.Background(), sessionID)
	if err != nil {
		return err
	}

	pe, err := podCtx.NewPodExec()
	if err != nil {
		return err
	}
	pe.Command = append([]string{"test", "-e"}, path)
	pe.Stdout = &bytes.Buffer{}
	pe.Stderr = &bytes.Buffer{}
	pe.Tty = false
	err = pe.Exec()
	if err != nil {
		return err
	}
	return nil
}

func checkFileSize(path, sessionID string, sizeLimit int) error {
	podCtx, err := sessions.NewStore().WebSocketScope().Get(context.Background(), sessionID)
	if err != nil {
		return err
	}

	pe, err := podCtx.NewPodExec()
	if err != nil {
		return err
	}
	pe.Command = []string{"stat", "-c", "%s", path}
	stdout := &bytes.Buffer{}
	pe.Stdout = stdout
	pe.Stderr = &bytes.Buffer{}
	pe.Tty = false
	err = pe.Exec()
	if err != nil {
		return err
	}
	// 解析文件大小, stdout 会返回 \r\n 或者 \n
	sizeText := strings.TrimSuffix(stdout.String(), "\n")
	sizeText = strings.TrimSuffix(sizeText, "\r")
	size, err := strconv.Atoi(sizeText)
	if err != nil {
		return err
	}
	if size > sizeLimit {
		return errors.Errorf("file size %d > %d", size, sizeLimit)
	}
	return nil
}

// CreatePortalSession xxx
func (s *service) CreatePortalSession(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)
	if authCtx.BindSession == nil {
		msg := i18n.GetMessage(c, "session_id不合法或已经过期")
		APIError(c, msg)
		return
	}

	podCtx := authCtx.BindSession

	sessionId, err := sessions.NewStore().WebSocketScope().Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage(c, "获取session失败{}", err)
		APIError(c, msg)
		return
	}

	lang := c.Query("lang")
	lang = strings.TrimSuffix(lang, "/")
	data := types.APIResponse{
		Data: map[string]string{
			"session_id": sessionId,
			"ws_url":     makeWebSocketURL(sessionId, lang, false),
		},
		Code:      types.NoError,
		Message:   i18n.GetMessage(c, "获取session成功"),
		RequestID: authCtx.RequestId,
	}
	c.JSON(http.StatusOK, data)
}

// CreateContainerPortalSession 创建 webconsole url api
func (s *service) CreateContainerPortalSession(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)

	consoleQuery := new(podmanager.OpenQuery)

	err := c.BindJSON(consoleQuery)
	if err != nil {
		msg := i18n.GetMessage(c, "请求参数错误{}", err)
		APIError(c, msg)
		return
	}

	if e := consoleQuery.Validate(); e != nil {
		msg := i18n.GetMessage(c, "请求参数错误{}", e)
		APIError(c, msg)
		return
	}

	// 自定义命令行
	commands, err := consoleQuery.SplitCommand()
	if err != nil {
		msg := i18n.GetMessage(c, "请求参数错误, command not valid{}", err)
		APIError(c, msg)
		return
	}

	podCtx, err := podmanager.QueryOpenPodCtx(c.Request.Context(), authCtx.ClusterId, consoleQuery)
	if err != nil {
		msg := i18n.GetMessage(c, "请求参数错误{}", err)
		APIError(c, msg)
		return
	}

	podCtx.ProjectId = authCtx.ProjectId
	// bkapigw 校验, 使用 Operator 做用户标识
	podCtx.Username = consoleQuery.Operator
	podCtx.ConnIdleTimeout = consoleQuery.ConnIdleTimeout
	podCtx.SessionTimeout = consoleQuery.SessionTimeout
	podCtx.Viewers = consoleQuery.Viewers

	if len(commands) > 0 {
		podCtx.Commands = commands
	}

	sessionId, err := sessions.NewStore().OpenAPIScope().Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage(c, "获取session失败{}", err)
		APIError(c, msg)
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
			msg := i18n.GetMessage(c, "获取session失败{}", err)
			APIError(c, msg)
			return
		}

		data["ws_url"] = makeWebSocketURL(wsSessionId, "", true)
	}

	respData := types.APIResponse{
		Data:      data,
		Code:      types.NoError,
		Message:   i18n.GetMessage(c, "获取session成功"),
		RequestID: authCtx.RequestId,
	}

	c.JSON(http.StatusOK, respData)
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
	rest.AbortWithBadRequestError(c, errors.New("Not implemented"))
}

// APIError 简易的错误返回
func APIError(c *gin.Context, msg string) {
	authCtx := route.MustGetAuthContext(c)

	data := types.APIResponse{
		Code:      types.ApiErrorCode,
		Message:   msg,
		RequestID: authCtx.RequestId,
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, data)
}
