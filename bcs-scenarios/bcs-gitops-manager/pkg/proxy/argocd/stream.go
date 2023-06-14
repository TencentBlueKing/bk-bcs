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
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

// * addinational stream path, application wrapper
// GET：/api/v1/stream/applications?projects={projects}，获取event事件流，强制启用projects
// GET：/api/v1/stream/applications/{name}/resource-tree，指定资源树事件流

// StreamPlugin for internal streaming
type StreamPlugin struct {
	*mux.Router

	appHandler *AppPlugin
	middleware MiddlewareInterface
}

// Init all project sub path handler
// project plugin is a subRouter, all path registered is relative
func (plugin *StreamPlugin) Init() error {
	// done(DeveloperJim): GET /api/v1/stream/applications?projects={projects}
	plugin.Path("").Methods("GET").Queries("projects", "{projects}").
		Handler(plugin.middleware.HttpWrapper(plugin.projectViewHandler))

	// done(DeveloperJim): GET /api/v1/stream/applications/{name}/resource-tree
	plugin.Path("/{name}/resource-tree").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.appHandler.applicationViewsHandler))
	blog.Infof("argocd stream applications plugin init successfully")
	return nil
}

func (plugin *StreamPlugin) projectViewHandler(ctx context.Context, r *http.Request) *httpResponse {
	projectName := r.URL.Query().Get("projects")
	_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, projectName, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check project '%s' permission failed", projectName),
		}
	}
	return nil
}
