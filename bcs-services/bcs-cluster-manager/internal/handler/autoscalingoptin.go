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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/autoscalingoption"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateAutoScalingOption implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateAutoScalingOption(ctx context.Context,
	req *cmproto.CreateAutoScalingOptionRequest, resp *cmproto.CreateAutoScalingOptionResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := autoscalingoption.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateAutoScalingOption", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action CreateAutoScalingOption, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateAutoScalingOption implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateAutoScalingOption(ctx context.Context,
	req *cmproto.UpdateAutoScalingOptionRequest, resp *cmproto.UpdateAutoScalingOptionResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := autoscalingoption.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateAutoScalingOption", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateAutoScalingOption, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteAutoScalingOption implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteAutoScalingOption(ctx context.Context,
	req *cmproto.DeleteAutoScalingOptionRequest, resp *cmproto.DeleteAutoScalingOptionResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := autoscalingoption.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteAutoScalingOption", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteAutoScalingOption, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetAutoScalingOption implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetAutoScalingOption(ctx context.Context,
	req *cmproto.GetAutoScalingOptionRequest, resp *cmproto.GetAutoScalingOptionResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := autoscalingoption.NewGetAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetAutoScalingOption", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetAutoScalingOption, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: GetAutoScalingOption, req %v, resp %v",
		reqID, req, resp)
	return nil
}

// ListAutoScalingOption implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListAutoScalingOption(ctx context.Context,
	req *cmproto.ListAutoScalingOptionRequest, resp *cmproto.ListAutoScalingOptionResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := autoscalingoption.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListAutoScalingOption", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListAutoScalingOption, req %v, resp.Code %d, resp.Message %s, resp.Data.Length",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListAutoScalingOption, req %v, resp %v", reqID, req, resp)
	return nil
}
