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

// Package analyze xx
package analyze

import (
	"context"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/secretstore"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/ccv3"
)

// AnalysisInterface defines the analysis interface
type AnalysisInterface interface {
	Init() error

	QueryArgoProjects(projs []string) ([]v1alpha1.AppProject, error)

	AnalysisProjectsAll() []*AnalysisProject
	AnalysisProject(ctx context.Context, argoProjs []v1alpha1.AppProject) ([]AnalysisProject, error)
	ResourceInfosAll() []ProjectResourceInfo

	TopProjects() []*AnalysisProjectOverview
	AnalysisOverview() (*AnalysisOverviewAll, error)

	Applications() ([]*ApplicationInfo, error)

	GetBusinessName(bizID int) string
}

// analysisHandler defines the handler for analysis
type analysisHandler struct {
	op          *options.AnalysisOptions
	storage     store.Store
	secretStore secretstore.SecretInterface
	db          dao.Interface

	bkccClient    ccv3.Interface
	businessCache *sync.Map

	cacheLock    sync.Mutex
	cache        []AnalysisProject
	resourceInfo []ProjectResourceInfo
}

var (
	once                  sync.Once
	globalAnalysisHandler *analysisHandler
)

// GlobalAnalysisHandler return the global AnalysisInterface instance
func GlobalAnalysisHandler() AnalysisInterface {
	return globalAnalysisHandler
}

// NewAnalysisHandler create AnalysisInterface instance
func NewAnalysisHandler() AnalysisInterface {
	if globalAnalysisHandler != nil {
		return globalAnalysisHandler
	}
	globalAnalysisHandler = &analysisHandler{
		op:            options.GlobalOptions(),
		businessCache: &sync.Map{},
	}
	return globalAnalysisHandler
}

// Init will init argo store and secret store. And then start goroutine to collect
// analysis and resource info to cache
func (h *analysisHandler) Init() error {
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

	once.Do(func() {
		go func() {
			analysisTicker := time.NewTicker(1 * time.Minute)
			defer analysisTicker.Stop()
			resourceTicker := time.NewTicker(15 * time.Minute)
			defer resourceTicker.Stop()

			globalAnalysisHandler.CollectAnalysisData()
			globalAnalysisHandler.CollectResourceInfo()
			blog.Infof("analysis cache started")
			for {
				select {
				case <-analysisTicker.C:
					globalAnalysisHandler.CollectAnalysisData()
				case <-resourceTicker.C:
					go globalAnalysisHandler.CollectResourceInfo()
				}
			}
		}()
	})
	return nil
}

// QueryArgoProjects query argo proects by name
func (h *analysisHandler) QueryArgoProjects(projs []string) ([]v1alpha1.AppProject, error) {
	result := make([]v1alpha1.AppProject, 0, len(projs))
	for i := range projs {
		proj, err := h.storage.GetProject(context.Background(), projs[i])
		if err != nil {
			return nil, errors.Wrapf(err, "get argo project '%s' failed", projs[i])
		}
		result = append(result, *proj)
	}
	return result, nil
}

// ProjectResourceInfo defines the resource-info of project
type ProjectResourceInfo struct {
	Name         string `json:"name"`
	ResourceAll  int64  `json:"resourceAll"`
	GameWorkload int64  `json:"gameWorkload"`
	Workload     int64  `json:"workload"`
	Pod          int64  `json:"pod"`
}

// ApplicationInfo defines application info
type ApplicationInfo struct {
	Name         string    `json:"name"`
	Cluster      string    `json:"cluster"`
	Repo         string    `json:"repo"`
	Sync         int64     `json:"sync"`
	LastSyncTime time.Time `json:"lastSyncTime"`
	ResourceInfo string    `json:"resourceInfo"`
}

// AnalysisProject defines the analysis data of project
type AnalysisProject struct {
	BizID       int64  `json:"bizID"`
	BizName     string `json:"bizName"`
	ProjectID   string `json:"projectID"`
	ProjectCode string `json:"projectCode"`
	ProjectName string `json:"projectName"`

	Clusters   []*AnalysisCluster `json:"clusters"`
	ClusterNum int                `json:"clusterNum"`

	ApplicationSets []*AnalysisApplicationSet `json:"applicationSets"`
	Applications    []*AnalysisApplication    `json:"applications"`
	ApplicationNum  int                       `json:"applicationNum,"`

	Secrets   []*AnalysisSecret `json:"secrets"`
	SecretNum int               `json:"secretNum,"`

	Repos   []*AnalysisRepo `json:"repos"`
	RepoNum int             `json:"repoNum,"`

	ActivityUsers   []*AnalysisActivityUser `json:"activityUsers"`
	ActivityUserNum int                     `json:"activityUserNum,"`

	Syncs     []*AnalysisSync `json:"syncs"`
	SyncTotal int64           `json:"syncTotal,"`
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
