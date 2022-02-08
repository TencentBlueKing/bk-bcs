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

package namespace

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction action for list namespace
type ListAction struct {
	ctx    context.Context
	model  store.ClusterManagerModel
	req    *cmproto.ListNamespaceReq
	resp   *cmproto.ListNamespaceResp
	nsList []*cmproto.Namespace
}

// NewListAction new list namespace action
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

func (la *ListAction) listQuotas(namespace, federationClusterID string) ([]*cmproto.ResourceQuota, error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"namespace":           namespace,
		"federationclusterid": federationClusterID,
	})
	quotas, err := la.model.ListQuota(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return nil, err
	}
	var retQuotaList []*cmproto.ResourceQuota
	for i := range quotas {
		retQuotaList = append(retQuotaList, &quotas[i])
	}
	return retQuotaList, nil
}

func (la *ListAction) listNamespaces() error {
	condM := operator.M{}
	if len(la.req.BusinessID) != 0 {
		condM["businessID"] = la.req.BusinessID
	}
	if len(la.req.ProjectID) != 0 {
		condM["projectID"] = la.req.ProjectID
	}
	if len(la.req.FederationClusterID) != 0 {
		condM["federationClusterID"] = la.req.FederationClusterID
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nsList, err := la.model.ListNamespace(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	for _, ns := range nsList {
		quotaList, err := la.listQuotas(ns.Name, ns.FederationClusterID)
		if err != nil {
			return err
		}
		retNs := &cmproto.Namespace{
			Name:                ns.Name,
			FederationClusterID: ns.FederationClusterID,
			ProjectID:           ns.ProjectID,
			BusinessID:          ns.BusinessID,
			Labels:              ns.Labels,
			MaxQuota:            ns.MaxQuota,
			CreateTime:          ns.CreateTime,
			UpdateTime:          ns.UpdateTime,
			QuotaList:           quotaList,
		}
		la.nsList = append(la.nsList, retNs)
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.nsList
}

// Handle handle list namespace request
func (la *ListAction) Handle(ctx context.Context, req *cmproto.ListNamespaceReq, resp *cmproto.ListNamespaceResp) {
	if req == nil || resp == nil {
		blog.Errorf("list namespace failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listNamespaces(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
