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
	"net/http"
	"sort"
	"time"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/pkg/manager/handler"
)

// ProjectObject defines the project object
type ProjectObject struct {
	BizID           int64  `json:"bizID"`
	BizName         string `json:"bizName"`
	ProjectID       string `json:"projectID"`
	ProjectCode     string `json:"projectCode"`
	ProjectName     string `json:"projectName"`
	OperatorGroup   string `json:"operatorGroup"`
	Applications    int    `json:"applications"`
	ApplicationSets int    `json:"applicationSets"`
	Clusters        int    `json:"clusters"`
	Repos           int    `json:"repos"`
	Secrets         int    `json:"secrets"`
	TotalSync       int64  `json:"totalSync"`
	TotalOperate    int64  `json:"totalOperate"`
	Active15DayUser int    `json:"active15DayUser"`
	LastSyncTime    string `json:"lastSyncTime"`
}

type projObjSort []*ProjectObject

// Len return the length of project object
func (a projObjSort) Len() int { return len(a) }

// Less return the less of project object
func (a projObjSort) Less(i, j int) bool {
	return a[i].Applications > a[j].Applications
}

// Swap item
func (a projObjSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// QueryProjects query projects
func (m *AnalysisManager) QueryProjects(writer http.ResponseWriter, request *http.Request) {
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	result := make([]*ProjectObject, 0)
	now := time.Now()
	for i := range projs {
		proj := projs[i]
		projObj := &ProjectObject{
			BizID:           proj.BizID,
			BizName:         proj.BizName,
			ProjectID:       proj.ProjectID,
			ProjectCode:     proj.ProjectCode,
			ProjectName:     proj.ProjectName,
			OperatorGroup:   projectGroup(&proj),
			Applications:    len(proj.Applications),
			ApplicationSets: len(proj.ApplicationSets),
			Clusters:        len(proj.Clusters),
			Repos:           len(proj.Repos),
			Secrets:         len(proj.Secrets),
		}
		lastSyncTime := time.Now().Add(-265 * 24 * time.Hour)
		for _, sync := range proj.Syncs {
			projObj.TotalSync += sync.SyncTotal
			updateTime, err := time.Parse(time.DateTime, sync.UpdateTime)
			if err != nil {
				continue
			}
			if updateTime.After(lastSyncTime) {
				lastSyncTime = updateTime
			}
		}
		projObj.LastSyncTime = lastSyncTime.Format(time.DateTime)
		for _, user := range proj.ActivityUsers {
			projObj.TotalOperate += user.OperateNum
			if isActiveUser(user, now.Add(-15*24*time.Hour)) {
				projObj.Active15DayUser++
			}
		}
		result = append(result, projObj)
	}
	sort.Sort(projObjSort(result))
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}

// UserObject defines the user object
type UserObject struct {
	UserName         string `json:"userName"`
	ChineseName      string `json:"chineseName"`
	ProjectCode      string `json:"projectCode"`
	ProjectName      string `json:"projectName"`
	OperateNum       int64  `json:"operateNum"`
	LastActivityTime string `json:"lastActivityTime"`

	Dept  string `json:"dept"`
	Group string `json:"group"`
}

type userObjSort []*UserObject

// Len return the length of user object
func (a userObjSort) Len() int { return len(a) }

// Less return the less of user object
func (a userObjSort) Less(i, j int) bool {
	activeA, _ := time.Parse(time.DateTime, a[i].LastActivityTime)
	activeB, _ := time.Parse(time.DateTime, a[j].LastActivityTime)
	return activeA.After(activeB)
}

// Swap item
func (a userObjSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// QueryUsers query users
func (m *AnalysisManager) QueryUsers(writer http.ResponseWriter, request *http.Request) {
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	result := make([]*UserObject, 0)
	for i := range projs {
		proj := projs[i]
		for _, user := range proj.ActivityUsers {
			result = append(result, &UserObject{
				UserName:         user.UserName,
				ChineseName:      user.ChineseName,
				ProjectCode:      proj.ProjectCode,
				ProjectName:      proj.ProjectName,
				OperateNum:       user.OperateNum,
				LastActivityTime: user.LastActivityTime,
				Dept:             userGroupOuter(user),
				Group:            userGroupInner(user),
			})
		}
	}
	sort.Sort(userObjSort(result))
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}

// ApplicationObject defines the application object
type ApplicationObject struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	Cluster      string `json:"cluster"`
	ProjectCode  string `json:"projectCode"`
	ProjectName  string `json:"projectName"`
	SyncTotal    int64  `json:"syncTotal"`
	LastSyncTime string `json:"lastSyncTime"`
}

type appObjSort []*ApplicationObject

// Len return the length of app object
func (a appObjSort) Len() int { return len(a) }

// Less return the less of app object
func (a appObjSort) Less(i, j int) bool {
	activeA, _ := time.Parse(time.DateTime, a[i].LastSyncTime)
	activeB, _ := time.Parse(time.DateTime, a[j].LastSyncTime)
	return activeA.After(activeB)
}

// Swap item
func (a appObjSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// QueryApplications query applications
func (m *AnalysisManager) QueryApplications(writer http.ResponseWriter, request *http.Request) {
	projs := m.returnAnalysisHandler(request).GetAnalysisProjects()
	result := make([]*ApplicationObject, 0)
	for i := range projs {
		proj := projs[i]
		syncMap := make(map[string]*handler.AnalysisSync)
		for _, sync := range proj.Syncs {
			syncMap[sync.Application] = sync
		}
		for _, app := range proj.Applications {
			appObj := &ApplicationObject{
				Name:        app.Name,
				Status:      app.Status,
				Cluster:     app.Cluster,
				ProjectCode: proj.ProjectCode,
				ProjectName: proj.ProjectName,
			}
			if appSync, ok := syncMap[app.Name]; ok {
				appObj.LastSyncTime = appSync.UpdateTime
				appObj.SyncTotal = appSync.SyncTotal
			}
			result = append(result, appObj)
		}
	}
	sort.Sort(appObjSort(result))
	m.httpJson(writer, &AnalysisResponse{
		Code: 0,
		Data: result,
	})
}
