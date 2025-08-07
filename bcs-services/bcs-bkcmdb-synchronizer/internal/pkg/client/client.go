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
	//pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	pmp "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/types"
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
	// GetHostInfo returns the Host information for the given list of host IPs.
	GetHostInfo([]string) (*[]HostData, error)
	// GetHostsByBiz returns the Host information for the given list of host IPs
	GetHostsByBiz(ctx context.Context, bkBizID int64, hostIP []string) (*[]HostData, error)

	// GetBcsNamespace returns the BCS namespace information for the given request.
	GetBcsNamespace(ctx context.Context, request *GetBcsNamespaceRequest, db *gorm.DB, withDB bool) (*[]bkcmdbkube.Namespace, error)
	// GetBcsNode returns the BCS node information for the given request.
	GetBcsNode(ctx context.Context, request *GetBcsNodeRequest, db *gorm.DB, withDB bool) (*[]bkcmdbkube.Node, error)
	// GetBcsWorkload returns the BCS workload information for the given request.
	GetBcsWorkload(ctx context.Context, request *GetBcsWorkloadRequest, db *gorm.DB, withDB bool) (*[]interface{}, error)
	// GetBcsPod returns the BCS pod information for the given request.
	GetBcsPod(ctx context.Context, request *GetBcsPodRequest, db *gorm.DB, withDB bool) (*[]bkcmdbkube.Pod, error)
	// GetBcsCluster returns the BCS cluster information for the given request.
	GetBcsCluster(ctx context.Context, request *GetBcsClusterRequest, db *gorm.DB, withDB bool) (*[]bkcmdbkube.Cluster, error)
	// GetBcsContainer retrieves the BCS container information based on the provided request.
	GetBcsContainer(ctx context.Context, request *GetBcsContainerRequest, db *gorm.DB, withDB bool) (*[]Container, error)

	// CreateBcsNode creates a new BCS node with the given request.
	CreateBcsNode(ctx context.Context, request *CreateBcsNodeRequest, db *gorm.DB) (*[]int64, error)
	// DeleteBcsNode deletes the BCS node with the given request.
	DeleteBcsNode(ctx context.Context, request *DeleteBcsNodeRequest, db *gorm.DB) error
	// UpdateBcsNode updates the BCS node with the given request.
	UpdateBcsNode(ctx context.Context, request *UpdateBcsNodeRequest, db *gorm.DB) error

	// CreateBcsNamespace creates a new BCS namespace with the given request.
	CreateBcsNamespace(ctx context.Context, request *CreateBcsNamespaceRequest, db *gorm.DB) (*[]int64, error)
	// DeleteBcsNamespace deletes the BCS namespace with the given request.
	DeleteBcsNamespace(ctx context.Context, request *DeleteBcsNamespaceRequest, db *gorm.DB) error
	// UpdateBcsNamespace updates the BCS namespace with the given request.
	UpdateBcsNamespace(ctx context.Context, request *UpdateBcsNamespaceRequest, db *gorm.DB) error

	// CreateBcsWorkload creates a new BCS workload with the given request.
	CreateBcsWorkload(ctx context.Context, request *CreateBcsWorkloadRequest, db *gorm.DB) (*[]int64, error)
	// DeleteBcsWorkload deletes the BCS workload with the given request.
	DeleteBcsWorkload(ctx context.Context, request *DeleteBcsWorkloadRequest, db *gorm.DB) error
	// UpdateBcsWorkload updates the BCS workload with the given request.
	UpdateBcsWorkload(ctx context.Context, request *UpdateBcsWorkloadRequest, db *gorm.DB) error

	// CreateBcsPod creates a new BCS pod with the given request.
	CreateBcsPod(ctx context.Context, request *CreateBcsPodRequest, db *gorm.DB) (*[]int64, error)
	// DeleteBcsPod deletes the BCS pod with the given request.
	DeleteBcsPod(ctx context.Context, request *DeleteBcsPodRequest, db *gorm.DB) error

	// CreateBcsCluster creates a new BCS cluster with the given request.
	CreateBcsCluster(ctx context.Context, request *CreateBcsClusterRequest, db *gorm.DB) (int64, error)
	// UpdateBcsCluster updates the BCS cluster with the given request.
	UpdateBcsCluster(ctx context.Context, request *UpdateBcsClusterRequest, db *gorm.DB) error
	// DeleteBcsCluster deletes the BCS cluster with the given request.
	DeleteBcsCluster(ctx context.Context, request *DeleteBcsClusterRequest, db *gorm.DB) error
	// UpdateBcsClusterType updates the BCS cluster type with the given request.
	UpdateBcsClusterType(ctx context.Context, request *UpdateBcsClusterTypeRequest, db *gorm.DB) error
	// DeleteBcsClusterAll 删除所有的BCS集群，根据给定的请求。
	// 参数request包含了删除BCS集群所需的信息。
	// 参数db是gorm数据库连接实例，用于执行数据库操作。
	DeleteBcsClusterAll(request *DeleteBcsClusterAllRequest, db *gorm.DB) error
}
