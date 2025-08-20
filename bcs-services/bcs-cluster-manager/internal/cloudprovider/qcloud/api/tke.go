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

package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// NewTkeClient init Tke client
func NewTkeClient(opt *cloudprovider.CommonOption) (*TkeClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	// qcloud account
	credential := common.NewCredential(opt.Account.SecretID, opt.Account.SecretKey)
	cpf := profile.NewClientProfile()
	if opt.CommonConf.CloudInternalEnable {
		cpf.HttpProfile.Endpoint = opt.CommonConf.CloudDomain
	}

	// tke client
	cli, err := tke.NewClient(credential, opt.Region, cpf)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}
	commonCli, err := NewClient(credential, opt.Region, cpf)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}

	return &TkeClient{tke: cli, tkeCommon: commonCli}, nil
}

// TkeClient xxx
type TkeClient struct {
	tke       *tke.Client
	tkeCommon *Client
}

// TKE cluster relative interface

// CreateTKECluster create tke cluster
func (cli *TkeClient) CreateTKECluster(createReq *CreateClusterRequest) (*CreateClusterResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req, err := generateClusterRequestInfo(createReq)
	if err != nil {
		return nil, err
	}

	tags, _ := json.Marshal(createReq.ClusterBasic.TagSpecification)
	blog.Infof("CreateTKECluster tags %s", tags)

	resp, err := cli.tke.CreateCluster(req)
	if err != nil {
		blog.Errorf("CreateTKECluster client CreateCluster[%s] failed: %v", createReq.ClusterBasic.ClusterName, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("CreateTKECluster client CreateCluster[%s] but lost response information",
			createReq.ClusterBasic.ClusterName)
		return nil, cloudprovider.ErrCloudLostResponse
	}

	// check response data
	blog.Infof("RequestId[%s] tke client DescribeClusters[%s] response successful",
		*response.RequestId, createReq.ClusterBasic.ClusterName)

	if *response.ClusterId == "" {
		return nil, fmt.Errorf("CreateTKECluster client CreateCluster[%s] failed: clusterID is empty",
			createReq.ClusterBasic.ClusterName)
	}

	return &CreateClusterResponse{ClusterID: *response.ClusterId}, nil
}

// GetTKECluster get tke cluster info
func (cli *TkeClient) GetTKECluster(clusterID string) (*tke.Cluster, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	if len(clusterID) == 0 {
		return nil, fmt.Errorf("GetTKECluster failed: clusterID is empty")
	}

	// create cluster request
	req := tke.NewDescribeClustersRequest()
	req.ClusterIds = append(req.ClusterIds, common.StringPtr(clusterID))

	resp, err := cli.tke.DescribeClusters(req)
	if err != nil {
		blog.Errorf("GetTKECluster client DescribeClusters[%s] failed: %v", clusterID, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("GetTKECluster client DescribeClusters[%s] but lost response information", clusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeClusters[%s] response num %d",
		*response.RequestId, clusterID, *response.TotalCount,
	)

	if *response.TotalCount == 0 || len(response.Clusters) == 0 {
		return nil, fmt.Errorf("GetTKECluster client DescribeClusters[%s] response data empty", clusterID)
	}

	return response.Clusters[0], nil
}

// ListTKECluster get tke cluster list, region parameter init tke client
func (cli *TkeClient) ListTKECluster() ([]*tke.Cluster, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	var (
		initOffset     int64
		clusterList    = make([]*tke.Cluster, 0)
		clusterListLen = 100
	)

	for {
		if clusterListLen != 100 {
			break
		}
		req := tke.NewDescribeClustersRequest()
		req.Offset = common.Int64Ptr(initOffset)
		req.Limit = common.Int64Ptr(int64(100))

		resp, err := cli.tke.DescribeClusters(req)
		if err != nil {
			return nil, err
		}
		// check response
		response := resp.Response
		if response == nil {
			return nil, cloudprovider.ErrCloudLostResponse
		}
		blog.Info("ListTKECluster %v DescribeClusters success", *response.RequestId, *response.TotalCount)

		clusterList = append(clusterList, response.Clusters...)
		clusterListLen = len(response.Clusters)
		initOffset += 100
	}

	return clusterList, nil
}

// DeleteTKECluster delete cluster bu clusterID, deleteMode: terminate retain
func (cli *TkeClient) DeleteTKECluster(clusterID string, deleteMode DeleteMode) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}

	if len(clusterID) == 0 {
		return fmt.Errorf("DeleteTKECluster failed: clusterID is empty")
	}

	if deleteMode != Terminate && deleteMode != Retain && deleteMode != "" {
		return fmt.Errorf("DeleteTKECluster[%s] invalid deleteMode[%s]", clusterID, deleteMode)
	}

	// default deleteMode
	if deleteMode == "" {
		deleteMode = Retain
	}

	// create cluster request
	req := tke.NewDeleteClusterRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.InstanceDeleteMode = common.StringPtr(deleteMode.String())

	resp, err := cli.tke.DeleteCluster(req)
	if err != nil {
		blog.Errorf("DeleteTKECluster client DeleteCluster[%s] failed: %v", clusterID, err)
		return err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("DeleteTKECluster client DeleteCluster[%s] but lost response information", clusterID)
		return cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeClusters[%s] response successful",
		*response.RequestId, clusterID)

	return nil
}

// TKE node relative interface

// QueryTkeClusterAllInstances query all cluster instances
func (cli *TkeClient) QueryTkeClusterAllInstances(ctx context.Context, clusterID string,
	filter QueryFilter) ([]*InstanceInfo, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	if len(clusterID) == 0 {
		return nil, fmt.Errorf("GetTKECluster failed: clusterID is empty")
	}

	traceID := utils.GetTraceIDFromContext(ctx)

	var (
		initOffset      int64
		instanceList    = make([]*InstanceInfo, 0)
		instanceListLen = 100
		instanceIDList  []string
	)

	for {
		if instanceListLen != 100 {
			break
		}
		req := tke.NewDescribeClusterInstancesRequest()
		req.ClusterId = common.StringPtr(clusterID)
		req.InstanceRole = common.StringPtr(ALL.String())
		if filter != nil {
			req.Filters = filter.BuildFilters()
		}
		req.Offset = common.Int64Ptr(initOffset)
		req.Limit = common.Int64Ptr(int64(100))

		resp, err := cli.tke.DescribeClusterInstances(req)
		if err != nil {
			return nil, err
		}
		// check response
		response := resp.Response
		if response == nil {
			return nil, cloudprovider.ErrCloudLostResponse
		}
		blog.Infof("traceID[%s] RequestId[%s] tke client DescribeClusterInstances[%s:%d] response successful",
			traceID, *response.RequestId, clusterID, *response.TotalCount)

		for _, instance := range response.InstanceSet {
			instanceList = append(instanceList, &InstanceInfo{
				InstanceID:         utils.StringPtrToString(instance.InstanceId),
				InstanceIP:         utils.StringPtrToString(instance.LanIP),
				InstanceRole:       utils.StringPtrToString(instance.InstanceRole),
				InstanceState:      utils.StringPtrToString(instance.InstanceState),
				NodePoolId:         utils.StringPtrToString(instance.NodePoolId),
				AutoscalingGroupId: utils.StringPtrToString(instance.AutoscalingGroupId),
			})

			instanceIDList = append(instanceIDList, utils.StringPtrToString(instance.InstanceId))
		}

		instanceListLen = len(response.InstanceSet)
		initOffset += 100
	}

	blog.Infof("traceID[%s] QueryTkeClusterAllInstances[%+v]", traceID, instanceIDList)
	return instanceList, nil
}

// QueryTkeClusterInstances query cluster specified instances, attention limit max 100.
// if limit > 100, need to split chunks
func (cli *TkeClient) QueryTkeClusterInstances(clusterReq *DescribeClusterInstances) ([]*tke.Instance, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	if len(clusterReq.ClusterID) == 0 || len(clusterReq.InstanceIDs) == 0 {
		return nil, fmt.Errorf("GetTKECluster failed: clusterID or InstanceIDs is empty")
	}

	req := tke.NewDescribeClusterInstancesRequest()
	req.ClusterId = common.StringPtr(clusterReq.ClusterID)

	if len(clusterReq.InstanceIDs) > 0 {
		req.InstanceIds = common.StringPtrs(clusterReq.InstanceIDs)
	}

	req.InstanceRole = common.StringPtr(WORKER.String())
	if len(clusterReq.InstanceRole) > 0 {
		req.InstanceRole = common.StringPtr(clusterReq.InstanceRole.String())
	}
	req.Limit = common.Int64Ptr(limit)

	// tke DescribeClusterInstances
	resp, err := cli.tke.DescribeClusterInstances(req)
	if err != nil {
		blog.Errorf("QueryTkeClusterInstances client DescribeClusterInstances[%s] failed: %v", clusterReq.ClusterID, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("QueryTkeClusterInstances client DescribeClusterInstances[%s] but lost response information",
			clusterReq.ClusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeClusterInstances[%s] response num %d",
		*response.RequestId, clusterReq.ClusterID, *response.TotalCount,
	)

	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		return nil, fmt.Errorf("QueryTkeClusterInstances client DescribeClusterInstances[%s] response data empty",
			clusterReq.ClusterID)
	}

	return response.InstanceSet, nil
}

// DeleteTkeClusterInstance delete tke cluster instance
func (cli *TkeClient) DeleteTkeClusterInstance(deleteReq *DeleteInstancesRequest) (*DeleteInstancesResult, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	err := deleteReq.validate()
	if err != nil {
		return nil, err
	}

	// default deleteMode
	if deleteReq.DeleteMode == "" {
		deleteReq.DeleteMode = Retain
	}

	// delete tke cluster instance request
	req := tke.NewDeleteClusterInstancesRequest()
	req.ClusterId = common.StringPtr(deleteReq.ClusterID)
	req.InstanceDeleteMode = common.StringPtr(deleteReq.DeleteMode.String())
	req.ForceDelete = common.BoolPtr(deleteReq.ForceDelete)

	if len(deleteReq.Instances) > 0 {
		req.InstanceIds = common.StringPtrs(deleteReq.Instances)
	}

	// tke DeleteClusterInstances
	resp, err := cli.tke.DeleteClusterInstances(req)
	if err != nil {
		blog.Errorf("DeleteTkeClusterInstance client DeleteClusterInstances[%s] failed: %v", deleteReq.ClusterID, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("DeleteTkeClusterInstance client DeleteTkeClusterInstance[%s] but lost response information",
			deleteReq.ClusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DeleteTkeClusterInstance[%s] response successful",
		*response.RequestId, deleteReq.ClusterID)

	result := &DeleteInstancesResult{
		Success:  common.StringValues(response.SuccInstanceIds),
		Failure:  common.StringValues(response.FailedInstanceIds),
		NotFound: common.StringValues(response.NotFoundInstanceIds),
	}

	return result, nil
}

// AddExistedInstancesToCluster add node to cluster
func (cli *TkeClient) AddExistedInstancesToCluster(addReq *AddExistedInstanceReq) (*AddExistedInstanceRsp, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	err := addReq.validate()
	if err != nil {
		return nil, err
	}

	req := generateAddExistedInstancesReq(addReq)

	blog.Infof("AddExistedInstancesToCluster request %+v", req)

	// tke AddExistedInstances
	resp, err := cli.tke.AddExistedInstances(req)
	if err != nil {
		blog.Errorf("AddExistedInstancesToCluster client AddExistedInstances[%s] failed: %v", addReq.ClusterID, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("AddExistedInstancesToCluster client AddExistedInstances[%s] but lost response information",
			addReq.ClusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}

	blog.Infof("RequestId[%s] tke client AddExistedInstancesToCluster[%s] response successful",
		*response.RequestId, addReq.ClusterID)

	result := &AddExistedInstanceRsp{
		FailedInstanceIDs: common.StringValues(response.FailedInstanceIds),
		FailedReasons:     common.StringValues(response.FailedReasons),

		SuccessInstanceIDs: common.StringValues(response.SuccInstanceIds),
		TimeoutInstanceIDs: common.StringValues(response.TimeoutInstanceIds),
	}

	return result, nil
}

// DescribeInstanceCreateProgress describe instance create progress
func (cli *TkeClient) DescribeInstanceCreateProgress(clusterId, instanceId string) (string, error) {
	req := NewDescribeInstanceCreateProgressRequest()
	req.ClusterId = common.StringPtr(clusterId)
	req.InstanceId = common.StringPtr(instanceId)

	resp, err := cli.tkeCommon.DescribeInstanceCreateProgress(req)
	if err != nil {
		blog.Errorf("DescribeInstanceCreateProgress[%s:%s] failed: %v", clusterId, instanceId, err)
		return "", err
	}

	if resp == nil || resp.Response == nil {
		return "", fmt.Errorf("response empty")
	}

	var (
		failedReason string
	)
	for i := range resp.Response.Progress {
		if resp.Response.Progress[i] != nil && resp.Response.Progress[i].Status != nil &&
			*resp.Response.Progress[i].Status == FailedInstanceTke.String() {
			failedReason = *resp.Response.Progress[i].Message
			break
		}
	}

	return failedReason, nil
}

// TKE network relative interface

// EnableTKEVpcCniMode enable vpc-cni plugin: tke-route-eni开启的是策略路由模式，tke-direct-eni开启的是独立网卡模式
func (cli *TkeClient) EnableTKEVpcCniMode(input *EnableVpcCniInput) error {
	req := tke.NewEnableVpcCniNetworkTypeRequest()
	req.ClusterId = &input.TkeClusterID
	req.VpcCniType = &input.VpcCniType
	// 是否开启固定IP模式
	req.EnableStaticIp = &input.EnableStaticIp
	// 容器子网
	if len(input.SubnetsIDs) > 0 {
		req.Subnets = common.StringPtrs(input.SubnetsIDs)
	}
	// 固定IP模式下，Pod销毁后退还IP的时间，传参必须大于300；不传默认IP永不销毁。
	if input.ExpiredSeconds >= 0 {
		if input.ExpiredSeconds < 300 {
			input.ExpiredSeconds = 300
		}
		req.ExpiredSeconds = common.Uint64Ptr(uint64(input.ExpiredSeconds))
	}

	// tke EnableVpcCniNetworkType
	resp, err := cli.tke.EnableVpcCniNetworkType(req)
	if err != nil {
		if resp != nil && resp.Response != nil {
			return fmt.Errorf("enable vpc-cni failed: %v, request id: %v", err, resp.Response.RequestId)
		}
		return fmt.Errorf("enable vpc-cni failed: %v", err)
	}

	blog.Infof("EnableTKEVpcCniMode successful, requestID[%s]", *resp.Response.RequestId)
	fmt.Println(*resp.Response.RequestId)
	return nil
}

// GetEnableVpcCniProgress enable vpc-cni progress
func (cli *TkeClient) GetEnableVpcCniProgress(clusterID string) (*GetEnableVpcCniProgressOutput, error) {
	req := tke.NewDescribeEnableVpcCniProgressRequest()
	req.ClusterId = &clusterID

	// tke DescribeEnableVpcCniProgress
	resp, err := cli.tke.DescribeEnableVpcCniProgress(req)
	if err != nil {
		if resp != nil && resp.Response != nil {
			return nil, fmt.Errorf("query vpc-cni progress failed: %v, request id: %v", err, *resp.Response.RequestId)
		}
		return nil, fmt.Errorf("query vpc-cni progress failed: %v", err)
	}

	// Status: Running/Succeed/Failed, return Message when task Failed
	return &GetEnableVpcCniProgressOutput{
		Status:    *resp.Response.Status,
		Message:   *resp.Response.ErrorMessage,
		RequestID: *resp.Response.RequestId,
	}, nil
}

// AddVpcCniSubnets add vpc-cni mode subnet
func (cli *TkeClient) AddVpcCniSubnets(input *AddVpcCniSubnetsInput) error {
	req := tke.NewAddVpcCniSubnetsRequest()
	req.ClusterId = &input.ClusterID
	req.VpcId = &input.VpcID
	if len(input.SubnetIDs) > 0 {
		req.SubnetIds = common.StringPtrs(input.SubnetIDs)
	}

	// tke AddVpcCniSubnets
	resp, err := cli.tke.AddVpcCniSubnets(req)
	if err != nil {
		if resp != nil && resp.Response != nil {
			return fmt.Errorf("add vpc-cni subnets failed: %v, request id: %v", err, *resp.Response.RequestId)
		}
		return fmt.Errorf("add vpc-cni subnets failed: %v", err)
	}

	return nil
}

// CloseVpcCniMode close extra vpc-cni mode
func (cli *TkeClient) CloseVpcCniMode(clusterID string) error {
	req := tke.NewDisableVpcCniNetworkTypeRequest()
	req.ClusterId = &clusterID

	// tke DisableVpcCniNetworkType
	resp, err := cli.tke.DisableVpcCniNetworkType(req)
	if err != nil {
		if resp != nil && resp.Response != nil {
			return fmt.Errorf("close vpc-cni mode failed: %v, request id: %v", err, *resp.Response.RequestId)
		}
		return fmt.Errorf("close vpc-cni mode failed: %v", err)
	}

	return nil
}

// AddClusterCIDR cluster add cidr
func (cli *TkeClient) AddClusterCIDR(clusterId string, cidrs []string, ignore bool) error {
	req := tke.NewAddClusterCIDRRequest()
	req.ClusterId = common.StringPtr(clusterId)
	if len(cidrs) > 0 {
		req.ClusterCIDRs = common.StringPtrs(cidrs)
	}
	req.IgnoreClusterCIDRConflict = common.BoolPtr(ignore)

	resp, err := cli.tke.AddClusterCIDR(req)
	if err != nil {
		if resp != nil && resp.Response != nil {
			return fmt.Errorf("add cluster[%s] cidr failed: %v, request id: %v", clusterId, err, *resp.Response.RequestId)
		}
		return fmt.Errorf("add cluster[%s] cidr failed: %v", clusterId, err)
	}

	blog.Infof("cluster[%s] add cidrs[%v] success", clusterId, cidrs)

	return nil
}

// DescribeVpcCniPodLimits 查询机型可支持的最大VPC-CNI模式Pod数量
func (cli *TkeClient) DescribeVpcCniPodLimits(zone string, instanceType string) (*DescribeVpcCniPodLimitsOut, error) {
	req := tke.NewDescribeVpcCniPodLimitsRequest()
	req.Zone = common.StringPtr(zone)
	req.InstanceType = common.StringPtr(instanceType)

	resp, err := cli.tke.DescribeVpcCniPodLimits(req)
	if err != nil {
		if resp != nil && resp.Response != nil {
			return nil, fmt.Errorf("DescribeVpcCniPodLimits[%s:%s] failed: %v, requestID %v", zone, instanceType,
				err, resp.Response.RequestId)
		}

		return nil, fmt.Errorf("DescribeVpcCniPodLimits[%s:%s] failed: %v", zone, instanceType, err)
	}

	if *resp.Response.TotalCount == 0 || len(resp.Response.PodLimitsInstanceSet) == 0 {
		return nil, fmt.Errorf("DescribeVpcCniPodLimits[%s:%s] empty", zone, instanceType)
	}

	blog.Infof("DescribeVpcCniPodLimits[%s:%s] successful", zone, instanceType)

	return &DescribeVpcCniPodLimitsOut{
		Zone:         zone,
		InstanceType: instanceType,
		Limits: PodLimits{
			RouterEniNonStaticIP: *resp.Response.PodLimitsInstanceSet[0].PodLimits.TKERouteENINonStaticIP,
			RouterEniStaticIP:    *resp.Response.PodLimitsInstanceSet[0].PodLimits.TKERouteENIStaticIP,
			directEni:            *resp.Response.PodLimitsInstanceSet[0].PodLimits.TKEDirectENI,
		},
	}, nil
}

// TKE other relative interface

// GetTKEClusterVersions get tke cluster versions
func (cli *TkeClient) GetTKEClusterVersions() ([]*Versions, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := tke.NewDescribeVersionsRequest()

	// tke DescribeVersions
	resp, err := cli.tke.DescribeVersions(req)
	if err != nil {
		blog.Errorf("GetTKEClusterVersions client DescribeVersions failed: %v", err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("GetTKEClusterVersions client DescribeVersions but lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeVersions response successful", *response.RequestId)

	if *response.TotalCount == 0 || len(response.VersionInstanceSet) == 0 {
		return nil, fmt.Errorf("GetTKEClusterVersions client DescribeVersions response data empty")
	}

	versions := make([]*Versions, 0)
	for i := range response.VersionInstanceSet {
		versions = append(versions, &Versions{
			Name:    *response.VersionInstanceSet[i].Name,
			Version: *response.VersionInstanceSet[i].Version,
		})
	}

	return versions, nil
}

// GetTKEClusterKubeConfig get clusterKubeConfig: isExtranet internal/external kubeConfig
func (cli *TkeClient) GetTKEClusterKubeConfig(clusterID string, isExtranet bool) (string, error) {
	if cli == nil {
		return "", cloudprovider.ErrServerIsNil
	}

	if clusterID == "" {
		return "", fmt.Errorf("clusterID is null")
	}

	req := tke.NewDescribeClusterKubeconfigRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.IsExtranet = common.BoolPtr(isExtranet)

	// tke DescribeClusterKubeconfig
	resp, err := cli.tke.DescribeClusterKubeconfig(req)
	if err != nil {
		blog.Errorf("GetTKEClusterKubeConfig client DescribeClusterKubeconfig failed: %v", err)
		return "", err
	}

	if resp.Response == nil {
		blog.Errorf("GetTKEClusterKubeConfig client DescribeClusterKubeconfig but lost response information")
		return "", cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeClusterKubeconfig response successful", *resp.Response.RequestId)
	baseRet := base64.StdEncoding.EncodeToString([]byte(*resp.Response.Kubeconfig))

	return baseRet, nil
}

// GetClusterEndpointStatus 查询集群访问端口状态
func (cli *TkeClient) GetClusterEndpointStatus(clusterID string, isExtranet bool) (EndpointStatus, error) {
	if cli == nil {
		return "", cloudprovider.ErrServerIsNil
	}
	if clusterID == "" {
		return "", fmt.Errorf("clusterID is null")
	}

	req := tke.NewDescribeClusterEndpointStatusRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.IsExtranet = common.BoolPtr(isExtranet)

	// tke DescribeClusterEndpointStatus
	resp, err := cli.tke.DescribeClusterEndpointStatus(req)
	if err != nil {
		blog.Errorf("GetClusterEndpointStatus client DescribeClusterEndpointStatus failed: %v", err)
		return "", err
	}

	if resp.Response == nil {
		blog.Errorf("GetClusterEndpointStatus client DescribeClusterEndpointStatus but lost response information")
		return "", cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeClusterEndpointStatus response successful", *resp.Response.RequestId)

	if resp.Response.Status == nil {
		blog.Errorf("GetClusterEndpointStatus client DescribeClusterEndpointStatus failed: %v", "status nil")
		return "", cloudprovider.ErrCloudLostResponse
	}

	return EndpointStatus(*resp.Response.Status), nil
}

// CreateClusterEndpoint 创建集群访问端口,默认开启公网访问
func (cli *TkeClient) CreateClusterEndpoint(clusterID string, config ClusterEndpointConfig) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}
	if clusterID == "" {
		return fmt.Errorf("clusterID is null")
	}

	// 开启内网/外网访问的端口
	req := config.getEndpointConfig(clusterID)

	// tke CreateClusterEndpoint
	resp, err := cli.tke.CreateClusterEndpoint(req)
	if err != nil {
		blog.Errorf("client CreateClusterEndpoint failed: %v", err)
		return err
	}

	// check response data
	blog.Infof("RequestId[%s] tke client CreateClusterEndpoint response successful", *resp.Response.RequestId)

	return nil
}

// DeleteClusterEndpoint 删除集群访问端口,默认开启公网访问
func (cli *TkeClient) DeleteClusterEndpoint(clusterID string, isExtranet bool) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}
	if clusterID == "" {
		return fmt.Errorf("clusterID is null")
	}

	req := tke.NewDeleteClusterEndpointRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.IsExtranet = common.BoolPtr(isExtranet)

	// tke DeleteClusterEndpoint
	resp, err := cli.tke.DeleteClusterEndpoint(req)
	if err != nil {
		blog.Errorf("client DeleteClusterEndpoint failed: %v", err)
		return err
	}

	// check response data
	blog.Infof("RequestId[%s] tke client DeleteClusterEndpoint response successful", *resp.Response.RequestId)

	return nil
}

// GetTKEClusterImages get tke cluster images info
func (cli *TkeClient) GetTKEClusterImages() ([]*Images, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := tke.NewDescribeImagesRequest()

	// tke DescribeImages
	resp, err := cli.tke.DescribeImages(req)
	if err != nil {
		blog.Errorf("GetTKEClusterImages client DescribeImages failed: %v", err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("GetTKEClusterImages client DescribeImages but lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeImages response successful", *response.RequestId)

	if *response.TotalCount == 0 || len(response.ImageInstanceSet) == 0 {
		return nil, fmt.Errorf("GetTKEClusterImages client DescribeImages response data empty")
	}

	images := make([]*Images, 0)
	for i := range response.ImageInstanceSet {
		images = append(images, &Images{
			OsName:  *response.ImageInstanceSet[i].OsName,
			ImageID: *response.ImageInstanceSet[i].ImageId,
		})
	}

	return images, nil
}

// CreateClusterNodePool create cluster node pool, return cluster node pool id
func (cli *TkeClient) CreateClusterNodePool(nodePool *CreateNodePoolInput) (string, error) {
	blog.Infof("CreateClusterNodePool input: %", utils.ToJSONString(nodePool))
	if *nodePool.LaunchConfigurePara.InternetAccessible.InternetChargeType == InternetChargeTypeBandwidthPrepaid {
		nodePool.LaunchConfigurePara.InternetAccessible.InternetChargeType = common.StringPtr(InternetChargeTypeBandwidthPostpaidByHour) // nolint
	}
	req := generateClusterNodePool(nodePool)
	if req == nil {
		blog.Errorf("CreateClusterNodePool failed: generateClusterNodePool failed, CreateClusterNodePoolRequest is nil")
		return "", fmt.Errorf("CreateClusterNodePool failed: CreateClusterNodePoolRequest is nil")
	}
	if len(*req.AutoScalingGroupPara) == 0 {
		blog.Errorf("CreateClusterNodePool failed: AutoScalingGroupPara is empty")
		return "", fmt.Errorf("CreateClusterNodePool failed: AutoScalingGroupPara is empty")
	}
	if len(*req.LaunchConfigurePara) == 0 {
		blog.Errorf("CreateClusterNodePool failed: LaunchConfigurePara is empty")
		return "", fmt.Errorf("CreateClusterNodePool failed: LaunchConfigurePara is empty")
	}

	// tke CreateClusterNodePool
	resp, err := cli.tke.CreateClusterNodePool(req)

	if err != nil {
		blog.Errorf("CreateClusterNodePool client CreateClusterNodePool failed: %v", err)
		return "", err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("CreateClusterNodePool client CreateClusterNodePool but lost response information")
		return "", cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("RequestId[%s] tke client CreateClusterNodePool response successful", *resp.Response.RequestId)
	return *resp.Response.NodePoolId, nil
}

// DescribeClusterNodePools describe cluster node pools
func (cli *TkeClient) DescribeClusterNodePools(clusterID string, filters []*Filter) ([]*tke.NodePool, int, error) {
	blog.Infof("DescribeClusterNodePools input: clusterID[%s], filters[%s]", clusterID, utils.ToJSONString(filters))
	req := tke.NewDescribeClusterNodePoolsRequest()
	req.ClusterId = common.StringPtr(clusterID)
	if len(filters) > 0 {
		req.Filters = make([]*tke.Filter, 0)
		for _, v := range filters {
			req.Filters = append(req.Filters, &tke.Filter{
				Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
		}
	}

	// tke DescribeClusterNodePools
	resp, err := cli.tke.DescribeClusterNodePools(req)
	if err != nil {
		blog.Errorf("DescribeClusterNodePools client DescribeClusterNodePools failed: %v", err)
		return nil, 0, err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("DescribeClusterNodePools client DescribeClusterNodePools but lost response information")
		return nil, 0, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("RequestId[%s] tke client DescribeClusterNodePools response successful", *resp.Response.RequestId)
	return resp.Response.NodePoolSet, int(*resp.Response.TotalCount), nil
}

// DescribeClusterNodePoolDetail describe cluster node pool detail
func (cli *TkeClient) DescribeClusterNodePoolDetail(clusterID string, nodePoolID string) (*tke.NodePool, error) {
	blog.Infof("DescribeClusterNodePoolDetail, clusterID: %s, nodePoolID: %s", clusterID, nodePoolID)
	req := tke.NewDescribeClusterNodePoolDetailRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.NodePoolId = common.StringPtr(nodePoolID)

	// tke DescribeClusterNodePoolDetail
	resp, err := cli.tke.DescribeClusterNodePoolDetail(req)
	if err != nil {
		blog.Errorf("DescribeClusterNodePoolDetail failed: %v", err)
		return nil, err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("DescribeClusterNodePoolDetail resp is nil")
		return nil, cloudprovider.ErrCloudLostResponse
	}

	blog.Infof("RequestId[%s] tke client DescribeClusterNodePoolDetail response successful", *resp.Response.RequestId)
	return resp.Response.NodePool, nil
}

// ModifyClusterNodePool modify cluster node pool
func (cli *TkeClient) ModifyClusterNodePool(req *tke.ModifyClusterNodePoolRequest) error {
	blog.Infof("ModifyClusterNodePool request: %s", utils.ToJSONString(req))

	// tke ModifyClusterNodePool
	resp, err := cli.tke.ModifyClusterNodePool(req)
	if err != nil {
		blog.Errorf("ModifyClusterNodePool failed: %v", err)
		return err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("ModifyClusterNodePool resp is nil")
		return fmt.Errorf("ModifyClusterNodePool resp is nil")
	}
	blog.Infof("RequestId[%s] tke client ModifyClusterNodePool response successful", *resp.Response.RequestId)
	return nil
}

// DeleteClusterNodePool delete cluster node pool
func (cli *TkeClient) DeleteClusterNodePool(clusterID string, nodePoolIDs []string, keepInstance bool) error {
	blog.Infof("DeleteClusterNodePool input: clusterID: %s, nodePoolIDs: %s, keepInstance: %t", clusterID,
		utils.ToJSONString(nodePoolIDs), keepInstance)
	req := tke.NewDeleteClusterNodePoolRequest()
	req.ClusterId = common.StringPtr(clusterID)
	if len(nodePoolIDs) > 0 {
		req.NodePoolIds = common.StringPtrs(nodePoolIDs)
	}
	req.KeepInstance = common.BoolPtr(keepInstance)

	// tke DeleteClusterNodePool
	resp, err := cli.tke.DeleteClusterNodePool(req)
	if err != nil {
		blog.Errorf("DeleteClusterNodePool failed: %v", err)
		return err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("DeleteClusterNodePool resp is nil")
		return fmt.Errorf("DeleteClusterNodePool resp is nil")
	}
	blog.Infof("RequestId[%s] tke client DeleteClusterNodePool response successful", *resp.Response.RequestId)
	return nil
}

// ModifyNodePoolCapacity modify node pool desired capacity about asg
func (cli *TkeClient) ModifyNodePoolCapacity(clusterID string, nodePoolID string,
	desiredCapacity int64) error {
	blog.Infof("ModifyNodePoolDesiredCapacityAboutAsg input: clusterID: %s, nodePoolID: %s, desiredCapacity: %d",
		clusterID, nodePoolID, desiredCapacity)
	req := tke.NewModifyNodePoolDesiredCapacityAboutAsgRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.NodePoolId = common.StringPtr(nodePoolID)
	req.DesiredCapacity = common.Int64Ptr(desiredCapacity)

	// tke ModifyNodePoolDesiredCapacityAboutAsg
	resp, err := cli.tke.ModifyNodePoolDesiredCapacityAboutAsg(req)
	if err != nil {
		blog.Errorf("ModifyNodePoolDesiredCapacityAboutAsg failed: %v", err)
		return err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("ModifyNodePoolDesiredCapacityAboutAsg resp is nil")
		return fmt.Errorf("ModifyNodePoolDesiredCapacityAboutAsg resp is nil")
	}
	blog.Infof("RequestId[%s] tke client ModifyNodePoolDesiredCapacityAboutAsg response successful",
		*resp.Response.RequestId)
	return nil
}

// ModifyNodePoolInstanceTypes modify node pool instance types
func (cli *TkeClient) ModifyNodePoolInstanceTypes(clusterID string, nodePoolID string, instanceTypes []string) error {
	blog.Infof("ModifyNodePoolInstanceTypes input: clusterID: %s, nodePoolID: %s, instanceTypes: %s",
		clusterID, nodePoolID, utils.ToJSONString(instanceTypes))
	req := tke.NewModifyNodePoolInstanceTypesRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.NodePoolId = common.StringPtr(nodePoolID)
	if len(instanceTypes) > 0 {
		req.InstanceTypes = common.StringPtrs(instanceTypes)
	}

	// tke ModifyNodePoolInstanceTypes
	resp, err := cli.tke.ModifyNodePoolInstanceTypes(req)
	if err != nil {
		blog.Errorf("ModifyNodePoolInstanceTypes failed: %v", err)
		return err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("ModifyNodePoolInstanceTypes resp is nil")
		return fmt.Errorf("ModifyNodePoolInstanceTypes resp is nil")
	}
	blog.Infof("RequestId[%s] tke client ModifyNodePoolInstanceTypes response successful", *resp.Response.RequestId)
	return nil
}

// RemoveNodeFromNodePool remove node from node pool
func (cli *TkeClient) RemoveNodeFromNodePool(clusterID string, nodePoolID string, nodeIDs []string) error {
	blog.Infof("RemoveNodeFromNodePool input: clusterID: %s, nodePoolID: %s, nodeIDs: %s", clusterID, nodePoolID,
		utils.ToJSONString(nodeIDs))
	req := tke.NewRemoveNodeFromNodePoolRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.NodePoolId = common.StringPtr(nodePoolID)
	allNodes := utils.SplitStringsChunks(nodeIDs, limit)
	for _, v := range allNodes {
		if len(v) > 0 {
			req.InstanceIds = common.StringPtrs(v)
			// tke RemoveNodeFromNodePool
			resp, err := cli.tke.RemoveNodeFromNodePool(req)
			if err != nil {
				blog.Errorf("RemoveNodeFromNodePool failed: %v", err)
				return err
			}
			if resp == nil || resp.Response == nil {
				blog.Errorf("RemoveNodeFromNodePool resp is nil")
				return fmt.Errorf("RemoveNodeFromNodePool resp is nil")
			}
			blog.Infof("RequestId[%s] tke client RemoveNodeFromNodePool response successful", *resp.Response.RequestId)
		}
	}
	return nil
}

// AddNodeToNodePool add node to node pool
func (cli *TkeClient) AddNodeToNodePool(clusterID string, nodePoolID string, nodeIDs []string) error {
	blog.Infof("AddNodeToNodePool input: clusterID: %s, nodePoolID: %s, nodeIDs: %s", clusterID, nodePoolID,
		utils.ToJSONString(nodeIDs))
	req := tke.NewAddNodeToNodePoolRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.NodePoolId = common.StringPtr(nodePoolID)
	allNodes := utils.SplitStringsChunks(nodeIDs, limit)
	for _, v := range allNodes {
		if len(v) > 0 {
			req.InstanceIds = common.StringPtrs(v)
			// tke AddNodeToNodePool
			resp, err := cli.tke.AddNodeToNodePool(req)
			if err != nil {
				blog.Errorf("AddNodeToNodePool failed: %v", err)
				return err
			}
			if resp == nil || resp.Response == nil {
				blog.Errorf("AddNodeToNodePool resp is nil")
				return fmt.Errorf("AddNodeToNodePool resp is nil")
			}
			blog.Infof("RequestId[%s] tke client AddNodeToNodePool response successful", *resp.Response.RequestId)
		}
	}
	return nil
}

// GetNodeGroupInstances describe nodegroup instances
func (cli *TkeClient) GetNodeGroupInstances(clusterID, nodeGroupID string) ([]*tke.Instance, error) {
	blog.Infof("GetNodeGroupInstances input: clusterID: %s, nodeGroupID", clusterID, nodeGroupID)
	req := tke.NewDescribeClusterInstancesRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.Limit = common.Int64Ptr(limit)
	req.Filters = make([]*tke.Filter, 0)
	req.Filters = append(req.Filters, &tke.Filter{
		Name: common.StringPtr("nodepool-id"), Values: common.StringPtrs([]string{nodeGroupID})})
	req.Filters = append(req.Filters, &tke.Filter{
		Name: common.StringPtr("nodepool-instance-type"), Values: common.StringPtrs([]string{"ALL"})})
	got, total := 0, 0
	first := true
	ins := make([]*tke.Instance, 0)
	for got < total || first {
		first = false
		req.Offset = common.Int64Ptr(int64(got))
		// tke DescribeClusterInstances
		resp, err := cli.tke.DescribeClusterInstances(req)
		if err != nil {
			blog.Errorf("DescribeClusterInstances failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeClusterInstances resp is nil")
			return nil, fmt.Errorf("DescribeClusterInstances resp is nil")
		}
		blog.Infof("DescribeClusterInstances success, requestID: %s", resp.Response.RequestId)
		ins = append(ins, resp.Response.InstanceSet...)
		got += len(resp.Response.InstanceSet)
		total = int(*resp.Response.TotalCount)
	}
	return ins, nil
}

// TKE集群支持第三方节点功能

// EnableExternalNodeSupport 开启关闭集群第三方节点池特性
func (cli *TkeClient) EnableExternalNodeSupport(clusterID string, config EnableExternalNodeConfig) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return fmt.Errorf("EnableExternalNodeSupport failed: clusterID is empty")
	}

	err := config.validate()
	if err != nil {
		return err
	}

	req := NewEnableExternalNodeSupportRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.ClusterExternalConfig = &ClusterExternalConfig{
		NetworkType: common.StringPtr(config.NetworkType),
		SubnetId:    common.StringPtr(config.SubnetId),
		ClusterCIDR: common.StringPtr(config.ClusterCIDR),
		Enabled:     common.BoolPtr(config.Enabled),
	}

	resp, err := cli.tkeCommon.EnableExternalNodeSupport(req)
	if err != nil {
		blog.Errorf("EnableExternalNodeSupport[%s] failed: %v", clusterID, err)
		return err
	}
	// check response data
	blog.Infof("RequestId[%s] tke client EnableExternalNodeSupport[%s] success",
		*resp.Response.RequestId, clusterID)

	return nil
}

// DescribeExternalNodeScript 获取第三方节点添加脚本
func (cli *TkeClient) DescribeExternalNodeScript(clusterID string,
	config DescribeExternalNodeScriptConfig) (*DescribeExternalNodeScriptResponseParams, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return nil, fmt.Errorf("DescribeExternalNodeScript failed: clusterID is empty")
	}
	err := config.validate()
	if err != nil {
		return nil, err
	}

	req := NewDescribeExternalNodeScriptRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.NodePoolId = common.StringPtr(config.NodePoolId)
	if len(config.Interface) > 0 {
		req.Interface = common.StringPtr(config.Interface)
	}
	if len(config.Name) > 0 {
		req.Name = common.StringPtr(config.Name)
	}
	if config.Internal {
		req.Internal = common.BoolPtr(true)
	}

	resp, err := cli.tkeCommon.DescribeExternalNodeScript(req)
	if err != nil {
		blog.Errorf("DescribeExternalNodeScript[%s] failed: %v", clusterID, err)
		return nil, err
	}

	return resp.Response, nil
}

// DeleteExternalNode 删除第三方节点
func (cli *TkeClient) DeleteExternalNode(clusterID string, config DeleteExternalNodeConfig) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return fmt.Errorf("DeleteExternalNode failed: clusterID is empty")
	}
	err := config.validate()
	if err != nil {
		return err
	}

	req := NewDeleteExternalNodeRequest()
	req.ClusterId = common.StringPtr(clusterID)
	if len(config.Names) > 0 {
		req.Names = common.StringPtrs(config.Names)
	}
	req.Force = common.BoolPtr(config.Force)

	resp, err := cli.tkeCommon.DeleteExternalNode(req)
	if err != nil {
		blog.Errorf("DeleteExternalNode[%s] failed: %v", clusterID, err)
		return err
	}
	blog.Infof("RequestId[%s] tke client DeleteExternalNode[%s] success", *resp.Response.RequestId, clusterID)

	return nil
}

// DeleteExternalNodePool 删除第三方节点池
func (cli *TkeClient) DeleteExternalNodePool(clusterID string, config DeleteExternalNodePoolConfig) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return fmt.Errorf("DeleteExternalNodePool failed: clusterID is empty")
	}
	err := config.validate()
	if err != nil {
		return err
	}

	req := NewDeleteExternalNodePoolRequest()
	req.ClusterId = common.StringPtr(clusterID)
	if len(config.NodePoolIds) > 0 {
		req.NodePoolIds = common.StringPtrs(config.NodePoolIds)
	}
	req.Force = common.BoolPtr(config.Force)

	resp, err := cli.tkeCommon.DeleteExternalNodePool(req)
	if err != nil {
		blog.Errorf("DeleteExternalNodePool[%s] failed: %v", clusterID, err)
		return err
	}
	blog.Infof("RequestId[%s] tke client DeleteExternalNodePool[%s] success", *resp.Response.RequestId, clusterID)

	return nil
}

// DescribeExternalNodePools 查看第三方节点池列表
func (cli *TkeClient) DescribeExternalNodePools(clusterID string) ([]*ExternalNodePool, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return nil, fmt.Errorf("DescribeExternalNodePools failed: clusterID is empty")
	}

	req := NewDescribeExternalNodePoolsRequest()
	req.ClusterId = common.StringPtr(clusterID)

	resp, err := cli.tkeCommon.DescribeExternalNodePools(req)
	if err != nil {
		blog.Errorf("DescribeExternalNodePools[%s] failed: %v", clusterID, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("DescribeExternalNodePools[%s] but lost response information", clusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeExternalNodePools[%s] response num %d",
		*response.RequestId, clusterID, *response.TotalCount,
	)

	return response.NodePoolSet, nil
}

// DescribeExternalNodeSupportConfig 查看开启第三方节点池配置信息
func (cli *TkeClient) DescribeExternalNodeSupportConfig(
	clusterID string) (*DescribeExternalNodeConfigInfoResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return nil, fmt.Errorf("DescribeExternalNodeSupportConfig failed: clusterID is empty")
	}

	req := NewDescribeExternalNodeSupportConfigRequest()
	req.ClusterId = common.StringPtr(clusterID)

	resp, err := cli.tkeCommon.DescribeExternalNodeSupportConfig(req)
	if err != nil {
		blog.Errorf("DescribeExternalNodeSupportConfig[%s] failed: %v", clusterID, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("DescribeExternalNodeSupportConfig[%s] but lost response information", clusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeExternalNodeSupportConfig[%s] success",
		*response.RequestId, clusterID)

	externalNodeConfig := &DescribeExternalNodeConfigInfoResponse{
		ClusterCIDR:  *response.ClusterCIDR,
		NetworkType:  *response.NetworkType,
		SubnetId:     *response.SubnetId,
		Enabled:      *response.Enabled,
		AS:           *response.AS,
		SwitchIP:     *response.SwitchIP,
		Status:       *response.Status,
		FailedReason: *response.FailedReason,
		Master:       *response.Master,
		Proxy:        *response.Proxy,
	}

	return externalNodeConfig, nil
}

// CreateExternalNodePool 创建第三方节点池
func (cli *TkeClient) CreateExternalNodePool(clusterID string, config CreateExternalNodePoolConfig) (string, error) {
	if cli == nil {
		return "", cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return "", fmt.Errorf("CreateExternalNodePool failed: clusterID is empty")
	}

	err := config.validate()
	if err != nil {
		return "", err
	}
	req := config.transToTkeExternalNodeConfig(clusterID)

	resp, err := cli.tkeCommon.CreateExternalNodePool(req)
	if err != nil {
		blog.Errorf("CreateExternalNodePool[%s] failed: %v", clusterID, err)
		return "", err
	}
	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("CreateExternalNodePool[%s] but lost response information", clusterID)
		return "", cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client CreateExternalNodePool[%s] success",
		*response.RequestId, clusterID)

	return *response.NodePoolId, nil
}

// ModifyExternalNodePool 修改第三方节点池
func (cli *TkeClient) ModifyExternalNodePool(clusterID string, config ModifyExternalNodePoolConfig) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return fmt.Errorf("CreateExternalNodePool failed: clusterID is empty")
	}

	err := config.validate()
	if err != nil {
		return err
	}

	req := NewModifyExternalNodePoolRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.NodePoolId = common.StringPtr(config.NodePoolId)
	if len(config.Name) > 0 {
		req.Name = common.StringPtr(config.Name)
	}
	if len(config.Labels) > 0 {
		req.Labels = config.Labels
	}
	if len(config.Taints) > 0 {
		req.Taints = config.Taints
	}

	resp, err := cli.tkeCommon.ModifyExternalNodePool(req)
	if err != nil {
		blog.Errorf("ModifyExternalNodePool[%s] failed: %v", clusterID, err)
		return err
	}
	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("ModifyExternalNodePool[%s] but lost response information", clusterID)
		return cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client ModifyExternalNodePool[%s] success",
		*response.RequestId, clusterID)

	return nil
}

// DescribeExternalNode 查看第三方节点列表
func (cli *TkeClient) DescribeExternalNode(
	clusterID string, config DescribeExternalNodeConfig) ([]ExternalNodeInfo, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return nil, fmt.Errorf("DescribeExternalNode failed: clusterID is empty")
	}
	err := config.validate()
	if err != nil {
		return nil, err
	}

	req := NewDescribeExternalNodeRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.NodePoolId = common.StringPtr(config.NodePoolId)
	if len(config.Names) > 0 {
		req.Names = common.StringPtrs(config.Names)
	}

	resp, err := cli.tkeCommon.DescribeExternalNode(req)
	if err != nil {
		blog.Errorf("DescribeExternalNode[%s] failed: %v", clusterID, err)
		return nil, err
	}
	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("DescribeExternalNode[%s] but lost response information", clusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeExternalNode[%s] success: %v",
		*response.RequestId, clusterID, *response.TotalCount)

	externalNodes := make([]ExternalNodeInfo, 0)
	for i := range response.Nodes {
		externalNodes = append(externalNodes, ExternalNodeInfo{
			Name:          *response.Nodes[i].Name,
			NodePoolId:    *response.Nodes[i].NodePoolId,
			IP:            *response.Nodes[i].IP,
			Location:      *response.Nodes[i].Location,
			Status:        *response.Nodes[i].Status,
			CreatedTime:   *response.Nodes[i].CreatedTime,
			Reason:        *response.Nodes[i].Reason,
			Unschedulable: *response.Nodes[i].Unschedulable,
		})
	}

	return externalNodes, nil
}

func (cli *TkeClient) getCommonImages() ([]*OSImage, error) {
	req := NewDescribeOSImagesRequest()

	// tke DescribeOSImages
	resp, err := cli.tkeCommon.DescribeOSImages(req)
	if err != nil {
		blog.Errorf("DescribeOsImages failed: %v", err)
		return nil, err
	}
	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("DescribeOsImages but lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] tke client DescribeOsImages success: %v",
		*response.RequestId, *response.TotalCount)

	return response.OSImageSeriesSet, nil
}

// DescribeOsImages pull common images
func (cli *TkeClient) DescribeOsImages(provider, clusterID string, bcsImageNameList []string,
	opt *cloudprovider.CommonOption) ([]*OSImage, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	images := make([]*OSImage, 0)

	switch provider {
	case icommon.MarketImageProvider:
		// market image provider
		for _, v := range utils.ImageOsList {
			if provider == v.Provider {
				images = append(images, &OSImage{
					SeriesName:      &v.SeriesName,
					Alias:           &v.Alias,
					Arch:            &v.Arch,
					OsName:          &v.OsName,
					OsCustomizeType: &v.OsCustomizeType,
					Status:          &v.Status,
					ImageId:         &v.ImageID,
				})
			}
		}
		return images, nil
	case icommon.PublicImageProvider:
		// public image provider
		return cli.getCommonImages()
	case icommon.PrivateImageProvider:
		// private image provider
		cvmImages, err := getCvmImagesByImageType(provider, opt)
		if err != nil {
			return nil, fmt.Errorf("DescribeOsImages[%s] DescribeImages failed: %v", provider, err)
		}

		for i := range cvmImages {
			images = append(images, &OSImage{
				Alias:   cvmImages[i].ImageName,
				Arch:    cvmImages[i].Architecture,
				OsName:  cvmImages[i].OsName,
				Status:  cvmImages[i].ImageState,
				ImageId: cvmImages[i].ImageId,
			})
		}
		return images, nil
	case icommon.BCSImageProvider:
		// bcs image provider
		if len(bcsImageNameList) > 0 {
			for _, imageName := range bcsImageNameList {
				image, err := getCvmImageByImageName(imageName, opt)
				if err != nil {
					return nil, fmt.Errorf("qcloud getCvmImageByImageName[%s] failed: %v", imageName, err)
				}

				images = append(images, &OSImage{
					Alias:   image.ImageName,
					Arch:    image.Architecture,
					OsName:  image.OsName,
					Status:  image.ImageState,
					ImageId: image.ImageId,
				})
			}
		}
		return images, nil
	case icommon.ClusterImageProvider:
		// cluster image provider
		if clusterID != "" {
			cls, _ := cloudprovider.GetClusterByID(clusterID)
			if cls != nil {
				clusterImageOs := cls.GetClusterBasicSettings().GetOS()
				image, err := getCvmImageByImageName(clusterImageOs, opt)
				if err != nil {
					return nil, fmt.Errorf("qcloud clusterImageOs getCvmImageByImageName[%s] failed: %v", clusterImageOs, err)
				}

				images = append(images, &OSImage{
					Alias:   image.ImageName,
					Arch:    image.Architecture,
					OsName:  image.OsName,
					Status:  image.ImageState,
					ImageId: image.ImageId,
				})
			}
		}
		return images, nil
	default:
	}

	return nil, fmt.Errorf("not supported image provider[%s]", provider)
}

// AcquireClusterAdminRole 获取账号 tke:admin 权限
func (cli *TkeClient) AcquireClusterAdminRole(clusterID string) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return fmt.Errorf("AcquireClusterAdminRole failed: clusterID is empty")
	}

	req := tke.NewAcquireClusterAdminRoleRequest()
	req.ClusterId = common.StringPtr(clusterID)

	resp, err := cli.tke.AcquireClusterAdminRole(req)
	if err != nil {
		blog.Errorf("AcquireClusterAdminRole[%s] failed: %v", clusterID, err)
		return err
	}

	// check response
	if resp == nil || resp.Response == nil {
		blog.Errorf("AcquireClusterAdminRole[%s] but lost response information", clusterID)
		return cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("RequestId[%s] tke client AcquireClusterAdminRole response successful", *resp.Response.RequestId)

	return nil
}

// DescribeClusterEndpoints 获取集群访问地址
func (cli *TkeClient) DescribeClusterEndpoints(clusterID string) (*ClusterEndpointInfo, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	if len(clusterID) == 0 {
		return nil, fmt.Errorf("DescribeClusterEndpoints failed: clusterID is empty")
	}

	req := tke.NewDescribeClusterEndpointsRequest()
	req.ClusterId = common.StringPtr(clusterID)

	resp, err := cli.tke.DescribeClusterEndpoints(req)
	if err != nil {
		blog.Errorf("DescribeClusterEndpoints[%s] failed: %v", clusterID, err)
		return nil, err
	}

	// check response
	if resp == nil || resp.Response == nil {
		blog.Errorf("DescribeClusterEndpoints[%s] but lost response information", clusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("RequestId[%s] tke client DescribeClusterEndpoints response successful", *resp.Response.RequestId)

	endpointInfo := resp.Response

	return &ClusterEndpointInfo{
		CertClusterAuthority:    utils.StringPtrToString(endpointInfo.CertificationAuthority),
		ClusterExternalEndpoint: utils.StringPtrToString(endpointInfo.ClusterExternalEndpoint),
		ClusterIntranetEndpoint: utils.StringPtrToString(endpointInfo.ClusterIntranetEndpoint),
		ClusterExternalDomain:   utils.StringPtrToString(endpointInfo.ClusterExternalDomain),
		ClusterIntranetDomain:   utils.StringPtrToString(endpointInfo.ClusterIntranetDomain),
		ClusterDomain:           utils.StringPtrToString(endpointInfo.ClusterDomain),
		SecurityGroup:           utils.StringPtrToString(endpointInfo.SecurityGroup),
		ClusterExternalACL:      common.StringValues(endpointInfo.ClusterExternalACL),
	}, nil
}

/* Addon相关接口 */

// GetTkeAppChartVersionByName 获取AppChart版本
func (cli *TkeClient) GetTkeAppChartVersionByName(clusterType string, appName string) (string, error) {
	if cli == nil {
		return "", cloudprovider.ErrServerIsNil
	}
	if clusterType == "" {
		clusterType = TkeClusterType
	}

	req := tke.NewGetTkeAppChartListRequest()

	resp, err := cli.tke.GetTkeAppChartList(req)
	if err != nil {
		blog.Errorf("GetTkeAppChartList[%s:%s] failed: %v", clusterType, appName, err)
		return "", err
	}

	// check response
	if resp == nil || resp.Response == nil {
		blog.Errorf("GetTkeAppChartList[%s:%s] but lost response information", clusterType, appName)
		return "", cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("RequestId[%s] tke client GetTkeAppChartList response successful", *resp.Response.RequestId)

	appChart := resp.Response

	for i := range appChart.AppCharts {
		if *appChart.AppCharts[i].Name == appName {
			return *appChart.AppCharts[i].LatestVersion, nil
		}
	}

	return "", nil
}
