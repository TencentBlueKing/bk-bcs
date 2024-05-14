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

package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// UpdateVirtualClusterQuotaAction action for update virtual cluster namespace quota
type UpdateVirtualClusterQuotaAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	k8sOp *clusterops.K8SOperator

	cluster *cmproto.Cluster

	req  *cmproto.UpdateVirtualClusterQuotaReq
	resp *cmproto.UpdateVirtualClusterQuotaResp
}

// NewUpdateVirtualClusterQuotaAction update virtual cluster namespace quota action
func NewUpdateVirtualClusterQuotaAction(model store.ClusterManagerModel,
	k8sOp *clusterops.K8SOperator) *UpdateVirtualClusterQuotaAction {
	return &UpdateVirtualClusterQuotaAction{
		model: model,
		k8sOp: k8sOp,
	}
}

func (ua *UpdateVirtualClusterQuotaAction) validate() error {
	err := ua.req.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (ua *UpdateVirtualClusterQuotaAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create virtual cluster request
func (ua *UpdateVirtualClusterQuotaAction) Handle(ctx context.Context, req *cmproto.UpdateVirtualClusterQuotaReq,
	resp *cmproto.UpdateVirtualClusterQuotaResp) {
	if req == nil || resp == nil {
		blog.Errorf("create virtual cluster failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	var err error

	// create validate cluster
	if err = ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	ua.cluster, err = actions.GetClusterInfoByClusterID(ua.model, ua.req.ClusterID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	var nsInfo cmproto.NamespaceInfo
	err = utils.ToStringObject([]byte(ua.cluster.ExtraInfo[common.VClusterNamespaceInfo]), &nsInfo)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// update quota
	err = ua.k8sOp.UpdateResourceQuota(ctx, ua.cluster.SystemID, clusterops.ResourceQuotaInfo{
		Name:        nsInfo.Name,
		CpuRequests: ua.req.Quota.CpuRequests,
		CpuLimits:   ua.req.Quota.CpuLimits,
		MemRequests: ua.req.Quota.MemoryRequests,
		MemLimits:   ua.req.Quota.MemoryLimits,
	})
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	nsInfo.Quota.CpuRequests = ua.req.Quota.CpuRequests
	nsInfo.Quota.CpuLimits = ua.req.Quota.CpuLimits
	nsInfo.Quota.MemoryRequests = ua.req.Quota.MemoryRequests
	nsInfo.Quota.MemoryLimits = ua.req.Quota.MemoryLimits

	ua.cluster.ExtraInfo[common.VClusterNamespaceInfo] = utils.ToJSONString(nsInfo)
	err = ua.model.UpdateCluster(ctx, ua.cluster)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
