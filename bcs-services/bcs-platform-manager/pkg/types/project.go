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

// Package types pod types
package types

// ListProjectsReq request for list projects
type ListProjectsReq struct {
	ProjectIDs  string `json:"projectIDs" in:"query=projectIDs"`
	Names       string `json:"names" in:"query=names"`
	ProjectCode string `json:"projectCode" in:"query=projectCode"`
	SearchName  string `json:"searchName" in:"query=searchName"`
	Kind        string `json:"kind" in:"query=kind"`
	Offset      int64  `json:"offset" in:"query=offset"`
	Limit       int64  `json:"limit" in:"query=limit"`
	All         bool   `json:"all" in:"query=all"`
	BusinessID  string `json:"businessID" in:"query=businessID"`
}

// ListProjectsResp response for list project
type ListProjectsResp struct {
	Total   uint32              `json:"total"`
	Results []*ListProjectsData `json:"results"`
}

// ListProjectsData response for list project
type ListProjectsData struct {
	ProjectID    string `json:"projectID"`
	Name         string `json:"name"`
	ProjectCode  string `json:"projectCode"`
	Description  string `json:"description"`
	Creator      string `json:"creator"`
	IsOffline    bool   `json:"isOffline"`
	BusinessID   string `json:"businessID"`
	BusinessName string `json:"businessName"`
	Managers     string `json:"managers"`
	CreateTime   string `json:"createTime"`
	Link         string `json:"link"`
}

// GetProjectsReq request for get project
type GetProjectsReq struct {
	ProjectIDOrCode string `json:"projectIDOrCode" in:"path=projectIDOrCode"`
}

// GetProjectsResp response for get project
type GetProjectsResp struct {
	CreateTime   string            `json:"createTime"`
	UpdateTime   string            `json:"updateTime"`
	Creator      string            `json:"creator"`
	Updater      string            `json:"updater"`
	Managers     string            `json:"managers"`
	ProjectID    string            `json:"projectID"`
	Name         string            `json:"name"`
	ProjectCode  string            `json:"projectCode"`
	UseBKRes     bool              `json:"useBKRes"`
	Description  string            `json:"description"`
	IsOffline    bool              `json:"isOffline"`
	Kind         string            `json:"kind"`
	BusinessID   string            `json:"businessID"`
	BusinessName string            `json:"businessName"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
}

// UpdateProjectReq request for update project
type UpdateProjectReq struct {
	ProjectID   string            `json:"projectID" in:"path=projectID"`
	Managers    string            `json:"managers"`
	BusinessID  string            `json:"businessID"`
	Name        string            `json:"name"`
	ProjectCode string            `json:"projectCode"`
	UseBKRes    bool              `json:"useBKRes"`
	Description string            `json:"description"`
	Kind        string            `json:"kind"`
	IsOffline   bool              `json:"isOffline"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// UpdateProjectManagersReq request for update project managers
type UpdateProjectManagersReq struct {
	ProjectID string `json:"projectID" in:"path=projectID"`
	Managers  string `json:"managers"`
}

// UpdateProjectBusinessReq request for update project business
type UpdateProjectBusinessReq struct {
	ProjectID  string `json:"projectID" in:"path=projectID"`
	BusinessID string `json:"businessID"`
}

// UpdateProjectIsOfflineReq request for update project isoffline
type UpdateProjectIsOfflineReq struct {
	ProjectID string `json:"projectID" in:"path=projectID"`
	IsOffline bool   `json:"isOffline"`
}
