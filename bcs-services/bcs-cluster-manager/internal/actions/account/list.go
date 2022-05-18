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
 *
 */

package account

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction action for list online cloudAccount
type ListAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req              *cmproto.ListCloudAccountRequest
	resp             *cmproto.ListCloudAccountResponse
	cloudAccountList []*cmproto.CloudAccountInfo
}

// NewListAction create list action for cluster account list
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listCloudAccount() error {
	condM := make(operator.M)
	//! we don't setting bson tag in proto file
	//! all fields are in lowcase
	if len(la.req.CloudID) != 0 {
		condM["cloudid"] = la.req.CloudID
	}
	if len(la.req.AccountID) != 0 {
		condM["accountid"] = la.req.AccountID
	}
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	cloudAccounts, err := la.model.ListCloudAccount(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}

	// get cloud AccountInfo and relative cluster
	for i := range cloudAccounts {
		if !cloudAccounts[i].Enable {
			continue
		}
		clusterIDs, err := getRelativeClustersByAccountID(la.ctx, la.model, CloudAccount{
			CloudID:   cloudAccounts[i].CloudID,
			AccountID: cloudAccounts[i].AccountID,
		})
		if err != nil {
			blog.Errorf("getRelativeClustersByAccountID[%s] failed: %v", cloudAccounts[i].AccountID, err)
		}
		cloudAccounts[i].Account.SecretKey = shieldReturnCloudKey(cloudAccounts[i].Account.SecretKey)

		cloudAccountInfo := &cmproto.CloudAccountInfo{
			Account:  &cloudAccounts[i],
			Clusters: clusterIDs,
		}
		la.cloudAccountList = append(la.cloudAccountList, cloudAccountInfo)
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.cloudAccountList
}

// Handle handle list cluster account list
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListCloudAccountRequest, resp *cmproto.ListCloudAccountResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cloudAccount failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listCloudAccount(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// shieldReturnCloudKey return key by '***'
func shieldReturnCloudKey(key string) string {
	keyBytes := []byte(key)
	if len(keyBytes) <= 4 {
		return string(keyBytes)
	}
	size := len(keyBytes)

	resultKeys := make([]byte, 0)
	for i := range keyBytes {
		if i < 2 || i >= (size-2) {
			resultKeys = append(resultKeys, keyBytes[i])
			continue
		}

		resultKeys = append(resultKeys, '*')
	}

	return string(resultKeys)
}
