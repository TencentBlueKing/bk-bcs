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
 *
 */

package api

import (
	"encoding/base64"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
)

// NewTkeClient init Tke client
func NewTkeClient(opt *cloudprovider.CommonOption) (*TkeClient, error) {
	if opt == nil || len(opt.Key) == 0 || len(opt.Secret) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}
	credential := common.NewCredential(opt.Key, opt.Secret)
	cpf := profile.NewClientProfile()
	if opt.CommonConf.CloudInternalEnable {
		cpf.HttpProfile.Endpoint = opt.CommonConf.CloudDomain
	}

	cli, err := tke.NewClient(credential, opt.Region, cpf)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}

	return &TkeClient{tke: cli}, nil
}

// TkeClient xxx
type TkeClient struct {
	tke *tke.Client
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

	resp, err := cli.tke.CreateCluster(req)
	if err != nil {
		blog.Errorf("CreateTKECluster client CreateCluster[%s] failed: %v", createReq.ClusterBasic.ClusterName, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("CreateTKECluster client CreateCluster[%s] but lost response information", createReq.ClusterBasic.ClusterName)
		return nil, cloudprovider.ErrCloudLostResponse
	}

	//check response data
	blog.Infof("RequestId[%s] tke client DescribeClusters[%s] response successful",
		response.RequestId, createReq.ClusterBasic.ClusterName)

	if *response.ClusterId == "" {
		return nil, fmt.Errorf("CreateTKECluster client CreateCluster[%s] failed: clusterID is empty", createReq.ClusterBasic.ClusterName)
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
	//check response data
	blog.Infof("RequestId[%s] tke client DescribeClusters[%s] response num %d",
		response.RequestId, clusterID, *response.TotalCount,
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
		initOffset      int64
		clusterList  = make([]*tke.Cluster, 0)
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
		//check response
		response := resp.Response
		if response == nil {
			return nil, cloudprovider.ErrCloudLostResponse
		}

		clusterList = append(clusterList, response.Clusters...)
		clusterListLen = len(response.Clusters)
		initOffset = initOffset + 100
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
		deleteMode = Terminate
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
	//check response data
	blog.Infof("RequestId[%s] tke client DescribeClusters[%s] response successful",
		response.RequestId, clusterID)

	return nil
}

// TKE node relative interface

// QueryTkeClusterAllInstances query all cluster instances
func (cli *TkeClient) QueryTkeClusterAllInstances(clusterID string) ([]*InstanceInfo, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	if len(clusterID) == 0 {
		return nil, fmt.Errorf("GetTKECluster failed: clusterID is empty")
	}

	var (
		initOffset      int64
		instanceIDList  = make([]*InstanceInfo, 0)
		instanceListLen = 100
	)

	for {
		if instanceListLen != 100 {
			break
		}
		req := tke.NewDescribeClusterInstancesRequest()
		req.ClusterId = common.StringPtr(clusterID)
		req.InstanceRole = common.StringPtr(ALL.String())
		req.Offset = common.Int64Ptr(initOffset)
		req.Limit = common.Int64Ptr(int64(100))

		resp, err := cli.tke.DescribeClusterInstances(req)
		if err != nil {
			return nil, err
		}
		//check response
		response := resp.Response
		if response == nil {
			return nil, cloudprovider.ErrCloudLostResponse
		}

		for _, instance := range response.InstanceSet {
			instanceIDList = append(instanceIDList, &InstanceInfo{
				InstanceID:    *instance.InstanceId,
				InstanceIP:    *instance.LanIP,
				InstanceRole:  *instance.InstanceRole,
				InstanceState: *instance.InstanceState,
			})
		}

		instanceListLen = len(response.InstanceSet)
		initOffset = initOffset + 100
	}

	return instanceIDList, nil
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
	req.InstanceIds = common.StringPtrs(clusterReq.InstanceIDs)

	req.InstanceRole = common.StringPtr(WORKER.String())
	if len(clusterReq.InstanceRole) > 0 {
		req.InstanceRole = common.StringPtr(clusterReq.InstanceRole.String())
	}
	req.Limit = common.Int64Ptr(limit)

	resp, err := cli.tke.DescribeClusterInstances(req)
	if err != nil {
		blog.Errorf("QueryTkeClusterInstances client DescribeClusterInstances[%s] failed: %v", clusterReq.ClusterID, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("QueryTkeClusterInstances client DescribeClusterInstances[%s] but lost response information", clusterReq.ClusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] tke client DescribeClusterInstances[%s] response num %d",
		response.RequestId, clusterReq.ClusterID, *response.TotalCount,
	)

	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		return nil, fmt.Errorf("QueryTkeClusterInstances client DescribeClusterInstances[%s] response data empty", clusterReq.ClusterID)
	}

	return response.InstanceSet, nil
}

// DeleteTkeClusterInstance delete tke cluster instance, no limit
func (cli *TkeClient) DeleteTkeClusterInstance(deleteReq *DeleteInstancesRequest) (*DeleteInstancesResult, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	err := deleteReq.validateDeleteClusterInstanceRequest()
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
	req.InstanceIds = common.StringPtrs(deleteReq.Instances)

	resp, err := cli.tke.DeleteClusterInstances(req)
	if err != nil {
		blog.Errorf("DeleteTkeClusterInstance client DeleteClusterInstances[%s] failed: %v", deleteReq.ClusterID, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("DeleteTkeClusterInstance client DeleteCluster[%s] but lost response information", deleteReq.ClusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] tke client DeleteCluster[%s] response successful",
		response.RequestId, deleteReq.ClusterID)

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
	resp, err := cli.tke.AddExistedInstances(req)
	if err != nil {
		blog.Errorf("AddExistedInstancesToCluster client AddExistedInstances[%s] failed: %v", addReq.ClusterID, err)
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("AddExistedInstancesToCluster client AddExistedInstances[%s] but lost response information", addReq.ClusterID)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] tke client AddExistedInstances[%s] response successful",
		response.RequestId, addReq.ClusterID)

	result := &AddExistedInstanceRsp{
		FailedInstanceIDs: common.StringValues(response.FailedInstanceIds),
		FailedReasons:     common.StringValues(response.FailedReasons),

		SuccessInstanceIDs: common.StringValues(response.SuccInstanceIds),
		TimeoutInstanceIDs: common.StringValues(response.TimeoutInstanceIds),
	}

	return result, nil
}

// TKE network relative interface

// EnableTKEVpcCniMode enable vpc-cni plugin
func (cli *TkeClient) EnableTKEVpcCniMode(input *EnableVpcCniInput) error {
	req := tke.NewEnableVpcCniNetworkTypeRequest()
	req.ClusterId = &input.TkeClusterID
	req.VpcCniType = &input.VpcCniType
	req.EnableStaticIp = &input.EnableStaticIP
	req.Subnets = common.StringPtrs(input.SubnetsIDs)
	req.ExpiredSeconds = common.Uint64Ptr(uint64(input.ExpiredSeconds))

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

	resp, err := cli.tke.DescribeEnableVpcCniProgress(req)
	if err != nil {
		if resp != nil && resp.Response != nil {
			return nil, fmt.Errorf("query vpc-cni progress failed: %v, request id: %v", err, resp.Response.RequestId)
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
	req.SubnetIds = common.StringPtrs(input.SubnetIDs)

	resp, err := cli.tke.AddVpcCniSubnets(req)
	if err != nil {
		if resp != nil && resp.Response != nil {
			return fmt.Errorf("add vpc-cni subnets failed: %v, request id: %v", err, resp.Response.RequestId)
		}
		return fmt.Errorf("add vpc-cni subnets failed: %v", err)
	}

	return nil
}

// CloseVpcCniMode close extra vpc-cni mode
func (cli *TkeClient) CloseVpcCniMode(clusterID string) error {
	req := tke.NewDisableVpcCniNetworkTypeRequest()
	req.ClusterId = &clusterID

	resp, err := cli.tke.DisableVpcCniNetworkType(req)
	if err != nil {
		if resp != nil && resp.Response != nil {
			return fmt.Errorf("close vpc-cni mode failed: %v, request id: %v", err, resp.Response.RequestId)
		}
		return fmt.Errorf("close vpc-cni mode failed: %v", err)
	}

	return nil
}

// TKE other relative interface

// GetTKEClusterVersions get tke cluster versions
func (cli *TkeClient) GetTKEClusterVersions() ([]*Versions, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := tke.NewDescribeVersionsRequest()
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
	//check response data
	blog.Infof("RequestId[%s] tke client DescribeVersions response successful", response.RequestId)

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
func (cli *TkeClient) GetTKEClusterKubeConfig(clusterID string, isExtranet bool) (string, error){
	if cli == nil {
		return "", cloudprovider.ErrServerIsNil
	}

	req := tke.NewDescribeClusterKubeconfigRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.IsExtranet = common.BoolPtr(isExtranet)

	resp, err := cli.tke.DescribeClusterKubeconfig(req)
	if err != nil {
		blog.Errorf("GetTKEClusterKubeConfig client DescribeClusterKubeconfig failed: %v", err)
		return "", err
	}

	if resp.Response == nil {
		blog.Errorf("GetTKEClusterKubeConfig client DescribeClusterKubeconfig but lost response information")
		return "", cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] tke client DescribeClusterKubeconfig response successful", resp.Response.RequestId)
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

	resp, err := cli.tke.DescribeClusterEndpointStatus(req)
	if err != nil {
		blog.Errorf("GetClusterEndpointStatus client DescribeClusterEndpointStatus failed: %v", err)
		return "", err
	}

	if resp.Response == nil {
		blog.Errorf("GetClusterEndpointStatus client DescribeClusterEndpointStatus but lost response information")
		return "", cloudprovider.ErrCloudLostResponse
	}
	//check response data
	blog.Infof("RequestId[%s] tke client DescribeClusterEndpointStatus response successful", *resp.Response.RequestId)

	if resp.Response.Status == nil {
		blog.Errorf("GetClusterEndpointStatus client DescribeClusterEndpointStatus failed: %v", "status nil")
		return "", cloudprovider.ErrCloudLostResponse
	}

	return EndpointStatus(*resp.Response.Status), nil
}

// CreateClusterEndpoint 创建集群访问端口,默认开启公网访问
func (cli *TkeClient) CreateClusterEndpoint(clusterID string) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}
	if clusterID == "" {
		return fmt.Errorf("clusterID is null")
	}

	req := tke.NewCreateClusterEndpointRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.IsExtranet = common.BoolPtr(true)

	resp, err := cli.tke.CreateClusterEndpoint(req)
	if err != nil {
		blog.Errorf("client CreateClusterEndpoint failed: %v", err)
		return err
	}

	//check response data
	blog.Infof("RequestId[%s] tke client CreateClusterEndpoint response successful", *resp.Response.RequestId)

	return nil
}

// DeleteClusterEndpoint 删除集群访问端口, 默认开启公网访问
func (cli *TkeClient) DeleteClusterEndpoint(clusterID string) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}
	if clusterID == "" {
		return fmt.Errorf("clusterID is null")
	}

	req := tke.NewDeleteClusterEndpointRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.IsExtranet = common.BoolPtr(true)

	resp, err := cli.tke.DeleteClusterEndpoint(req)
	if err != nil {
		blog.Errorf("client DeleteClusterEndpoint failed: %v", err)
		return err
	}

	//check response data
	blog.Infof("RequestId[%s] tke client DeleteClusterEndpoint response successful", *resp.Response.RequestId)

	return nil
}

// GetTKEClusterImages get tke cluster images info
func (cli *TkeClient) GetTKEClusterImages() ([]*Images, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := tke.NewDescribeImagesRequest()
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
	//check response data
	blog.Infof("RequestId[%s] tke client DescribeImages response successful", response.RequestId)

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
