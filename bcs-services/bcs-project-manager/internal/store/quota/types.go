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

// Package quota xxx
package quota

const (
	// FieldKeyQuotaId id
	FieldKeyQuotaId = "quotaId"
	// FieldKeyProjectId projectId
	FieldKeyProjectId = "projectId"
	// FieldKeyProjectCode projectCode
	FieldKeyProjectCode = "projectCode"
	// FieldKeyQuotaName quota name
	FieldKeyQuotaName = "quotaName"
	// FieldKeyClusterId clusterId
	FieldKeyClusterId = "clusterId"
	// FieldKeyBusinessId clusterId
	FieldKeyBusinessId = "businessId"
	// FieldKeyIsDeleted isDeleted
	FieldKeyIsDeleted = "isDeleted"
	// FieldKeyStatus status
	FieldKeyStatus = "status"
	// FieldKeyDeleteTime deleteTime
	FieldKeyDeleteTime = "deleteTime"
	// FieldKeyCreateTime createTime
	FieldKeyCreateTime = "createTime"
	// FieldKeyUpdateTime updateTime
	FieldKeyUpdateTime = "updateTime"
	// FieldKeyQuotaType quotaType
	FieldKeyQuotaType = "quotaType"
	// FieldKeyQuota quota
	FieldKeyQuota = "quota"
)

// GetSortField field sort, default descend
func GetSortField(field string) map[string]int {
	return map[string]int{field: -1}
}

// ProjectQuotaType project quota type
type ProjectQuotaType string

// String to string
func (qt ProjectQuotaType) String() string {
	return string(qt)
}

var (
	// Host host 整机资源类型
	Host ProjectQuotaType = "host"
	// Common common 通用资源类型(包括区分不同类型的通用资源)
	Common ProjectQuotaType = "common"
	// Shared shared 共享集群资源类型
	Shared ProjectQuotaType = "shared"
	// Federation federation 共享集群资源类型
	Federation ProjectQuotaType = "federation"
)

// ProjectQuotaStatusType quota status
type ProjectQuotaStatusType string

// String to string
func (qst ProjectQuotaStatusType) String() string {
	return string(qst)
}

var (
	// Running quota normal
	Running ProjectQuotaStatusType = "RUNNING"
	// Creating quota creating
	Creating ProjectQuotaStatusType = "CREATING"
	// CreateFailure quota create failure
	CreateFailure ProjectQuotaStatusType = "CREATE-FAILURE"
	// Deleting quota deleting
	Deleting ProjectQuotaStatusType = "DELETING"
	// DeleteFailure quota delete failure
	DeleteFailure ProjectQuotaStatusType = "DELETE-FAILURE"
	// Deleted quota deleted
	Deleted ProjectQuotaStatusType = "DELETED"
)

// ProjectQuota xxx
type ProjectQuota struct {
	CreateTime  int64                  `json:"createTime" bson:"createTime"`
	UpdateTime  int64                  `json:"updateTime" bson:"updateTime"`
	DeleteTime  int64                  `json:"deleteTime" bson:"deleteTime"`
	Creator     string                 `json:"creator" bson:"creator"`
	Updater     string                 `json:"updater" bson:"updater"`
	QuotaId     string                 `json:"quotaId" bson:"quotaId"`
	QuotaName   string                 `json:"quotaName" bson:"quotaName"`
	Description string                 `json:"description" bson:"description"`
	Provider    string                 `json:"provider" bson:"provider"`
	QuotaType   ProjectQuotaType       `json:"quotaType" bson:"quotaType"`
	Quota       *QuotaResource         `json:"quota" bson:"quota"`
	ProjectId   string                 `json:"projectId" bson:"projectId"`
	ProjectCode string                 `json:"projectCode" bson:"projectCode"`
	ClusterId   string                 `json:"clusterId" bson:"clusterId"`
	Namespace   string                 `json:"namespace" bson:"namespace"`
	BusinessId  string                 `json:"businessId" bson:"businessId"`
	IsDeleted   bool                   `json:"isDeleted" bson:"isDeleted"`
	Status      ProjectQuotaStatusType `json:"status" bson:"status"`
	Labels      map[string]string      `json:"labels" bson:"labels"`
}

// QuotaResource quota resource type (包含整机资源类型 / 通用资源类型(按照匹配优先级确定项目维度、共享集群维度通用资源))
type QuotaResource struct { // nolint
	HostResources *HostConfig `json:"hostResources" bson:"hostResources"`
	Cpu           *DeviceInfo `json:"cpu" bson:"cpu"`
	Mem           *DeviceInfo `json:"mem" bson:"mem"`
	Gpu           *DeviceInfo `json:"gpu" bson:"gpu"`
}

// HostConfig host instance resource config
type HostConfig struct {
	Region       string       `json:"region" bson:"region"`
	InstanceType string       `json:"instanceType" bson:"instanceType"`
	Cpu          uint32       `json:"cpu" bson:"cpu"`
	Mem          uint32       `json:"mem" bson:"mem"`
	Gpu          uint32       `json:"gpu" bson:"gpu"`
	ZoneId       string       `json:"zoneId" bson:"zoneId"`
	ZoneName     string       `json:"zoneName" bson:"zoneName"`
	QuotaNum     uint32       `json:"quotaNum" bson:"quotaNum"`
	UsedNum      uint32       `json:"usedNum" bson:"usedNum"`
	SystemDisk   DeviceDisk   `json:"systemDisk" bson:"systemDisk"`
	DataDisks    []DeviceDisk `json:"dataDisks" bson:"dataDisks"`
}

// DeviceDisk structure storage
type DeviceDisk struct {
	Type string `json:"type" bson:"type"`
	Size string `json:"size" bson:"size"`
}

// DeviceInfo device info, DeviceType 区分quota的类型(intel/amd), 为空则为通用的cpu核心数
type DeviceInfo struct {
	DeviceType  string            `json:"deviceType" bson:"deviceType"`
	DeviceQuota string            `json:"deviceQuota" bson:"deviceQuota"`
	Attributes  map[string]string `json:"attributes" bson:"attributes"`
}
