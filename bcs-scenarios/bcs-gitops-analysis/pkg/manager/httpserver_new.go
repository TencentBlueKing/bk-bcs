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
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/manager/handler"
)

// RawData return all data with raw format
func (m *AnalysisManager) RawData(writer http.ResponseWriter, request *http.Request) {
	result := m.returnAnalysisHandler(request).GetAnalysisProjects()
	m.httpJson(writer, result)
}

// AnalysisOverview defines the overview for analysis
type AnalysisOverview struct {
	Projects        int   `json:"projects"`
	Clusters        int   `json:"clusters"`
	Applications    int   `json:"applications"`
	ApplicationSets int   `json:"applicationSets"`
	SyncTotal       int64 `json:"syncTotal"`
	UserOperate     int64 `json:"userOperate"`
}

// OverviewNew 运营总览数据
func (m *AnalysisManager) OverviewNew(writer http.ResponseWriter, request *http.Request) {
	result := &AnalysisOverview{}
	projects := m.returnAnalysisHandler(request).GetAnalysisProjects()
	for i := range projects {
		proj := projects[i]
		if (len(proj.Applications) + len(proj.ApplicationSets)) != 0 {
			result.Projects++
		}
		result.Clusters += len(proj.Clusters)
		result.Applications += len(proj.Applications)
		result.ApplicationSets += len(proj.ApplicationSets)
		for _, sync := range proj.Syncs {
			result.SyncTotal += sync.SyncTotal
		}
		for _, user := range proj.ActivityUsers {
			result.UserOperate += user.OperateNum
		}
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: []*AnalysisOverview{result},
	})
}

// AnalysisResourceInfo defines the resource info object
type AnalysisResourceInfo struct {
	Name         string `json:"name"`
	ResourceAll  int64  `json:"resourceAll"`
	GameWorkload int64  `json:"gameWorkload"`
	Workload     int64  `json:"workload"`
	Pod          int64  `json:"pod"`
}

// ManagedResourceNew 纳管资源统计
func (m *AnalysisManager) ManagedResourceNew(writer http.ResponseWriter, request *http.Request) {
	result := &AnalysisResourceInfo{}
	rsInfos := m.returnAnalysisHandler(request).GetResourceInfo()
	for i := range rsInfos {
		rsInfo := rsInfos[i]
		result.ResourceAll += rsInfo.ResourceAll
		result.GameWorkload += rsInfo.GameWorkload
		result.Workload += rsInfo.Workload
		result.Pod += rsInfo.Pod
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: []*AnalysisResourceInfo{result},
	})
}

// ProjectManagedResourceNew project managed resources
func (m *AnalysisManager) ProjectManagedResourceNew(writer http.ResponseWriter, request *http.Request) {
	rsInfos := m.returnAnalysisHandler(request).GetResourceInfo()
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: rsInfos,
	})
}

// AnalysisProjectOverview defines the project overview
type AnalysisProjectOverview struct {
	Activity1DayProjects  int `json:"activity1DayProjects"`
	Activity7DayProjects  int `json:"activity7DayProjects"`
	Activity30DayProjects int `json:"activity30DayProjects"`
}

func isActiveProject(proj *handler.AnalysisProject, until time.Time) bool {
	for _, sync := range proj.Syncs {
		updateTime, err := time.Parse(time.DateTime, sync.UpdateTime)
		if err != nil {
			continue
		}
		if until.Before(updateTime) {
			return true
		}
	}
	return false
}

func (m *AnalysisManager) handleProjectOverview(request *http.Request) *AnalysisProjectOverview {
	result := &AnalysisProjectOverview{}
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	now := time.Now()
	for i := range projs {
		proj := projs[i]
		if isActiveProject(&proj, now.Add(-24*time.Hour)) {
			result.Activity1DayProjects++
		}
		if isActiveProject(&proj, now.Add(-24*7*time.Hour)) {
			result.Activity7DayProjects++
		}
		if isActiveProject(&proj, now.Add(-24*30*time.Hour)) {
			result.Activity30DayProjects++
		}
	}
	return result
}

func projectGroup(proj *handler.AnalysisProject) string {
	result := make([]string, 0)
	if proj.GroupLevel3 != "" {
		result = append(result, proj.GroupLevel3)
	}
	if proj.GroupLevel4 != "" {
		result = append(result, proj.GroupLevel4)
	}
	if proj.GroupLevel5 != "" {
		result = append(result, proj.GroupLevel5)
	}
	if len(result) == 0 {
		return ""
	}
	return strings.Join(result, "/")
}

func isActiveUser(user *handler.AnalysisActivityUser, until time.Time) bool {
	lastActivityTime, err := time.Parse(time.DateTime, user.LastActivityTime)
	if err != nil {
		return false
	}
	if until.Before(lastActivityTime) {
		return true
	}
	return false
}

func isManageDeptUser(user *handler.AnalysisActivityUser, proj *handler.AnalysisProject) bool {
	if user.GroupLevel0 == proj.GroupLevel0 && user.GroupLevel1 == proj.GroupLevel1 &&
		user.GroupLevel2 == proj.GroupLevel2 {
		return true
	}
	return false
}

func isManageDeptOperator(user *handler.AnalysisActivityUser, proj *handler.AnalysisProject) bool {
	if user.GroupLevel0 == proj.GroupLevel0 && user.GroupLevel1 == proj.GroupLevel1 &&
		user.GroupLevel2 == proj.GroupLevel2 && user.GroupLevel3 == proj.GroupLevel3 &&
		user.GroupLevel4 == proj.GroupLevel4 {
		return true
	}
	return false
}

// AnalysisProjectGroup defines the project group
type AnalysisProjectGroup struct {
	Group      string   `json:"group"`
	Projects   []string `json:"projects"`
	ProjectNum int      `json:"projectNum"`
	// 项目主管部门活跃用户
	ManageDeptUsers int `json:"manageDeptUsers"`
	// 项目主管部门活跃运维
	ManageDeptOperatorMap map[string]struct{} `json:"-"`
	ManageDeptOperators   []string            `json:"manageDeptOperators"`
	// 非项目主管部门活跃用户
	NotManagedDeptUsers int `json:"notManagedDeptUsers"`
}

type projectGroupSort []*AnalysisProjectGroup

// Len return the len for project-group
func (a projectGroupSort) Len() int { return len(a) }

// Less return the less for project-group
func (a projectGroupSort) Less(i, j int) bool {
	return len(a[i].Projects) > len(a[j].Projects)
}

// Swap item
func (a projectGroupSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// ProjectGroups return project groups
func (m *AnalysisManager) ProjectGroups(writer http.ResponseWriter, request *http.Request) {
	groupProjects := make(map[string]*AnalysisProjectGroup)
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	now := time.Now()
	for i := range projs {
		proj := projs[i]
		group := projectGroup(&proj)
		if group == "" {
			blog.Warnf("project '%s' not have group", proj.ProjectName)
			continue
		}
		if len(proj.Applications)+len(proj.ApplicationSets) == 0 {
			continue
		}
		groupProj, ok := groupProjects[group]
		if !ok {
			groupProjects[group] = &AnalysisProjectGroup{
				ManageDeptOperatorMap: make(map[string]struct{}),
			}
			groupProj = groupProjects[group]
		}
		groupProj.Group = group
		var active = "不活跃"
		if isActive := isActiveProject(&proj, now.Add(-24*15*time.Hour)); isActive {
			active = "活跃"
		}
		groupProj.Projects = append(groupProj.Projects, fmt.Sprintf("%s(%s)", proj.ProjectName, active))
		for _, user := range proj.ActivityUsers {
			if !isActiveUser(user, now.Add(-24*7*time.Hour)) {
				continue
			}
			if isManageDeptUser(user, &proj) {
				groupProj.ManageDeptUsers++
			} else {
				groupProj.NotManagedDeptUsers++
			}
			if isManageDeptOperator(user, &proj) {
				groupProj.ManageDeptOperatorMap[user.UserName] = struct{}{}
			}
		}
		groupProj.ProjectNum = len(groupProj.Projects)
	}
	result := make([]*AnalysisProjectGroup, 0, len(groupProjects))
	for _, gp := range groupProjects {
		for k := range gp.ManageDeptOperatorMap {
			gp.ManageDeptOperators = append(gp.ManageDeptOperators, k)
		}
		result = append(result, gp)
	}
	sort.Sort(projectGroupSort(result))
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}

// AnalysisProjectDetail defines the project detail
type AnalysisProjectDetail struct {
	BizName            string `json:"bizName"`
	OperatorGroup      string `json:"operatorGroup"`
	ProjectCode        string `json:"projectCode"`
	ProjectName        string `json:"projectName"`
	Clusters           int    `json:"clusters"`
	Applications       int    `json:"applications"`
	ActiveApplications int    `json:"activeApplications"`
	// 项目主管部门用户
	ManageDeptUsers int `json:"manageDeptUsers"`
	// 非项目主管部门用户
	NotManageDeptUsers int `json:"notManageDeptUsers"`
	// 项目主管部门活跃运维
	ManageDeptOperators []string `json:"manageDeptOperators"`
}

type projectDetailSort []*AnalysisProjectDetail

// Len return the length for project detail
func (a projectDetailSort) Len() int { return len(a) }

// Less return less for project detail
func (a projectDetailSort) Less(i, j int) bool {
	return (a[i].ManageDeptUsers + a[i].NotManageDeptUsers) > (a[j].ManageDeptUsers + a[j].NotManageDeptUsers)
}

// Swap item
func (a projectDetailSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func isActiveApplication(appSync *handler.AnalysisSync, until time.Time) bool {
	updateTime, err := time.Parse(time.DateTime, appSync.UpdateTime)
	if err != nil {
		return false
	}
	if until.Before(updateTime) {
		return true
	}
	return false
}

// ProjectRank return project rank
func (m *AnalysisManager) ProjectRank(writer http.ResponseWriter, request *http.Request) {
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	result := make([]*AnalysisProjectDetail, 0, len(projs))
	now := time.Now()
	for i := range projs {
		proj := projs[i]
		projDetail := &AnalysisProjectDetail{
			BizName:       proj.BizName,
			OperatorGroup: projectGroup(&proj),
			ProjectCode:   proj.ProjectCode,
			ProjectName:   proj.ProjectName,
			Clusters:      len(proj.Clusters),
			Applications:  len(proj.Applications),
		}
		for _, sync := range proj.Syncs {
			if isActiveApplication(sync, now.Add(-24*7*time.Hour)) {
				projDetail.ActiveApplications++
			}
		}
		for _, user := range proj.ActivityUsers {
			if !isActiveUser(user, now.Add(-24*7*time.Hour)) {
				continue
			}
			if isManageDeptUser(user, &proj) {
				projDetail.ManageDeptUsers++
			} else {
				projDetail.NotManageDeptUsers++
			}
			if isManageDeptOperator(user, &proj) {
				projDetail.ManageDeptOperators = append(projDetail.ManageDeptOperators, user.UserName)
			}
		}
		result = append(result, projDetail)
	}
	sort.Sort(projectDetailSort(result))
	limit := request.URL.Query().Get("limit")
	var limitInt = 10
	if limit != "" {
		t, err := strconv.Atoi(limit)
		if err == nil {
			limitInt = t
		}
	}
	if len(result) >= limitInt {
		result = result[:limitInt]
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}

// AnalysisUserOverview defines the overview for user
type AnalysisUserOverview struct {
	Activity1DayUsers  int `json:"activity1DayUsers"`
	Activity7DayUsers  int `json:"activity7DayUsers"`
	Activity30DayUsers int `json:"activity30DayUsers"`
}

// UserOverview return the user overview
func (m *AnalysisManager) UserOverview(writer http.ResponseWriter, request *http.Request) {
	users := make(map[string]struct{})
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	result := &AnalysisUserOverview{}
	now := time.Now()
	for i := range projs {
		proj := projs[i]
		if len(proj.Applications)+len(proj.ApplicationSets) == 0 {
			continue
		}
		for _, user := range proj.ActivityUsers {
			if _, ok := users[user.UserName]; ok {
				continue
			}
			users[user.UserName] = struct{}{}
			if isActiveUser(user, now.Add(-24*time.Hour)) {
				result.Activity1DayUsers++
			}
			if isActiveUser(user, now.Add(-24*7*time.Hour)) {
				result.Activity7DayUsers++
			}
			if isActiveUser(user, now.Add(-24*30*time.Hour)) {
				result.Activity30DayUsers++
			}
		}
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: []*AnalysisUserOverview{result},
	})
}

// AnalysisUserGroup defines the user group
type AnalysisUserGroup struct {
	Group string `json:"group"`
	Users int    `json:"users"`
}

type userGroupSort []*AnalysisUserGroup

// Len return the length of user group
func (a userGroupSort) Len() int { return len(a) }

// Less return the less for user group
func (a userGroupSort) Less(i, j int) bool {
	return a[i].Users > a[j].Users
}

// Swap item
func (a userGroupSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func userGroupOuter(user *handler.AnalysisActivityUser) string {
	if user.GroupLevel2 == "" {
		return "others"
	}
	return user.GroupLevel2
}

// UserGroupDeptOuter return the user group dept outer
func (m *AnalysisManager) UserGroupDeptOuter(writer http.ResponseWriter, request *http.Request) {
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	groups := make(map[string]*AnalysisUserGroup)
	appearUsers := make(map[string]struct{})
	now := time.Now()
	for i := range projs {
		proj := projs[i]
		if len(proj.Applications)+len(proj.ApplicationSets) == 0 {
			continue
		}
		for _, user := range proj.ActivityUsers {
			ug := userGroupOuter(user)
			aug, ok := groups[ug]
			if !ok {
				groups[ug] = &AnalysisUserGroup{}
				aug = groups[ug]
			}
			aug.Group = ug
			if _, ok = appearUsers[user.UserName]; ok {
				continue
			}
			if isActiveUser(user, now.Add(-7*24*time.Hour)) {
				aug.Users++
				appearUsers[user.UserName] = struct{}{}
			}
		}
	}
	result := make([]*AnalysisUserGroup, 0, len(groups))
	for _, g := range groups {
		if g.Users == 0 {
			continue
		}
		result = append(result, g)
	}
	sort.Sort(userGroupSort(result))
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}

// AnalysisUserOperateGroup defines the user group for operate
type AnalysisUserOperateGroup struct {
	Group   string `json:"group"`
	Operate int64  `json:"operate"`
}

// UserGroupDeptOperate return the user group dept user_operate
func (m *AnalysisManager) UserGroupDeptOperate(writer http.ResponseWriter, request *http.Request) {
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	groups := make(map[string]*AnalysisUserOperateGroup)
	appearUsers := make(map[string]struct{})
	now := time.Now()
	for i := range projs {
		proj := projs[i]
		if len(proj.Applications)+len(proj.ApplicationSets) == 0 {
			continue
		}
		for _, user := range proj.ActivityUsers {
			ug := userGroupOuter(user)
			aug, ok := groups[ug]
			if !ok {
				groups[ug] = &AnalysisUserOperateGroup{}
				aug = groups[ug]
			}
			aug.Group = ug
			if _, ok = appearUsers[user.UserName]; ok {
				continue
			}
			if isActiveUser(user, now.Add(-7*24*time.Hour)) {
				aug.Operate += user.OperateNum
				appearUsers[user.UserName] = struct{}{}
			}
		}
	}
	result := make([]*AnalysisUserOperateGroup, 0, len(groups))
	for _, g := range groups {
		if g.Operate == 0 {
			continue
		}
		result = append(result, g)
	}
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}

func userGroupInner(user *handler.AnalysisActivityUser) string {
	result := make([]string, 0)
	if user.GroupLevel3 != "" {
		result = append(result, user.GroupLevel3)
	}
	if user.GroupLevel4 != "" {
		result = append(result, user.GroupLevel4)
	}
	if user.GroupLevel5 != "" {
		result = append(result, user.GroupLevel5)
	}
	if len(result) == 0 {
		return ""
	}
	return strings.Join(result, "/")
}

// UserGroupDeptInner return the user group dept inner
func (m *AnalysisManager) UserGroupDeptInner(writer http.ResponseWriter, request *http.Request) {
	// 待查询的部门
	queryDept := request.URL.Query().Get("dept")
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	groups := make(map[string]*AnalysisUserGroup)
	appearUsers := make(map[string]struct{})
	now := time.Now()
	for i := range projs {
		proj := projs[i]
		if len(proj.Applications)+len(proj.ApplicationSets) == 0 {
			continue
		}
		for _, user := range proj.ActivityUsers {
			if queryDept != user.GroupLevel2 {
				continue
			}
			ug := userGroupInner(user)
			aug, ok := groups[ug]
			if !ok {
				groups[ug] = &AnalysisUserGroup{}
				aug = groups[ug]
			}
			aug.Group = ug
			if _, ok = appearUsers[user.UserName]; ok {
				continue
			}
			if isActiveUser(user, now.Add(-7*24*time.Hour)) {
				aug.Users++
				appearUsers[user.UserName] = struct{}{}
			}
		}
	}
	result := make([]*AnalysisUserGroup, 0, len(groups))
	for _, g := range groups {
		if g.Users == 0 {
			continue
		}
		result = append(result, g)
	}
	sort.Sort(userGroupSort(result))
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}
