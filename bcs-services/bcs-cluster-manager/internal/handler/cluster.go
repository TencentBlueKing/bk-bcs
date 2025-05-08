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

package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	clusterac "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateCluster(ctx context.Context,
	req *cmproto.CreateClusterReq, resp *cmproto.CreateClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewCreateAction(cm.model, cm.locker)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// CreateVirtualCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateVirtualCluster(ctx context.Context,
	req *cmproto.CreateVirtualClusterReq, resp *cmproto.CreateVirtualClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewCreateVirtualClusterAction(cm.model, cm.locker)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateVirtualCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateVirtualCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// CheckCloudKubeConfig implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CheckCloudKubeConfig(ctx context.Context,
	req *cmproto.KubeConfigReq, resp *cmproto.KubeConfigResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	ca := clusterac.NewCheckKubeAction()
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CheckCloudKubeConfig", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CheckCloudKubeConfig, req %v, resp %v", reqID, req, resp)
	return nil
}

// CheckCloudKubeConfigConnect implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CheckCloudKubeConfigConnect(ctx context.Context,
	req *cmproto.KubeConfigConnectReq, resp *cmproto.KubeConfigConnectResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	ca := clusterac.NewCheckKubeConnectAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CheckCloudKubeConfig", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CheckCloudKubeConfig, req %v, resp %v", reqID, req, resp)
	return nil
}

// ImportCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ImportCluster(ctx context.Context,
	req *cmproto.ImportClusterReq, resp *cmproto.ImportClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewImportAction(cm.model, cm.locker)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ImportCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ImportCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// RetryCreateClusterTask implements interface cmproto.ClusterManagerServer for retry create task
func (cm *ClusterManager) RetryCreateClusterTask(ctx context.Context,
	req *cmproto.RetryCreateClusterReq, resp *cmproto.RetryCreateClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewRetryCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("RetryCreateCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: RetryCreateCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateCluster(ctx context.Context,
	req *cmproto.UpdateClusterReq, resp *cmproto.UpdateClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateClusterModule implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateClusterModule(ctx context.Context,
	req *cmproto.UpdateClusterModuleRequest, resp *cmproto.UpdateClusterModuleResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewUpdateClusterModuleAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateClusterModule", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateClusterModule, req %v, resp %v", reqID, req, resp)
	return nil
}

// AddNodesToCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) AddNodesToCluster(ctx context.Context,
	req *cmproto.AddNodesRequest, resp *cmproto.AddNodesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewAddNodesAction(cm.model)
	// 新的接口请求内容不变，直接复用；V1 版本仍然保持返回任务作为 Data，V2 版本返回任务列表作为 Data
	newReq := &cmproto.AddNodesV2Request{
		ClusterID:         req.ClusterID,
		Nodes:             req.Nodes,
		InitLoginPassword: req.InitLoginPassword,
		NodeGroupID:       req.NodeGroupID,
		OnlyCreateInfo:    req.OnlyCreateInfo,
		Operator:          req.Operator,
		NodeTemplateID:    req.NodeTemplateID,
		IsExternalNode:    req.IsExternalNode,
		Login:             req.Login,
		Advance:           req.Advance,
	}
	newResp := &cmproto.AddNodesV2Response{}
	ca.Handle(ctx, newReq, newResp)
	resp.Code = newResp.Code
	if len(newResp.Data) > 0 {
		resp.Data = newResp.Data[0]
	}
	resp.Message = newResp.Message
	resp.Result = newResp.Result
	metrics.ReportAPIRequestMetric("AddNodesToCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: AddNodesToCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// AddNodesToClusterV2 implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) AddNodesToClusterV2(ctx context.Context,
	req *cmproto.AddNodesV2Request, resp *cmproto.AddNodesV2Response) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewAddNodesAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("AddNodesToClusterV2", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: AddNodesToClusterV2, req %v, resp %v", reqID, req, resp)
	return nil
}

// BatchDeleteNodesFromCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) BatchDeleteNodesFromCluster(ctx context.Context,
	req *cmproto.BatchDeleteClusterNodesRequest, resp *cmproto.BatchDeleteClusterNodesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewBatchDeleteClusterNodesAction(cm.model, cm.locker)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("BatchDeleteClusterNodes", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: BatchDeleteClusterNodes, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteNodesFromCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteNodesFromCluster(ctx context.Context,
	req *cmproto.DeleteNodesRequest, resp *cmproto.DeleteNodesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewDeleteNodesAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteNodesFromCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteNodesFromCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteCluster(ctx context.Context,
	req *cmproto.DeleteClusterReq, resp *cmproto.DeleteClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewDeleteAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteVirtualCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteVirtualCluster(ctx context.Context,
	req *cmproto.DeleteVirtualClusterReq, resp *cmproto.DeleteVirtualClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewDeleteVirtualAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteVirtualCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteVirtualCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetCluster(ctx context.Context,
	req *cmproto.GetClusterReq, resp *cmproto.GetClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewGetAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetCluster, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: GetCluster, req %v, resp %v",
		reqID, req, resp)
	return nil
}

// GetClustersMetaData implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetClustersMetaData(ctx context.Context,
	req *cmproto.GetClustersMetaDataRequest, resp *cmproto.GetClustersMetaDataResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := clusterac.NewGetClustersMetaDataAction(cm.model, cm.kubeOp)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetClustersMetaData", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetClustersMetaData, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListProjectCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListProjectCluster(ctx context.Context,
	req *cmproto.ListProjectClusterReq, resp *cmproto.ListProjectClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	ca := clusterac.NewListProjectClusterAction(cm.model, cm.iam)
	ca.Handle(ctx, req, resp)

	metrics.ReportAPIRequestMetric("ListProjectCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListProjectCluster, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListProjectCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCluster(ctx context.Context,
	req *cmproto.ListClusterReq, resp *cmproto.ListClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewListAction(cm.model, cm.iam)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCluster, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCommonCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCommonCluster(ctx context.Context,
	req *cmproto.ListCommonClusterReq, resp *cmproto.ListCommonClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewListCommonClusterAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCommonCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCommonCluster, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCommonCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// InitFederationCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) InitFederationCluster(ctx context.Context,
	req *cmproto.InitFederationClusterReq, resp *cmproto.InitFederationClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	metrics.ReportAPIRequestMetric("InitFederationCluster", "grpc", "notimplemented", start)
	blog.Infof("reqID: %s, action: InitFederationCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// AddFederatedCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) AddFederatedCluster(ctx context.Context,
	req *cmproto.AddFederatedClusterReq, resp *cmproto.AddFederatedClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	fa := clusterac.NewFederateAction(cm.model)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("AddFederatedCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: AddFederatedCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetNode implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetNode(ctx context.Context,
	req *cmproto.GetNodeRequest, resp *cmproto.GetNodeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := clusterac.NewGetNodeAction(cm.model)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetNode", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetNoe, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetNodeInfo implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetNodeInfo(ctx context.Context,
	req *cmproto.GetNodeInfoRequest, resp *cmproto.GetNodeInfoResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := clusterac.NewGetNodeInfoAction(cm.model)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetNodeInfo", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetNodeInfo, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateNode implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateNode(ctx context.Context,
	req *cmproto.UpdateNodeRequest, resp *cmproto.UpdateNodeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := clusterac.NewUpdateNodeAction(cm.model)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateNode", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateNode, req %v, resp %v", reqID, req, resp)
	return nil
}

// CheckNodeInCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CheckNodeInCluster(ctx context.Context,
	req *cmproto.CheckNodesRequest, resp *cmproto.CheckNodesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := clusterac.NewCheckNodeAction(cm.model, cm.kubeOp)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CheckNodeInCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CheckNodeInCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListNodesInCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNodesInCluster(ctx context.Context,
	req *cmproto.ListNodesInClusterRequest, resp *cmproto.ListNodesInClusterResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	fa := clusterac.NewListNodesInClusterAction(cm.model, cm.kubeOp)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListNodesInCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListNodesInCluster, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListNodesInCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListMastersInCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListMastersInCluster(ctx context.Context,
	req *cmproto.ListMastersInClusterRequest, resp *cmproto.ListMastersInClusterResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	fa := clusterac.NewListMastersInClusterAction(cm.model, cm.kubeOp)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListMastersInCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListMastersInCluster, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListMastersInCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateVirtualClusterQuota implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateVirtualClusterQuota(ctx context.Context,
	req *cmproto.UpdateVirtualClusterQuotaReq, resp *cmproto.UpdateVirtualClusterQuotaResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := clusterac.NewUpdateVirtualClusterQuotaAction(cm.model, cm.kubeOp)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateVirtualClusterQuota", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateVirtualClusterQuota, req %v, resp %v", reqID, req, resp)
	return nil
}

// AddSubnetToCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) AddSubnetToCluster(ctx context.Context,
	req *cmproto.AddSubnetToClusterReq, resp *cmproto.AddSubnetToClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := clusterac.NewAddSubnetToClusterAction(cm.model)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("AddSubnetToCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: AddSubnetToCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// SwitchClusterUnderlayNetwork implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) SwitchClusterUnderlayNetwork(ctx context.Context,
	req *cmproto.SwitchClusterUnderlayNetworkReq, resp *cmproto.SwitchClusterUnderlayNetworkResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := clusterac.NewSwitchClusterUnderlayNetworkAction(cm.model)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("SwitchClusterUnderlayNetwork", "grpc",
		strconv.Itoa(int(resp.Code)), start)

	blog.Infof("reqID: %s, action: SwitchClusterUnderlayNetwork, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetClusterSharedProject implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetClusterSharedProject(ctx context.Context,
	req *cmproto.GetClusterSharedProjectRequest, resp *cmproto.GetClusterSharedProjectResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := clusterac.NewGetClusterSharedProjectAction(cm.model)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetClusterSharedProject", "grpc",
		strconv.Itoa(int(resp.Code)), start)

	blog.Infof("reqID: %s, action: GetClusterSharedProject, req %v, resp %v", reqID, req, resp)
	return nil
}
