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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/namespacequota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateNamespaceQuota implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateNamespaceQuota(ctx context.Context,
	req *cmproto.CreateNamespaceQuotaReq, resp *cmproto.CreateNamespaceQuotaResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := namespacequota.NewCreateAction(cm.model, cm.kubeOp)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("createquota", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: createquota", reqID)
	blog.V(3).Infof("reqID: %s, action: createquota, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateNamespaceQuota implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateNamespaceQuota(ctx context.Context,
	req *cmproto.UpdateNamespaceQuotaReq, resp *cmproto.UpdateNamespaceQuotaResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ua := namespacequota.NewUpdateAction(cm.model, cm.kubeOp)
	ua.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("updatequota", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: updatequota", reqID)
	blog.V(3).Infof("reqID: %s, action: updatequota, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteNamespaceQuota implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteNamespaceQuota(ctx context.Context,
	req *cmproto.DeleteNamespaceQuotaReq, resp *cmproto.DeleteNamespaceQuotaResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ua := namespacequota.NewDeleteAction(cm.model, cm.kubeOp)
	ua.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("deletequota", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: deletequota", reqID)
	blog.V(3).Infof("reqID: %s, action: deletequota, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetNamespaceQuota implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetNamespaceQuota(ctx context.Context,
	req *cmproto.GetNamespaceQuotaReq, resp *cmproto.GetNamespaceQuotaResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := namespacequota.NewGetAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("getquota", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: getquota", reqID)
	blog.V(3).Infof("reqID: %s, action: getquota, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListNamespaceQuota implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNamespaceQuota(ctx context.Context,
	req *cmproto.ListNamespaceQuotaReq, resp *cmproto.ListNamespaceQuotaResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := namespacequota.NewListAction(cm.model)
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("listquota", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: listquota", reqID)
	blog.V(3).Infof("reqID: %s, action: listquota, req %v, resp %v", reqID, req, resp)
	return nil
}
