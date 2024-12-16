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

package manager

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/bkm"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/manager/handler"
)

var (
	internalTarget = "gitops"
	externalTarget = "gitops-external"
)

// collectAnalysis will collect internal and external analysis data
func (m *AnalysisManager) collectAnalysis() {
	if m.op.IsExternal {
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		projects := m.internalAnalysisHandler.GetAnalysisProjects()
		m.handleCollectAnalysis(internalTarget, projects)
	}()
	go func() {
		defer wg.Done()
		projects := m.externalAnalysisHandler.GetAnalysisProjects()
		m.handleCollectAnalysis(externalTarget, projects)
	}()
	wg.Wait()
}

// handleCollectAnalysis will calculate the analysis from projects data. And then send
// the data as metrics to bkmonitor platform.
func (m *AnalysisManager) handleCollectAnalysis(target string, projects []handler.AnalysisProject) {
	var effectiveProjects, active1DayProjects, active7DayProjects int
	var active3MonthUsers, active3MonthManageDeptUser, active3MonthNotManageDeptUser int
	var active1DayUsers, active1DayManageDeptUser, active1DayNotManageDeptUser int
	var active7DayUsers, active7DayManageDeptUser, active7DayNotManageDeptUser int
	now := time.Now()
	appear1DayUsers := make(map[string]struct{})
	appear7DayUsers := make(map[string]struct{})
	appear3MonthUsers := make(map[string]struct{})
	for i := range projects {
		proj := projects[i]
		if (len(proj.Applications) + len(proj.ApplicationSets)) != 0 {
			effectiveProjects++
		}
		if isActiveProject(&proj, now.Add(-24*time.Hour)) {
			active1DayProjects++
		}
		if isActiveProject(&proj, now.Add(-24*7*time.Hour)) {
			active7DayProjects++
		}
		for _, user := range proj.ActivityUsers {
			if isActiveUser(user, now.Add(-24*time.Hour)) {
				if _, ok := appear1DayUsers[user.UserName]; !ok {
					active1DayUsers++
					appear1DayUsers[user.UserName] = struct{}{}
					if isManageDeptUser(user, &proj) {
						active1DayManageDeptUser++
					} else {
						active1DayNotManageDeptUser++
					}
				}
			}
			if isActiveUser(user, now.Add(-24*7*time.Hour)) {
				if _, ok := appear7DayUsers[user.UserName]; !ok {
					active7DayUsers++
					appear7DayUsers[user.UserName] = struct{}{}
					if isManageDeptUser(user, &proj) {
						active7DayManageDeptUser++
					} else {
						active7DayNotManageDeptUser++
					}
				}
			}
			if isActiveUser(user, now.Add(-24*30*3*time.Hour)) {
				if _, ok := appear3MonthUsers[user.UserName]; !ok {
					active3MonthUsers++
					appear3MonthUsers[user.UserName] = struct{}{}
					if isManageDeptUser(user, &proj) {
						active3MonthManageDeptUser++
					} else {
						active3MonthNotManageDeptUser++
					}
				}
			}
		}
	}
	bkmMessage := &bkm.BKMonitorMessage{
		DataID:      m.op.BKMonitorPushDataID,
		AccessToken: m.op.BKMonitorPushToken,
		Data: []*bkm.BKMonitorMessageData{
			{
				Metrics: map[string]interface{}{
					"effective_projects":                  effectiveProjects,  // 有效接入项目数量
					"active_1day_projects":                active1DayProjects, // 最近 1 天活跃项目
					"active_7day_projects":                active7DayProjects, // 最近 7 天活跃项目
					"active_3month_users":                 active3MonthUsers,  // 最近 3 月用户
					"active_3month_manage_dept_users":     active3MonthManageDeptUser,
					"active_3month_not_manage_dept_users": active3MonthNotManageDeptUser,
					"active_1day_users":                   active1DayUsers, // 最近 1 天用户
					"active_1day_manage_dept_users":       active1DayManageDeptUser,
					"active_1day_not_manage_dept_users":   active1DayNotManageDeptUser,
					"active_7day_users":                   active7DayUsers, // 最近 7 天用户
					"active_7day_manage_dept_users":       active7DayManageDeptUser,
					"active_7day_not_manage_dept_users":   active7DayNotManageDeptUser,
				},
				Dimension: map[string]string{},
				Target:    target,
				Timestamp: time.Now().UnixMilli(),
			},
		},
	}
	userOperate := make([]*bkm.BKMonitorMessageData, 0)
	for i := range projects {
		proj := projects[i]
		for _, user := range proj.ActivityUsers {
			operateMetric := &bkm.BKMonitorMessageData{
				Metrics: map[string]interface{}{
					"user_operate": user.OperateNum,
				},
				Dimension: map[string]string{
					"project":   proj.ProjectName,
					"user":      user.GroupLevel2 + "/" + user.UserName,
					"user_dept": user.GroupLevel2,
				},
				Target:    target,
				Timestamp: time.Now().UnixMilli(),
			}
			bkmMessage.Data = append(bkmMessage.Data, operateMetric)
			userOperate = append(userOperate, operateMetric)
		}
		for _, appSync := range proj.Syncs {
			bkmMessage.Data = append(bkmMessage.Data, &bkm.BKMonitorMessageData{
				Metrics: map[string]interface{}{
					"app_sync": appSync.SyncTotal,
				},
				Dimension: map[string]string{
					"project":     proj.ProjectName,
					"application": appSync.Application,
				},
				Target:    target,
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}
	uoBS, _ := json.Marshal(userOperate)
	blog.Infof("collect analysis for '%s' success, userOperate: %s", target, string(uoBS))
	m.bkmClient.Push(bkmMessage)
}
