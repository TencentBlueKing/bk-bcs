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

package cluster

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// DeleteAction action for delete cluster
type DeleteAction struct {
	ctx       context.Context
	model     store.ClusterManagerModel
	req       *cmproto.DeleteClusterReq
	resp      *cmproto.DeleteClusterResp
	quotaList []types.NamespaceQuota
}

// NewDeleteAction delete cluster action
func NewDeleteAction(model store.ClusterManagerModel) *DeleteAction {
	return &DeleteAction{
		model: model,
	}
}

func (da *DeleteAction) validate() error {
	if err := da.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (da *DeleteAction) queryQuotas() error {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"clusterID": da.req.ClusterID,
	})
	quotaList, err := da.model.ListQuota(da.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	da.quotaList = quotaList
	return nil
}

func (da *DeleteAction) deleteCluster() error {
	if !da.req.IsForced {
		if err := da.queryQuotas(); err != nil {
			return fmt.Errorf("query quotas in delete cluster failed, err %s", err.Error())
		}
		if len(da.quotaList) != 0 {
			return fmt.Errorf("cannot delete cluster, there is quots in cluster")
		}
		if err := da.model.DeleteCluster(da.ctx, da.req.ClusterID); err != nil {
			return err
		}
		return nil
	}
	// delete all namespace quotas related to certain cluster
	if err := da.model.BatchDeleteQuotaByCluster(da.ctx, da.req.ClusterID); err != nil {
		return err
	}
	// delete cluster
	if err := da.model.DeleteCluster(da.ctx, da.req.ClusterID); err != nil {
		return err
	}
	return nil
}

func (da *DeleteAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == types.BcsErrClusterManagerSuccess)
}

// Handle delete cluster request
func (da *DeleteAction) Handle(ctx context.Context, req *cmproto.DeleteClusterReq, resp *cmproto.DeleteClusterResp) {
	if req == nil || resp == nil {
		blog.Errorf("delete cluster failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		da.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// delete cluster
	if err := da.deleteCluster(); err != nil {
		da.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	da.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)
	return
}
