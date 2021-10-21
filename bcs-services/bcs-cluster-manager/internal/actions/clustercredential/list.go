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

package clustercredential

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// ListAction action for list online cluster credential
type ListAction struct {
	ctx                   context.Context
	model                 store.ClusterManagerModel
	req                   *cmproto.ListClusterCredentialReq
	resp                  *cmproto.ListClusterCredentialResp
	clusterCredentialList []*cmproto.ClusterCredential
}

// NewListAction create list action for cluster credential
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

func (la *ListAction) listClusterCredential() error {
	condM := make(operator.M)
	if len(la.req.ServerKey) != 0 {
		condM["serverKey"] = la.req.ServerKey
	}
	if len(la.req.ClusterID) != 0 {
		condM["clusterID"] = la.req.ClusterID
	}
	if len(la.req.ClientMode) != 0 {
		condM["clientModule"] = la.req.ClientMode
	}
	if len(la.req.ConnectMode) != 0 {
		condM["connectMode"] = la.req.ConnectMode
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	clusterCredentialList, err := la.model.ListClusterCredential(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	for _, clusterCred := range clusterCredentialList {
		la.clusterCredentialList = append(la.clusterCredentialList, &cmproto.ClusterCredential{
			ServerKey:     clusterCred.ServerKey,
			ClusterID:     clusterCred.ClusterID,
			ClientModule:  clusterCred.ClientModule,
			ServerAddress: clusterCred.ServerAddress,
			CaCertData:    clusterCred.CaCertData,
			UserToken:     clusterCred.UserToken,
			ClusterDomain: clusterCred.ClusterDomain,
			ConnectMode:   clusterCred.ConnectMode,
			CreateTime:    clusterCred.CreateTime.String(),
			UpdateTime:    clusterCred.UpdateTime.String(),
		})
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == types.BcsErrClusterManagerSuccess)
	la.resp.Data = la.clusterCredentialList
}

// Handle handle list cluster credential
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListClusterCredentialReq, resp *cmproto.ListClusterCredentialResp) {

	if req == nil || resp == nil {
		blog.Errorf("list cluster credentials failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listClusterCredential(); err != nil {
		la.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)
	return
}
