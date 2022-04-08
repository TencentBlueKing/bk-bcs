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

package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// handlerPath 定义采样路由struct
type handlerPath struct {
	sync.Map
}

// get 获取path
func (hp *handlerPath) get(handler string) string {
	v, ok := hp.Load(handler)
	if !ok {
		return ""
	}
	return v.(string)
}

// set 保存path到sync.Map
func (hp *handlerPath) set(ri gin.RouteInfo) {
	hp.Store(ri.Handler, ri.Path)
}

type Option func(*GinPrometheus)

type GinPrometheus struct {
	engine  *gin.Engine
	ignored map[string]bool
	pathMap *handlerPath
	updated bool
}

// Ignore 添加忽略的路径
func Ignore(path ...string) Option {
	return func(gp *GinPrometheus) {
		for _, p := range path {
			gp.ignored[p] = true
		}
	}
}

// New new gin prometheus
func New(e *gin.Engine, options ...Option) *GinPrometheus {
	if e == nil {
		return nil
	}

	gp := &GinPrometheus{
		engine: e,
		ignored: map[string]bool{
			bcsWebSocketHandlerPath: true,
		},
		pathMap: &handlerPath{},
	}

	for _, o := range options {
		o(gp)
	}
	return gp
}

// updatePath 更新path
func (gp *GinPrometheus) updatePath() {
	gp.updated = true
	for _, ri := range gp.engine.Routes() {
		gp.pathMap.set(ri)
	}
}

// Middleware set gin middleware
func (gp *GinPrometheus) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !gp.updated {
			gp.updatePath()
		}
		// 过滤请求
		if gp.ignored[c.FullPath()] {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		ReportAPIRequestMetric(c.HandlerName(), c.Request.Method, gp.respStatusTransform(c.Writer.Status()), start)
	}
}

// respStatusTransform api返回状态转换
func (gp *GinPrometheus) respStatusTransform(status int) string {
	if status == http.StatusOK {
		return SucStatus
	}
	return ErrStatus
}
