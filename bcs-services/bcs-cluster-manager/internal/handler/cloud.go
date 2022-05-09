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

package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/cloud"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateCloud implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateCloud(ctx context.Context,
	req *cmproto.CreateCloudRequest, resp *cmproto.CreateCloudResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloud.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateCloud", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateCloud, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateCloud implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateCloud(ctx context.Context,
	req *cmproto.UpdateCloudRequest, resp *cmproto.UpdateCloudResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloud.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateCloud", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateCloud, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteCloud implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteCloud(ctx context.Context,
	req *cmproto.DeleteCloudRequest, resp *cmproto.DeleteCloudResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloud.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteCloud", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteCloud, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetCloud implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetCloud(ctx context.Context,
	req *cmproto.GetCloudRequest, resp *cmproto.GetCloudResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloud.NewGetAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetCloud", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetCloud, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: GetCloud, req %v, resp %v",
		reqID, req, resp)
	return nil
}

// ListCloud implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloud(ctx context.Context,
	req *cmproto.ListCloudRequest, resp *cmproto.ListCloudResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloud.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloud", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloud, req %v, resp.Code %d, resp.Message %s, resp.Data.Length",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloud, req %v, resp %v", reqID, req, resp)
	return nil
}
