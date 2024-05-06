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
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/internal/dao"
)

// AnalysisProjectsAll return all projects analysis overview
func (h *analysisHandler) AnalysisProjectsAll() []*AnalysisProject {
	data := h.getCacheAnalysis()
	result := make([]*AnalysisProject, 0, len(data))
	for i := range data {
		result = append(result, &data[i])
	}
	return result
}

// AnalysisProject return analysis overview by projects
func (h *analysisHandler) AnalysisProject(ctx context.Context, argoProjs []v1alpha1.AppProject) (
	[]AnalysisProject, error) {
	result := make([]AnalysisProject, 0, len(argoProjs))
	for i := range argoProjs {
		anaProj, err := h.collectProjectAnalysis(ctx, &argoProjs[i])
		if err != nil {
			return nil, errors.Wrapf(err, "handle project '%s' analysis failed", argoProjs[i].Name)
		}
		result = append(result, *anaProj)
	}
	return result, nil
}

// ResourceInfosAll return the resource infos all
func (h *analysisHandler) ResourceInfosAll() []ProjectResourceInfo {
	return h.getCacheResourceInfo()
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

// GetBusinessName return the business name by biz id
func (h *analysisHandler) GetBusinessName(bizID int) string {
	v, ok := h.businessCache.Load(bizID)
	if ok {
		return v.(string)
	}
	cs, err := h.bkccClient.SearchBusiness([]int64{int64(bizID)})
	if err != nil {
		blog.Errorf("search business '%d' from cc failed: %s", bizID, err.Error())
		return ""
	}
	if len(cs) != 1 {
		blog.Errorf("search business '%d' from cc return '%d' items", bizID, len(cs))
		return ""
	}
	h.businessCache.Store(bizID, cs[0].BkBizName)
	return cs[0].BkBizName
}

// TopProjects returns the top projects
func (h *analysisHandler) TopProjects() []*AnalysisProjectOverview {
	data := h.getCacheAnalysis()
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
		item.BizName = h.GetBusinessName(item.BizID)
	}
	return result
}

// AnalysisOverview returns the overview of analysis
func (h *analysisHandler) AnalysisOverview() (*AnalysisOverviewAll, error) {
	data := h.getCacheAnalysis()
	if len(data) == 0 {
		return nil, nil
	}
	result := &AnalysisOverviewAll{
		ProjectNum:       len(data),
		ProjectSyncTotal: make(map[string]int64),
	}

	bizMap := make(map[int64]struct{})
	userMap := make(map[string]struct{})
	sevenDayAgo := time.Now().Add(-168 * time.Hour)
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	effectiveBiz := make(map[int64]struct{})
	effectiveCluster := make(map[string]struct{})
	for i := range data {
		item := &data[i]
		if _, ok := bizMap[item.BizID]; !ok {
			result.BizNum++
			bizMap[item.BizID] = struct{}{}
		}
		if (len(item.Applications) + len(item.ApplicationSets)) > 0 {
			result.EffectiveProjectNum++
			effectiveBiz[item.BizID] = struct{}{}
		}
		result.ClusterNum += len(item.Clusters)
		result.ApplicationSetNum += len(item.ApplicationSets)
		result.ApplicationNum += len(item.Applications)
		result.SecretNum += len(item.Secrets)
		result.RepoNum += len(item.Repos)
		for _, app := range item.Applications {
			effectiveCluster[app.Cluster] = struct{}{}
		}
		var projectSync int64
		for _, appSync := range item.Syncs {
			result.SyncTotal += appSync.SyncTotal
			projectSync += appSync.SyncTotal
		}
		result.ProjectSyncTotal[item.ProjectCode] = projectSync

		for _, atUser := range item.ActivityUsers {
			result.UserOperateNum += atUser.OperateNum
			if _, ok := userMap[atUser.UserName]; ok {
				continue
			}
			if atUser.LastActivityTime.After(sevenDayAgo) {
				result.Activity7DayUserNum++
				if atUser.LastActivityTime.After(oneDayAgo) {
					result.Activity1DayUserNum++
				}
				userMap[atUser.UserName] = struct{}{}
			}
		}

		for _, appSync := range item.Syncs {
			if appSync.UpdateTime.After(sevenDayAgo) {
				result.Activity7DayProjectNum++
				break
			}
		}
		for _, appSync := range item.Syncs {
			if appSync.UpdateTime.After(oneDayAgo) {
				result.Activity1DayProjectNum++
				break
			}
		}
	}
	result.EffectiveBizNum = len(effectiveBiz)
	result.EffectiveClusterNum = len(effectiveCluster)
	return result, nil
}

// Applications query the application info
func (h *analysisHandler) Applications() ([]*ApplicationInfo, error) {
	apps := h.storage.AllApplications()
	result := make([]*ApplicationInfo, 0, len(apps))
	syncs, err := h.listSyncs("")
	if err != nil {
		return nil, errors.Wrapf(err, "list sync infos failed")
	}
	syncMap := make(map[string]*dao.SyncInfo)
	for i := range syncs {
		syncMap[syncs[i].Application] = &syncs[i]
	}
	ris, err := h.db.ListResourceInfosByProject(nil)
	if err != nil {
		return nil, errors.Wrapf(err, "list resource infos failed")
	}
	riMap := make(map[string]map[string]int64)
	for i := range ris {
		ri := ris[i]
		m := make(map[string]int64)
		if err = json.Unmarshal(utils.StringToSliceByte(ri.Resources), &m); err != nil {
			continue
		}
		riMap[ri.Application] = m
	}
	for _, app := range apps {
		var appInfo *ApplicationInfo
		if app.Spec.Source != nil {
			appInfo = &ApplicationInfo{
				Name:    app.Name,
				Cluster: getClusterIDFromServer(app.Spec.Destination.Server),
				Repo:    app.Spec.Source.RepoURL,
			}
		} else {
			appInfo = &ApplicationInfo{
				Name:    app.Name,
				Cluster: getClusterIDFromServer(app.Spec.Destination.Server),
				Repo:    app.Spec.Sources[0].RepoURL,
			}
		}
		sync, ok := syncMap[app.Name]
		if ok {
			appInfo.Sync = sync.SyncTotal
			appInfo.LastSyncTime = sync.UpdateTime
		}
		ri, ok := riMap[app.Name]
		if ok {
			resourceAll := ri[ResourceInfoAll]
			resourceGW := ri[ResourceInfoGameWorkload]
			resourceWK := ri[ResourceInfoWorkload]
			resourcePod := ri[ResourceInfoPod]
			appInfo.ResourceInfo = fmt.Sprintf("总资源: %d, 游戏负载: %d, 普通负载: %d, 纳管 Pod: %d",
				resourceAll, resourceGW, resourceWK, resourcePod)
		}
		result = append(result, appInfo)
	}
	return result, nil
}

func getClusterIDFromServer(server string) string {
	t := strings.Split(server, "/")
	clusterID := t[len(t)-1]
	if strings.HasPrefix(clusterID, "BCS-K8S-") {
		return clusterID
	}
	return server
}
