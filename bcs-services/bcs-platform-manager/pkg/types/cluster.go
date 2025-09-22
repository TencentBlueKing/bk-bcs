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

// ListClusterReq list cluster request
type ListClusterReq struct {
	ProjectID  string `json:"projectID" in:"query=projectID"`
	BusinessID string `json:"businessID" in:"query=businessID"`
	Provider   string `json:"provider" in:"query=provider"`
	SortKey    string `json:"sortKey" in:"query=sortKey"`
	SortWay    string `json:"sortWay" in:"query=sortWay"` // asc or desc
}

// ListClusterRsp list cluster response
type ListClusterRsp struct {
	ClusterID       string            `json:"clusterID"`
	ClusterName     string            `json:"clusterName"`
	Provider        string            `json:"provider"`
	Region          string            `json:"region"`
	VpcID           string            `json:"vpcID"`
	ProjectID       string            `json:"projectID"`
	BusinessID      string            `json:"businessID"`
	Environment     string            `json:"environment"`
	EngineType      string            `json:"engineType"`
	ClusterType     string            `json:"clusterType"`
	Label           map[string]string `json:"label"`
	Creator         string            `json:"creator"`
	CreateTime      string            `json:"createTime"`
	UpdateTime      string            `json:"updateTime"`
	SystemID        string            `json:"systemID"`
	ManageType      string            `json:"manageType"`
	Status          string            `json:"status"`
	Updater         string            `json:"updater"`
	NetworkType     string            `json:"networkType"`
	ModuleID        string            `json:"moduleID"`
	IsCommonCluster bool              `json:"isCommonCluster"`
	Description     string            `json:"description"`
	ClusterCategory string            `json:"clusterCategory"`
	IsShared        bool              `json:"isShared"`
	Link            string            `json:"link"`
	SortKey         string            `json:"-"`
	SortWay         string            `json:"-"`
}

// UpdateClusterOperatorReq update cluster operator request
type UpdateClusterOperatorReq struct {
	ClusterID string `json:"clusterID" in:"path=clusterID"`
	Creator   string `json:"creator"`
	Updater   string `json:"updater"`
}

// UpdateClusterProjectBusinessReq update cluster projectID or businessID request
type UpdateClusterProjectBusinessReq struct {
	ClusterID  string `json:"clusterID" in:"path=clusterID"`
	ProjectID  string `json:"projectID"`
	BusinessID string `json:"businessID"`
}
