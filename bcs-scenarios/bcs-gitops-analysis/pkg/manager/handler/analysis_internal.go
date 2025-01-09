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

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/secretstore"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/ccv3"
)

// AnalysisHandler defines the analysis handler for internal
type AnalysisHandler struct {
	op          *options.AnalysisOptions
	storage     store.Store
	secretStore secretstore.SecretInterface
	db          dao.Interface
	bkccClient  ccv3.Interface

	cacheLock         sync.Mutex
	cache             []AnalysisProject
	resourceInfo      []AnalysisProjectResourceInfo
	bizDeptInfoCache  *sync.Map
	userDeptInfoCache *sync.Map
}

// NewAnalysisHandler create AnalysisInterface instance
func NewAnalysisHandler() *AnalysisHandler {
	return &AnalysisHandler{
		op:                options.GlobalOptions(),
		bizDeptInfoCache:  &sync.Map{},
		userDeptInfoCache: &sync.Map{},
	}
}

// Init will init argo store and secret store. And then start goroutine to collect
// analysis and resource info to cache
func (h *AnalysisHandler) Init() error {
	h.storage = store.NewStore(&store.Options{
		Service:        h.op.ArgoConfig.ArgoService,
		User:           h.op.ArgoConfig.ArgoUser,
		Pass:           h.op.ArgoConfig.ArgoPass,
		AdminNamespace: h.op.ArgoConfig.AdminNamespace,
		Cache:          true,
		CacheHistory:   false,
	})
	if err := h.storage.Init(); err != nil {
		return errors.Wrapf(err, "init argocd stroe failed")
	}
	h.secretStore = secretstore.NewSecretStore(&h.op.SecretConfig)
	h.db = dao.GlobalDB()
	h.bkccClient = ccv3.NewHandler()

	go func() {
		analysisTicker := time.NewTicker(1 * time.Minute)
		defer analysisTicker.Stop()
		resourceTicker := time.NewTicker(15 * time.Minute)
		defer resourceTicker.Stop()

		h.analysisProjects()
		h.analysisResourceInfos()
		blog.Infof("analysis cache started")
		for {
			select {
			case <-analysisTicker.C:
				h.analysisProjects()
			case <-resourceTicker.C:
				h.analysisResourceInfos()
			}
		}
	}()
	return nil
}

// GetAnalysisProjects return anlysis projects data
func (h *AnalysisHandler) GetAnalysisProjects() []AnalysisProject {
	h.cacheLock.Lock()
	defer h.cacheLock.Unlock()

	result := make([]AnalysisProject, 0, len(h.cache))
	for i := range h.cache {
		item := h.cache[i]
		result = append(result, *(&item).DeepCopy())
	}
	return result
}

// GetResourceInfo return resource info data
func (h *AnalysisHandler) GetResourceInfo() []AnalysisProjectResourceInfo {
	return h.resourceInfo
}

func fillProjectDeptInfo(anaProj *AnalysisProject, bizDeptInfo *ccv3.BusinessDeptInfo) {
	anaProj.BizID = bizDeptInfo.BKBizID
	anaProj.BizName = bizDeptInfo.BKBizName
	anaProj.GroupLevel0 = bizDeptInfo.Level0
	anaProj.GroupLevel1 = bizDeptInfo.Level1
	anaProj.GroupLevel2 = bizDeptInfo.Level2
	anaProj.GroupLevel3 = bizDeptInfo.Level3
	anaProj.GroupLevel4 = bizDeptInfo.Level4
	anaProj.GroupLevel5 = bizDeptInfo.Level5
}

// analysisProjects collect every projects analysis data
func (h *AnalysisHandler) analysisProjects() {
	ctx := context.Background()
	projList, err := h.storage.ListProjectsWithoutAuth(ctx)
	if err != nil {
		blog.Errorf("list projects without auth failed: %s", err.Error())
		return
	}
	argoProjects := projList.Items
	result := make([]AnalysisProject, 0, len(argoProjects))
	var parallel = 10
	var wg sync.WaitGroup
	wg.Add(parallel)
	for i := 0; i < parallel; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := idx; j < len(argoProjects); j += parallel {
				argoProj := argoProjects[j]
				anaProj, err := h.collectProjectAnalysis(ctx, &argoProj)
				if err != nil {
					blog.Errorf("collect project '%s' analysis data failed: %s", argoProj.Name, err.Error())
					continue
				}
				result = append(result, *anaProj)
			}
		}(i)
	}
	wg.Wait()

	blog.Infof("new collect projects(%d) analysis data success", len(argoProjects))
	h.cacheLock.Lock()
	h.cache = result
	h.cacheLock.Unlock()
}

// CollectResourceInfo collect resource info with multi-goroutines.
func (h *AnalysisHandler) analysisResourceInfos() {
	apps := h.storage.AllApplications()
	var parallel = 5
	var wg sync.WaitGroup
	wg.Add(parallel)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	blog.Infof("collect applications(%d) resource info started", len(apps))
	for i := 0; i < parallel; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := idx; j < len(apps); j += parallel {
				app := apps[j]
				err := h.collectAppResourceInfo(ctx, app)
				if err != nil {
					blog.Errorf("analysis collect app '%s' resource info failed: %s", app.Name, err.Error())
					continue
				}
			}
		}(i)
	}
	wg.Wait()
	h.cacheResourceInfoData()
	blog.Infof("collect applications(%d) resource info success", len(apps))
}

// collectProjectAnalysis 计算对应项目的运营数据, 获取项目下的所有数据
// nolint
func (h *AnalysisHandler) collectProjectAnalysis(ctx context.Context, argoProj *v1alpha1.AppProject) (
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
	// should change to real bizID if is-external
	if h.op.IsExternal {
		bizInfos, err := h.bkccClient.SearchBusiness([]int64{bizID})
		if err != nil {
			return nil, errors.Wrapf(err, "get business failed")
		}
		if len(bizInfos) != 1 {
			return nil, fmt.Errorf("get business return not 1 but %d", len(bizInfos))
		}
		bizID = bizInfos[0].BkBizId
	}
	result := &AnalysisProject{
		BizID:       bizID,
		ProjectID:   argoProj.Annotations[common.ProjectIDKey],
		ProjectName: argoProj.Annotations[common.ProjectAliaName],
		ProjectCode: argoProj.Name,
	}
	if err := h.listApplicationSets(ctx, result, argoProj); err != nil {
		return nil, err
	}
	if err := h.listApplications(ctx, result, argoProj); err != nil {
		return nil, err
	}
	if err := h.listClusters(ctx, result, argoProj); err != nil {
		return nil, err
	}
	if err := h.listSecrets(ctx, result, argoProj); err != nil {
		return nil, err
	}
	if err := h.listRepos(ctx, result, argoProj); err != nil {
		return nil, err
	}
	if err := h.listSyncs(ctx, result, argoProj); err != nil {
		return nil, errors.Wrapf(err, "list sync infos failed")
	}
	if err := h.listActivityUsers(ctx, result, argoProj); err != nil {
		return nil, errors.Wrapf(err, "list activity users failed")
	}
	v, ok := h.bizDeptInfoCache.Load(bizID)
	if ok {
		bizDeptInfo := v.(*ccv3.BusinessDeptInfo)
		fillProjectDeptInfo(result, bizDeptInfo)
		return result, nil
	}
	businessInfo, err := h.bkccClient.GetBizDeptInfo([]int64{bizID})
	if err != nil {
		blog.Warnf("get project '%s' business ‘%d’ dept-info failed: %s", argoProj.Name, bizID, err.Error())
	} else {
		if bizDeptInfo, ok := businessInfo[bizID]; ok {
			h.bizDeptInfoCache.Store(bizID, bizDeptInfo)
			fillProjectDeptInfo(result, bizDeptInfo)
		}
	}
	return result, nil
}

// listApplicationSets list all applicationsets
func (h *AnalysisHandler) listApplicationSets(ctx context.Context, result *AnalysisProject,
	argoProj *v1alpha1.AppProject) error {
	applicationSets, err := h.storage.ListApplicationSets(ctx, &appsetpkg.ApplicationSetListQuery{
		Projects: []string{argoProj.Name},
	})
	if err != nil {
		return errors.Wrapf(err, "list applicationsets failed")
	}
	result.ApplicationSets = make([]*AnalysisApplicationSet, 0, len(applicationSets.Items))
	for i := range applicationSets.Items {
		result.ApplicationSets = append(result.ApplicationSets, &AnalysisApplicationSet{
			Name: applicationSets.Items[i].Name,
		})
	}
	return nil
}

// listApplications list all applications
func (h *AnalysisHandler) listApplications(ctx context.Context, result *AnalysisProject,
	argoProj *v1alpha1.AppProject) error {
	apps, err := h.storage.ListApplications(ctx, &appclient.ApplicationQuery{
		Projects: []string{argoProj.Name},
	})
	if err != nil {
		return errors.Wrapf(err, "list applications failed")
	}
	result.Applications = make([]*AnalysisApplication, 0, len(apps.Items))
	for i := range apps.Items {
		result.Applications = append(result.Applications, &AnalysisApplication{
			Name:    apps.Items[i].Name,
			Status:  string(apps.Items[i].Status.Health.Status),
			Cluster: apps.Items[i].Spec.Destination.Server,
		})
	}
	return nil
}

// listClusters list all clusters
func (h *AnalysisHandler) listClusters(ctx context.Context, result *AnalysisProject,
	argoProj *v1alpha1.AppProject) error {
	clusters, err := h.storage.ListClustersByProject(ctx, result.ProjectID)
	if err != nil {
		return errors.Wrapf(err, "list clusters by project '%s' failed", argoProj.Name)
	}
	result.Clusters = make([]*AnalysisCluster, 0, len(clusters.Items))
	for i := range clusters.Items {
		result.Clusters = append(result.Clusters, &AnalysisCluster{
			ClusterName:   clusters.Items[i].Annotations[common.ClusterAliaName],
			ClusterServer: clusters.Items[i].Server,
			ClusterID:     clusters.Items[i].Name,
		})
	}
	return nil
}

// listSecrets list all secrets
func (h *AnalysisHandler) listSecrets(ctx context.Context, result *AnalysisProject,
	argoProj *v1alpha1.AppProject) error {
	projectSecrets, err := h.secretStore.ListProjectSecrets(ctx, argoProj.Name)
	if err != nil {
		return errors.Wrapf(err, "list secrets failed")
	}
	result.Secrets = make([]*AnalysisSecret, 0, len(projectSecrets))
	for i := range projectSecrets {
		result.Secrets = append(result.Secrets, &AnalysisSecret{
			Name: projectSecrets[i],
		})
	}
	return nil
}

// listRepos list all repos
func (h *AnalysisHandler) listRepos(ctx context.Context, result *AnalysisProject,
	argoProj *v1alpha1.AppProject) error {
	repos, err := h.storage.ListRepository(ctx, []string{argoProj.Name})
	if err != nil {
		return errors.Wrapf(err, "list repository failed")
	}
	result.Repos = make([]*AnalysisRepo, 0, len(repos.Items))
	for i := range repos.Items {
		result.Repos = append(result.Repos, &AnalysisRepo{
			RepoName: repos.Items[i].Name,
			RepoUrl:  repos.Items[i].Repo,
		})
	}
	return nil
}

// listSyncs return syncs total for project
func (h *AnalysisHandler) listSyncs(ctx context.Context, result *AnalysisProject,
	argoProj *v1alpha1.AppProject) error {
	syncs, err := h.db.ListSyncInfosForProject(argoProj.Name)
	if err != nil {
		return errors.Wrapf(err, "list sync infos failed")
	}
	result.Syncs = make([]*AnalysisSync, 0, len(syncs))
	for i := range syncs {
		result.Syncs = append(result.Syncs, &AnalysisSync{
			Application: syncs[i].Application,
			Cluster:     syncs[i].Cluster,
			SyncTotal:   syncs[i].SyncTotal,
			UpdateTime:  syncs[i].UpdateTime.Format("2006-01-02 15:04:05"),
		})
	}
	return nil
}

// fillUserDeptInfo fill user's dept info
func fillUserDeptInfo(user *AnalysisActivityUser, userDeptInfo *ccv3.UserDeptInfo) {
	user.ChineseName = userDeptInfo.ChineseName
	user.GroupLevel0 = userDeptInfo.Level0
	user.GroupLevel1 = userDeptInfo.Level1
	user.GroupLevel2 = userDeptInfo.Level2
	user.GroupLevel3 = userDeptInfo.Level3
	user.GroupLevel4 = userDeptInfo.Level4
	user.GroupLevel5 = userDeptInfo.Level5
}

// listActivityUsers return activity users for project
func (h *AnalysisHandler) listActivityUsers(ctx context.Context, result *AnalysisProject,
	argoProj *v1alpha1.AppProject) error {
	users, err := h.db.ListActivityUser(argoProj.Name)
	if err != nil {
		return errors.Wrapf(err, "list sync infos failed")
	}

	result.ActivityUsers = make([]*AnalysisActivityUser, 0, len(users))
	for i := range users {
		username := users[i].UserName
		if common.IsAdminUser(username) {
			continue
		}
		activityUser := &AnalysisActivityUser{
			UserName:         users[i].UserName,
			Project:          users[i].Project,
			OperateNum:       users[i].OperateNum,
			LastActivityTime: users[i].LastActivityTime.Format("2006-01-02 15:04:05"),
		}
		result.ActivityUsers = append(result.ActivityUsers, activityUser)
		if strings.Contains(activityUser.UserName, "@") || strings.Contains(activityUser.UserName, "_") {
			continue
		}
		v, ok := h.userDeptInfoCache.Load(activityUser.UserName)
		if ok {
			fillUserDeptInfo(activityUser, v.(*ccv3.UserDeptInfo))
			continue
		}
		var userDeptInfo *ccv3.UserDeptInfo
		if userDeptInfo, err = h.bkccClient.GetUserDeptInfo(activityUser.UserName); err != nil {
			blog.Warnf("query user '%s' dept info failed: %s", activityUser.UserName, err.Error())
		} else {
			h.userDeptInfoCache.Store(activityUser.UserName, userDeptInfo)
			fillUserDeptInfo(activityUser, userDeptInfo)
		}
	}
	return nil
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

// collectAppResourceInfo will get the resource-tree with application. Parse the resource-tree to
// collect the information which we need
func (h *AnalysisHandler) collectAppResourceInfo(ctx context.Context, app *v1alpha1.Application) error {
	resourceTree, err := h.storage.GetApplicationResourceTree(ctx, app.Name)
	if err != nil {
		return errors.Wrapf(err, "get  resource tree failed")
	}
	result := make(map[string]int)
	all := 0
	for i := range resourceTree.Nodes {
		node := resourceTree.Nodes[i]
		all++
		kind := strings.ToLower(node.Kind)
		if utils.StringsContainsOr(kind, "gamedeployment", "gamestatefulset") {
			result[ResourceInfoGameWorkload]++
			continue
		}
		if utils.StringsContainsOr(kind, "statefulset", "deployment", "job", "cronjob", "daemonset") {
			result[ResourceInfoWorkload]++
			continue
		}
		if kind == "pod" {
			result[ResourceInfoPod]++
			continue
		}
	}
	result[ResourceInfoAll] = all
	riJSON, _ := json.Marshal(result)
	if err = h.db.SaveOrUpdateResourceInfo(&dao.ResourceInfo{
		Project:     app.Spec.Project,
		Application: app.Name,
		Resources:   utils.SliceByteToString(riJSON),
	}); err != nil {
		return errors.Wrapf(err, "save resource info failed")
	}
	return nil
}

// cacheResourceInfoData cache the data for resource-info
func (h *AnalysisHandler) cacheResourceInfoData() {
	ris, err := h.db.ListResourceInfosByProject(nil)
	if err != nil {
		blog.Errorf("analysis list resource infos failed: %s", err.Error())
		return
	}
	apps := h.storage.AllApplications()
	appMap := make(map[string]struct{})
	for _, app := range apps {
		appMap[app.Name] = struct{}{}
	}
	result := make(map[string]*AnalysisProjectResourceInfo)
	for i := range ris {
		ri := ris[i]
		_, ok := appMap[ri.Application]
		if !ok {
			continue
		}
		m := make(map[string]int64)
		if err = json.Unmarshal(utils.StringToSliceByte(ri.Resources), &m); err != nil {
			continue
		}
		pri, ok := result[ri.Project]
		if !ok {
			result[ri.Project] = &AnalysisProjectResourceInfo{
				Name:         ri.Project,
				ResourceAll:  m[ResourceInfoAll],
				GameWorkload: m[ResourceInfoGameWorkload],
				Workload:     m[ResourceInfoWorkload],
				Pod:          m[ResourceInfoPod],
			}
		} else {
			pri.ResourceAll += m[ResourceInfoAll]
			pri.GameWorkload += m[ResourceInfoGameWorkload]
			pri.Workload += m[ResourceInfoWorkload]
			pri.Pod += m[ResourceInfoPod]
		}
	}
	data := make([]AnalysisProjectResourceInfo, 0, len(result))
	for _, pri := range result {
		data = append(data, *pri)
	}

	h.cacheLock.Lock()
	h.resourceInfo = data
	h.cacheLock.Unlock()
}
