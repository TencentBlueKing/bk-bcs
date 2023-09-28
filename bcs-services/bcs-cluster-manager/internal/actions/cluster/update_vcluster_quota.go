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

func (ca *UpdateVirtualClusterQuotaAction) validate() error {
	err := ca.req.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (ca *UpdateVirtualClusterQuotaAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create virtual cluster request
func (ca *UpdateVirtualClusterQuotaAction) Handle(ctx context.Context, req *cmproto.UpdateVirtualClusterQuotaReq,
	resp *cmproto.UpdateVirtualClusterQuotaResp) {
	if req == nil || resp == nil {
		blog.Errorf("create virtual cluster failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	var err error

	// create validate cluster
	if err = ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	ca.cluster, err = actions.GetClusterInfoByClusterID(ca.model, ca.req.ClusterID)
	if err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	var nsInfo cmproto.NamespaceInfo
	err = utils.ToStringObject([]byte(ca.cluster.ExtraInfo[common.VClusterNamespaceInfo]), &nsInfo)
	if err != nil {
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// update quota
	err = ca.k8sOp.UpdateResourceQuota(ctx, ca.cluster.SystemID, clusterops.ResourceQuotaInfo{
		Name:        nsInfo.Name,
		CpuRequests: ca.req.Quota.CpuRequests,
		CpuLimits:   ca.req.Quota.CpuLimits,
		MemRequests: ca.req.Quota.MemoryRequests,
		MemLimits:   ca.req.Quota.MemoryLimits,
	})
	if err != nil {
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	nsInfo.Quota.CpuRequests = ca.req.Quota.CpuRequests
	nsInfo.Quota.CpuLimits = ca.req.Quota.CpuLimits
	nsInfo.Quota.MemoryRequests = ca.req.Quota.MemoryRequests
	nsInfo.Quota.MemoryLimits = ca.req.Quota.MemoryLimits

	ca.cluster.ExtraInfo[common.VClusterNamespaceInfo] = utils.ToJSONString(nsInfo)
	err = ca.model.UpdateCluster(ctx, ca.cluster)
	if err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
