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
	clustercredac "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/clustercredential"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// ListClusterCredential implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListClusterCredential(ctx context.Context,
	req *cmproto.ListClusterCredentialReq, resp *cmproto.ListClusterCredentialResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := clustercredac.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("listclustercredential", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: listclustercredential", reqID)
	blog.V(3).Infof("reqID: %s, action: listclustercredential, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetClusterCredential implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetClusterCredential(ctx context.Context,
	req *cmproto.GetClusterCredentialReq, resp *cmproto.GetClusterCredentialResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := clustercredac.NewGetAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("getclustercredential", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: getclustercredential", reqID)
	blog.V(3).Infof("reqID: %s, action: getclustercredential, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateClusterCredential implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateClusterCredential(ctx context.Context,
	req *cmproto.UpdateClusterCredentialReq, resp *cmproto.UpdateClusterCredentialResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ua := clustercredac.NewUpdateAction(cm.model)
	ua.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("updateclustercredential", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: updateclustercredential", reqID)
	blog.V(3).Infof("reqID: %s, action: updateclustercredential, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteClusterCredential implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteClusterCredential(ctx context.Context,
	req *cmproto.DeleteClusterCredentialReq, resp *cmproto.DeleteClusterCredentialResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	da := clustercredac.NewDeleteAction(cm.model)
	da.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("deleteclustercredential", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: deleteclustercredential", reqID)
	blog.V(3).Infof("reqID: %s, action: deleteclustercredential, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateClusterKubeConfig implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateClusterKubeConfig(ctx context.Context,
	req *cmproto.UpdateClusterKubeConfigReq, resp *cmproto.UpdateClusterKubeConfigResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	ca := clustercredac.NewUpdateKubeconfigAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateCloudKubeConfig", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateCloudKubeConfig, req %v, resp %v", reqID, req, resp)
	return nil
}
