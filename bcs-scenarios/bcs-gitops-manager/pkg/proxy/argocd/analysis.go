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

package argocd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/analysis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// AnalysisPlugin for gitops analysis
type AnalysisPlugin struct {
	*mux.Router
	storage        store.Store
	middleware     mw.MiddlewareInterface
	analysisClient analysis.AnalysisInterface
}

// Init init the http route for analysis
func (plugin *AnalysisPlugin) Init() error {
	plugin.Path("").Methods("GET").Handler(plugin.middleware.HttpWrapper(plugin.analysis))
	blog.Infof("argocd analysis init successfully")
	return nil
}

// AnalysisResponse defines the response of analysis data
type AnalysisResponse struct {
	Code      int32              `json:"code"`
	Message   string             `json:"message"`
	RequestID string             `json:"requestID"`
	Data      []*ProjectAnalysis `json:"data"`
}

// ProjectAnalysis defines the analysis data of project
type ProjectAnalysis struct {
	BizID           int64                     `json:"bizID"`
	BizName         string                    `json:"bizName"`
	ProjectID       string                    `json:"projectID"`
	ProjectCode     string                    `json:"projectCode"`
	ProjectName     string                    `json:"projectName"`
	Clusters        []*AnalysisCluster        `json:"clusters"`
	ApplicationSets []*AnalysisApplicationSet `json:"applicationSets"`
	Applications    []*AnalysisApplication    `json:"applications"`
	Secrets         []*AnalysisSecret         `json:"secrets"`
	Repos           []*AnalysisRepo           `json:"repos"`
	ActivityUsers   []*AnalysisActivityUser   `json:"activityUsers"`
	Syncs           []*AnalysisSync           `json:"syncs"`
}

// AnalysisCluster defines the cluster info
type AnalysisCluster struct {
	ClusterName   string `json:"clusterName"`
	ClusterServer string `json:"clusterServer"`
	ClusterID     string `json:"clusterID"`
}

// AnalysisApplicationSet defines the applicationSet info
type AnalysisApplicationSet struct {
	Name string `json:"name"`
}

// AnalysisApplication defines the application info
type AnalysisApplication struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// AnalysisSecret defines the secret info
type AnalysisSecret struct {
	Name string `json:"name"`
}

// AnalysisRepo defines the repo info
type AnalysisRepo struct {
	RepoName string `json:"repoName"`
	RepoUrl  string `json:"repoUrl"`
}

// AnalysisActivityUser defines the user info
type AnalysisActivityUser struct {
	UserName         string `json:"userName"`
	OperateNum       int64  `json:"operateNum"`
	LastActivityTime string `json:"lastActivityTime"`
}

// AnalysisSync defines the sync info
type AnalysisSync struct {
	Application string `json:"application"`
	Cluster     string `json:"cluster"`
	SyncTotal   int64  `json:"syncTotal"`
}

func (plugin *AnalysisPlugin) analysis(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projects, err := plugin.checkQuery(r)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, err)
	}
	result := make([]*ProjectAnalysis, 0, len(projects))
	for _, argoProj := range projects {
		var projAna *ProjectAnalysis
		projAna, err = plugin.handleProject(r.Context(), &argoProj)
		if err != nil {
			return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
		}
		result = append(result, projAna)
	}
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      result,
	})
}

func (plugin *AnalysisPlugin) handleProject(ctx context.Context,
	argoProj *v1alpha1.AppProject) (*ProjectAnalysis, error) {
	var bizID int64
	bizIDStr := argoProj.Annotations[common.ProjectBusinessIDKey]
	if bizIDStr != "" {
		var err error
		bizID, err = strconv.ParseInt(bizIDStr, 0, 64)
		if err != nil {
			blog.Warnf("project '%s' with businessID '%s' parse failed", argoProj.Name, bizIDStr)
		}
	}
	result := &ProjectAnalysis{
		BizID:       bizID,
		BizName:     argoProj.Annotations[common.ProjectBusinessName],
		ProjectID:   argoProj.Annotations[common.ProjectIDKey],
		ProjectName: argoProj.Annotations[common.ProjectAliaName],
		ProjectCode: argoProj.Name,
	}
	appsets, err := plugin.storage.ListApplicationSets(ctx, &appsetpkg.ApplicationSetListQuery{
		Projects: []string{argoProj.Name},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "list applicationsets failed")
	}
	result.ApplicationSets = make([]*AnalysisApplicationSet, 0, len(appsets.Items))
	for i := range appsets.Items {
		result.ApplicationSets = append(result.ApplicationSets, &AnalysisApplicationSet{
			Name: appsets.Items[i].Name,
		})
	}

	apps, err := plugin.storage.ListApplications(ctx, &appclient.ApplicationQuery{
		Projects: []string{argoProj.Name},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "list applications failed")
	}
	result.Applications = make([]*AnalysisApplication, 0, len(apps.Items))
	for i := range apps.Items {
		result.Applications = append(result.Applications, &AnalysisApplication{
			Name: apps.Items[i].Name,
		})
	}

	clusters, err := plugin.storage.ListClustersByProject(ctx, argoProj.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "list clusters by project '%s' failed", argoProj.Name)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "list clusters failed")
	}
	result.Clusters = make([]*AnalysisCluster, 0, len(apps.Items))
	for i := range clusters.Items {
		result.Clusters = append(result.Clusters, &AnalysisCluster{
			ClusterName:   clusters.Items[i].Annotations[common.ClusterAliaName],
			ClusterServer: clusters.Items[i].Server,
			ClusterID:     clusters.Items[i].Name,
		})
	}

	repos, err := plugin.storage.ListRepository(ctx, []string{argoProj.Name})
	if err != nil {
		return nil, errors.Wrapf(err, "list repository failed")
	}
	result.Repos = make([]*AnalysisRepo, 0, len(repos.Items))
	for i := range repos.Items {
		result.Repos = append(result.Repos, &AnalysisRepo{
			RepoName: repos.Items[i].Name,
			RepoUrl:  repos.Items[i].Repo,
		})
	}

	syncs, err := plugin.analysisClient.ListSyncs(argoProj.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "list sync infos failed")
	}
	result.Syncs = make([]*AnalysisSync, 0, len(syncs))
	for i := range syncs {
		result.Syncs = append(result.Syncs, &AnalysisSync{
			Application: syncs[i].Application,
			Cluster:     syncs[i].Cluster,
			SyncTotal:   syncs[i].SyncTotal,
		})
	}

	users, err := plugin.analysisClient.ListActivityUsers(argoProj.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "list activity users failed")
	}
	result.ActivityUsers = make([]*AnalysisActivityUser, 0, len(users))
	for i := range users {
		result.ActivityUsers = append(result.ActivityUsers, &AnalysisActivityUser{
			UserName:         users[i].UserName,
			OperateNum:       users[i].OperateNum,
			LastActivityTime: users[i].LastActivityTime.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

// checkQuery it will check the permission with projects
func (plugin *AnalysisPlugin) checkQuery(r *http.Request) ([]v1alpha1.AppProject, error) {
	projects := r.URL.Query()["projects"]
	if len(projects) != 0 {
		result := make([]v1alpha1.AppProject, 0, len(projects))
		for i := range projects {
			projectName := projects[i]
			argoProj, statusCode, err := plugin.middleware.CheckProjectPermission(r.Context(),
				projectName, iam.ProjectView)
			if statusCode != http.StatusOK {
				return nil, errors.Wrapf(err, "check project '%s' permission failed", projectName)
			}
			result = append(result, *argoProj)
		}
		return result, nil
	}

	user := mw.User(r.Context())
	if user.ClientID != proxy.AdminClientUser && user.ClientID != proxy.AdminGitOpsUser {
		return nil, fmt.Errorf("query param 'projects' cannot be empty")
	}
	projList, _, err := plugin.middleware.ListProjectsWithoutAuth(r.Context())
	if err != nil {
		return nil, err
	}
	return projList.Items, nil
}
