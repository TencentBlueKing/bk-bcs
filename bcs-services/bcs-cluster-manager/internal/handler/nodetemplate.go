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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/nodetemplate"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateNodeTemplate implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateNodeTemplate(ctx context.Context,
	req *cmproto.CreateNodeTemplateRequest, resp *cmproto.CreateNodeTemplateResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodetemplate.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateNodeTemplate", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateNodeTemplate, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateNodeTemplate implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateNodeTemplate(ctx context.Context,
	req *cmproto.UpdateNodeTemplateRequest, resp *cmproto.UpdateNodeTemplateResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodetemplate.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateNodeTemplate", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateNodeTemplate, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteNodeTemplate implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteNodeTemplate(ctx context.Context,
	req *cmproto.DeleteNodeTemplateRequest, resp *cmproto.DeleteNodeTemplateResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodetemplate.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteNodeTemplate", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteNodeTemplate, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListNodeTemplate implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNodeTemplate(ctx context.Context,
	req *cmproto.ListNodeTemplateRequest, resp *cmproto.ListNodeTemplateResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := nodetemplate.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListNodeTemplate", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListNodeTemplate, req %v, resp.Code %d, r"+
		"esp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListNodeTemplate, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetNodeTemplate implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetNodeTemplate(ctx context.Context,
	req *cmproto.GetNodeTemplateRequest, resp *cmproto.GetNodeTemplateResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := nodetemplate.NewGetAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetNodeTemplate", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetNodeTemplate, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: GetNodeTemplate, req %v, resp %v", reqID, req, resp)
	return nil
}
