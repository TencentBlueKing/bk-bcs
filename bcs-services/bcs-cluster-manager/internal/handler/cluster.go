/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
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

// AddNodesToCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) AddNodesToCluster(ctx context.Context,
	req *cmproto.AddNodesRequest, resp *cmproto.AddNodesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewAddNodesAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("AddNodesToCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: AddNodesToCluster, req %v, resp %v", reqID, req, resp)
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
	ca := clusterac.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteCluster, req %v, resp %v", reqID, req, resp)
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

// ListCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCluster(ctx context.Context,
	req *cmproto.ListClusterReq, resp *cmproto.ListClusterResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clusterac.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCluster, req %v, resp.Code %d, resp.Message %s, resp.Data.Length",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCluster, req %v, resp %v", reqID, req, resp)
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
	na := clusterac.NewCheckNodeAction(cm.model)
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
	fa := clusterac.NewListNodesInClusterAction(cm.model)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListNodesInCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListNodesInCluster, req %v, resp.Code %d, resp.Message %s, resp.Data.Length",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListNodesInCluster, req %v, resp %v", reqID, req, resp)
	return nil
}
