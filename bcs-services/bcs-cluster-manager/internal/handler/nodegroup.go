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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/nodegroup"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateNodeGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateNodeGroup(ctx context.Context,
	req *cmproto.CreateNodeGroupRequest, resp *cmproto.CreateNodeGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewCreateAction(cm.model, cm.locker)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateNodeGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateNodeGroup, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateNodeGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateNodeGroup(ctx context.Context,
	req *cmproto.UpdateNodeGroupRequest, resp *cmproto.UpdateNodeGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateNodeGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateNodeGroup, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteNodeGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteNodeGroup(ctx context.Context,
	req *cmproto.DeleteNodeGroupRequest, resp *cmproto.DeleteNodeGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteNodeGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteNodeGroup, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetNodeGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetNodeGroup(ctx context.Context,
	req *cmproto.GetNodeGroupRequest, resp *cmproto.GetNodeGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewGetAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetNodeGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetNodeGroup, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: GetNodeGroup, req %v, resp %v",
		reqID, req, resp)
	return nil
}

// ListClusterNodeGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListClusterNodeGroup(ctx context.Context,
	req *cmproto.ListClusterNodeGroupRequest, resp *cmproto.ListClusterNodeGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	ca := nodegroup.NewListClusterGroupAction(cm.model)
	ca.Handle(ctx, req, resp)

	metrics.ReportAPIRequestMetric("ListClusterNodeGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListClusterNodeGroup, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListClusterNodeGroup, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListNodeGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNodeGroup(ctx context.Context,
	req *cmproto.ListNodeGroupRequest, resp *cmproto.ListNodeGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewListAction(cm.model)
	newReq := &cmproto.ListNodeGroupV2Request{
		Name:      req.Name,
		ClusterID: req.ClusterID,
		Region:    req.Region,
		ProjectID: req.ProjectID,
		Page:      0,
		Limit:     0,
	}
	newResp := &cmproto.ListNodeGroupV2Response{}
	ca.Handle(ctx, newReq, newResp)
	resp.Code = newResp.Code
	resp.Message = newResp.Message
	resp.Result = newResp.Result
	resp.Data = newResp.Data.GetResults()
	metrics.ReportAPIRequestMetric("ListNodeGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListNodeGroup, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListNodeGroup, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListNodeGroupV2 implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNodeGroupV2(ctx context.Context,
	req *cmproto.ListNodeGroupV2Request, resp *cmproto.ListNodeGroupV2Response) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListNodeGroupV2", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListNodeGroupV2, req %v, resp.Code %d, "+
		"resp.Message %s", reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: ListNodeGroupV2, req %v, resp %v", reqID, req, resp)
	return nil
}

// RecommendNodeGroupConf implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) RecommendNodeGroupConf(ctx context.Context,
	req *cmproto.RecommendNodeGroupConfReq, resp *cmproto.RecommendNodeGroupConfResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewRecommendNodeGroupConfAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("RecommendNodeGroupConf", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: RecommendNodeGroupConf, req %v, resp %v", reqID, req, resp)
	return nil
}

// MoveNodesToGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) MoveNodesToGroup(ctx context.Context,
	req *cmproto.MoveNodesToGroupRequest, resp *cmproto.MoveNodesToGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewMoveNodeAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("MoveNodesToGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: MoveNodesToGroup, req %v, resp %v", reqID, req, resp)
	return nil
}

// RemoveNodesFromGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) RemoveNodesFromGroup(ctx context.Context,
	req *cmproto.RemoveNodesFromGroupRequest, resp *cmproto.RemoveNodesFromGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewRemoveNodeAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("RemoveNodesFromGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: RemoveNodesFromGroup, req %v, resp %v", reqID, req, resp)
	return nil
}

// CleanNodesInGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CleanNodesInGroup(ctx context.Context,
	req *cmproto.CleanNodesInGroupRequest, resp *cmproto.CleanNodesInGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewCleanNodesAction(cm.model, cm.locker)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CleanNodesInGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CleanNodesInGroup, req %v, resp %v", reqID, req, resp)
	return nil
}

// CleanNodesInGroupV2 implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CleanNodesInGroupV2(ctx context.Context,
	req *cmproto.CleanNodesInGroupV2Request, resp *cmproto.CleanNodesInGroupV2Response) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	// 新的缩容接口从 query 拿参数，但主体逻辑使用 v1 的缩容接口，为了遵守 HTTP DELETE 规范, 不从 body 传参
	ca := nodegroup.NewCleanNodesAction(cm.model, cm.locker)
	newReq := &cmproto.CleanNodesInGroupRequest{
		ClusterID:   req.ClusterID,
		Nodes:       strings.Split(req.Nodes, ","),
		NodeGroupID: req.NodeGroupID,
		Operator:    req.Operator,
	}
	newResp := &cmproto.CleanNodesInGroupResponse{}
	ca.Handle(ctx, newReq, newResp)
	resp.Code = newResp.Code
	resp.Data = newResp.Data
	resp.Message = newResp.Message
	resp.Result = newResp.Result
	metrics.ReportAPIRequestMetric("CleanNodesInGroupV2", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CleanNodesInGroupV2, req %v, resp %v", reqID, utils.ToJSONString(req), resp)
	return nil
}

// ListNodesInGroup implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNodesInGroup(ctx context.Context,
	req *cmproto.GetNodeGroupRequest, resp *cmproto.ListNodesInGroupResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewListNodesAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListNodesInGroup", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListNodesInGroup, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListNodesInGroup, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListNodesInGroupV2 implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNodesInGroupV2(ctx context.Context,
	req *cmproto.ListNodesInGroupV2Request, resp *cmproto.ListNodesInGroupV2Response) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewListNodesV2Action(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListNodesInGroupV2", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListNodesInGroupV2, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListNodesInGroupV2, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateGroupDesiredNode implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateGroupDesiredNode(ctx context.Context,
	req *cmproto.UpdateGroupDesiredNodeRequest, resp *cmproto.UpdateGroupDesiredNodeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewUpdateDesiredNodeAction(cm.model, cm.locker)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateGroupDesiredNode", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateGroupDesiredNode, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateGroupDesiredSize implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateGroupDesiredSize(ctx context.Context,
	req *cmproto.UpdateGroupDesiredSizeRequest, resp *cmproto.UpdateGroupDesiredSizeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewUpdateDesiredSizeAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateGroupDesiredSize", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateGroupDesiredSize, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateGroupMinMaxSize implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateGroupMinMaxSize(ctx context.Context,
	req *cmproto.UpdateGroupMinMaxSizeRequest, resp *cmproto.UpdateGroupMinMaxSizeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewUpdateGroupMinMaxAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateGroupMinMaxSize", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateGroupMinMaxSize, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateGroupAsTimeRange implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateGroupAsTimeRange(ctx context.Context,
	req *cmproto.UpdateGroupAsTimeRangeRequest, resp *cmproto.UpdateGroupAsTimeRangeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewUpdateGroupAsTimeRangeAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateGroupAsTimeRange", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateGroupAsTimeRange, req %v, resp %v", reqID, req, resp)
	return nil
}

// EnableNodeGroupAutoScale implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) EnableNodeGroupAutoScale(ctx context.Context,
	req *cmproto.EnableNodeGroupAutoScaleRequest, resp *cmproto.EnableNodeGroupAutoScaleResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewEnableNodeGroupAutoScaleAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("EnableNodeGroupAutoScale", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: EnableNodeGroupAutoScale, req %v, resp %v", reqID, req, resp)
	return nil
}

// DisableNodeGroupAutoScale implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DisableNodeGroupAutoScale(ctx context.Context,
	req *cmproto.DisableNodeGroupAutoScaleRequest, resp *cmproto.DisableNodeGroupAutoScaleResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodegroup.NewDisableNodeGroupAutoScaleAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DisableNodeGroupAutoScale", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DisableNodeGroupAutoScale, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetExternalNodeScriptByGroupID implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetExternalNodeScriptByGroupID(ctx context.Context,
	req *cmproto.GetExternalNodeScriptRequest, resp *cmproto.GetExternalNodeScriptResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := nodegroup.NewGetExternalNodesScriptAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetExternalNodeScriptByGroupID", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetExternalNodeScriptByGroupID, req %v, resp %v", reqID, req, resp)
	return nil
}

// TransNodeGroupToNodeTemplate implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) TransNodeGroupToNodeTemplate(ctx context.Context,
	req *cmproto.TransNodeGroupToNodeTemplateRequest, resp *cmproto.TransNodeGroupToNodeTemplateResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ta := nodegroup.NewTransNgToNtAction(cm.model)
	ta.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("TransNodeGroupToNodeTemplate", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: TransNodeGroupToNodeTemplate, req %v, resp %v", reqID, req, resp)
	return nil
}
