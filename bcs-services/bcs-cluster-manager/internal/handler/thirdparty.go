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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/thirdparty"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// GetBkSopsTemplateList implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetBkSopsTemplateList(ctx context.Context,
	req *cmproto.GetBkSopsTemplateListRequest, resp *cmproto.GetBkSopsTemplateListResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewListTemplateListAction()
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetBkSopsTemplateList", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetBkSopsTemplateList, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: GetBkSopsTemplateList, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetBkSopsTemplateInfo implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetBkSopsTemplateInfo(ctx context.Context,
	req *cmproto.GetBkSopsTemplateInfoRequest, resp *cmproto.GetBkSopsTemplateInfoResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := thirdparty.NewGetTemplateInfoAction()
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetBkSopsTemplateInfo", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetBkSopsTemplateInfo, req %v, resp.Code %d, resp.Message %s",
		reqID, req, resp.Code, resp.Message)
	blog.V(5).Infof("reqID: %s, action: GetBkSopsTemplateInfo, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetInnerTemplateValues implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetInnerTemplateValues(ctx context.Context,
	req *cmproto.GetInnerTemplateValuesRequest, resp *cmproto.GetInnerTemplateValuesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ga := thirdparty.NewGetTemplateValuesAction(cm.model)
	ga.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetInnerTemplateValues", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetInnerTemplateValues, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: GetInnerTemplateValues, req %v, resp %v", reqID, req, resp)
	return nil
}

// DebugBkSopsTask implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DebugBkSopsTask(ctx context.Context,
	req *cmproto.DebugBkSopsTaskRequest, resp *cmproto.DebugBkSopsTaskResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	da := thirdparty.NewDebugBkSopsTaskAction(cm.model)
	da.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DebugBkSopsTask", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DebugBkSopsTask, req %v, resp.Code %d, resp.Message %s, resp.Data %v",
		reqID, req, resp.Code, resp.Message, resp.Data)
	blog.V(5).Infof("reqID: %s, action: DebugBkSopsTask, req %v, resp %v", reqID, req, resp)
	return nil
}

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

// GetProviderResourceUsage implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetProviderResourceUsage(ctx context.Context,
	req *cmproto.GetProviderResourceUsageRequest, resp *cmproto.GetProviderResourceUsageResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewGetProviderResourceUsageAction(cm.model)
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetProviderResourceUsage", "grpc",
		strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: GetProviderResourceUsage, req %v", reqID, req)
	return nil
}

// GetBatchCustomSetting implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetBatchCustomSetting(ctx context.Context,
	req *cmproto.GetBatchCustomSettingRequest, resp *cmproto.GetBatchCustomSettingResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewGetCustomSettingAction()
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetBatchCustomSetting", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: GetBatchCustomSetting, req %v", reqID, req)
	return nil
}

// GetBizTopologyHost implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetBizTopologyHost(ctx context.Context,
	req *cmproto.GetBizTopologyHostRequest, resp *cmproto.GetBizTopologyHostResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewGetBizInstanceTopoAction()
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetBizTopologyHost", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: GetBizTopologyHost, req %v", reqID, req)
	return nil
}

// GetTopologyNodes implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetTopologyNodes(ctx context.Context,
	req *cmproto.GetTopologyNodesRequest, resp *cmproto.GetTopologyNodesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewGetTopoNodesAction(cm.model)
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetTopologyNodes", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: GetTopologyNodes, req %v", reqID, req)
	return nil
}

// GetTopologyHostIdsNodes implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetTopologyHostIdsNodes(ctx context.Context,
	req *cmproto.GetTopologyHostIdsNodesRequest, resp *cmproto.GetTopologyHostIdsNodesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewGetTopologyHostIdsNodesAction()
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetTopologyHostIdsNodes", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: GetTopologyHostIdsNodes, req %v", reqID, req)
	return nil
}

// GetHostsDetails implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetHostsDetails(ctx context.Context,
	req *cmproto.GetHostsDetailsRequest, resp *cmproto.GetHostsDetailsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewGetHostsDetailsAction()
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetHostsDetailsAction", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: GetHostsDetailsAction, req %v", reqID, req)
	return nil
}

// GetScopeHostCheck implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetScopeHostCheck(ctx context.Context,
	req *cmproto.GetScopeHostCheckRequest, resp *cmproto.GetScopeHostCheckResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	la := thirdparty.NewGetScopeHostCheckAction()
	la.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetScopeHostCheck", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.V(3).Infof("reqID: %s, action: GetScopeHostCheck, req %v", reqID, req)
	return nil
}
