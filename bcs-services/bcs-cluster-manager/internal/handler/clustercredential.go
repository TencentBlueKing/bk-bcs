/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
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
	start := time.Now()
	ca := clustercredac.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("listclustercredential", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: listclustercredential, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// GetClusterCredential implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetClusterCredential(ctx context.Context,
	req *cmproto.GetClusterCredentialReq, resp *cmproto.GetClusterCredentialResp) error {
	start := time.Now()
	ga := clustercredac.NewGetAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("getclustercredential", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: getclustercredential, req %v, resp %v", req.Seq, req, resp)
	return nil
}

// UpdateClusterCredential implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateClusterCredential(ctx context.Context,
	req *cmproto.UpdateClusterCredentialReq, resp *cmproto.UpdateClusterCredentialResp) error {
	start := time.Now()
	ua := clustercredac.NewUpdateAction(cm.model)
	ua.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("updateclustercredential", "grpc", strconv.Itoa(int(resp.ErrCode)), start)
	blog.V(5).Infof("seq: %d, action: updateclustercredential, req %v, resp %v", req.Seq, req, resp)
	return nil
}
