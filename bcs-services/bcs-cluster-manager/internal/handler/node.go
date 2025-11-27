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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/node"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// RecordNodeInfo implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) RecordNodeInfo(ctx context.Context,
	req *cmproto.RecordNodeInfoRequest, resp *cmproto.CommonResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := node.NewRecordNodeDataAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("RecordNodeInfo", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: RecordNodeInfo, req %v, resp %v", reqID, req, resp)
	return nil
}

// CordonNode implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CordonNode(ctx context.Context,
	req *cmproto.CordonNodeRequest, resp *cmproto.CordonNodeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := node.NewCordonNodeAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CordonNode", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CordonNode, req %v, resp %v", reqID, req, resp)
	return nil
}

// UnCordonNode implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UnCordonNode(ctx context.Context,
	req *cmproto.UnCordonNodeRequest, resp *cmproto.UnCordonNodeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := node.NewUnCordonNodeAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UnCordonNode", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UnCordonNode, req %v, resp %v", reqID, req, resp)
	return nil
}

// DrainNode implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DrainNode(ctx context.Context,
	req *cmproto.DrainNodeRequest, resp *cmproto.DrainNodeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := node.NewDrainNodeAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DrainNode", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DrainNode, req %v, resp %v", reqID,
		utils.ToJSONString(req), utils.ToJSONString(resp))
	return nil
}

// CheckDrainNode implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CheckDrainNode(ctx context.Context,
	req *cmproto.CheckDrainNodeRequest, resp *cmproto.CheckDrainNodeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := node.NewCheckDrainNodeAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CheckDrainNode", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CheckDrainNode, req %v, resp %v", reqID,
		utils.ToJSONString(req), utils.ToJSONString(resp))
	return nil
}

// UpdateNodeAnnotations implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateNodeAnnotations(ctx context.Context,
	req *cmproto.UpdateNodeAnnotationsRequest, resp *cmproto.UpdateNodeAnnotationsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := node.NewUpdateNodeAnnotationsAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateNodeAnnotations", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateNodeAnnotations, req %v, resp %v", reqID, utils.ToJSONString(req),
		utils.ToJSONString(resp))
	return nil
}

// UpdateNodeLabels implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateNodeLabels(ctx context.Context,
	req *cmproto.UpdateNodeLabelsRequest, resp *cmproto.UpdateNodeLabelsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := node.NewUpdateNodeLabelsAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateNodeLabels", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateNodeLabels, req %v, resp %v", reqID, utils.ToJSONString(req),
		utils.ToJSONString(resp))
	return nil
}

// UpdateNodeTaints implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateNodeTaints(ctx context.Context,
	req *cmproto.UpdateNodeTaintsRequest, resp *cmproto.UpdateNodeTaintsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := node.NewUpdateNodeTaintsAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateNodeTaints", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateNodeTaints, req %v, resp %v", reqID, utils.ToJSONString(req),
		utils.ToJSONString(resp))
	return nil
}

// ListCloudNodePublicPrefix implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudNodePublicPrefix(ctx context.Context,
	req *cmproto.ListCloudNodePublicPrefixRequest, resp *cmproto.ListCloudNodePublicPrefixResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := node.NewListCloudNodePublicPrefixAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudNodePublicPrefix", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudNodePublicPrefix, req %v, resp %v", reqID, utils.ToJSONString(req),
		utils.ToJSONString(resp))
	return nil
}

// ListClusterNodes implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListClusterNodes(ctx context.Context,
	req *cmproto.ListClusterNodesRequest, resp *cmproto.ListClusterNodesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := node.NewListClusterNodeAction(cm.model)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListClusterNodes", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListClusterNodes, req %v, resp %v", reqID, req, resp)
	return nil
}

// SyncClusterNodes implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) SyncClusterNodes(ctx context.Context,
	req *cmproto.SyncClusterNodesRequest, resp *cmproto.SyncClusterNodesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := node.NewSyncClusterNodesAction(cm.model, cm.kubeOp)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("SyncClusterNodes", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: SyncClusterNodes, req %v, resp %v", reqID, req, resp)
	return nil
}
