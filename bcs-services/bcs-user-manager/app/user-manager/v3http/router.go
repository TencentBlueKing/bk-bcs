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

// Package v3http xxx
package v3http

import (
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/middleware"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v3http/activity"
)

// InitV3Routers init v3 version route,
func InitV3Routers(ws *restful.WebService) {
	ws.Filter(middleware.RequestIDFilter)
	ws.Filter(middleware.ProjectFilter)
	ws.Filter(middleware.TracingFilter)
	ws.Filter(middleware.LoggingFilter)

	initActivityLogRouters(ws)
}

// initActivityLogRouters init activity log api routers
func initActivityLogRouters(ws *restful.WebService) {
	ws.Route(auth.ManagerAuthFunc(ws.POST("/activity_logs")).To(activity.PushActivities))
	ws.Route(auth.ProjectViewFunc(auth.TokenAuthenticateV2Func(ws.GET("/projects/{project_code}/activity_logs"))).
		To(activity.SearchActivities))
}
