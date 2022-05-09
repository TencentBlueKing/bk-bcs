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
	"encoding/json"

	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// MockCm mock cm
type MockCm struct {
	mock.Mock
}

// NewMockCm new mock cm
func NewMockCm() cm.ClusterManagerClient {
	mockCm := &MockCm{}
	rawClusters := []byte("{\"code\":0,\"message\":\"success\",\"result\":true,\"data\":[{\"clusterID\":\"BCS-MESOS-10039\",\"clusterName\":\"mesos测试\",\"federationClusterID\":\"\",\"provider\":\"bluekingCloud\",\"region\":\"sz\",\"vpcID\":\"\",\"projectID\":\"ab2b254938e84f6b86b466cc22e730b1\",\"businessID\":\"100148\",\"environment\":\"stag\",\"engineType\":\"mesos\",\"isExclusive\":true,\"clusterType\":\"single\",\"labels\":{\"bellketest\":\"test\"},\"creator\":\"boweiguan\",\"createTime\":\"2021-11-23T21:59:32+08:00\",\"updateTime\":\"2021-11-23T22:06:27+08:00\",\"bcsAddons\":{},\"extraAddons\":{},\"systemID\":\"\",\"manageType\":\"INDEPENDENT_CLUSTER\",\"master\":{\"\":{\"nodeID\":\"\",\"innerIP\":\"\",\"instanceType\":\"\",\"CPU\":0,\"mem\":0,\"GPU\":0,\"status\":\"RUNNING\",\"zoneID\":\"\",\"nodeGroupID\":\"\",\"clusterID\":\"\",\"VPC\":\"\",\"region\":\"\",\"passwd\":\"\",\"zone\":0,\"deviceID\":\"\"}},\"networkSettings\":{\"clusterIPv4CIDR\":\"\",\"serviceIPv4CIDR\":\"\",\"maxNodePodNum\":0,\"maxServiceNum\":0,\"enableVPCCni\":false,\"eniSubnetIDs\":[],\"subnetSource\":null,\"isStaticIpMode\":false,\"claimExpiredSeconds\":0,\"multiClusterCIDR\":[],\"cidrStep\":0},\"clusterBasicSettings\":{\"OS\":\"Tencent tlinux release 2.2 (Final)\",\"version\":\"1.16\",\"clusterTags\":{},\"versionName\":\"BCS-MESOS-10039\"},\"clusterAdvanceSettings\":{\"IPVS\":false,\"containerRuntime\":\"\",\"runtimeVersion\":\"\",\"extraArgs\":{}},\"nodeSettings\":{\"dockerGraphPath\":\"/data/bcs/service/docker\",\"mountTarget\":\"/data\",\"unSchedulable\":1,\"labels\":{},\"extraArgs\":{}},\"status\":\"RUNNING\",\"updater\":\"boweiguan\",\"networkType\":\"overlay\",\"autoGenerateMasterNodes\":false,\"template\":[],\"extraInfo\":{},\"moduleID\":\"\",\"extraClusterID\":\"\",\"isCommonCluster\":false,\"description\":\"\",\"clusterCategory\":\"builder\",\"is_shared\":false},{\"clusterID\":\"BCS-K8S-15091\",\"clusterName\":\"先不要使用此集群\",\"federationClusterID\":\"\",\"provider\":\"bluekingCloud\",\"region\":\"sz\",\"vpcID\":\"\",\"projectID\":\"b37778ec757544868a01e1f01f07037f\",\"businessID\":\"100148\",\"environment\":\"stag\",\"engineType\":\"k8s\",\"isExclusive\":true,\"clusterType\":\"single\",\"labels\":{},\"creator\":\"bellkeyang\",\"createTime\":\"2019-10-15T15:48:19+08:00\",\"updateTime\":\"2022-03-02T12:05:57+08:00\",\"bcsAddons\":{},\"extraAddons\":{},\"systemID\":\"\",\"manageType\":\"INDEPENDENT_CLUSTER\",\"master\":{\"\":{\"nodeID\":\"\",\"innerIP\":\"\",\"instanceType\":\"\",\"CPU\":0,\"mem\":0,\"GPU\":0,\"status\":\"RUNNING\",\"zoneID\":\"\",\"nodeGroupID\":\"\",\"clusterID\":\"\",\"VPC\":\"\",\"region\":\"\",\"passwd\":\"\",\"zone\":0,\"deviceID\":\"\"}},\"networkSettings\":{\"clusterIPv4CIDR\":\"\",\"serviceIPv4CIDR\":\"\",\"maxNodePodNum\":0,\"maxServiceNum\":0,\"enableVPCCni\":false,\"eniSubnetIDs\":[],\"subnetSource\":null,\"isStaticIpMode\":false,\"claimExpiredSeconds\":0,\"multiClusterCIDR\":[],\"cidrStep\":0},\"clusterBasicSettings\":{\"OS\":\"Tencent tlinux release 2.2 (Final)\",\"version\":\"1.16\",\"clusterTags\":{},\"versionName\":\"BCS-K8S-15091\"},\"clusterAdvanceSettings\":{\"IPVS\":false,\"containerRuntime\":\"\",\"runtimeVersion\":\"\",\"extraArgs\":{}},\"nodeSettings\":{\"dockerGraphPath\":\"/data/bcs/service/docker\",\"mountTarget\":\"/data\",\"unSchedulable\":1,\"labels\":{},\"extraArgs\":{}},\"status\":\"RUNNING\",\"updater\":\"bellkeyang\",\"networkType\":\"overlay\",\"autoGenerateMasterNodes\":false,\"template\":[],\"extraInfo\":{},\"moduleID\":\"\",\"extraClusterID\":\"\",\"isCommonCluster\":false,\"description\":\"自动化测试信息\",\"clusterCategory\":\"builder\",\"is_shared\":false}],\"clusterPerm\":{\"BCS-K8S-15171\":{\"policy\":{\"cluster_create\":false,\"cluster_delete\":false,\"cluster_manage\":false,\"cluster_use\":true,\"cluster_view\":true,\"create\":false,\"delete\":false,\"deploy\":false,\"download\":false,\"edit\":false,\"use\":true,\"view\":true}}},\"clusterExtraInfo\":{\"BCS-K8S-15091\":{\"canDeleted\":false,\"providerType\":\"k8s\"},\"BCS-MESOS-10039\":{\"canDeleted\":true,\"providerType\":\"k8s\"}},\"permissions\":{},\"web_annotations\":null}")
	clusterListRsp := &cm.ListClusterResp{}
	json.Unmarshal(rawClusters, clusterListRsp)
	mockCm.On("GetCluster", &cm.GetClusterReq{}).
		Return(&cm.GetClusterResp{}, nil)
	mockCm.On("ListCluster", &cm.ListClusterReq{}).
		Return(clusterListRsp, nil)
	return mockCm
}

// CreateCluster mock cm create cluster
func (m *MockCm) CreateCluster(ctx context.Context, in *cm.CreateClusterReq,
	opts ...grpc.CallOption) (*cm.CreateClusterResp, error) {
	return nil, nil
}

// RetryCreateClusterTask mock cm
func (m *MockCm) RetryCreateClusterTask(ctx context.Context, in *cm.RetryCreateClusterReq,
	opts ...grpc.CallOption) (*cm.RetryCreateClusterResp, error) {
	return nil, nil
}

// ImportCluster mock cm
func (m *MockCm) ImportCluster(ctx context.Context, in *cm.ImportClusterReq,
	opts ...grpc.CallOption) (*cm.ImportClusterResp, error) {
	return nil, nil
}

// UpdateCluster mock cm
func (m *MockCm) UpdateCluster(ctx context.Context, in *cm.UpdateClusterReq,
	opts ...grpc.CallOption) (*cm.UpdateClusterResp, error) {
	return nil, nil
}

// AddNodesToCluster mock cm
func (m *MockCm) AddNodesToCluster(ctx context.Context, in *cm.AddNodesRequest,
	opts ...grpc.CallOption) (*cm.AddNodesResponse, error) {
	return nil, nil
}

// DeleteNodesFromCluster mock cm
func (m *MockCm) DeleteNodesFromCluster(ctx context.Context, in *cm.DeleteNodesRequest,
	opts ...grpc.CallOption) (*cm.DeleteNodesResponse, error) {
	return nil, nil
}

// ListNodesInCluster mock cm
func (m *MockCm) ListNodesInCluster(ctx context.Context, in *cm.ListNodesInClusterRequest,
	opts ...grpc.CallOption) (*cm.ListNodesInClusterResponse, error) {
	return nil, nil
}

// DeleteCluster mock cm
func (m *MockCm) DeleteCluster(ctx context.Context, in *cm.DeleteClusterReq,
	opts ...grpc.CallOption) (*cm.DeleteClusterResp, error) {
	return nil, nil
}

// GetCluster mock cm
func (m *MockCm) GetCluster(ctx context.Context, in *cm.GetClusterReq,
	opts ...grpc.CallOption) (*cm.GetClusterResp, error) {
	args := m.Called(in)
	return args.Get(0).(*cm.GetClusterResp), args.Error(1)
}

// ListCluster mock cm
func (m *MockCm) ListCluster(ctx context.Context, in *cm.ListClusterReq,
	opts ...grpc.CallOption) (*cm.ListClusterResp, error) {
	args := m.Called(in)
	return args.Get(0).(*cm.ListClusterResp), args.Error(1)
}

// * node management

// GetNode mock cm
func (m *MockCm) GetNode(ctx context.Context, in *cm.GetNodeRequest,
	opts ...grpc.CallOption) (*cm.GetNodeResponse, error) {
	return nil, nil
}

// UpdateNode mock cm
func (m *MockCm) UpdateNode(ctx context.Context, in *cm.UpdateNodeRequest,
	opts ...grpc.CallOption) (*cm.UpdateNodeResponse, error) {
	return nil, nil
}

// CheckNodeInCluster mock cm
func (m *MockCm) CheckNodeInCluster(ctx context.Context, in *cm.CheckNodesRequest,
	opts ...grpc.CallOption) (*cm.CheckNodesResponse, error) {
	return nil, nil
}

// * cluster credential management

// GetClusterCredential mock cm
func (m *MockCm) GetClusterCredential(ctx context.Context, in *cm.GetClusterCredentialReq,
	opts ...grpc.CallOption) (*cm.GetClusterCredentialResp, error) {
	return nil, nil
}

// UpdateClusterCredential mock cm
func (m *MockCm) UpdateClusterCredential(ctx context.Context, in *cm.UpdateClusterCredentialReq,
	opts ...grpc.CallOption) (*cm.UpdateClusterCredentialResp, error) {
	return nil, nil
}

// DeleteClusterCredential mock cm
func (m *MockCm) DeleteClusterCredential(ctx context.Context, in *cm.DeleteClusterCredentialReq,
	opts ...grpc.CallOption) (*cm.DeleteClusterCredentialResp, error) {
	return nil, nil
}

// ListClusterCredential mock cm
func (m *MockCm) ListClusterCredential(ctx context.Context, in *cm.ListClusterCredentialReq,
	opts ...grpc.CallOption) (*cm.ListClusterCredentialResp, error) {
	return nil, nil
}

// * federation cluster management

// InitFederationCluster mock cm
func (m *MockCm) InitFederationCluster(ctx context.Context, in *cm.InitFederationClusterReq,
	opts ...grpc.CallOption) (*cm.InitFederationClusterResp, error) {
	return nil, nil
}

// AddFederatedCluster mock cm
func (m *MockCm) AddFederatedCluster(ctx context.Context, in *cm.AddFederatedClusterReq,
	opts ...grpc.CallOption) (*cm.AddFederatedClusterResp, error) {
	return nil, nil
}

// * namespace management *

// CreateNamespace mock cm
func (m *MockCm) CreateNamespace(ctx context.Context, in *cm.CreateNamespaceReq,
	opts ...grpc.CallOption) (*cm.CreateNamespaceResp, error) {
	return nil, nil
}

// UpdateNamespace mock cm
func (m *MockCm) UpdateNamespace(ctx context.Context, in *cm.UpdateNamespaceReq,
	opts ...grpc.CallOption) (*cm.UpdateNamespaceResp, error) {
	return nil, nil
}

// DeleteNamespace mock cm
func (m *MockCm) DeleteNamespace(ctx context.Context, in *cm.DeleteNamespaceReq,
	opts ...grpc.CallOption) (*cm.DeleteNamespaceResp, error) {
	return nil, nil
}

// GetNamespace mock cm
func (m *MockCm) GetNamespace(ctx context.Context, in *cm.GetNamespaceReq,
	opts ...grpc.CallOption) (*cm.GetNamespaceResp, error) {
	return nil, nil
}

// ListNamespace mock cm
func (m *MockCm) ListNamespace(ctx context.Context, in *cm.ListNamespaceReq,
	opts ...grpc.CallOption) (*cm.ListNamespaceResp, error) {
	return nil, nil
}

// * NamespaceQuota management *

// CreateNamespaceQuota mock cm
func (m *MockCm) CreateNamespaceQuota(ctx context.Context, in *cm.CreateNamespaceQuotaReq,
	opts ...grpc.CallOption) (*cm.CreateNamespaceQuotaResp, error) {
	return nil, nil
}

// UpdateNamespaceQuota mock cm
func (m *MockCm) UpdateNamespaceQuota(ctx context.Context, in *cm.UpdateNamespaceQuotaReq,
	opts ...grpc.CallOption) (*cm.UpdateNamespaceQuotaResp, error) {
	return nil, nil
}

// DeleteNamespaceQuota mock cm
func (m *MockCm) DeleteNamespaceQuota(ctx context.Context, in *cm.DeleteNamespaceQuotaReq,
	opts ...grpc.CallOption) (*cm.DeleteNamespaceQuotaResp, error) {
	return nil, nil
}

// GetNamespaceQuota mock cm
func (m *MockCm) GetNamespaceQuota(ctx context.Context, in *cm.GetNamespaceQuotaReq,
	opts ...grpc.CallOption) (*cm.GetNamespaceQuotaResp, error) {
	return nil, nil
}

// ListNamespaceQuota mock cm
func (m *MockCm) ListNamespaceQuota(ctx context.Context, in *cm.ListNamespaceQuotaReq,
	opts ...grpc.CallOption) (*cm.ListNamespaceQuotaResp, error) {
	return nil, nil
}

// CreateNamespaceWithQuota mock cm
func (m *MockCm) CreateNamespaceWithQuota(ctx context.Context, in *cm.CreateNamespaceWithQuotaReq,
	opts ...grpc.CallOption) (*cm.CreateNamespaceWithQuotaResp, error) {
	return nil, nil
}

// * project information management *

// CreateProject mock cm
func (m *MockCm) CreateProject(ctx context.Context, in *cm.CreateProjectRequest,
	opts ...grpc.CallOption) (*cm.CreateProjectResponse, error) {
	return nil, nil
}

// UpdateProject mock cm
func (m *MockCm) UpdateProject(ctx context.Context, in *cm.UpdateProjectRequest,
	opts ...grpc.CallOption) (*cm.UpdateProjectResponse, error) {
	return nil, nil
}

// DeleteProject mock cm
func (m *MockCm) DeleteProject(ctx context.Context, in *cm.DeleteProjectRequest,
	opts ...grpc.CallOption) (*cm.DeleteProjectResponse, error) {
	return nil, nil
}

// GetProject mock cm
func (m *MockCm) GetProject(ctx context.Context, in *cm.GetProjectRequest,
	opts ...grpc.CallOption) (*cm.GetProjectResponse, error) {
	return nil, nil
}

// ListProject mock cm
func (m *MockCm) ListProject(ctx context.Context, in *cm.ListProjectRequest,
	opts ...grpc.CallOption) (*cm.ListProjectResponse, error) {
	return nil, nil
}

// * Cloud information management *

// CreateCloud mock cm
func (m *MockCm) CreateCloud(ctx context.Context, in *cm.CreateCloudRequest,
	opts ...grpc.CallOption) (*cm.CreateCloudResponse, error) {
	return nil, nil
}

// UpdateCloud mock cm
func (m *MockCm) UpdateCloud(ctx context.Context, in *cm.UpdateCloudRequest,
	opts ...grpc.CallOption) (*cm.UpdateCloudResponse, error) {
	return nil, nil
}

// DeleteCloud mock cm
func (m *MockCm) DeleteCloud(ctx context.Context, in *cm.DeleteCloudRequest,
	opts ...grpc.CallOption) (*cm.DeleteCloudResponse, error) {
	return nil, nil
}

// GetCloud mock cm
func (m *MockCm) GetCloud(ctx context.Context, in *cm.GetCloudRequest,
	opts ...grpc.CallOption) (*cm.GetCloudResponse, error) {
	return nil, nil
}

// ListCloud mock cm
func (m *MockCm) ListCloud(ctx context.Context, in *cm.ListCloudRequest,
	opts ...grpc.CallOption) (*cm.ListCloudResponse, error) {
	return nil, nil
}

// * Cloud VPC information management *

// CreateCloudVPC mock cm
func (m *MockCm) CreateCloudVPC(ctx context.Context, in *cm.CreateCloudVPCRequest,
	opts ...grpc.CallOption) (*cm.CreateCloudVPCResponse, error) {
	return nil, nil
}

// UpdateCloudVPC mock cm
func (m *MockCm) UpdateCloudVPC(ctx context.Context, in *cm.UpdateCloudVPCRequest,
	opts ...grpc.CallOption) (*cm.UpdateCloudVPCResponse, error) {
	return nil, nil
}

// DeleteCloudVPC mock cm
func (m *MockCm) DeleteCloudVPC(ctx context.Context, in *cm.DeleteCloudVPCRequest,
	opts ...grpc.CallOption) (*cm.DeleteCloudVPCResponse, error) {
	return nil, nil
}

// ListCloudVPC mock cm
func (m *MockCm) ListCloudVPC(ctx context.Context, in *cm.ListCloudVPCRequest,
	opts ...grpc.CallOption) (*cm.ListCloudVPCResponse, error) {
	return nil, nil
}

// ListCloudRegions mock cm
func (m *MockCm) ListCloudRegions(ctx context.Context, in *cm.ListCloudRegionsRequest,
	opts ...grpc.CallOption) (*cm.ListCloudRegionsResponse, error) {
	return nil, nil
}

// GetVPCCidr mock cm
func (m *MockCm) GetVPCCidr(ctx context.Context, in *cm.GetVPCCidrRequest,
	opts ...grpc.CallOption) (*cm.GetVPCCidrResponse, error) {
	return nil, nil
}

// * NodeGroup information management *

// CreateNodeGroup mock cm
func (m *MockCm) CreateNodeGroup(ctx context.Context, in *cm.CreateNodeGroupRequest,
	opts ...grpc.CallOption) (*cm.CreateNodeGroupResponse, error) {
	return nil, nil
}

// UpdateNodeGroup mock cm
func (m *MockCm) UpdateNodeGroup(ctx context.Context, in *cm.UpdateNodeGroupRequest,
	opts ...grpc.CallOption) (*cm.UpdateNodeGroupResponse, error) {
	return nil, nil
}

// DeleteNodeGroup mock cm
func (m *MockCm) DeleteNodeGroup(ctx context.Context, in *cm.DeleteNodeGroupRequest,
	opts ...grpc.CallOption) (*cm.DeleteNodeGroupResponse, error) {
	return nil, nil
}

// GetNodeGroup mock cm
func (m *MockCm) GetNodeGroup(ctx context.Context, in *cm.GetNodeGroupRequest,
	opts ...grpc.CallOption) (*cm.GetNodeGroupResponse, error) {
	return nil, nil
}

// ListNodeGroup mock cm
func (m *MockCm) ListNodeGroup(ctx context.Context, in *cm.ListNodeGroupRequest,
	opts ...grpc.CallOption) (*cm.ListNodeGroupResponse, error) {
	return nil, nil
}

// MoveNodesToGroup mock cm
func (m *MockCm) MoveNodesToGroup(ctx context.Context, in *cm.MoveNodesToGroupRequest,
	opts ...grpc.CallOption) (*cm.MoveNodesToGroupResponse, error) {
	return nil, nil
}

// RemoveNodesFromGroup mock cm
func (m *MockCm) RemoveNodesFromGroup(ctx context.Context, in *cm.RemoveNodesFromGroupRequest,
	opts ...grpc.CallOption) (*cm.RemoveNodesFromGroupResponse, error) {
	return nil, nil
}

// CleanNodesInGroup mock cm
func (m *MockCm) CleanNodesInGroup(ctx context.Context, in *cm.CleanNodesInGroupRequest,
	opts ...grpc.CallOption) (*cm.CleanNodesInGroupResponse, error) {
	return nil, nil
}

// ListNodesInGroup mock cm
func (m *MockCm) ListNodesInGroup(ctx context.Context, in *cm.GetNodeGroupRequest,
	opts ...grpc.CallOption) (*cm.ListNodesInGroupResponse, error) {
	return nil, nil
}

// UpdateGroupDesiredNode mock cm
func (m *MockCm) UpdateGroupDesiredNode(ctx context.Context, in *cm.UpdateGroupDesiredNodeRequest,
	opts ...grpc.CallOption) (*cm.UpdateGroupDesiredNodeResponse, error) {
	return nil, nil
}

// UpdateGroupDesiredSize mock cm
func (m *MockCm) UpdateGroupDesiredSize(ctx context.Context, in *cm.UpdateGroupDesiredSizeRequest,
	opts ...grpc.CallOption) (*cm.UpdateGroupDesiredSizeResponse, error) {
	return nil, nil
}

// * Task information management *

// CreateTask mock cm
func (m *MockCm) CreateTask(ctx context.Context, in *cm.CreateTaskRequest,
	opts ...grpc.CallOption) (*cm.CreateTaskResponse, error) {
	return nil, nil
}

// RetryTask mock cm
func (m *MockCm) RetryTask(ctx context.Context, in *cm.RetryTaskRequest,
	opts ...grpc.CallOption) (*cm.RetryTaskResponse, error) {
	return nil, nil
}

// UpdateTask mock cm
func (m *MockCm) UpdateTask(ctx context.Context, in *cm.UpdateTaskRequest,
	opts ...grpc.CallOption) (*cm.UpdateTaskResponse, error) {
	return nil, nil
}

// DeleteTask mock cm
func (m *MockCm) DeleteTask(ctx context.Context, in *cm.DeleteTaskRequest,
	opts ...grpc.CallOption) (*cm.DeleteTaskResponse, error) {
	return nil, nil
}

// GetTask mock cm
func (m *MockCm) GetTask(ctx context.Context, in *cm.GetTaskRequest,
	opts ...grpc.CallOption) (*cm.GetTaskResponse, error) {
	return nil, nil
}

// ListTask mock cm
func (m *MockCm) ListTask(ctx context.Context, in *cm.ListTaskRequest,
	opts ...grpc.CallOption) (*cm.ListTaskResponse, error) {
	return nil, nil
}

// * ClusterAutoScalingOption information management *

// CreateAutoScalingOption mock cm
func (m *MockCm) CreateAutoScalingOption(ctx context.Context, in *cm.CreateAutoScalingOptionRequest,
	opts ...grpc.CallOption) (*cm.CreateAutoScalingOptionResponse, error) {
	return nil, nil
}

// UpdateAutoScalingOption mock cm
func (m *MockCm) UpdateAutoScalingOption(ctx context.Context, in *cm.UpdateAutoScalingOptionRequest,
	opts ...grpc.CallOption) (*cm.UpdateAutoScalingOptionResponse, error) {
	return nil, nil
}

// DeleteAutoScalingOption mock cm
func (m *MockCm) DeleteAutoScalingOption(ctx context.Context, in *cm.DeleteAutoScalingOptionRequest,
	opts ...grpc.CallOption) (*cm.DeleteAutoScalingOptionResponse, error) {
	return nil, nil
}

// GetAutoScalingOption mock cm
func (m *MockCm) GetAutoScalingOption(ctx context.Context, in *cm.GetAutoScalingOptionRequest,
	opts ...grpc.CallOption) (*cm.GetAutoScalingOptionResponse, error) {
	return nil, nil
}

// ListAutoScalingOption mock cm
func (m *MockCm) ListAutoScalingOption(ctx context.Context, in *cm.ListAutoScalingOptionRequest,
	opts ...grpc.CallOption) (*cm.ListAutoScalingOptionResponse, error) {
	return nil, nil
}
