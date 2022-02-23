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

package handler

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/web"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"

	"github.com/gin-gonic/gin"
)

func NewRouteRegistrar() route.Registrar {
	return BcsWebconsole{}
}

func (e BcsWebconsole) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/api/ping", e.Ping)

}

func (e *BcsWebconsole) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func Register(opts *route.Options) error {
	router := opts.Router
	h := NewRouteRegistrar()
	h.RegisterRoute(router.Group(""))
	for _, r := range []route.Registrar{
		web.NewRouteRegistrar(opts),
		api.NewRouteRegistrar(opts),
	} {
		r.RegisterRoute(router.Group(""))
	}
	return nil
}
