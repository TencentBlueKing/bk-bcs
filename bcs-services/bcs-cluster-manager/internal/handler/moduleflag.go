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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/moduleflag"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateCloudModuleFlag implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateCloudModuleFlag(ctx context.Context,
	req *cmproto.CreateCloudModuleFlagRequest, resp *cmproto.CreateCloudModuleFlagResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := moduleflag.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateCloudModuleFlag", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateCloudModuleFlag, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateCloudModuleFlag implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateCloudModuleFlag(ctx context.Context,
	req *cmproto.UpdateCloudModuleFlagRequest, resp *cmproto.UpdateCloudModuleFlagResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := moduleflag.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateCloudModuleFlag", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateCloudModuleFlag, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteCloudModuleFlag implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteCloudModuleFlag(ctx context.Context,
	req *cmproto.DeleteCloudModuleFlagRequest, resp *cmproto.DeleteCloudModuleFlagResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := moduleflag.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteCloudModuleFlag", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteCloudModuleFlag, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudModuleFlag implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudModuleFlag(ctx context.Context,
	req *cmproto.ListCloudModuleFlagRequest, resp *cmproto.ListCloudModuleFlagResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := moduleflag.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudModuleFlag", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudModuleFlag, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudModuleFlag, req %v, resp %v", reqID, req, resp)
	return nil
}
