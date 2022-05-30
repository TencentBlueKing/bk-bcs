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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/cloudresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// GetCloudRegions implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetCloudRegions(ctx context.Context,
	req *cmproto.GetCloudRegionsRequest, resp *cmproto.GetCloudRegionsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudresource.NewGetCloudRegionsAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetCloudRegions", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetCloudRegions, req %v, resp.Code %d, resp.Message %s, resp.Data %v",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: GetCloudRegions, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetCloudRegionZones implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetCloudRegionZones(ctx context.Context,
	req *cmproto.GetCloudRegionZonesRequest, resp *cmproto.GetCloudRegionZonesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudresource.NewGetCloudRegionZonesAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetCloudRegionZones", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetCloudRegions, req %v, resp.Code %d, resp.Message %s, resp.Data %v",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: GetCloudRegionZones, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudRegionCluster implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudRegionCluster(ctx context.Context,
	req *cmproto.ListCloudRegionClusterRequest, resp *cmproto.ListCloudRegionClusterResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudresource.NewListCloudClusterAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudRegionCluster", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudRegionCluster, req %v, resp.Code %d, resp.Message %s, resp.Data %v",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudRegionCluster, req %v, resp %v", reqID, req, resp)
	return nil
}
