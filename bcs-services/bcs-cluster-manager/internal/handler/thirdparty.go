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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/thirdparty"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// ListBKCloud implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListBKCloud(ctx context.Context,
	req *cmproto.ListBKCloudRequest, resp *cmproto.CommonListResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewListBKCloudAction()
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListBKCloud", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: ListBKCloud, req %v", reqID, req)
	return nil
}

// ListCCTopology implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCCTopology(ctx context.Context,
	req *cmproto.ListCCTopologyRequest, resp *cmproto.CommonResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewListCCTopologyAction(cm.model)
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCCTopology", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: ListCCTopology, req %v", reqID, req)
	return nil
}
