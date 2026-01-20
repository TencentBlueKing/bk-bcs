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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/cloudvpc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// CreateCloudVPC implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CreateCloudVPC(ctx context.Context,
	req *cmproto.CreateCloudVPCRequest, resp *cmproto.CreateCloudVPCResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudvpc.NewCreateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("CreateCloudVPC", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CreateCloudVPC, req %v, resp %v", reqID, req, resp)
	return nil
}

// UpdateCloudVPC implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) UpdateCloudVPC(ctx context.Context,
	req *cmproto.UpdateCloudVPCRequest, resp *cmproto.UpdateCloudVPCResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudvpc.NewUpdateAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("UpdateCloudVPC", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: UpdateCloudVPC, req %v, resp %v", reqID, req, resp)
	return nil
}

// DeleteCloudVPC implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) DeleteCloudVPC(ctx context.Context,
	req *cmproto.DeleteCloudVPCRequest, resp *cmproto.DeleteCloudVPCResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudvpc.NewDeleteAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("DeleteCloudVPC", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: DeleteCloudVPC, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudVPC implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudVPC(ctx context.Context,
	req *cmproto.ListCloudVPCRequest, resp *cmproto.ListCloudVPCResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudvpc.NewListAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudVPC", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudVPC, req %v, resp.Code %d, resp.Message %s, resp.Data.Length %v",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudVPC, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudRegions implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudRegions(ctx context.Context,
	req *cmproto.ListCloudRegionsRequest, resp *cmproto.ListCloudRegionsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudvpc.NewListRegionsAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudRegions", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudRegions, req %v, resp.Code %d, resp.Message %s, resp.Data.Length %v",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudRegions, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetVPCCidr implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetVPCCidr(ctx context.Context,
	req *cmproto.GetVPCCidrRequest, resp *cmproto.GetVPCCidrResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudvpc.NewGetVPCCidrAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudRegions", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudRegions, req %v, resp.Code %d, resp.Message %s, resp.Data.Length %v",
		reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudRegions, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudSubnets implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudSubnets(ctx context.Context,
	req *cmproto.ListCloudSubnetsRequest, resp *cmproto.ListCloudSubnetsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudvpc.NewListSubnetsAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudSubnets", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudSubnets, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudSubnets, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudVpcs implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudVpcs(ctx context.Context,
	req *cmproto.ListCloudVpcsRequest, resp *cmproto.ListCloudVpcsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	fa := cloudvpc.NewListCloudVpcsAction(cm.model)
	fa.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudVpcs", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudVpcs, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudVpcs, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListCloudSecurityGroups implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListCloudSecurityGroups(ctx context.Context,
	req *cmproto.ListCloudSecurityGroupsRequest, resp *cmproto.ListCloudSecurityGroupsResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	ca := cloudvpc.NewListSecurityGroupsAction(cm.model)
	ca.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListCloudSecurityGroups", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListCloudSecurityGroups, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data.Length %v", reqID, req, resp.Code, resp.Message, len(resp.Data))
	blog.V(5).Infof("reqID: %s, action: ListCloudSecurityGroups, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetCloudBandwidthPackages implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetCloudBandwidthPackages(ctx context.Context,
	req *cmproto.GetCloudBandwidthPackagesRequest, resp *cmproto.GetCloudBandwidthPackagesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	ca := cloudvpc.NewGGetCloudBandwidthPackagesAction(cm.model)
	ca.Handle(ctx, req, resp)

	metrics.ReportAPIRequestMetric("GetCloudBandwidthPackages", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetCloudBandwidthPackages, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data %v", reqID, req, resp.Code, resp.Message, resp.Data) // nolint
	blog.V(5).Infof("reqID: %s, action: GetCloudBandwidthPackages, req %v, resp %v", reqID, req, resp)
	return nil
}

// CheckCidrConflictFromVpc implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) CheckCidrConflictFromVpc(ctx context.Context,
	req *cmproto.CheckCidrConflictFromVpcRequest, resp *cmproto.CheckCidrConflictFromVpcResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}

	start := time.Now()
	ca := cloudvpc.NewCheckCidrConflictAction(cm.model)
	ca.Handle(ctx, req, resp)

	metrics.ReportAPIRequestMetric("CheckCidrConflictFromVpc", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: CheckCidrConflictFromVpc, req %v, resp.Code %d, "+
		"resp.Message %s, resp.Data %v", reqID, req, resp.Code, resp.Message, resp.Data)
	blog.V(5).Infof("reqID: %s, action: CheckCidrConflictFromVpc, req %v, resp %v", reqID, req, resp)
	return nil
}

// GetMasterSuggestedMachines implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) GetMasterSuggestedMachines(ctx context.Context,
	req *cmproto.GetMasterSuggestedMachinesRequest, resp *cmproto.GetMasterSuggestedMachinesResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := cloudvpc.NewSuggestMasterMachinesAction(cm.model)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("GetMasterSuggestedMachines", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: GetMasterSuggestedMachines, req %v, resp %v", reqID, req, resp)
	return nil
}

// ListRecommendCloudVpcCidr implements interface cmproto.ClusterManagerServer
func (cm *ClusterManager) ListRecommendCloudVpcCidr(ctx context.Context,
	req *cmproto.ListRecommendCloudVpcCidrRequest, resp *cmproto.ListRecommendCloudVpcCidrResponse) error {
	reqID, err := requestIDFromContext(ctx)
	if err != nil {
		return err
	}
	start := time.Now()
	na := cloudvpc.NewListRecommendCloudVpcCidrAction(cm.model)
	na.Handle(ctx, req, resp)
	metrics.ReportAPIRequestMetric("ListRecommendCloudVpcCidr", "grpc", strconv.Itoa(int(resp.Code)), start)
	blog.Infof("reqID: %s, action: ListRecommendCloudVpcCidr, req %v, resp %v", reqID, req, resp)
	return nil
}
