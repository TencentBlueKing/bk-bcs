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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/cloudresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// GetServiceRoles implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetServiceRoles(ctx context.Context,
	req *cmproto.GetServiceRolesRequest, resp *cmproto.GetServiceRolesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := cloudresource.NewGetServiceRolesAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetServiceRoles", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetServiceRoles, req %v, resp.Code %d, resp.Message %s, resp.Data %v",
		reqID, req, resp.Code, resp.Message, resp.Data)
	blog.V(5).Infof("reqID: %s, action: GetServiceRoles, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetResourceGroups implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetResourceGroups(ctx context.Context,
	req *cmproto.GetResourceGroupsRequest, resp *cmproto.GetResourceGroupsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := cloudresource.NewGetResourceGroupsAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetResourceGroups", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetResourceGroups, req %v, resp.Code %d, resp.Message %s, resp.Data %v",
		reqID, req, resp.Code, resp.Message, resp.Data)
	blog.V(5).Infof("reqID: %s, action: GetResourceGroups, req %v, resp %v", reqID, req, resp)
	return nil
}

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
		reqID, req, resp.Code, resp.Message, resp.Data)
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
		reqID, req, resp.Code, resp.Message, resp.Data)
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
		reqID, req, resp.Code, resp.Message, resp.Data)
	blog.V(5).Infof("reqID: %s, action: ListCloudRegionCluster, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudInstanceTypes implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudInstanceTypes(ctx context.Context,
	req *cmproto.ListCloudInstanceTypeRequest, resp *cmproto.ListCloudInstanceTypeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	fa := cloudresource.NewListNodeTypeAction(cm.model)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudInstanceTypes", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudInstanceTypes, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudInstanceTypes, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudInstances implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudInstances(ctx context.Context,
	req *cmproto.ListCloudInstancesRequest, resp *cmproto.ListCloudInstancesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	fa := cloudresource.NewListCloudInstancesAction(cm.model)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudInstances", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudInstances, req %v, resp.Code %d, resp.Message %s, resp.Data.Length: %v",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudInstances, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudOsImage implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudOsImage(ctx context.Context,
	req *cmproto.ListCloudOsImageRequest, resp *cmproto.ListCloudOsImageResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	fa := cloudresource.NewListCloudOsImageAction(cm.model)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudOsImage", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudOsImage, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudOsImage, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudRuntimeInfo implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudRuntimeInfo(ctx context.Context,
	req *cmproto.ListCloudRuntimeInfoRequest, resp *cmproto.ListCloudRuntimeInfoResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	fa := cloudresource.NewListCloudRuntimeInfoAction(cm.model)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudRuntimeInfo", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudRuntimeInfo, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudRuntimeInfo, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudProjects implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudProjects(ctx context.Context,
	req *cmproto.ListCloudProjectsRequest, resp *cmproto.ListCloudProjectsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	fa := cloudresource.NewListCloudProjectsAction(cm.model)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudProjects", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudProjects, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudProjects, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListKeypairs implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListKeypairs(ctx context.Context,
	req *cmproto.ListKeyPairsRequest, resp *cmproto.ListKeyPairsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudresource.NewListKeyPairsAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListKeypairs", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListKeypairs, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListKeypairs, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetCloudAccountType implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetCloudAccountType(ctx context.Context,
	req *cmproto.GetCloudAccountTypeRequest, resp *cmproto.GetCloudAccountTypeResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	ca := cloudresource.NewGetCloudAccountTypeAction(cm.model)
	ca.Handle(ctx, req, resp)

	metrics.ReportAPIRequestMetric("GetCloudAccountType", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetCloudAccountType, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data %v", reqID, req, resp.Code, resp.Message, resp.Data)
	blog.V(5).Infof("reqID: %s, action: GetCloudAccountType, req %v, resp %v", reqID, req, resp)
	return nil
}
