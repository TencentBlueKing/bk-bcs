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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/health"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// Health implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) Health(ctx context.Context,
	req *cmproto.HealthRequest, resp *cmproto.HealthResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := health.NewHealthAction()
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("Health", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: Health, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: Health, req %v, resp %v",
		reqID, req, resp)

	return nil
}

// CleanDbHistoryData implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CleanDbHistoryData(ctx context.Context,
	req *cmproto.CleanDbHistoryDataRequest, resp *cmproto.CleanDbHistoryDataResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := health.NewCleanDBDataAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CleanDbHistoryData", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CleanDbHistoryData, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: CleanDbHistoryData, req %v, resp %v",
		reqID, req, resp)

	return nil
}
