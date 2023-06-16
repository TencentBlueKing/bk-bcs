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

// ProjectPlugin for internal project authorization
type ProjectPlugin struct {
	*mux.Router
	middleware MiddlewareInterface
}

// Init all project sub path handler
// project plugin is a subRouter, all path registered is relative
func (plugin *ProjectPlugin) Init() error {
	// not allow requests
	// POST /api/v1/projects
	plugin.Path("").Methods("POST").HandlerFunc(plugin.forbidden)
	// DELETE and Update /api/v1/projects/{name}
	plugin.HandleFunc("/{name}", plugin.forbidden).Methods("DELETE")
	plugin.HandleFunc("/{name}", plugin.forbidden).Methods("PUT")

	// requests by authorization
	// GET /api/v1/projects
	plugin.Path("").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.listProjectsHandler))
	// GET /api/v1/projects/{name}
	plugin.Path("/{name}").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.projectViewsHandler))
	// GET /api/v1/projects/{name}/{details}:
	// detailed, events, syncwindows, globalprojects all in one handler
	plugin.Path("/{name}/{details}").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.projectViewsHandler))

	// deny token access
	plugin.PathPrefix("/{name}/roles/").HandlerFunc(plugin.forbidden)

	blog.Infof("argocd project plugin init successfully")
	return nil
}

// forbidden specified path
func (plugin *ProjectPlugin) forbidden(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

// GET /api/v1/projects
func (plugin *ProjectPlugin) listProjectsHandler(ctx context.Context, r *http.Request) *httpResponse {
	projectList, statusCode, err := plugin.middleware.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list projects failed"),
		}
	}
	return &httpResponse{
		statusCode: statusCode,
		obj:        projectList,
	}
}

// handle path projectPermission belows:
// GET /api/v1/projects/{name}
// GET /api/v1/projects/{name}/detailed
// GET /api/v1/projects/{name}/events
// GET /api/v1/projects/{name}/globalprojects
// GET /api/v1/projects/{name}/syncwindows
func (plugin *ProjectPlugin) projectViewsHandler(ctx context.Context, r *http.Request) *httpResponse {
	projectName := mux.Vars(r)["name"]
	_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, projectName, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check project '%s' view permission failed", projectName),
		}
	}
	return nil
}
