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
	"fmt"
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
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := namespace.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("createnamespace", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: createnamespace", reqID)
	blog.V(3).Infof("reqID: %s, action: createnamespace, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateNamespace implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateNamespace(ctx context.Context,
	req *cmproto.UpdateNamespaceReq, resp *cmproto.UpdateNamespaceResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ua := namespace.NewUpdateAction(cm.model)
	ua.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("updatenamespace", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: updatenamespace", reqID)
	blog.V(3).Infof("reqID: %s, action: updatenamespace, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteNamespace implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteNamespace(ctx context.Context,
	req *cmproto.DeleteNamespaceReq, resp *cmproto.DeleteNamespaceResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := namespace.NewDeleteAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("deletenamespace", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: deletenamespace", reqID)
	blog.V(3).Infof("reqID: %s, action: deletenamespace, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetNamespace implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetNamespace(ctx context.Context,
	req *cmproto.GetNamespaceReq, resp *cmproto.GetNamespaceResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := namespace.NewGetAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("getnamespace", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: getnamespace", reqID)
	blog.V(3).Infof("reqID: %s, action: getnamespace, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListNamespace implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNamespace(ctx context.Context,
	req *cmproto.ListNamespaceReq, resp *cmproto.ListNamespaceResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := namespace.NewListAction(cm.model)
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("listnamespace", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: listnamespace", reqID)
	blog.V(3).Infof("reqID: %s, action: listnamespace, req %v, resp %v", reqID, req, resp)
	return nil
}

// CreateNamespaceWithQuota implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateNamespaceWithQuota(ctx context.Context,
	req *cmproto.CreateNamespaceWithQuotaReq, resp *cmproto.CreateNamespaceWithQuotaResp) error {
	if resp == nil {
		return fmt.Errorf("incoming resp struct is empty when CreateNamespaceWithQuotaResp")
	}
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := namespace.NewCreateQuotaAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("createnamespacequota", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: createnamespacequota", reqID)
	blog.V(3).Infof("reqID: %s, action: createnamespacequota, req %v, resp %v", reqID, req, resp)
	return nil
}
