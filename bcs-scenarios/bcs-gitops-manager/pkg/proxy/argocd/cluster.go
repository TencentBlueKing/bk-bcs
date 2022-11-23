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

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
)

// ClusterPlugin for internal cluster authorization
type ClusterPlugin struct {
	*mux.Router

	Session    *Session
	Permission *cluster.BCSClusterPerm
	option     *proxy.GitOpsOptions
}

// Init implementation for plugin
func (plugin *ClusterPlugin) Init() error {
	// init BCSClusterPerm
	plugin.Permission = cluster.NewBCSClusterPermClient(plugin.option.IAMClient)

	// not allow requests
	// POST /api/v1/clusters, create cluster, deny
	plugin.Path("").Methods("POST").HandlerFunc(plugin.forbidden)
	// DELETE and Update /api/v1/clusters/{name}, deny
	plugin.HandleFunc("/{name}", plugin.forbidden).Methods("DELETE")
	plugin.HandleFunc("/{name}", plugin.forbidden).Methods("PUT")

	// requests by authorization, addinational URL managed by BCS
	// GET /api/v1/clusters?project=bcs-project, list all projects,
	plugin.Path("").Methods("GET").Queries("projects", "{projects}").HandlerFunc(plugin.listClustersHandler)

	// GET /api/v1/clusters/{name}, get specified cluster info
	plugin.HandleFunc("/{name}", plugin.clusterViewHandler).Methods("GET")

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
func (plugin *ClusterPlugin) listClustersHandler(w http.ResponseWriter, r *http.Request) {
	// check header user info
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("request %s get jwt token failure, %s", r.URL.Path, err.Error())
		http.Error(w,
			fmt.Sprintf("Bad Request: %s", err.Error()),
			http.StatusBadRequest,
		)
		return
	}
	params := mux.Vars(r)
	project, ok := params["projects"]
	if !ok {
		blog.Errorf("request %s lost additional project info", r.URL.Path)
		http.Error(w, "Bad Request: lost project", http.StatusBadRequest)
		return
	}
	// get all clusters from argocd
	clusterList, err := plugin.option.Storage.ListCluster(r.Context())
	if err != nil {
		blog.Errorf("request %s get all clusters from argocd failed, %s", r.URL.Path, err.Error())
		http.Error(w, "gitops storage failure", http.StatusInternalServerError)
		return
	}
	projectID := ""
	controlledClusters := make(map[string]v1alpha1.Cluster)
	clusterIDs := make([]string, 0)
	for _, cls := range clusterList.Items {
		if cls.Project != project {
			continue
		}
		controllID := common.GetBCSProjectID(cls.Annotations)
		if controllID == "" {
			continue
		}
		if projectID == "" {
			projectID = controllID
		}
		controlledClusters[cls.Name] = cls
		clusterIDs = append(clusterIDs, cls.Name)
	}
	if len(controlledClusters) == 0 {
		blog.Infof("request %s found no project %s clusters", r.URL.Path, project)
		proxy.JSONResponse(w, &v1alpha1.ClusterList{})
		return
	}
	// list permission verify
	action := string(cluster.ClusterView)
	result, err := plugin.Permission.GetMultiClusterMultiActionPermission(
		user.GetUser(), projectID, clusterIDs, []string{action})
	if err != nil {
		blog.Errorf("request %s authenticate clusters %+v failure, %s", r.URL.Path, clusterIDs, err.Error())
		http.Error(w, "authentication failure", http.StatusInternalServerError)
		return
	}
	// response filter data
	finnals := make([]v1alpha1.Cluster, 0)
	for clusterName, permit := range result {
		if permit[action] {
			cluster := controlledClusters[clusterName]
			finnals = append(finnals, cluster)
			blog.Infof("cluster %s for user %s view permission pass", cluster.Name, user.GetUser())
		}
	}
	blog.Infof("user %s request %s, project %s, %d clusters retrive",
		user.GetUser(), r.URL.Path, project, len(finnals))
	clusterList.Items = finnals
	proxy.JSONResponse(w, clusterList)
}

// GET /api/v1/clusters/{name}
func (plugin *ClusterPlugin) clusterViewHandler(w http.ResponseWriter, r *http.Request) {
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("ClusterPlugin get jwt info from request %s failed, %s", r.URL.Path, err.Error())
		http.Error(w,
			fmt.Sprintf("Bad Request: %s", err.Error()),
			http.StatusBadRequest,
		)
		return
	}
	name := mux.Vars(r)["name"]
	cluster, err := plugin.option.Storage.GetCluster(r.Context(), name)
	if err != nil {
		blog.Errorf("ClusterPlugin Serve %s get project info failed, %s", r.URL.Path, err.Error())
		http.Error(w, "gitops cluster storage failure", http.StatusInternalServerError)
		return
	}
	if cluster == nil {
		blog.Errorf("ClusterPlugin Serve %s get no cluster %s", r.URL.Path, name)
		http.Error(w, "gitops resource Not Found", http.StatusNotFound)
		return
	}
	projectID := common.GetBCSProjectID(cluster.Annotations)
	if projectID == "" {
		blog.Errorf("ClusterPlugin Serve %s get no bcs project control information", r.URL.Path)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	permit, _, err := plugin.Permission.CanViewCluster(user.GetUser(), projectID, cluster.Name)
	if err != nil {
		blog.Errorf("Cluster permission validate user %s for %s failed, %s", user.GetUser(), r.URL.Path, err.Error())
		http.Error(w, "Unauthorized: Auth center failure", http.StatusInternalServerError)
		return
	}
	blog.Infof("user %s request %s, permission %t", user.GetUser(), r.URL.Path, permit)
	if !permit {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// authorized then go through reverse proxy
	plugin.Session.ServeHTTP(w, r)
}
