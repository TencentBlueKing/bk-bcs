/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mock

import (
	"context"

	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockCm mock cm
type MockCm struct {
	mock.Mock
}

// CreateCluster mock cm create cluster
func (m *MockCm) CreateCluster(ctx context.Context, in *cm.CreateClusterReq, opts ...grpc.CallOption) (*cm.CreateClusterResp, error) {
}

// RetryCreateClusterTask mock cm
func (m *MockCm) RetryCreateClusterTask(ctx context.Context, in *cm.RetryCreateClusterReq, opts ...grpc.CallOption) (*cm.RetryCreateClusterResp, error) {
}

// ImportCluster mock cm
func (m *MockCm) ImportCluster(ctx context.Context, in *cm.ImportClusterReq, opts ...grpc.CallOption) (*cm.ImportClusterResp, error) {
}

// UpdateCluster mock cm
func (m *MockCm) UpdateCluster(ctx context.Context, in *cm.UpdateClusterReq, opts ...grpc.CallOption) (*cm.UpdateClusterResp, error) {
}

// AddNodesToCluster mock cm
func (m *MockCm) AddNodesToCluster(ctx context.Context, in *cm.AddNodesRequest, opts ...grpc.CallOption) (*cm.AddNodesResponse, error) {
}

// DeleteNodesFromCluster mock cm
func (m *MockCm) DeleteNodesFromCluster(ctx context.Context, in *cm.DeleteNodesRequest, opts ...grpc.CallOption) (*cm.DeleteNodesResponse, error) {
}

// ListNodesInCluster mock cm
func (m *MockCm) ListNodesInCluster(ctx context.Context, in *cm.ListNodesInClusterRequest, opts ...grpc.CallOption) (*cm.ListNodesInClusterResponse, error) {
}

// DeleteCluster mock cm
func (m *MockCm) DeleteCluster(ctx context.Context, in *cm.DeleteClusterReq, opts ...grpc.CallOption) (*cm.DeleteClusterResp, error) {
}

// GetCluster mock cm
func (m *MockCm) GetCluster(ctx context.Context, in *cm.GetClusterReq, opts ...grpc.CallOption) (*cm.GetClusterResp, error) {
}

// ListCluster mock cm
func (m *MockCm) ListCluster(ctx context.Context, in *cm.ListClusterReq, opts ...grpc.CallOption) (*cm.ListClusterResp, error) {
}

// * node management

// GetNode mock cm
func (m *MockCm) GetNode(ctx context.Context, in *cm.GetNodeRequest, opts ...grpc.CallOption) (*cm.GetNodeResponse, error) {
}

// UpdateNode mock cm
func (m *MockCm) UpdateNode(ctx context.Context, in *cm.UpdateNodeRequest, opts ...grpc.CallOption) (*cm.UpdateNodeResponse, error) {
}

// CheckNodeInCluster mock cm
func (m *MockCm) CheckNodeInCluster(ctx context.Context, in *cm.CheckNodesRequest, opts ...grpc.CallOption) (*cm.CheckNodesResponse, error) {
}

// * cluster credential management

// GetClusterCredential mock cm
func (m *MockCm) GetClusterCredential(ctx context.Context, in *cm.GetClusterCredentialReq, opts ...grpc.CallOption) (*cm.GetClusterCredentialResp, error) {
}

// UpdateClusterCredential mock cm
func (m *MockCm) UpdateClusterCredential(ctx context.Context, in *cm.UpdateClusterCredentialReq, opts ...grpc.CallOption) (*cm.UpdateClusterCredentialResp, error) {
}

// DeleteClusterCredential mock cm
func (m *MockCm) DeleteClusterCredential(ctx context.Context, in *cm.DeleteClusterCredentialReq, opts ...grpc.CallOption) (*cm.DeleteClusterCredentialResp, error) {
}

// ListClusterCredential mock cm
func (m *MockCm) ListClusterCredential(ctx context.Context, in *cm.ListClusterCredentialReq, opts ...grpc.CallOption) (*cm.ListClusterCredentialResp, error) {
}

// * federation cluster management

// InitFederationCluster mock cm
func (m *MockCm) InitFederationCluster(ctx context.Context, in *cm.InitFederationClusterReq, opts ...grpc.CallOption) (*cm.InitFederationClusterResp, error) {
}

// AddFederatedCluster mock cm
func (m *MockCm) AddFederatedCluster(ctx context.Context, in *cm.AddFederatedClusterReq, opts ...grpc.CallOption) (*cm.AddFederatedClusterResp, error) {
}

// * namespace management *

// CreateNamespace mock cm
func (m *MockCm) CreateNamespace(ctx context.Context, in *cm.CreateNamespaceReq, opts ...grpc.CallOption) (*cm.CreateNamespaceResp, error) {
}

// UpdateNamespace mock cm
func (m *MockCm) UpdateNamespace(ctx context.Context, in *cm.UpdateNamespaceReq, opts ...grpc.CallOption) (*cm.UpdateNamespaceResp, error) {
}

// DeleteNamespace mock cm
func (m *MockCm) DeleteNamespace(ctx context.Context, in *cm.DeleteNamespaceReq, opts ...grpc.CallOption) (*cm.DeleteNamespaceResp, error) {
}

// GetNamespace mock cm
func (m *MockCm) GetNamespace(ctx context.Context, in *cm.GetNamespaceReq, opts ...grpc.CallOption) (*cm.GetNamespaceResp, error) {
}

// ListNamespace mock cm
func (m *MockCm) ListNamespace(ctx context.Context, in *cm.ListNamespaceReq, opts ...grpc.CallOption) (*cm.ListNamespaceResp, error) {
}

// * NamespaceQuota management *

// CreateNamespaceQuota mock cm
func (m *MockCm) CreateNamespaceQuota(ctx context.Context, in *cm.CreateNamespaceQuotaReq, opts ...grpc.CallOption) (*cm.CreateNamespaceQuotaResp, error) {
}

// UpdateNamespaceQuota mock cm
func (m *MockCm) UpdateNamespaceQuota(ctx context.Context, in *cm.UpdateNamespaceQuotaReq, opts ...grpc.CallOption) (*cm.UpdateNamespaceQuotaResp, error) {
}

// DeleteNamespaceQuota mock cm
func (m *MockCm) DeleteNamespaceQuota(ctx context.Context, in *cm.DeleteNamespaceQuotaReq, opts ...grpc.CallOption) (*cm.DeleteNamespaceQuotaResp, error) {
}

// GetNamespaceQuota mock cm
func (m *MockCm) GetNamespaceQuota(ctx context.Context, in *cm.GetNamespaceQuotaReq, opts ...grpc.CallOption) (*cm.GetNamespaceQuotaResp, error) {
}

// ListNamespaceQuota mock cm
func (m *MockCm) ListNamespaceQuota(ctx context.Context, in *cm.ListNamespaceQuotaReq, opts ...grpc.CallOption) (*cm.ListNamespaceQuotaResp, error) {
}

// CreateNamespaceWithQuota mock cm
func (m *MockCm) CreateNamespaceWithQuota(ctx context.Context, in *cm.CreateNamespaceWithQuotaReq, opts ...grpc.CallOption) (*cm.CreateNamespaceWithQuotaResp, error) {
}

// * project information management *
// CreateProject mock cm
func (m *MockCm) CreateProject(ctx context.Context, in *cm.CreateProjectRequest, opts ...grpc.CallOption) (*cm.CreateProjectResponse, error) {
}

// UpdateProject mock cm
func (m *MockCm) UpdateProject(ctx context.Context, in *cm.UpdateProjectRequest, opts ...grpc.CallOption) (*cm.UpdateProjectResponse, error) {
}

// DeleteProject mock cm
func (m *MockCm) DeleteProject(ctx context.Context, in *cm.DeleteProjectRequest, opts ...grpc.CallOption) (*cm.DeleteProjectResponse, error) {
}

// GetProject mock cm
func (m *MockCm) GetProject(ctx context.Context, in *cm.GetProjectRequest, opts ...grpc.CallOption) (*cm.GetProjectResponse, error) {
}

// ListProject mock cm
func (m *MockCm) ListProject(ctx context.Context, in *cm.ListProjectRequest, opts ...grpc.CallOption) (*cm.ListProjectResponse, error) {
}

// * Cloud information management *

// CreateCloud mock cm
func (m *MockCm) CreateCloud(ctx context.Context, in *cm.CreateCloudRequest, opts ...grpc.CallOption) (*cm.CreateCloudResponse, error) {
}

// UpdateCloud mock cm
func (m *MockCm) UpdateCloud(ctx context.Context, in *cm.UpdateCloudRequest, opts ...grpc.CallOption) (*cm.UpdateCloudResponse, error) {
}

// DeleteCloud mock cm
func (m *MockCm) DeleteCloud(ctx context.Context, in *cm.DeleteCloudRequest, opts ...grpc.CallOption) (*cm.DeleteCloudResponse, error) {
}

// GetCloud mock cm
func (m *MockCm) GetCloud(ctx context.Context, in *cm.GetCloudRequest, opts ...grpc.CallOption) (*cm.GetCloudResponse, error) {
}

// ListCloud mock cm
func (m *MockCm) ListCloud(ctx context.Context, in *cm.ListCloudRequest, opts ...grpc.CallOption) (*cm.ListCloudResponse, error) {
}

// * Cloud VPC information management *

// CreateCloudVPC mock cm
func (m *MockCm) CreateCloudVPC(ctx context.Context, in *cm.CreateCloudVPCRequest, opts ...grpc.CallOption) (*cm.CreateCloudVPCResponse, error) {
}

// UpdateCloudVPC mock cm
func (m *MockCm) UpdateCloudVPC(ctx context.Context, in *cm.UpdateCloudVPCRequest, opts ...grpc.CallOption) (*cm.UpdateCloudVPCResponse, error) {
}

// DeleteCloudVPC mock cm
func (m *MockCm) DeleteCloudVPC(ctx context.Context, in *cm.DeleteCloudVPCRequest, opts ...grpc.CallOption) (*cm.DeleteCloudVPCResponse, error) {
}

// ListCloudVPC mock cm
func (m *MockCm) ListCloudVPC(ctx context.Context, in *cm.ListCloudVPCRequest, opts ...grpc.CallOption) (*cm.ListCloudVPCResponse, error) {
}

// ListCloudRegions mock cm
func (m *MockCm) ListCloudRegions(ctx context.Context, in *cm.ListCloudRegionsRequest, opts ...grpc.CallOption) (*cm.ListCloudRegionsResponse, error) {
}

// GetVPCCidr mock cm
func (m *MockCm) GetVPCCidr(ctx context.Context, in *cm.GetVPCCidrRequest, opts ...grpc.CallOption) (*cm.GetVPCCidrResponse, error) {
}

// * NodeGroup information management *

// CreateNodeGroup mock cm
func (m *MockCm) CreateNodeGroup(ctx context.Context, in *cm.CreateNodeGroupRequest, opts ...grpc.CallOption) (*cm.CreateNodeGroupResponse, error) {
}

// UpdateNodeGroup mock cm
func (m *MockCm) UpdateNodeGroup(ctx context.Context, in *cm.UpdateNodeGroupRequest, opts ...grpc.CallOption) (*cm.UpdateNodeGroupResponse, error) {
}

// DeleteNodeGroup mock cm
func (m *MockCm) DeleteNodeGroup(ctx context.Context, in *cm.DeleteNodeGroupRequest, opts ...grpc.CallOption) (*cm.DeleteNodeGroupResponse, error) {
}

// GetNodeGroup mock cm
func (m *MockCm) GetNodeGroup(ctx context.Context, in *cm.GetNodeGroupRequest, opts ...grpc.CallOption) (*cm.GetNodeGroupResponse, error) {
}

// ListNodeGroup mock cm
func (m *MockCm) ListNodeGroup(ctx context.Context, in *cm.ListNodeGroupRequest, opts ...grpc.CallOption) (*cm.ListNodeGroupResponse, error) {
}

// MoveNodesToGroup mock cm
func (m *MockCm) MoveNodesToGroup(ctx context.Context, in *cm.MoveNodesToGroupRequest, opts ...grpc.CallOption) (*cm.MoveNodesToGroupResponse, error) {
}

// RemoveNodesFromGroup mock cm
func (m *MockCm) RemoveNodesFromGroup(ctx context.Context, in *cm.RemoveNodesFromGroupRequest, opts ...grpc.CallOption) (*cm.RemoveNodesFromGroupResponse, error) {
}

// CleanNodesInGroup mock cm
func (m *MockCm) CleanNodesInGroup(ctx context.Context, in *cm.CleanNodesInGroupRequest, opts ...grpc.CallOption) (*cm.CleanNodesInGroupResponse, error) {
}

// ListNodesInGroup mock cm
func (m *MockCm) ListNodesInGroup(ctx context.Context, in *cm.GetNodeGroupRequest, opts ...grpc.CallOption) (*cm.ListNodesInGroupResponse, error) {
}

// UpdateGroupDesiredNode mock cm
func (m *MockCm) UpdateGroupDesiredNode(ctx context.Context, in *cm.UpdateGroupDesiredNodeRequest, opts ...grpc.CallOption) (*cm.UpdateGroupDesiredNodeResponse, error) {
}

// UpdateGroupDesiredSize mock cm
func (m *MockCm) UpdateGroupDesiredSize(ctx context.Context, in *cm.UpdateGroupDesiredSizeRequest, opts ...grpc.CallOption) (*cm.UpdateGroupDesiredSizeResponse, error) {
}

// * Task information management *

// CreateTask mock cm
func (m *MockCm) CreateTask(ctx context.Context, in *cm.CreateTaskRequest, opts ...grpc.CallOption) (*cm.CreateTaskResponse, error) {
}

// RetryTask mock cm
func (m *MockCm) RetryTask(ctx context.Context, in *cm.RetryTaskRequest, opts ...grpc.CallOption) (*cm.RetryTaskResponse, error) {
}

// UpdateTask mock cm
func (m *MockCm) UpdateTask(ctx context.Context, in *cm.UpdateTaskRequest, opts ...grpc.CallOption) (*cm.UpdateTaskResponse, error) {
}

// DeleteTask mock cm
func (m *MockCm) DeleteTask(ctx context.Context, in *cm.DeleteTaskRequest, opts ...grpc.CallOption) (*cm.DeleteTaskResponse, error) {
}

// GetTask mock cm
func (m *MockCm) GetTask(ctx context.Context, in *cm.GetTaskRequest, opts ...grpc.CallOption) (*cm.GetTaskResponse, error) {
}

// ListTask mock cm
func (m *MockCm) ListTask(ctx context.Context, in *cm.ListTaskRequest, opts ...grpc.CallOption) (*cm.ListTaskResponse, error) {
}

// * ClusterAutoScalingOption information management *

// CreateAutoScalingOption mock cm
func (m *MockCm) CreateAutoScalingOption(ctx context.Context, in *cm.CreateAutoScalingOptionRequest, opts ...grpc.CallOption) (*cm.CreateAutoScalingOptionResponse, error) {
}

// UpdateAutoScalingOption mock cm
func (m *MockCm) UpdateAutoScalingOption(ctx context.Context, in *cm.UpdateAutoScalingOptionRequest, opts ...grpc.CallOption) (*cm.UpdateAutoScalingOptionResponse, error) {
}

// DeleteAutoScalingOption mock cm
func (m *MockCm) DeleteAutoScalingOption(ctx context.Context, in *cm.DeleteAutoScalingOptionRequest, opts ...grpc.CallOption) (*cm.DeleteAutoScalingOptionResponse, error) {
}

// GetAutoScalingOption mock cm
func (m *MockCm) GetAutoScalingOption(ctx context.Context, in *cm.GetAutoScalingOptionRequest, opts ...grpc.CallOption) (*cm.GetAutoScalingOptionResponse, error) {
}

// ListAutoScalingOption mock cm
func (m *MockCm) ListAutoScalingOption(ctx context.Context, in *cm.ListAutoScalingOptionRequest, opts ...grpc.CallOption) (*cm.ListAutoScalingOptionResponse, error) {
}
