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

// ClusterPlugin for internal cluster authorization
type ClusterPlugin struct {
	*mux.Router
	middleware MiddlewareInterface
}

// Init implementation for plugin
func (plugin *ClusterPlugin) Init() error {
	// not allow requests
	// POST /api/v1/clusters, create cluster, deny
	plugin.Path("").Methods("POST").HandlerFunc(plugin.forbidden)
	// DELETE and Update /api/v1/clusters/{name}, deny
	plugin.HandleFunc("/{name}", plugin.forbidden).Methods("DELETE")
	plugin.HandleFunc("/{name}", plugin.forbidden).Methods("PUT")

	// requests by authorization, addinational URL managed by BCS
	// GET /api/v1/clusters?project=bcs-project, list all projects,
	plugin.Path("").Methods("GET").Queries("projects", "{projects}").
		Handler(plugin.middleware.HttpWrapper(plugin.listClustersHandler))

	// GET /api/v1/clusters/{name}, get specified cluster info
	plugin.Path("/{name}").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.clusterViewHandler))

	// GET /api/v1/clusters/{name}/{details}
	// details: invalidate-cache, rotate-auth
	plugin.HandleFunc("/{name}/invalidate-cache", plugin.forbidden).Methods("GET")
	plugin.HandleFunc("/{name}/rotate-auth", plugin.forbidden).Methods("GET")

	blog.Infof("argocd cluster plugin init successfully")
	return nil
}

// forbidden specified path
func (plugin *ClusterPlugin) forbidden(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

// GET /api/v1/clusters
func (plugin *ClusterPlugin) listClustersHandler(ctx context.Context, r *http.Request) *httpResponse {
	params := mux.Vars(r)
	projectName, ok := params["projects"]
	if !ok || projectName == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        errors.Errorf("lost projects param"),
		}
	}
	clusterList, statusCode, err := plugin.middleware.ListClusters(ctx, []string{projectName})
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list clusters by project '%s' failed", projectName),
		}
	}
	return &httpResponse{
		statusCode: statusCode,
		obj:        clusterList,
	}
}

// GET /api/v1/clusters/{name}
func (plugin *ClusterPlugin) clusterViewHandler(ctx context.Context, r *http.Request) *httpResponse {
	clusterName := mux.Vars(r)["name"]
	statusCode, err := plugin.middleware.CheckClusterPermission(ctx, clusterName, iam.ClusterView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check cluster '%s' view permision failed", clusterName),
		}
	}
	return nil
}
