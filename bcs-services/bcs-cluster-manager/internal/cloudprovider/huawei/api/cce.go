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

// Package api xxx
package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cce "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/region"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// NewCceClient init cce client
func NewCceClient(opt *cloudprovider.CommonOption) (*CceClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	projectID, err := GetProjectIDByRegion(opt)
	if err != nil {
		return nil, err
	}
	auth, err := getProjectAuth(opt.Account.SecretID, opt.Account.SecretKey, projectID)
	if err != nil {
		return nil, err
	}

	rn, err := region.SafeValueOf(opt.Region)
	if err != nil {
		return nil, err
	}

	// 创建hc client
	hcClient, err := cce.CceClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	return &CceClient{cce.NewCceClient(hcClient)}, nil
}

// CceClient cce client
type CceClient struct {
	cce *cce.CceClient
}

// ListCceCluster get cce cluster list, region parameter init tke client
func (cli *CceClient) ListCceCluster(filter *ClusterFilterCond) (*[]model.Cluster, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ListClustersRequest{}
	var (
		detail = "true"
	)
	req.Detail = &detail

	if filter != nil {
		if len(filter.Status) > 0 {
			req.Status = GetClusterStatus(filter.Status)
		}
		if len(filter.Type) > 0 {
			req.Type = GetClusterType(filter.Type)
		}
		if len(filter.Version) > 0 {
			req.Version = &filter.Version
		}
	}

	rsp, err := cli.cce.ListClusters(&req)
	if err != nil {
		return nil, err
	}

	return rsp.Items, nil
}

// CreateCluster create cce cluster
func (cli *CceClient) CreateCluster(req *CreateClusterRequest) (*model.CreateClusterResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	return cli.cce.CreateCluster(req.Trans2CreateClusterRequest())
}

// DeleteCceCluster delete cce cluster
func (cli *CceClient) DeleteCceCluster(clusterID string) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}

	_, err := cli.cce.DeleteCluster(&model.DeleteClusterRequest{
		ClusterId: clusterID,
	})
	return err
}

// GetCceCluster get cce cluster
func (cli *CceClient) GetCceCluster(clusterID string) (*model.ShowClusterResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ShowClusterRequest{
		ClusterId: clusterID,
		Detail:    common.StringPtr("true"),
	}
	rsp, err := cli.cce.ShowCluster(&req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// ShowCceClusterEndpoints get cce cluster endpoints
func (cli *CceClient) ShowCceClusterEndpoints(clusterID string) (*model.ShowClusterEndpointsResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ShowClusterEndpointsRequest{
		ClusterId: clusterID,
	}
	rsp, err := cli.cce.ShowClusterEndpoints(&req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// CreateKubernetesClusterCert 获取集群证书
func (cli *CceClient) CreateKubernetesClusterCert(clsId string, duration int32) (
	*model.CreateKubernetesClusterCertResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	if clsId == "" {
		return nil, fmt.Errorf("clusterId empty")
	}
	// duration 集群证书有效时间，最小值为1天，最大值为5年，因此取值范围为1-1827, -1则为最大值5年
	if duration != -1 {
		if duration < 1 {
			duration = 1
		}
		if duration > 1827 {
			duration = 1827
		}
	}

	request := &model.CreateKubernetesClusterCertRequest{
		ClusterId: clsId,
	}
	request.Body = &model.CertDuration{
		Duration: duration,
	}

	response, err := cli.cce.CreateKubernetesClusterCert(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetClusterKubeConfig 获取cce集群 kubeconfig, 返回值base64编码. isExtranet: true 外部 false 内部
func (cli *CceClient) GetClusterKubeConfig(clusterId string, isExtranet bool) (string, error) {
	cert, err := cli.CreateKubernetesClusterCert(clusterId, -1)
	if err != nil {
		return "", err
	}

	return getCceClusterKubeConfig(cert, isExtranet)
}

func getCceClusterKubeConfig(cert *model.CreateKubernetesClusterCertResponse, isExtranet bool) (string, error) {
	var (
		kubeConfigType = "internalCluster"
	)
	if isExtranet {
		kubeConfigType = "externalClusterTLSVerify"
	}

	clusterName := kubeConfigType
	contextName := kubeConfigType
	authName := "user"

	cluster := types.NamedCluster{
		Name: clusterName,
		Cluster: func() types.ClusterInfo {
			clusters := *cert.Clusters

			for i := range clusters {
				if *clusters[i].Name == clusterName {
					certLocal, _ := base64.StdEncoding.DecodeString(*clusters[i].Cluster.CertificateAuthorityData)

					return types.ClusterInfo{
						Server: *clusters[i].Cluster.Server,
						InsecureSkipTLSVerify: func() bool {
							if clusters[i].Cluster != nil && clusters[i].Cluster.InsecureSkipTlsVerify != nil {
								return *clusters[i].Cluster.InsecureSkipTlsVerify
							}
							return false
						}(),
						CertificateAuthorityData: certLocal,
					}
				}
			}
			return types.ClusterInfo{}
		}(),
	}
	auth := types.NamedAuthInfo{
		Name: authName,
		AuthInfo: func() types.AuthInfo {
			users := *cert.Users
			if len(users) == 0 {
				return types.AuthInfo{}
			}

			clientCert, _ := base64.StdEncoding.DecodeString(*users[0].User.ClientCertificateData)
			clientKey, _ := base64.StdEncoding.DecodeString(*users[0].User.ClientKeyData)

			return types.AuthInfo{
				ClientCertificateData: clientCert,
				ClientKeyData:         clientKey,
			}
		}(),
	}
	context := types.NamedContext{
		Name: contextName,
		Context: types.Context{
			Cluster:  clusterName,
			AuthInfo: authName,
		},
	}

	config := types.Config{
		Kind:           *cert.Kind,
		APIVersion:     *cert.ApiVersion,
		Clusters:       []types.NamedCluster{cluster},
		AuthInfos:      []types.NamedAuthInfo{auth},
		Contexts:       []types.NamedContext{context},
		CurrentContext: contextName,
	}
	bt, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(bt), nil
}

// ShowJob 获取任务信息；创建、删除集群时，查询相应任务的进度。创建、删除节点时，查询相应任务的进度。
func (cli *CceClient) ShowJob(jobId string) (*model.ShowJobResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ShowJobRequest{
		JobId: jobId,
	}
	rsp, err := cli.cce.ShowJob(&req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// UpdateClusterEip 绑定、解绑集群公网apiserver地址
func (cli *CceClient) UpdateClusterEip(clsId string, req UpdateClusterEipRequest) (
	*model.UpdateClusterEipResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	if clsId == "" {
		return nil, fmt.Errorf("clusterId empty")
	}
	err := req.validate()
	if err != nil {
		return nil, err
	}
	reqEip := req.trans2ClusterEipRequest(clsId)

	rsp, err := cli.cce.UpdateClusterEip(reqEip)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// UpdateCluster 更新指定集群
func (cli *CceClient) UpdateCluster(clsId string, req UpdateClusterRequest) (
	*model.UpdateClusterResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	if clsId == "" {
		return nil, fmt.Errorf("clusterId empty")
	}
	updateReq := req.trans2ClusterRequest(clsId)

	rsp, err := cli.cce.UpdateCluster(updateReq)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// ListClusterNodes get cluster all nodes
func (cli *CceClient) ListClusterNodes(clusterId string) ([]model.Node, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	rsp, err := cli.cce.ListNodes(&model.ListNodesRequest{
		ClusterId: clusterId,
	})
	if err != nil {
		return nil, err
	}

	return *rsp.Items, nil
}

// ListClusterNodePoolNodes get cluster node pool all nodes
func (cli *CceClient) ListClusterNodePoolNodes(clusterId, nodePoolId string) ([]model.Node, error) {
	nodes, err := cli.ListClusterNodes(clusterId)
	if err != nil {
		return nil, err
	}

	nodePoolNodes := make([]model.Node, 0)
	for _, v := range nodes {
		if id, ok := v.Metadata.Annotations[NodePoolIdKey]; ok {
			if id == nodePoolId {
				nodePoolNodes = append(nodePoolNodes, v)
			}
		}
	}

	return nodePoolNodes, nil
}

// AddNode 纳管节点/上架已有节点
func (cli *CceClient) AddNode(req *model.AddNodeRequest) (string, error) {
	resp, err := cli.cce.AddNode(req)
	if err != nil {
		return "", err
	}

	return *resp.Jobid, nil
}

// CreateNode 创建节点
func (cli *CceClient) CreateNode() {}

// DeleteNode 删除节点
func (cli *CceClient) DeleteNode(clsId, nodeId string, isNodePool bool) (string, error) {
	if clsId == "" || nodeId == "" {
		return "", fmt.Errorf("clusterId or nodeId empty")
	}

	request := &model.DeleteNodeRequest{
		ClusterId: clsId,
		NodeId:    nodeId,
	}
	if !isNodePool {
		scaleDown := model.GetDeleteNodeRequestNodepoolScaleDownEnum().NO_SCALE_DOWN
		request.NodepoolScaleDown = &scaleDown
	}

	response, err := cli.cce.DeleteNode(request)
	if err != nil {
		return "", err
	}

	return *response.Status.JobID, nil
}

// RemoveNode 移除节点，在指定集群下移除节点
func (cli *CceClient) RemoveNode(clsId string, data RemoveNodesRequest) (string, error) {
	request, err := data.trans2RemoveNodesRequest(clsId)
	if err != nil {
		return "", err
	}

	response, err := cli.cce.RemoveNode(request)
	if err != nil {
		return "", err
	}

	return *response.Status.JobID, nil
}

// ShowNode 获取指定节点
func (cli *CceClient) ShowNode(clusterId, nodeId string) (*model.Node, error) {
	request := &model.ShowNodeRequest{
		ClusterId: clusterId,
		NodeId:    nodeId,
	}
	response, err := cli.cce.ShowNode(request)
	if err != nil {
		return nil, err
	}

	return &model.Node{
		Kind:       response.Kind,
		ApiVersion: response.ApiVersion,
		Metadata:   response.Metadata,
		Spec:       response.Spec,
		Status:     response.Status,
	}, nil
}

// CreateClusterNodePool create cluster node pool
func (cli *CceClient) CreateClusterNodePool(data *CreateNodePoolRequest) (*model.CreateNodePoolResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	rsp, err := cli.cce.CreateNodePool(data.trans2NodePoolTemplate())
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// GetClusterNodePool get cluster node pool
func (cli *CceClient) GetClusterNodePool(clusterId, nodePoolId string) (*model.NodePool, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	rsp, err := cli.cce.ShowNodePool(&model.ShowNodePoolRequest{
		ClusterId:  clusterId,
		NodepoolId: nodePoolId,
	})
	if err != nil {
		return nil, err
	}

	return &model.NodePool{
		Kind:       *rsp.Kind,
		ApiVersion: *rsp.ApiVersion,
		Metadata:   rsp.Metadata,
		Spec:       rsp.Spec,
		Status:     rsp.Status,
	}, nil
}

func (cli *CceClient) ListClusterNodeGroups(clusterId string) ([]model.NodePool, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	rsp, err := cli.cce.ListNodePools(&model.ListNodePoolsRequest{
		ClusterId: clusterId,
	})
	if err != nil {
		return nil, err
	}

	nodePools := make([]model.NodePool, 0)
	for _, item := range *rsp.Items {
		tmp := item
		nodePools = append(nodePools, model.NodePool{
			Kind:       tmp.Kind,
			ApiVersion: tmp.ApiVersion,
			Metadata:   tmp.Metadata,
			Spec:       tmp.Spec,
			Status:     tmp.Status,
		})
	}

	return nodePools, nil
}

// UpdateNodePool 全量更新接口
func (cli *CceClient) UpdateNodePool(clsId, nodePoolId string, data UpdateNodePoolRequest) (
	*model.UpdateNodePoolResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req, err := data.trans2ModelUpdateNodePoolRequest(clsId, nodePoolId)
	if err != nil {
		return nil, err
	}

	rsp, err := cli.cce.UpdateNodePool(req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// UpdateNodePoolV2 华为云原生全量更新接口
func (cli *CceClient) UpdateNodePoolV2(data *model.UpdateNodePoolRequest) (
	*model.UpdateNodePoolResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	rsp, err := cli.cce.UpdateNodePool(data)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// UpdateNodePoolDesiredNodes update nodePool desired desiredSize nodes count
func (cli *CceClient) UpdateNodePoolDesiredNodes(clusterId, nodePoolId string, desiredSize int32, inc bool) (
	*model.UpdateNodePoolResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	if clusterId == "" || nodePoolId == "" {
		return nil, fmt.Errorf("cluster or nodePool empty")
	}

	var (
		taints            []model.Taint
		k8sTags           map[string]string
		autoscalingConfig = &model.NodePoolNodeAutoscaling{}
	)

	nodePool, err := cli.GetClusterNodePool(clusterId, nodePoolId)
	if err != nil {
		return nil, fmt.Errorf("updateDesiredNodes get cluster nodePool err: %s", err)
	}

	if nodePool != nil && nodePool.Spec != nil && nodePool.Spec.NodeTemplate != nil &&
		nodePool.Spec.NodeTemplate.Taints != nil {
		taints = *nodePool.Spec.NodeTemplate.Taints
	}

	if nodePool != nil && nodePool.Spec != nil && nodePool.Spec.NodeTemplate != nil &&
		nodePool.Spec.NodeTemplate.K8sTags != nil {
		k8sTags = nodePool.Spec.NodeTemplate.K8sTags
	}

	if nodePool != nil && nodePool.Spec != nil && nodePool.Spec.Autoscaling != nil {
		autoscalingConfig = nodePool.Spec.Autoscaling
	}

	if inc {
		desiredSize += *nodePool.Spec.InitialNodeCount
	}

	req := &model.UpdateNodePoolRequest{
		ClusterId:  clusterId,
		NodepoolId: nodePoolId,
		Body: &model.NodePoolUpdate{
			Metadata: &model.NodePoolMetadataUpdate{
				Name: nodePool.Metadata.Name,
			},
			Spec: &model.NodePoolSpecUpdate{
				NodeTemplate: &model.NodeSpecUpdate{
					Taints:  taints,
					K8sTags: k8sTags,
				},
				InitialNodeCount: desiredSize,
				Autoscaling:      autoscalingConfig,
			},
		},
	}

	rsp, err := cli.cce.UpdateNodePool(req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// DeleteNodePool 删除指定的节点池
func (cli *CceClient) DeleteNodePool(clsId, nodePoolId string) error {
	if cli == nil {
		return cloudprovider.ErrServerIsNil
	}

	if clsId == "" || nodePoolId == "" {
		return fmt.Errorf("cluster or nodePool empty")
	}

	request := &model.DeleteNodePoolRequest{
		ClusterId:  clsId,
		NodepoolId: nodePoolId,
	}
	_, err := cli.cce.DeleteNodePool(request)
	if err != nil {
		return err
	}

	return nil
}

// RemoveNodePoolNodes remove node pool nodes
func (cli *CceClient) RemoveNodePoolNodes(clusterId string, nodeIds []string, password string) error {
	if len(nodeIds) == 0 {
		return nil
	}

	pw, err := Crypt(password)
	if err != nil {
		return err
	}

	nodes := make([]model.NodeItem, 0)
	for _, v := range nodeIds {
		nodes = append(nodes, model.NodeItem{
			Uid: v,
		})
	}

	_, err = cli.cce.RemoveNode(&model.RemoveNodeRequest{
		ClusterId: clusterId,
		Body: &model.RemoveNodesTask{
			Spec: &model.RemoveNodesSpec{
				Login: &model.Login{
					UserPassword: &model.UserPassword{
						Password: pw,
					},
				},
				Nodes: nodes,
			},
		},
	})

	return err
}

// CleanNodePoolNodes delete node pool nodes
func (cli *CceClient) CleanNodePoolNodes(clusterId string, nodeIds []string) error {
	if len(nodeIds) == 0 {
		return nil
	}

	for _, nodeId := range nodeIds {
		jobId, err := cli.DeleteNode(clusterId, nodeId, true)
		if err != nil {
			return fmt.Errorf("删除节点[%s]失败, error: %s", nodeId, err)
		}

		blog.Infof("CleanNodePoolNodes[%s] DeleteNode[%s] jobId[%s] success", clusterId, nodeId, jobId)
	}

	return nil
}

// 节点池配置管理实现 实现组件参数管理
