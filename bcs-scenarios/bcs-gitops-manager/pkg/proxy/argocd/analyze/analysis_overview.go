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
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/ccv3"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/secretstore"
)

// AnalysisOverview 采集
type AnalysisOverview interface {
	AnalysisProjectAll() []*AnalysisProject
	AnalysisProject(ctx context.Context, argoProjs []v1alpha1.AppProject) ([]AnalysisProject, error)

	TopProjects() []*AnalysisProjectOverview
	OverviewAllInternal() *AnalysisOverviewAll
	OverviewAllExternal() *AnalysisOverviewAll

	ResourceInfos() []ProjectResourceInfo

	BKMonitorCommonGet(message *BKMonitorGetMessage) ([]*AnalysisBKMonitorSeries, error)
	BKMTopActivityProjectsInternal() ([]*AnalysisBKMonitorActivityProject, error)
	BKMTopActivityProjectsExternal() ([]*AnalysisBKMonitorActivityProject, error)
}

type analysisOverviewClient struct {
	op          *options.Options
	storage     store.Store
	db          dao.Interface
	secretStore secretstore.SecretInterface
	bkmClient   *bkMonitorClient

	bkccClient    ccv3.Interface
	businessCache *sync.Map

	cacheLock        sync.Mutex
	internalData     []AnalysisProject
	externalData     []AnalysisProject
	resourceInfoData []ProjectResourceInfo
}

var (
	analysisOverviewOnce         sync.Once
	globalAnalysisOverviewClient *analysisOverviewClient
)

// NewAnalysisOverview create the analysis overview client
func NewAnalysisOverview() AnalysisOverview {
	if globalAnalysisOverviewClient != nil {
		return globalAnalysisOverviewClient
	}
	analysisOverviewOnce.Do(func() {
		globalAnalysisOverviewClient = &analysisOverviewClient{
			op:          options.GlobalOptions(),
			db:          dao.GlobalDB(),
			storage:     store.GlobalStore(),
			secretStore: secretstore.NewSecretStore(),
			bkmClient: &bkMonitorClient{
				op: options.GlobalOptions(),
			},
			bkccClient:    ccv3.NewHandler(),
			businessCache: &sync.Map{},
		}
		go func() {
			overviewTicker := time.NewTicker(1 * time.Minute)
			defer overviewTicker.Stop()
			resourceTicker := time.NewTicker(15 * time.Minute)
			defer resourceTicker.Stop()

			globalAnalysisOverviewClient.collectAllOverviewData()
			globalAnalysisOverviewClient.refreshResourceInfoData()
			for {
				select {
				case <-overviewTicker.C:
					globalAnalysisOverviewClient.collectAllOverviewData()
				case <-resourceTicker.C:
					go globalAnalysisOverviewClient.collectAllResourceInfo()
				}
			}
		}()
	})
	return globalAnalysisOverviewClient
}

// AnalysisProject defines the analysis data of project
type AnalysisProject struct {
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
	Name    string `json:"name"`
	Status  string `json:"status"`
	Cluster string `json:"cluster"`
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
	UserName   string `json:"userName"`
	OperateNum int64  `json:"operateNum"`
	// 兼容以前运营数据格式
	LastActivityTimeStr string    `json:"lastActivityTime,omitempty"`
	LastActivityTime    time.Time `json:"lastActivityTimeTime,omitempty"`
}

// AnalysisSync defines the sync info
type AnalysisSync struct {
	Application string `json:"application"`
	Cluster     string `json:"cluster"`
	SyncTotal   int64  `json:"syncTotal"`

	// 兼容以前运营数据格式
	UpdateTimeStr string    `json:"updateTime,omitempty"`
	UpdateTime    time.Time `json:"updateTimeTime,omitempty"`
}

// AnalysisProjectOverview defines the project overview
type AnalysisProjectOverview struct {
	BizID               int    `json:"bizID"`
	BizName             string `json:"bizName"`
	ProjectCode         string `json:"projectCode"`
	ProjectName         string `json:"projectName"`
	ClusterNum          int    `json:"clusterNum"`
	ApplicationNum      int    `json:"applicationNum"`
	ActivityApplication int    `json:"activityApplication"`
	ActivityUserNum     int    `json:"activityUserNum"`
}

// AnalysisOverviewAll defines the overview
type AnalysisOverviewAll struct {
	Type string `json:"type,omitempty"`

	BizNum              int              `json:"bizNum"`
	EffectiveBizNum     int              `json:"effectiveBizNum"`
	ProjectNum          int              `json:"projectNum"`
	EffectiveProjectNum int              `json:"effectiveProjectNum"`
	ClusterNum          int              `json:"clusterNum"`
	EffectiveClusterNum int              `json:"effectiveClusterNum"`
	ApplicationSetNum   int              `json:"applicationSetNum"`
	ApplicationNum      int              `json:"applicationNum"`
	SyncTotal           int64            `json:"syncTotal"`
	ProjectSyncTotal    map[string]int64 `json:"projectSyncTotal"`
	SecretNum           int              `json:"secretNum"`
	RepoNum             int              `json:"repoNum"`
	UserOperateNum      int64            `json:"userOperateNum"`

	Activity7DayUserNum    int `json:"activity7DayUserNum"`
	Activity1DayUserNum    int `json:"activity1DayUserNum"`
	Activity7DayProjectNum int `json:"activity7DayProjectNum"`
	Activity1DayProjectNum int `json:"activity1DayProjectNum"`
}

// ProjectResourceInfo defines the resource-info of project
type ProjectResourceInfo struct {
	Name         string `json:"name"`
	ResourceAll  int64  `json:"resourceAll"`
	GameWorkload int64  `json:"gameWorkload"`
	Workload     int64  `json:"workload"`
	Pod          int64  `json:"pod"`
}

type analysisProjectOverviewSort []AnalysisProjectOverview

// Len defines the len of overview sort
func (s analysisProjectOverviewSort) Len() int {
	return len(s)
}

// Less defines the less of overview
func (s analysisProjectOverviewSort) Less(i, j int) bool {
	return s[i].ApplicationNum > s[j].ApplicationNum
}

// Swap the items
func (s analysisProjectOverviewSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// AnalysisProjectAll returns the project analysis data
func (c *analysisOverviewClient) AnalysisProjectAll() []*AnalysisProject {
	data := c.getInternalData()
	result := make([]*AnalysisProject, 0, len(data))
	for i := range data {
		result = append(result, &data[i])
	}
	return result
}

// AnalysisProject 根据项目列表，返回对应运营数据
func (c *analysisOverviewClient) AnalysisProject(ctx context.Context, argoProjs []v1alpha1.AppProject) (
	[]AnalysisProject, error) {
	result := make([]AnalysisProject, 0, len(argoProjs))
	for i := range argoProjs {
		anaProj, err := c.handleProject(ctx, &argoProjs[i])
		if err != nil {
			return nil, errors.Wrapf(err, "handle project '%s' analysis failed", argoProjs[i].Name)
		}
		result = append(result, *anaProj)
	}
	return result, nil
}

// TopProjects 根据应用数量，获取排名靠前的项目
func (c *analysisOverviewClient) TopProjects() []*AnalysisProjectOverview {
	data := c.getInternalData()
	if len(data) == 0 {
		return nil
	}
	projects := make([]AnalysisProjectOverview, 0, len(data))
	for _, item := range data {
		analysisProject := AnalysisProjectOverview{
			BizID:          int(item.BizID),
			ProjectCode:    item.ProjectCode,
			ProjectName:    item.ProjectName,
			ClusterNum:     len(item.Clusters),
			ApplicationNum: len(item.Applications),
		}
		weekAgo := time.Now().Add(-168 * time.Hour)
		addedUser := make(map[string]struct{})
		for _, atUser := range item.ActivityUsers {
			if atUser.LastActivityTime.After(weekAgo) {
				if _, ok := addedUser[atUser.UserName]; !ok {
					analysisProject.ActivityUserNum++
					addedUser[atUser.UserName] = struct{}{}
				}
			}
		}
		currentAppMap := make(map[string]struct{})
		for _, app := range item.Applications {
			currentAppMap[app.Name] = struct{}{}
		}
		addedApplication := make(map[string]struct{})
		for _, atApp := range item.Syncs {
			if atApp.UpdateTime.After(weekAgo) {
				if _, ok := addedApplication[atApp.Application]; !ok {
					addedApplication[atApp.Application] = struct{}{}
					if _, ok = currentAppMap[atApp.Application]; ok {
						analysisProject.ActivityApplication++
					}
				}
			}
		}
		projects = append(projects, analysisProject)
	}
	sort.Sort(analysisProjectOverviewSort(projects))
	result := make([]*AnalysisProjectOverview, 0, 10)
	ans := 0
	for i := range projects {
		if projects[i].ActivityApplication == 0 {
			continue
		}
		result = append(result, &projects[i])
		ans++
		if ans == 10 {
			break
		}
	}
	for _, item := range result {
		v, ok := c.businessCache.Load(item.BizID)
		if ok {
			item.BizName = v.(string)
			continue
		}
		cs, err := c.bkccClient.SearchBusiness([]int64{int64(item.BizID)})
		if err != nil {
			blog.Errorf("analysis search business '%d' from cc failed: %s", item.BizID, err.Error())
			continue
		}
		if len(cs) != 1 {
			blog.Errorf("analysis search business '%d' from cc return '%d' items", item.BizID, len(cs))
			continue
		}
		c.businessCache.Store(item.BizID, cs[0].BkBizName)
		item.BizName = cs[0].BkBizName
	}
	return result
}

// OverviewAllExternal defines the external overview data
func (c *analysisOverviewClient) OverviewAllExternal() *AnalysisOverviewAll {
	return c.calculateOverviewAll(c.getExternalData())
}

// OverviewAllInternal defines the internal overview data
func (c *analysisOverviewClient) OverviewAllInternal() *AnalysisOverviewAll {
	return c.calculateOverviewAll(c.getInternalData())
}

// ResourceInfos return the managed=resources
func (c *analysisOverviewClient) ResourceInfos() []ProjectResourceInfo {
	return c.getResourceInfoData()
}

func (c *analysisOverviewClient) getInternalData() []AnalysisProject {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	result := append(make([]AnalysisProject, 0, len(c.internalData)), c.internalData...)
	return result
}

func (c *analysisOverviewClient) getExternalData() []AnalysisProject {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	result := append(make([]AnalysisProject, 0, len(c.externalData)), c.externalData...)
	return result
}

func (c *analysisOverviewClient) getResourceInfoData() []ProjectResourceInfo {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()
	result := append(make([]ProjectResourceInfo, 0, len(c.resourceInfoData)), c.resourceInfoData...)
	return result
}

// handleProject 计算对应项目的运营数据
// nolint
func (c *analysisOverviewClient) handleProject(ctx context.Context, argoProj *v1alpha1.AppProject) (
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
	appsets, err := c.storage.ListApplicationSets(ctx, &appsetpkg.ApplicationSetListQuery{
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

	apps, err := c.storage.ListApplications(ctx, &appclient.ApplicationQuery{
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

	clusters, err := c.storage.ListClustersByProject(ctx, result.ProjectID)
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

	projectSecrets, err := c.secretStore.ListProjectSecrets(ctx, argoProj.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "list secrets failed")
	}
	for i := range projectSecrets {
		result.Secrets = append(result.Secrets, &AnalysisSecret{
			Name: projectSecrets[i],
		})
	}

	repos, err := c.storage.ListRepository(ctx, []string{argoProj.Name})
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

	syncs, err := c.ListSyncs(argoProj.Name)
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

	users, err := c.ListActivityUsers(argoProj.Name)
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

// ListSyncs return syncs total for project
func (c *analysisOverviewClient) ListSyncs(proj string) ([]dao.SyncInfo, error) {
	syncs, err := c.db.ListSyncInfosForProject(proj)
	if err != nil {
		return nil, errors.Wrapf(err, "list sync infos failed")
	}
	return syncs, nil
}

// ListActivityUsers return activity users for project
func (c *analysisOverviewClient) ListActivityUsers(proj string) ([]dao.ActivityUser, error) {
	users, err := c.db.ListActivityUser(proj)
	if err != nil {
		return nil, errors.Wrapf(err, "list sync infos failed")
	}
	return users, nil
}

// AnalysisBKMonitorSeries defines the series of bkmonitor search
type AnalysisBKMonitorSeries struct {
	Timestamp int64 `json:"timestamp"`
	Value     int64 `json:"value"`
}

// AnalysisBKMonitorActivityProject defines the project message
type AnalysisBKMonitorActivityProject struct {
	Project string `json:"project"`
	Value   int64  `json:"value"`
}

// BKMonitorCommonGet common query for bkmonitor
func (c *analysisOverviewClient) BKMonitorCommonGet(message *BKMonitorGetMessage) ([]*AnalysisBKMonitorSeries, error) {
	message.Complete()
	resp, err := c.bkmClient.Get(message)
	if err != nil {
		return nil, errors.Wrapf(err, "bkmonitor get failed with promql '%s'", message.PromQL)
	}
	if len(resp.Series) != 1 {
		return nil, errors.Errorf("bkmonitor return series not 1 but '%d'", len(resp.Series))
	}
	result := make([]*AnalysisBKMonitorSeries, 0, len(resp.Series[0].Values))
	for i := range resp.Series[0].Values {
		v := resp.Series[0].Values[i]
		if len(v) != 2 {
			return nil, errors.Errorf("bkmonitor return series values not 2 but '%d': %v", len(v), v)
		}
		result = append(result, &AnalysisBKMonitorSeries{
			Timestamp: time.UnixMilli(v[0]).Add(8 * time.Hour).UnixNano(),
			Value:     v[1],
		})
	}
	return result, nil
}

// BKMonitorTopActivityProjectsInternal get top activity projects from bkmonitor
func (c *analysisOverviewClient) BKMTopActivityProjectsInternal() ([]*AnalysisBKMonitorActivityProject, error) {
	message := &BKMonitorGetMessage{
		// nolint
		PromQL: `topk(10, max by (project) (increase(custom:GitOpsOperationData:project_sync{target="internal"}[1440m])) != 0)`,
		Step:   "3600s",
	}
	return c.calculateBKMonitorActivityProjects(message)
}

// BKMonitorTopActivityProjectsExternal get top activity projects from bkmonitor
func (c *analysisOverviewClient) BKMTopActivityProjectsExternal() ([]*AnalysisBKMonitorActivityProject, error) {
	message := &BKMonitorGetMessage{
		// nolint
		PromQL: `topk(10, max by (project) (increase(custom:GitOpsOperationData:project_sync{target="external"}[1440m])) != 0)`,
		Step:   "3600s",
	}
	return c.calculateBKMonitorActivityProjects(message)
}

func (c *analysisOverviewClient) calculateBKMonitorActivityProjects(message *BKMonitorGetMessage) (
	[]*AnalysisBKMonitorActivityProject, error) {
	message.Complete()
	resp, err := c.bkmClient.Get(message)
	if err != nil {
		return nil, errors.Wrapf(err, "bkmonitor get failed with promql '%s'", message.PromQL)
	}
	if len(resp.Series) == 0 {
		return nil, nil
	}
	var lastTime int64
	result := make([]*AnalysisBKMonitorActivityProject, 0)
	for i := range resp.Series {
		series := resp.Series[i]
		if len(series.GroupValues) != 1 {
			return nil, errors.Errorf("bkmonitor top projects '%d' group values length not 1: %v",
				i, series.GroupValues)
		}
		if len(series.Values) == 0 {
			continue
		}
		if len(series.Values[len(series.Values)-1]) != 2 {
			return nil, errors.Errorf("bkmonitor top projects '%d' values last item not 2: %v",
				i, series.Values[len(series.Values)-1])
		}
		timeStamp := series.Values[len(series.Values)-1][0]
		if i == 0 {
			lastTime = timeStamp
		}
		if timeStamp != lastTime {
			continue
		}
		result = append(result, &AnalysisBKMonitorActivityProject{
			Project: series.GroupValues[0],
			Value:   series.Values[len(series.Values)-1][1],
		})
	}
	return result, nil
}
