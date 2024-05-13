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

// Package dao xxx
package dao

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
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

const (
	// PreferenceTypeApplication xxx
	PreferenceTypeApplication = "application"
)

// ResourcePreference defines the resource preference
type ResourcePreference struct {
	ID           int64  `json:"id" gorm:"column:id;primaryKey;type:int(11) AUTO_INCREMENT NOT NULL"`
	Project      string `json:"project" gorm:"index:idx_proj;column:project;type:varchar(256) NOT NULL"`
	ResourceType string `json:"resourceType" gorm:"index:idx_rt;column:resourceType;type:varchar(64) NOT NULL"`
	Name         string `json:"name" gorm:"index:idx_name;column:name;type:varchar(256) NOT NULL"`
	Collect      int64  `json:"collect" gorm:"column:collect;type:int(4) DEFAULT 0"`
}

// ApplicationHistoryManifest defines the manifest of application every history
type ApplicationHistoryManifest struct {
	ID              int64  `json:"id" gorm:"column:id;primaryKey;type:int(11) AUTO_INCREMENT NOT NULL"`
	Project         string `json:"project" gorm:"index:idx_proj;column:project;type:varchar(256) NOT NULL"`
	Name            string `json:"name" gorm:"index:idx_name;column:name;type:varchar(128) NOT NULL"`
	ApplicationUID  string `json:"applicationUID" gorm:"index:idx_uid;column:applicationUID;type:varchar(64) NOT NULL"` // nolint
	ApplicationYaml string `json:"applicationYaml" gorm:"column:applicationYaml;type:longtext NOT NULL"`

	Revision               string    `json:"revision" gorm:"column:revision;type:varchar(256) DEFAULT NULL"`
	Revisions              string    `json:"revisions" gorm:"column:revisions;type:varchar(512) DEFAULT NULL"`
	ManagedResources       string    `json:"managedResources" gorm:"column:managedResources;type:longtext NOT NULL"`
	HistoryID              int64     `json:"historyID" gorm:"column:historyID;type:int(11) NOT NULL"`
	HistoryDeployStartedAt time.Time `json:"historyDeployStartedAt" gorm:"column:historyDeployStartedAt;type:datetime NOT NULL"` // nolint
	HistoryDeployedAt      time.Time `json:"historyDeployedAt" gorm:"column:historyDeployedAt;type:datetime NOT NULL"`
	HistorySource          string    `json:"historySource" gorm:"column:historySource;type:longtext NOT NULL"`
	HistorySources         string    `json:"historySources" gorm:"column:historySources;type:longtext NOT NULL"`

	// 将大字符串存储成二进制的方式缩小磁盘占用
	ApplicationYamlBytes  []byte `json:"applicationYamlBytes" gorm:"column:applicationYamlBytes;type:LONGBLOB DEFAULT NULL"`   // nolint
	ManagedResourcesBytes []byte `json:"managedResourcesBytes" gorm:"column:managedResourcesBytes;type:LONGBLOB DEFAULT NULL"` // nolint
}

// Encode 很对大字符串在保存前进行压缩
func (ahm *ApplicationHistoryManifest) Encode() error {
	appYaml, err := utils.GzipEncode(utils.StringToSliceByte(ahm.ApplicationYaml))
	if err != nil {
		return errors.Wrapf(err, "gzip encode application yaml failed")
	}
	ahm.ApplicationYamlBytes = appYaml
	// make original field empty
	ahm.ApplicationYaml = ""

	managedResource, err := utils.GzipEncode(utils.StringToSliceByte(ahm.ManagedResources))
	if err != nil {
		return errors.Wrapf(err, "gzip encode application yaml failed")
	}
	ahm.ManagedResourcesBytes = managedResource
	// make original field empty
	ahm.ManagedResources = ""
	return nil
}

// GetManagedResources 获取解压缩后的 managed resources 字符串
func (ahm *ApplicationHistoryManifest) GetManagedResources() []byte {
	// 如果 bytes 为空，表明是历史遗留的大字符串
	if ahm.ManagedResourcesBytes == nil {
		return utils.StringToSliceByte(ahm.ManagedResources)
	}
	bs, err := utils.GzipDecode(ahm.ManagedResourcesBytes)
	if err != nil {
		blog.Warnf("gzip decode id '%d''s managed resources failed: %s", ahm.ID, err.Error())
		return utils.StringToSliceByte(ahm.ManagedResources)
	}
	return bs
}

// GetApplicationYaml 获取解压缩后的 application yaml 字符串
func (ahm *ApplicationHistoryManifest) GetApplicationYaml() string {
	// 如果 bytes 为空，表明是历史遗留的大字符串
	if ahm.ApplicationYamlBytes == nil {
		return ahm.ApplicationYaml
	}
	bs, err := utils.GzipDecode(ahm.ApplicationYamlBytes)
	if err != nil {
		blog.Warnf("gzip decode id '%d''s managed resources failed: %s", ahm.ID, err.Error())
		return ahm.ApplicationYaml
	}
	return utils.SliceByteToString(bs)
}

const (
	tableActivityUser       = "bcs_gitops_activity_user" // nolint
	tableResourcePreference = "bcs_gitops_resource_preference"
	tableHistoryManifest    = "bcs_gitops_app_history_manifest"
)

// Interface xxx interface
type Interface interface {
	Init() error

	UpdateActivityUserWithName(item *ActivityUserItem)

	SaveResourcePreference(prefer *ResourcePreference) error
	DeleteResourcePreference(project, resourceType, name string) error
	ListResourcePreferences(project, resourceType string) ([]ResourcePreference, error)

	SaveApplicationHistoryManifest(hm *ApplicationHistoryManifest) error
	GetApplicationHistoryManifest(appName, appUID string, historyID int64) (*ApplicationHistoryManifest, error)
	CheckAppHistoryManifestExist(appName, appUID string, historyID int64) (bool, error)
}
