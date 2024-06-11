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

// Package clusterManger cluster-service
package clusterManger

import (
	"context"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/roundtrip"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// CreateCloudAccount

// Service 对接 bcs-cluster-manager
type Service interface {
	/*
		云凭证导入
	*/
	// CreateCloudAccount 创建云账号信息
	CreateCloudAccount(ctx context.Context, req *CreateCloudAccountRequest) (*pb.CreateCloudAccountResponse, error)
	// DeleteCloudAccount 删除特定cloud 账号信息
	DeleteCloudAccount(ctx context.Context, req *DeleteCloudAccountRequest) (*pb.DeleteCloudAccountResponse, error)
	// UpdateCloudAccount 更新云账号信息
	UpdateCloudAccount(ctx context.Context, req *UpdateCloudAccountRequest) (*pb.UpdateCloudAccountResponse, error)
	// ListCloudAccount 查询Cloud账号列表
	ListCloudAccount(ctx context.Context, req *ListCloudAccountRequest) (*pb.ListCloudAccountResponse, error)

	/*
		集群管理
	*/
	// ImportCluster 集群创建（云凭证方式）
	ImportCluster(ctx context.Context, req *pb.ImportClusterReq) (*pb.ImportClusterResp, error)
	// CreateCluster  集群创建（直接创建）
	CreateCluster(ctx context.Context, req *pb.CreateClusterReq) (*pb.CreateClusterResp, error)
	// DeleteCluster  删除集群
	DeleteCluster(ctx context.Context, req *pb.DeleteClusterReq) (*pb.DeleteClusterResp, error)
	// UpdateCluster  更新集群
	UpdateCluster(ctx context.Context, req *pb.UpdateClusterReq) (*pb.UpdateClusterResp, error)
	// GetCluster  查询集群
	GetCluster(ctx context.Context, req *pb.GetClusterReq) (*pb.GetClusterResp, error)
	// ListProjectCluster 查询某个项目下的Cluster列表
	ListProjectCluster(ctx context.Context, req *pb.ListProjectClusterReq) (*pb.ListProjectClusterResp, error)

	/*
		云上资源查询(创建集群辅助接口)
	*/
	// ListCloudInstanceTypes 查询Node机型
	ListCloudInstanceTypes(ctx context.Context, req *pb.ListCloudInstanceTypeRequest) (*pb.ListCloudInstanceTypeResponse, error)
	// ListCloudSubnets 查询vpc子网列表
	ListCloudSubnets(ctx context.Context, req *pb.ListCloudSubnetsRequest) (*pb.ListCloudSubnetsResponse, error)
	// ListCloudSecurityGroups 查询安全组列表
	ListCloudSecurityGroups(ctx context.Context, req *pb.ListCloudSecurityGroupsRequest) (*pb.ListCloudSecurityGroupsResponse, error)
	// GetCloudRegions 查询cloud地域列表
	GetCloudRegions(ctx context.Context, req *pb.GetCloudRegionsRequest) (*pb.GetCloudRegionsResponse, error)
	// GetCloudRegionZones 查询cloud地域可用区列表
	GetCloudRegionZones(ctx context.Context, req *pb.GetCloudRegionZonesRequest) (*pb.GetCloudRegionZonesResponse, error)
	// ListCloudVpcs 获取云所属地域vpc列表
	ListCloudVpcs(ctx context.Context, req *pb.ListCloudVpcsRequest) (*pb.ListCloudVpcsResponse, error)
	// ListCloudProjects 获取云项目列表
	ListCloudProjects(ctx context.Context, req *pb.ListCloudProjectsRequest) (*pb.ListCloudProjectsResponse, error)
	// ListCloudOsImage 查询Node操作系统镜像列表
	ListCloudOsImage(ctx context.Context, req *pb.ListCloudOsImageRequest) (*pb.ListCloudOsImageResponse, error)
	// ListKeypairs 查询密钥对列表
	ListKeypairs(ctx context.Context, req *pb.ListKeyPairsRequest) (*pb.ListKeyPairsResponse, error)
	// GetCloudAccountType 查询云账号类型
	GetCloudAccountType(ctx context.Context, req *pb.GetCloudAccountTypeRequest) (*pb.GetCloudAccountTypeResponse, error)
	// GetCloudBandwidthPackages 查询云共享带宽包
	GetCloudBandwidthPackages(ctx context.Context, req *pb.GetCloudBandwidthPackagesRequest) (*pb.GetCloudBandwidthPackagesResponse, error)

	/*
		节点池
	*/
	// CreateNodeGroup 创建节点池，用于弹性伸缩。
	CreateNodeGroup(ctx context.Context, req *pb.CreateNodeGroupRequest) (*pb.CreateNodeGroupResponse, error)
	// DeleteNodeGroup 删除节点池
	DeleteNodeGroup(ctx context.Context, req *pb.DeleteNodeGroupRequest) (*pb.DeleteNodeGroupResponse, error)
	// UpdateNodeGroup 更新节点池
	UpdateNodeGroup(ctx context.Context, req *pb.UpdateNodeGroupRequest) (*pb.UpdateNodeGroupResponse, error)
	// UpdateGroupDesiredNode 更新期望节点数
	UpdateGroupDesiredNode(ctx context.Context, req *pb.UpdateGroupDesiredNodeRequest) (*pb.UpdateGroupDesiredNodeResponse, error)
	// UpdateGroupMinMaxSize 更新最小最大的扩容限额
	UpdateGroupMinMaxSize(ctx context.Context, req *pb.UpdateGroupMinMaxSizeRequest) (*pb.UpdateGroupMinMaxSizeResponse, error)
	// GetNodeGroup 查询指定NodeGroupID信息
	GetNodeGroup(ctx context.Context, req *pb.GetNodeGroupRequest) (*pb.GetNodeGroupResponse, error)
	// ListClusterNodeGroup 获取集群节点池列表
	ListClusterNodeGroup(ctx context.Context, req *pb.ListClusterNodeGroupRequest) (*pb.ListClusterNodeGroupResponse, error)
}

// NewService return Service
func NewService(config *options.Config, roundtrip roundtrip.Client) (Service, error) {
	if config == nil {
		return nil, errors.Errorf("config cannot be empty.")
	}
	if roundtrip == nil {
		return nil, errors.Errorf("roundtrip cannot be empty.")
	}

	h := &handler{
		config: config,
		// roundtrip
		roundtrip: roundtrip,
		// api
		backendApi: map[string]string{},
	}
	if err := h.init(); err != nil {
		return nil, errors.Wrapf(err, "init handler failed")
	}

	return h, nil
}

// handler impl Service
type handler struct {
	// config 配置
	config *options.Config

	// roundtrip http client
	roundtrip roundtrip.Client

	// backendApi 后端完整路径
	backendApi map[string]string
}

func (h *handler) init() error {
	apis := []string{
		// 云凭证导入
		createCloudAccountApi,
		deleteCloudAccountApi,
		updateCloudAccountApi,
		listCloudAccountApi,
		// 集群管理
		importClusterApi,
		createClusterApi,
		deleteClusterApi,
		updateClusterApi,
		getClusterApi,
		listProjectClusterApi,
		// 其他创建集群辅助接口
		listCloudInstanceTypeApi,
		listCloudSubnetsApi,
		listCloudSecurityGroupsApi,
		getCloudRegionsApi,
		getCloudRegionZonesApi,
		listCloudVpcsApi,
		listCloudProjectsApi,
		listCloudOsImageApi,
		listKeypairsApi,
		getCloudAccountTypeApi,
		getCloudBandwidthPackagesApi,
		// 节点池
		createNodeGroupApi,
		deleteNodeGroupApi,
		updateNodeGroupApi,
		updateGroupDesiredNodeApi,
		updateGroupMinMaxSizeApi,
		getNodeGroupApi,
		listClusterNodeGroupApi,
	}
	addr := h.config.BcsGatewayAddr

	for _, api := range apis {
		h.backendApi[api] = utils.PathJoin(addr, api)
	}

	return nil
}
