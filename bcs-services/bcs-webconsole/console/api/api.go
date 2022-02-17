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
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/web"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// Router is api router
type Router struct {
	sync.RWMutex
	conf    *config.ConsoleConfig
	backend manager.Manager
}

// NewRouter return api router
func NewRouter(b manager.Manager, conf *config.ConsoleConfig) *Router {
	r := &Router{
		backend: b,
		conf:    conf,
	}

	r.initRoutes()
	return r
}

// 声明session存储
var (
	store = sessions.NewFilesystemStore("./", securecookie.GenerateRandomKey(32),
		securecookie.GenerateRandomKey(32))
)

// 注册路由
func (r *Router) initRoutes() {

	//handler container web console
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.FS(web.FS)))

	// websocket
	mux.HandleFunc("/web_console/projects/clusters/ws", r.BCSWebSocketHandler) // ws连接

	// 对sessionID进行校验，返回ws地址
	mux.HandleFunc("/api/projects/clusters/web_console/session", r.WebConsoleSession)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", r.conf.Address, r.conf.Port),
		Handler: mux,
	}
	if r.conf.ServCert.IsSSL {
		tlsConf, err := ssl.ServerTslConf(r.conf.ServCert.CAFile, r.conf.ServCert.CertFile, r.conf.ServCert.KeyFile, r.conf.ServCert.CertPasswd)
		if err != nil {
			blog.Error("fail to load certfile, err:%s", err.Error())
			return
		}
		s.TLSConfig = tlsConf
		blog.Info("Start https service on(%s:%d)", r.conf.Address, r.conf.Port)
		go func() {
			err := s.ListenAndServeTLS("", "")
			blog.Errorf("tls server failed, err : %v", err)
		}()
	} else {
		blog.Info("Start http service on(%s:%d)", r.conf.Address, r.conf.Port)
		go func() {
			err := s.ListenAndServe()
			blog.Errorf("insecure server failed, err : %v", err)
		}()
	}
}

// WebConsoleSession 获取ws连接地址
func (r *Router) WebConsoleSession(w http.ResponseWriter, req *http.Request) {

	data := types.APIResponse{
		Code: 1, // TODO code待确认
		Data: map[string]string{},
	}

	session, err := store.Get(req, "sessionID")
	if err != nil {
		data.Message = "获取session失败！"
		manager.ResponseJSON(w, http.StatusBadRequest, data)
		return
	}

	projectID := req.URL.Query().Get("projects")
	clustersID := req.URL.Query().Get("clusters")

	podName, err := r.backend.GetK8sContext(context.Background(), projectID, clustersID)
	if err != nil {
		data.Message = "获取session失败！"
		manager.ResponseJSON(w, http.StatusBadRequest, data)
		return
	}
	// 把创建好的pod信息保存到用户数据
	userPodData := &types.UserPodData{
		ProjectID:  projectID,
		ClustersID: clustersID,
		PodName:    podName,
		SessionID:  session.ID,
		CrateTime:  time.Now(),
	}
	r.backend.WritePodData(userPodData)

	// TODO 封装获取wsURL方法
	wsUrl := "ws://127.0.0.1:8080/web_console/projects/clusters/ws?projectsID=%s&clustersID=%s&session_id=%s"
	wsUrl = fmt.Sprintf(wsUrl, projectID, clustersID, session.ID)
	data.Code = 0
	data.Message = "获取session成功"
	data.Data = map[string]string{
		"session_id": session.ID,
		"ws_url":     wsUrl,
	}

	manager.ResponseJSON(w, http.StatusOK, data)
}

func (r *Router) BCSWebSocketHandler(w http.ResponseWriter, req *http.Request) {

	data := types.APIResponse{
		Code: 1, // TODO code待确认
		Data: map[string]string{},
	}

	projectID := req.URL.Query().Get("projectsID")
	clustersID := req.URL.Query().Get("clustersID")

	// 获取这个用户的信息
	session, err := store.Get(req, "sessionID")
	if err != nil {
		data.Message = "获取session失败！"
		manager.ResponseJSON(w, http.StatusBadRequest, data)
		return
	}

	if session.IsNew {
		data.Message = "没有对应的pod资源！"
		manager.ResponseJSON(w, http.StatusBadRequest, data)
		return
	}

	podData, ok := r.backend.ReadPodData(session.ID, projectID, clustersID)
	if !ok {
		data.Message = "没有对应的pod资源！"
		manager.ResponseJSON(w, http.StatusBadRequest, data)
		return
	}

	webConsole := &types.WebSocketConfig{
		PodName:    podData.PodName,
		User:       podData.UserName,
		ClusterID:  clustersID,
		ProjectsID: projectID,
	}
	c := &gin.Context{
		Request: req,
		//Writer: w,
	}

	// handler container web console
	r.backend.StartExec(c, webConsole)
}
