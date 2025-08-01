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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/operationlog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// ListOperationLogs implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListOperationLogs(ctx context.Context,
	req *cmproto.ListOperationLogsRequest, resp *cmproto.ListOperationLogsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := operationlog.NewListOperationLogsAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListOperationLogs", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListOperationLogs, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListProjectOperationLogs implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListProjectOperationLogs(ctx context.Context,
	req *cmproto.ListOperationLogsRequest, resp *cmproto.ListOperationLogsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := operationlog.NewListOperationLogsAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListOperationLogsV2", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListOperationLogsV2, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListTaskStepLogs implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListTaskStepLogs(ctx context.Context,
	req *cmproto.ListTaskStepLogsRequest, resp *cmproto.ListTaskStepLogsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := operationlog.NewListTaskStepLogsAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListTaskStepLogs", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListTaskStepLogs, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListTaskRecords implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListTaskRecords(ctx context.Context,
	req *cmproto.ListTaskRecordsRequest, resp *cmproto.ListTaskRecordsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := operationlog.NewTaskRecordsAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListTaskRecords", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListTaskRecords, req %v, resp %v", reqID, req, resp)
	return nil
}
