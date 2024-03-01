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

package argocd

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/gorilla/mux"

	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
)

// MonitorPlugin for internal monitor-scenario authorization
type MonitorPlugin struct {
	*mux.Router
	middleware mw.MiddlewareInterface
}

// GET:		/api/v1/monitor/{biz_id}, 查询bizID下对应的所有AppMonitor信息
// DELETE:	/api/v1/monitor/{biz_id}/{scenario}, 删除bizID下安装的对应场景监控
// POST:	/api/v1/monitor/{biz_id}/{scenario}, 创建/更新bizID下安装的对应场景监控

// Init all project sub path handler
// project plugin is a subRouter, all path registered is relative
func (plugin *MonitorPlugin) Init() error {
	plugin.Path("/{biz_id}").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.monitorViewHandler))
	plugin.Path("/{biz_id}/{scenario}").Methods(http.MethodPost, http.MethodDelete).Handler(plugin.middleware.
		HttpWrapper(plugin.monitorOperateHandler))
	blog.Infof("monitor plugin init successfully")
	return nil
}

// GET /api/v1/monitor/{biz_id}
func (plugin *MonitorPlugin) monitorViewHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	bizID := mux.Vars(r)["biz_id"]
	statusCode, err := plugin.middleware.CheckBusinessPermission(r.Context(), bizID, iam.ProjectView)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, fmt.Errorf("check businessID '%s' permission failed: %w", bizID,
			err))
	}
	return r, mw.ReturnMonitorReverse()
}

func (plugin *MonitorPlugin) monitorOperateHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	bizID := mux.Vars(r)["biz_id"]
	statusCode, err := plugin.middleware.CheckBusinessPermission(r.Context(), bizID, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, fmt.Errorf("check businessID '%s' permission failed: %w", bizID,
			err))
	}
	return r, mw.ReturnMonitorReverse()
}
