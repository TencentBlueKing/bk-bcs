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
	"github.com/gorilla/mux"

	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
)

// PreCheckPlugin plugin
type PreCheckPlugin struct {
	*mux.Router
	permitChecker permitcheck.PermissionInterface
	middleware    mw.MiddlewareInterface
}

// GET:		/api/v1/monitor/{biz_id}, 查询bizID下对应的所有AppMonitor信息
// GET:		/api/v1/monitor/{biz_id}/list_scenario, 查询bizID下能安装的场景信息
// DELETE:	/api/v1/monitor/{biz_id}/{scenario}, 删除bizID下安装的对应场景监控
// POST:	/api/v1/monitor/{biz_id}/{scenario}, 创建/更新bizID下安装的对应场景监控

// Init all project sub path handler
// project plugin is a subRouter, all path registered is relative
func (plugin *PreCheckPlugin) Init() error {
	plugin.Path("/mr/info").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.preCheckViewHandler))
	plugin.Path("/record").Methods(http.MethodPost).Handler(plugin.middleware.HttpWrapper(plugin.common))
	plugin.Path("/task").Methods(http.MethodPut).Handler(plugin.middleware.HttpWrapper(plugin.common))
	plugin.Path("/task").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.common))
	plugin.Path("/tasks").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.common))
	blog.Infof("precheck plugin init successfully")
	return nil
}

// GET /api/v1/monitor/{biz_id}
func (plugin *PreCheckPlugin) preCheckViewHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	// repo := r.PathValue("repository")
	queryValues := r.URL.Query()
	repo := queryValues.Get("repository")
	// mrIID := r.PathValue("mrIID")
	_, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), repo, permitcheck.RepoViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, fmt.Errorf("check repo '%s' permission failed: %w", repo,
			err))
	}
	return r, mw.ReturnPreCheckReverse()
}

func (plugin *PreCheckPlugin) common(r *http.Request) (*http.Request, *mw.HttpResponse) {
	return r, mw.ReturnPreCheckReverse()
}
