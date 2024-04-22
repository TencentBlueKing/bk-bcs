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
	"time"
)

// calculateOverviewAll 根据各个项目的运营数据，汇总计算总的运营数据
func (c *analysisOverviewClient) calculateOverviewAll(data []AnalysisProject) *AnalysisOverviewAll {
	if len(data) == 0 {
		return nil
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
				if appSync.UpdateTime.After(oneDayAgo) {
					result.Activity1DayProjectNum++
				}
				break
			}
		}
	}
	result.EffectiveBizNum = len(effectiveBiz)
	result.EffectiveClusterNum = len(effectiveCluster)
	return result
}
