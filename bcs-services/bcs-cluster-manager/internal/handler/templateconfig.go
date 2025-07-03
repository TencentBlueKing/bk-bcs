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

// Package handler xxx
package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/templateconfig"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateTemplateConfig implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateTemplateConfig(ctx context.Context,
	req *cmproto.CreateTemplateConfigRequest, resp *cmproto.CreateTemplateConfigResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := templateconfig.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateTemplateConfig", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateTemplateConfig, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteTemplateConfig implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteTemplateConfig(ctx context.Context,
	req *cmproto.DeleteTemplateConfigRequest, resp *cmproto.DeleteTemplateConfigResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := templateconfig.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteTemplateConfig", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteTemplateConfig, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListTemplateConfig implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListTemplateConfig(ctx context.Context,
	req *cmproto.ListTemplateConfigRequest, resp *cmproto.ListTemplateConfigResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := templateconfig.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListTemplateConfig", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListTemplateConfig, req %v, resp.Code %d, r"+
		"esp.Message %s, resp.Data %v", reqID, req, resp.Code, resp.Message, resp.Data)
	blog.V(5).Infof("reqID: %s, action: ListTemplateConfig, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateTemplateConfig implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateTemplateConfig(ctx context.Context,
	req *cmproto.UpdateTemplateConfigRequest, resp *cmproto.UpdateTemplateConfigResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := templateconfig.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateTemplateConfig", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateTemplateConfig, req %v, resp %v", reqID, req, resp)
	return nil
}
