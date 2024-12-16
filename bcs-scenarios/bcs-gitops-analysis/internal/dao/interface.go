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

package dao

import (
	"time"
)

const (
	tableSyncInfo     = "bcs_gitops_sync_info"
	tableActivityUser = "bcs_gitops_activity_user" // nolint
	tableResourceInfo = "bcs_gitops_resource_info" // nolint
)

// ActivityUser defines the activity user
type ActivityUser struct {
	ID               int64     `json:"id" gorm:"column:id;primaryKey;type:int(11) AUTO_INCREMENT NOT NULL"`
	Project          string    `json:"project" gorm:"index:idx_proj;column:project;type:varchar(256) NOT NULL"`
	UserName         string    `json:"userName" gorm:"column:userName;type:varchar(256) NOT NULL"`
	OperateNum       int64     `json:"operateNum" gorm:"column:operateNum;type:int(11) DEFAULT 0"`
	LastActivityTime time.Time `json:"lastActivityTime" gorm:"column:lastActivityTime;type:datetime NOT NULL"`
}

// SyncInfo defines the sync info of every application and cluster
type SyncInfo struct {
	ID           int64     `json:"id" gorm:"column:id;primaryKey;type:int(11) AUTO_INCREMENT NOT NULL"`
	Project      string    `json:"project" gorm:"index:idx_proj;column:project;type:varchar(256) NOT NULL"`
	Application  string    `json:"application" gorm:"index:idx_app;column:application;type:varchar(256) NOT NULL"`
	Cluster      string    `json:"cluster" gorm:"index:idx_cls;column:cluster;type:varchar(256) NOT NULL"`
	SyncTotal    int64     `json:"syncTotal" gorm:"column:syncTotal;type:int(11) DEFAULT 0"`
	Phase        string    `json:"phase" gorm:"index:idx_phase;column:phase;type:varchar(64) DEFAULT NULL"`
	PreviousSync int64     `json:"previousSync" gorm:"column:previousSync;type:int(11) DEFAULT 0"`
	UpdateTime   time.Time `json:"updateTime" gorm:"column:updateTime;type:datetime NOT NULL"`
}

// ResourceInfo defines the resource info object
type ResourceInfo struct {
	ID          int64     `json:"id" gorm:"column:id;primaryKey;type:int(11) AUTO_INCREMENT NOT NULL"`
	Project     string    `json:"project" gorm:"index:idx_proj;column:project;type:varchar(256) NOT NULL"`
	Application string    `json:"application" gorm:"index:idx_app;column:application;type:varchar(256) NOT NULL"`
	Resources   string    `json:"resources" gorm:"column:resources;type:text NOT NULL"`
	UpdateTime  time.Time `json:"updateTime" gorm:"column:updateTime;type:datetime NOT NULL"`
}

// Interface defines the interface of dao
type Interface interface {
	Init() error

	GetSyncInfo(project, cluster, app, phase string) (*SyncInfo, error)
	ListSyncInfosForProject(project string) ([]SyncInfo, error)
	SaveSyncInfo(info *SyncInfo) error
	UpdateSyncInfo(info *SyncInfo) error

	ListActivityUser(project string) ([]ActivityUser, error)
	List7DayActivityUsers() ([]ActivityUser, error)

	SaveOrUpdateResourceInfo(info *ResourceInfo) error
	ListResourceInfosByProject(projects []string) ([]ResourceInfo, error)
	GetResourceInfo(project, app string) (*ResourceInfo, error)
}
