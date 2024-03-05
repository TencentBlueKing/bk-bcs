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

// Package server implements a http server for bcs-terraform-controller
package server

import (
	"net"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Handler server
type Handler interface {
	// Init handler
	Init() error
	// Run handler
	Run() error
}

// handler impl Handler
type handler struct {
	// addr http srv addr
	addr string
	// port http srv port
	port string
	// svr http server
	svr *http.Server
	// router gin router
	router *gin.Engine

	// client k8s api server client
	client client.Client
}

// NewHandler new handler
func NewHandler(addr, port string, client client.Client) Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	return &handler{
		router: router,
		addr:   addr,
		port:   port,
		client: client,
		svr:    new(http.Server),
	}
}

// registerRouter register
func (h *handler) registerRouter() {
	// 执行tf
	h.router.POST("/apply", h.Apply)
	// 创建plan
	h.router.POST("/plan", h.CreatePlan)

	// 查询所有的tf资源
	h.router.GET("/listTerraform", h.ListTerraform)
	// 查询所有的tf资源
	h.router.GET("/getTerraform", h.GetTerraform)
}

// Init handler
func (h *handler) Init() error {
	h.registerRouter()
	h.svr.Handler = h.router
	h.svr.Addr = net.JoinHostPort(h.addr, h.port)

	return nil
}

// Run handler
func (h *handler) Run() error {
	blog.Infof("http server listen: %s", h.svr.Addr)
	return h.svr.ListenAndServe()
}
