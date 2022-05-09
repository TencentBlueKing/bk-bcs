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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateTask implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateTask(ctx context.Context,
	req *cmproto.CreateTaskRequest, resp *cmproto.CreateTaskResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := task.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateTask", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action CreateTask, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateTask implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateTask(ctx context.Context,
	req *cmproto.UpdateTaskRequest, resp *cmproto.UpdateTaskResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := task.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateTask", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateTask, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteTask implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteTask(ctx context.Context,
	req *cmproto.DeleteTaskRequest, resp *cmproto.DeleteTaskResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := task.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteTask", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteTask, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetTask implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetTask(ctx context.Context,
	req *cmproto.GetTaskRequest, resp *cmproto.GetTaskResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := task.NewGetAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetTask", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetTask, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: GetTask, req %v, resp %v",
		reqID, req, resp)
	return nil
}

// ListTask implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListTask(ctx context.Context,
	req *cmproto.ListTaskRequest, resp *cmproto.ListTaskResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := task.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListTask", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListTask, req %v, resp.Code %d, resp.Message %s, resp.Data.Length",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListTask, req %v, resp %v", reqID, req, resp)
	return nil
}

// RetryTask implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) RetryTask(ctx context.Context,
	req *cmproto.RetryTaskRequest, resp *cmproto.RetryTaskResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := task.NewRetryAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("RetryTask", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: RetryTask, req %v, resp %v", reqID, req, resp)
	return nil
}
