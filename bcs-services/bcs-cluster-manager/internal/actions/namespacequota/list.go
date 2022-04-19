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

package namespacequota

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction is action for list namespace resource quota
type ListAction struct {
	ctx       context.Context
	model     store.ClusterManagerModel
	req       *cmproto.ListNamespaceQuotaReq
	resp      *cmproto.ListNamespaceQuotaResp
	quotaList []*cmproto.ResourceQuota
}

// NewListAction create new action for list namespace quota
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (la *ListAction) listQuotas() error {
	condM := operator.M{}
	if len(la.req.Namespace) != 0 {
		condM["namespace"] = la.req.Namespace
	}
	if len(la.req.FederationClusterID) != 0 {
		condM["federationclusterid"] = la.req.FederationClusterID
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	quotaList, err := la.model.ListQuota(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	for i := range quotaList {
		la.quotaList = append(la.quotaList, &quotaList[i])
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.quotaList
}

// Handle handle list quota request
func (la *ListAction) Handle(ctx context.Context,
	req *cmproto.ListNamespaceQuotaReq, resp *cmproto.ListNamespaceQuotaResp) {
	if req == nil || resp == nil {
		blog.Errorf("list namespace quota failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listQuotas(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
