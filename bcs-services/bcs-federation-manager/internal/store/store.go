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

// Package store xxx
package store

import "context"

var storeClient FederationMangerModel

// FederationListOptions list options for federation clusters
type FederationListOptions struct {
	Conditions map[string]string
}

// FederationClusterDeleteOptions options for federation cluster delete
type FederationClusterDeleteOptions struct {
	FederationClusterID string
	Updater             string
}

// SubClusterListOptions list options for sub clusters
type SubClusterListOptions struct {
	FederationClusterID string
	Conditions          map[string]string
}

// SubClusterDeleteOptions options for sub cluster delete
type SubClusterDeleteOptions struct {
	FederationClusterID string
	SubClusterID        string
	Updater             string
}

// FederationMangerModel store interface
type FederationMangerModel interface {
	// FederationCluster
	GetFederationCluster(ctx context.Context, clusterID string) (*FederationCluster, error)
	ListFederationClusters(ctx context.Context, opt *FederationListOptions) ([]*FederationCluster, error)
	CreateFederationCluster(ctx context.Context, cluster *FederationCluster) error
	DeleteFederationCluster(ctx context.Context, opt *FederationClusterDeleteOptions) error
	UpdateFederationCluster(ctx context.Context, cluster *FederationCluster, updater string) error

	// SubCluster
	GetSubCluster(ctx context.Context, fedClusterId, subClusterId string) (*SubCluster, error)
	ListSubClusters(ctx context.Context, opt *SubClusterListOptions) ([]*SubCluster, error)
	CreateSubCluster(ctx context.Context, cluster *SubCluster) error
	DeleteSubCluster(ctx context.Context, opt *SubClusterDeleteOptions) error
	UpdateSubCluster(ctx context.Context, cluster *SubCluster, updater string) error
}

// SetStoreModel set storage client
func SetStoreModel(client FederationMangerModel) {
	storeClient = client
}

// GetStoreModel get storage client
func GetStoreModel() FederationMangerModel {
	return storeClient
}
