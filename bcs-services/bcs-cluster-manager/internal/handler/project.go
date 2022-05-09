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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateProject implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateProject(ctx context.Context,
	req *cmproto.CreateProjectRequest, resp *cmproto.CreateProjectResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := project.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateProject", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action CreateProject, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateProject implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateProject(ctx context.Context,
	req *cmproto.UpdateProjectRequest, resp *cmproto.UpdateProjectResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := project.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateProject", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateProject, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteProject implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteProject(ctx context.Context,
	req *cmproto.DeleteProjectRequest, resp *cmproto.DeleteProjectResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := project.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteProject", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteProject, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetProject implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetProject(ctx context.Context,
	req *cmproto.GetProjectRequest, resp *cmproto.GetProjectResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := project.NewGetAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetProject", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetProject, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: GetProject, req %v, resp %v",
		reqID, req, resp)
	return nil
}

// ListProject implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListProject(ctx context.Context,
	req *cmproto.ListProjectRequest, resp *cmproto.ListProjectResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := project.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListProject", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListProject, req %v, resp.Code %d, resp.Message %s, resp.Data.Length",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListProject, req %v, resp %v", reqID, req, resp)
	return nil
}
