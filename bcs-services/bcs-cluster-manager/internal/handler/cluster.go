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
	start := time.Now()
	ca := clusterac.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("createcluster", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: createcluster, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// UpdateCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateCluster(ctx context.Context,
	req *cmproto.UpdateClusterReq, resp *cmproto.UpdateClusterResp) error {
	start := time.Now()
	ca := clusterac.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("updatecluster", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: updatecluster, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// DeleteCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteCluster(ctx context.Context,
	req *cmproto.DeleteClusterReq, resp *cmproto.DeleteClusterResp) error {
	start := time.Now()
	ca := clusterac.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("deletecluster", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: deletecluster, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// GetCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetCluster(ctx context.Context,
	req *cmproto.GetClusterReq, resp *cmproto.GetClusterResp) error {
	start := time.Now()
	ca := clusterac.NewGetAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("getcluster", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: getcluster, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// ListCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCluster(ctx context.Context,
	req *cmproto.ListClusterReq, resp *cmproto.ListClusterResp) error {
	start := time.Now()
	ca := clusterac.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("listcluster", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: listcluster, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// InitFederationCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) InitFederationCluster(ctx context.Context,
	req *cmproto.InitFederationClusterReq, resp *cmproto.InitFederationClusterResp) error {
	start := time.Now()
	metrics.ReportAPIRequestMetric("initfederationcluster", "grpc", "notimplemented", start)
	return nil
}

// AddFederatedCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) AddFederatedCluster(ctx context.Context,
	req *cmproto.AddFederatedClusterReq, resp *cmproto.AddFederatedClusterResp) error {
	start := time.Now()
	metrics.ReportAPIRequestMetric("addfederatedcluster", "grpc", "notimplemented", start)
	return nil
}
