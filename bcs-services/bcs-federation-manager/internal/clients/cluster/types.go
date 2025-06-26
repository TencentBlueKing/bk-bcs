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

// Package cluster xxx
package cluster

import (
	"time"
)

const (
	// default params for create federation proxy cluster in cluster-manager
	DefaultClusterDescription = "federation proxy cluster entry"
	DefaultProviderBlueking   = "bluekingCloud"
	DefaultRegionDefault      = "default"
	DefaultEnginTypeK8s       = "k8s"
	DefaultClusterTypeSingle  = "single"
	// bkmonitor will not install single cluster dataid
	DefaultClusterTypeFederation = "federation"
	DefaultManageTypeIndependent = "INDEPENDENT_CLUSTER"
	DefaultNetworkTypeOverlay    = "overlay"

	// DefaultProviderTencent provider tencentcloud
	DefaultProviderTencent = "tencentCloud"
	// ClusterSubnetsRegexPattern cluster subnet regex pattern
	ClusterSubnetsRegexPattern = `^BCS-K8S-`
)

const (
	// FederationClusterTaskIDLabelKey federation cluster task id label key, used for relate federation task
	FederationClusterTaskIDLabelKey = "federation.bkbcs.tencent.com/taskid"
	// FederationClusterTypeLabelKeyFedCluster federation cluster type label key, used for cluster manager identify cluster type
	FederationClusterTypeLabelKeyFedCluster  = "federation.bkbcs.tencent.com/is-fed-cluster"
	FederationClusterTypeLabelKeySubCluster  = "federation.bkbcs.tencent.com/is-sub-cluster"
	FederationClusterTypeLabelKeyHostCluster = "federation.bkbcs.tencent.com/is-host-cluster"
	FederationClusterTypeLabelValueTrue      = "true"
	FederationClusterTypeLabelValueFalse     = "false"

	// ClusterStatusRunning cluster running
	ClusterStatusRunning = "RUNNING"
	// ClusterStatusInitialization cluster initialization
	ClusterStatusInitialization = "INITIALIZATION"
	// ClusterStatusCreateFailure cluster failure
	ClusterStatusCreateFailure = "CREATE-FAILURE"
	// ClusterStatusDeleting cluster deleting
	ClusterStatusDeleting = "DELETING"
	// ClusterStatusDeleteFailure cluster delete failure
	ClusterStatusDeleteFailure = "DELETE-FAILURE"
	// ClusterStatusDeleted cluster deleted
	ClusterStatusDeleted = "DELETED"
	// TaskStatusINITIALIZING task initializing
	TaskStatusINITIALIZING = "INITIALIZING"
	// TaskStatusRUNNING task running
	TaskStatusRUNNING = "RUNNING"
)

// ClusterStatusList cluster status list
var ClusterStatusList = map[string]struct{}{
	ClusterStatusRunning:        {},
	ClusterStatusInitialization: {},
	ClusterStatusCreateFailure:  {},
	ClusterStatusDeleting:       {},
	ClusterStatusDeleteFailure:  {},
	ClusterStatusDeleted:        {},
}

// FederationNamespace federation namespace for query
type FederationNamespace struct {
	HostClusterId string    // HostClusterId host cluster id
	Namespace     string    // Namespace namespace name
	SubClusters   []string  // SubClusters sub cluster id list
	ProjectCode   string    // ProjectCode project code from namespace annotations
	CreatedTime   time.Time // CreatedTime created time from namespace obj
}

// FederationClusterCreateReq req for create federation proxy cluster
type FederationClusterCreateReq struct {
	BusinessId  string
	ProjectId   string
	ClusterName string
	Environment string
	Description string
	Labels      map[string]string
	Creator     string
}

// ResourceGetOptions get options for resource
type ResourceGetOptions struct {
	ClusterId    string
	Namespace    string
	Kind         string
	ResourceName string
}

// ResourceCreateOptions create options for resource
type ResourceCreateOptions struct {
	ClusterId    string
	Namespace    string
	Kind         string
	ResourceName string
}
