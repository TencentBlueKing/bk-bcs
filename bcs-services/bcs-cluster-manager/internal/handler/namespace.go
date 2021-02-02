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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateNamespace implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateNamespace(ctx context.Context,
	req *cmproto.CreateNamespaceReq, resp *cmproto.CreateNamespaceResp) error {
	start := time.Now()
	ca := namespace.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("createnamespace", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: createnamespace, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// UpdateNamespace implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateNamespace(ctx context.Context,
	req *cmproto.UpdateNamespaceReq, resp *cmproto.UpdateNamespaceResp) error {
	start := time.Now()
	ua := namespace.NewUpdateAction(cm.model)
	ua.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("updatenamespace", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: updatenamespace, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// DeleteNamespace implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteNamespace(ctx context.Context,
	req *cmproto.DeleteNamespaceReq, resp *cmproto.DeleteNamespaceResp) error {
	start := time.Now()
	ca := namespace.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("deletenamespace", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: deletenamespace, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// GetNamespace implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetNamespace(ctx context.Context,
	req *cmproto.GetNamespaceReq, resp *cmproto.GetNamespaceResp) error {
	start := time.Now()
	ga := namespace.NewGetAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("getnamespace", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: getnamespace, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// ListNamespace implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNamespace(ctx context.Context,
	req *cmproto.ListNamespaceReq, resp *cmproto.ListNamespaceResp) error {
	start := time.Now()
	la := namespace.NewListAction(cm.model)
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("listnamespace", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: listnamespace, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// CreateNamespaceWithQuota implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateNamespaceWithQuota(ctx context.Context,
	req *cmproto.CreateNamespaceWithQuotaReq, resp *cmproto.CreateNamespaceWithQuotaResp) error {
	return nil
}
