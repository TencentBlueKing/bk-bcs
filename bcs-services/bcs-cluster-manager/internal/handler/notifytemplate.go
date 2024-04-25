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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/notifytemplate"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateNotifyTemplate implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateNotifyTemplate(ctx context.Context,
	req *cmproto.CreateNotifyTemplateRequest, resp *cmproto.CreateNotifyTemplateResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := notifytemplate.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateNotifyTemplate", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateNotifyTemplate, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteNotifyTemplate implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteNotifyTemplate(ctx context.Context,
	req *cmproto.DeleteNotifyTemplateRequest, resp *cmproto.DeleteNotifyTemplateResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := notifytemplate.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteNotifyTemplate", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteNotifyTemplate, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListNotifyTemplate implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListNotifyTemplate(ctx context.Context,
	req *cmproto.ListNotifyTemplateRequest, resp *cmproto.ListNotifyTemplateResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := notifytemplate.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListNotifyTemplate", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListNotifyTemplate, req %v, resp.Code %d, r"+
		"esp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListNotifyTemplate, req %v, resp %v", reqID, req, resp)
	return nil
}
