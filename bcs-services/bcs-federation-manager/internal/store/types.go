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

package store

import "time"

const (
	RunningStatus      = "Running"
	DeletedStatus      = "Deleted"
	CreatingStatus     = "Creating"
	CreateFailedStatus = "CreateFailed"
	UnknownStatus      = "Unknown"
	DeletingStatus     = "Deleting"
)

// FederationCluster federation cluster form for store
type FederationCluster struct {
	FederationClusterID   string `json:"federation_cluster_id" bson:"federation_cluster_id"`
	FederationClusterName string `json:"federation_cluster_name" bson:"federation_cluster_name"`
	HostClusterID         string `json:"host_cluster_id" bson:"host_cluster_id"`
	ProjectCode           string `json:"project_code" bson:"project_code"`
	ProjectID             string `json:"project_id" bson:"project_id"`
	IsDeleted             bool   `json:"is_deleted" bson:"is_deleted"`
	Descriptions          string `json:"descriptions" bson:"descriptions"`
	Status                string `json:"status" bson:"status"`
	StatusMessage         string `json:"status_message" bson:"status_message"`

	Creator     string    `json:"creator" bson:"creator"`
	Updater     string    `json:"updater" bson:"updater"`
	CreatedTime time.Time `json:"created_time" bson:"created_time"`
	UpdatedTime time.Time `json:"updated_time" bson:"updated_time"`
	DeletedTime time.Time `json:"deleted_time" bson:"deleted_time"`

	// extra params such as params for register federation cluster
	Extras map[string]string `json:"extras" bson:"extras"`
}

const (
	// Keys for SubCluster Labels
	ArchKey         = "subscription.bkbcs.tencent.com/arch"
	AreaKey         = "subscription.bkbcs.tencent.com/area"
	ClusterTypeKey  = "subscription.bkbcs.tencent.com/clustertype"
	MixerClusterKey = "subscription.bkbcs.tencent.com/mixercluster"
	RegionKey       = "subscription.bkbcs.tencent.com/region"
	ResourceTypeKey = "subscription.bkbcs.tencent.com/resourcetype"
)

// SubCluster sub cluster form for store
type SubCluster struct {
	// The sub cluster may be managed by multiple federated clusters
	// and needs to form a unique ID with the federated cluster
	// format: federation_cluster_id/sub_cluster_id
	UID                 string `json:"uid" bson:"uid"`
	SubClusterID        string `json:"sub_cluster_id" bson:"sub_cluster_id"`
	SubClusterName      string `json:"sub_cluster_name" bson:"sub_cluster_name"`
	FederationClusterID string `json:"federation_cluster_id" bson:"federation_cluster_id"`
	HostClusterID       string `json:"host_cluster_id" bson:"host_cluster_id"`
	ProjectCode         string `json:"project_code" bson:"project_code"`
	ProjectID           string `json:"project_id" bson:"project_id"`
	IsDeleted           bool   `json:"is_deleted" bson:"is_deleted"`

	ClusternetClusterID        string `json:"clusternet_cluster_id" bson:"clusternet_cluster_id"`
	ClusternetClusterName      string `json:"clusternet_cluster_name" bson:"clusternet_cluster_name"`
	ClusternetClusterNamespace string `json:"clusternet_cluster_namespace" bson:"clusternet_cluster_namespace"`

	Descriptions  string `json:"descriptions" bson:"descriptions"`
	Status        string `json:"status" bson:"status"`
	StatusMessage string `json:"status_message" bson:"status_message"`

	Labels map[string]string `json:"labels" bson:"labels"`

	Creator     string    `json:"creator" bson:"creator"`
	Updater     string    `json:"updater" bson:"updater"`
	CreatedTime time.Time `json:"created_time" bson:"created_time"`
	UpdatedTime time.Time `json:"updated_time" bson:"updated_time"`
	DeletedTime time.Time `json:"deleted_time" bson:"deleted_time"`
}
