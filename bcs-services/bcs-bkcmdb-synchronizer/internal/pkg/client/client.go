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

// Package client define client interface
package client

import (
	"context"

	bkcmdbkube "configcenter/src/kube/types" // nolint
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"google.golang.org/grpc"
)

// Client is an interface that defines methods for getting client instances.
type Client interface {
	// GetClusterManagerConnWithURL returns a gRPC client connection with URL.
	GetClusterManagerConnWithURL() (*grpc.ClientConn, error)
	// GetClusterManagerClient returns the ClusterManagerClient instance.
	GetClusterManagerClient() (cmp.ClusterManagerClient, error)
	// GetClusterManagerConn returns a gRPC client connection.
	GetClusterManagerConn() (*grpc.ClientConn, error)
	// NewCMGrpcClientWithHeader returns a new ClusterManagerClientWithHeader instance.
	NewCMGrpcClientWithHeader(ctx context.Context, conn *grpc.ClientConn) *ClusterManagerClientWithHeader

	// GetProjectManagerConnWithURL returns a gRPC client connection with URL.
	GetProjectManagerConnWithURL() (*grpc.ClientConn, error)
	// GetProjectManagerClient returns the BCSProjectClient instance.
	GetProjectManagerClient() (pmp.BCSProjectClient, error)
	// GetProjectManagerConn returns a gRPC client connection.
	GetProjectManagerConn() (*grpc.ClientConn, error)
	// NewPMGrpcClientWithHeader returns a new ProjectManagerClientWithHeader instance.
	NewPMGrpcClientWithHeader(ctx context.Context, conn *grpc.ClientConn) *ProjectManagerClientWithHeader

	// GetStorageClient returns the Storage instance.
	GetStorageClient() (bcsapi.Storage, error)

	// GetCMDBClient returns the CMDBClient instance.
	GetCMDBClient() (CMDBClient, error)
}

// CMDBClient is an interface that defines methods for interacting with the CMDB.
type CMDBClient interface {
	// GetBS2IDByBizID returns the BS2 ID for the given Biz ID.
	GetBS2IDByBizID(int64) (int, error)
	// GetBizInfo returns the Business information for the given Biz ID.
	GetBizInfo(int64) (*Business, error)
	// GetHostInfo returns the Host information for the given list of host IDs.
	GetHostInfo([]string) (*[]HostData, error)

	// GetBcsNamespace returns the BCS namespace information for the given request.
	GetBcsNamespace(*GetBcsNamespaceRequest) (*[]bkcmdbkube.Namespace, error)
	// GetBcsNode returns the BCS node information for the given request.
	GetBcsNode(*GetBcsNodeRequest) (*[]bkcmdbkube.Node, error)
	// GetBcsWorkload returns the BCS workload information for the given request.
	GetBcsWorkload(*GetBcsWorkloadRequest) (*[]interface{}, error)
	// GetBcsPod returns the BCS pod information for the given request.
	GetBcsPod(*GetBcsPodRequest) (*[]bkcmdbkube.Pod, error)
	// GetBcsCluster returns the BCS cluster information for the given request.
	GetBcsCluster(*GetBcsClusterRequest) (*[]bkcmdbkube.Cluster, error)
	// GetBcsContainer retrieves the BCS container information based on the provided request.
	GetBcsContainer(request *GetBcsContainerRequest) (*[]bkcmdbkube.Container, error)

	// CreateBcsNode creates a new BCS node with the given request.
	CreateBcsNode(*CreateBcsNodeRequest) (*[]int64, error)
	// DeleteBcsNode deletes the BCS node with the given request.
	DeleteBcsNode(*DeleteBcsNodeRequest) error
	// UpdateBcsNode updates the BCS node with the given request.
	UpdateBcsNode(*UpdateBcsNodeRequest) error

	// CreateBcsNamespace creates a new BCS namespace with the given request.
	CreateBcsNamespace(*CreateBcsNamespaceRequest) (*[]int64, error)
	// DeleteBcsNamespace deletes the BCS namespace with the given request.
	DeleteBcsNamespace(*DeleteBcsNamespaceRequest) error
	// UpdateBcsNamespace updates the BCS namespace with the given request.
	UpdateBcsNamespace(*UpdateBcsNamespaceRequest) error

	// CreateBcsWorkload creates a new BCS workload with the given request.
	CreateBcsWorkload(*CreateBcsWorkloadRequest) (*[]int64, error)
	// DeleteBcsWorkload deletes the BCS workload with the given request.
	DeleteBcsWorkload(*DeleteBcsWorkloadRequest) error
	// UpdateBcsWorkload updates the BCS workload with the given request.
	UpdateBcsWorkload(*UpdateBcsWorkloadRequest) error

	// CreateBcsPod creates a new BCS pod with the given request.
	CreateBcsPod(*CreateBcsPodRequest) (*[]int64, error)
	// DeleteBcsPod deletes the BCS pod with the given request.
	DeleteBcsPod(*DeleteBcsPodRequest) error

	// CreateBcsCluster creates a new BCS cluster with the given request.
	CreateBcsCluster(*CreateBcsClusterRequest) (int64, error)
	// UpdateBcsCluster updates the BCS cluster with the given request.
	UpdateBcsCluster(*UpdateBcsClusterRequest) error
	// DeleteBcsCluster deletes the BCS cluster with the given request.
	DeleteBcsCluster(*DeleteBcsClusterRequest) error
	// UpdateBcsClusterType updates the BCS cluster type with the given request.
	UpdateBcsClusterType(request *UpdateBcsClusterTypeRequest) error

	DeleteBcsClusterAll(*DeleteBcsClusterAllRequest) error
}
