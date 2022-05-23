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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/account"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateCloudAccount implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateCloudAccount(ctx context.Context,
	req *cmproto.CreateCloudAccountRequest, resp *cmproto.CreateCloudAccountResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := account.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateCloudAccount", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateCloudAccount, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateCloudAccount implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateCloudAccount(ctx context.Context,
	req *cmproto.UpdateCloudAccountRequest, resp *cmproto.UpdateCloudAccountResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := account.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateCloudAccount", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateCloudAccount, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteCloudAccount implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteCloudAccount(ctx context.Context,
	req *cmproto.DeleteCloudAccountRequest, resp *cmproto.DeleteCloudAccountResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := account.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteCloudAccount", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteCloudAccount, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudVPC implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudAccount(ctx context.Context,
	req *cmproto.ListCloudAccountRequest, resp *cmproto.ListCloudAccountResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := account.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudAccount", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudAccount, req %v, resp.Code %d, resp.Message %s, resp.Data.Length",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudAccount, req %v, resp %v", reqID, req, resp)
	return nil
}
