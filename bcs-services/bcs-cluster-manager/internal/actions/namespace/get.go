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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// GetAction action for get cluster
type GetAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.GetNamespaceReq
	resp  *cmproto.GetNamespaceResp
	ns    *cmproto.Namespace
}

// NewGetAction create action for get namespace
func NewGetAction(model store.ClusterManagerModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

func (ga *GetAction) listQuotas(namespace, federationClusterID string) ([]*cmproto.ResourceQuota, error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"namespace":           namespace,
		"federationClusterID": federationClusterID,
	})
	quotas, err := ga.model.ListQuota(ga.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return nil, err
	}
	var retQuotaList []*cmproto.ResourceQuota
	now := time.Now().Format(time.RFC3339)
	for _, quota := range quotas {
		retQuotaList = append(retQuotaList, &cmproto.ResourceQuota{
			Namespace:           quota.Namespace,
			FederationClusterID: quota.FederationClusterID,
			ClusterID:           quota.ClusterID,
			Region:              quota.Region,
			ResourceQuota:       quota.ResourceQuota,
			CreateTime:          now,
			UpdateTime:          now,
		})
	}
	return retQuotaList, nil
}

func (ga *GetAction) getNamespace() error {
	ns, err := ga.model.GetNamespace(ga.ctx, ga.req.Name, ga.req.FederationClusterID)
	if err != nil {
		return err
	}
	quotaList, err := ga.listQuotas(ga.req.Name, ga.req.FederationClusterID)
	if err != nil {
		return err
	}
	ga.ns = &cmproto.Namespace{
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
	return nil
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ga.resp.Data = ga.ns
}

// Handle get namespace request
func (ga *GetAction) Handle(ctx context.Context, req *cmproto.GetNamespaceReq, resp *cmproto.GetNamespaceResp) {
	if req == nil || resp == nil {
		blog.Errorf("get namespace failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := req.Validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.getNamespace(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
