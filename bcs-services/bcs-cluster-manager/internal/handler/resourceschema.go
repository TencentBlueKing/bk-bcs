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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/resourceschema"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListResourceSchema implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListResourceSchema(ctx context.Context,
	req *cmproto.ListResourceSchemaRequest, resp *cmproto.CommonListResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := resourceschema.NewListAction(options.GetGlobalCMOptions().ResourceSchemaPath)
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListResourceSchema", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: ListResourceSchema, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetResourceSchema implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetResourceSchema(ctx context.Context,
	req *cmproto.GetResourceSchemaRequest, resp *cmproto.CommonResp) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := resourceschema.NewGetAction(options.GetGlobalCMOptions().ResourceSchemaPath)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetResourceSchema", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: GetResourceSchema, req %v, resp %s",
		reqID, req, utils.ToJSONString(resp))
	return nil
}
