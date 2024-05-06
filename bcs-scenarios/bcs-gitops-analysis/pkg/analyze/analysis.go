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

package analyze

import (
	"context"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/internal/dao"
)

func (h *analysisHandler) getCacheAnalysis() []AnalysisProject {
	h.cacheLock.Lock()
	defer h.cacheLock.Unlock()
	result := append(make([]AnalysisProject, 0, len(h.cache)), h.cache...)
	return result
}

// CollectAnalysisData collect all analysis data and save to cache
func (h *analysisHandler) CollectAnalysisData() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	projList, err := h.storage.ListProjectsWithoutAuth(ctx)
	if err != nil {
		panic(err)
	}
	argoProjs := projList.Items
	failedProjs := 0
	result := make([]AnalysisProject, 0, len(argoProjs))
	for i := range argoProjs {
		anaProj, err := h.collectProjectAnalysis(ctx, &argoProjs[i])
		if err != nil {
			failedProjs++
			blog.Errorf("collect project '%s' analysis data failed: %s", argoProjs[i].Name, err.Error())
			continue
		}
		result = append(result, *anaProj)
	}
	blog.Infof("collect projects(%d) analysis data success", len(argoProjs))
	if failedProjs > 0 {
		blog.Errorf("collect project analysis data will be give up because of failed projects")
	} else {
		h.cacheLock.Lock()
		h.cache = result
		h.cacheLock.Unlock()
	}
}

const (
	// ResourceInfoGameWorkload defines the gameworkload field
	ResourceInfoGameWorkload = "gameworkload"
	// ResourceInfoWorkload defines the workload field
	ResourceInfoWorkload = "workload"
	// ResourceInfoPod defines the pod field
	ResourceInfoPod = "pod"
	// ResourceInfoAll defines the all field
	ResourceInfoAll = "all"
)

// collectProjectAnalysis 计算对应项目的运营数据
// nolint
func (h *analysisHandler) collectProjectAnalysis(ctx context.Context, argoProj *v1alpha1.AppProject) (
	*AnalysisProject, error) {
	var bizID int64
	bizIDStr := argoProj.Annotations[common.ProjectBusinessIDKey]
	if bizIDStr != "" {
		var err error
		bizID, err = strconv.ParseInt(bizIDStr, 0, 64)
		if err != nil {
			blog.Warnf("project '%s' with businessID '%s' parse failed", argoProj.Name, bizIDStr)
		}
	}
	result := &AnalysisProject{
		BizID:       bizID,
		BizName:     argoProj.Annotations[common.ProjectBusinessName],
		ProjectID:   argoProj.Annotations[common.ProjectIDKey],
		ProjectName: argoProj.Annotations[common.ProjectAliaName],
		ProjectCode: argoProj.Name,
	}
	appsets, err := h.storage.ListApplicationSets(ctx, &appsetpkg.ApplicationSetListQuery{
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

	apps, err := h.storage.ListApplications(ctx, &appclient.ApplicationQuery{
		Projects: []string{argoProj.Name},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "list applications failed")
	}
	result.Applications = make([]*AnalysisApplication, 0, len(apps.Items))
	for i := range apps.Items {
		result.Applications = append(result.Applications, &AnalysisApplication{
			Name:    apps.Items[i].Name,
			Status:  string(apps.Items[i].Status.Health.Status),
			Cluster: apps.Items[i].Spec.Destination.Server,
		})
	}

	clusters, err := h.storage.ListClustersByProject(ctx, result.ProjectID)
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

	projectSecrets, err := h.secretStore.ListProjectSecrets(ctx, argoProj.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "list secrets failed")
	}
	for i := range projectSecrets {
		result.Secrets = append(result.Secrets, &AnalysisSecret{
			Name: projectSecrets[i],
		})
	}

	repos, err := h.storage.ListRepository(ctx, []string{argoProj.Name})
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

	syncs, err := h.listSyncs(argoProj.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "list sync infos failed")
	}
	result.Syncs = make([]*AnalysisSync, 0, len(syncs))
	for i := range syncs {
		result.Syncs = append(result.Syncs, &AnalysisSync{
			Application:   syncs[i].Application,
			Cluster:       syncs[i].Cluster,
			SyncTotal:     syncs[i].SyncTotal,
			UpdateTime:    syncs[i].UpdateTime,
			UpdateTimeStr: syncs[i].UpdateTime.Format("2006-01-02 15:04:05"),
		})
	}

	users, err := h.listActivityUsers(argoProj.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "list activity users failed")
	}
	result.ActivityUsers = make([]*AnalysisActivityUser, 0, len(users))
	for i := range users {
		username := users[i].UserName
		if common.IsAdminUser(username) {
			continue
		}
		result.ActivityUsers = append(result.ActivityUsers, &AnalysisActivityUser{
			UserName:            users[i].UserName,
			OperateNum:          users[i].OperateNum,
			LastActivityTime:    users[i].LastActivityTime,
			LastActivityTimeStr: users[i].LastActivityTime.Format("2006-01-02 15:04:05"),
		})
	}
	return result, nil
}

// listSyncs return syncs total for project
func (h *analysisHandler) listSyncs(proj string) ([]dao.SyncInfo, error) {
	syncs, err := h.db.ListSyncInfosForProject(proj)
	if err != nil {
		return nil, errors.Wrapf(err, "list sync infos failed")
	}
	return syncs, nil
}

// listActivityUsers return activity users for project
func (h *analysisHandler) listActivityUsers(proj string) ([]dao.ActivityUser, error) {
	users, err := h.db.ListActivityUser(proj)
	if err != nil {
		return nil, errors.Wrapf(err, "list sync infos failed")
	}
	return users, nil
}
