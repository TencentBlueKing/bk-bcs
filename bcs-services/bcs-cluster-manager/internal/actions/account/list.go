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
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction action for list online cloudAccount
type ListAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	iam   iam.PermClient

	req              *cmproto.ListCloudAccountRequest
	resp             *cmproto.ListCloudAccountResponse
	cloudAccountList []*cmproto.CloudAccountInfo
}

// NewListAction create list action for cluster account list
func NewListAction(model store.ClusterManagerModel, iam iam.PermClient) *ListAction {
	return &ListAction{
		model: model,
		iam:   iam,
	}
}

func (la *ListAction) listCloudAccount() error {
	condM := make(operator.M)
	// ! we don't setting bson tag in proto file
	// ! all fields are in lowcase
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

	accountList := make([]string, 0)
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
		cloudAccounts[i].Account, err = shieldCloudSecret(cloudAccounts[i].Account)
		if err != nil {
			blog.Errorf("shieldCloudSecret failed: %v", err)
		}

		cloudAccountInfo := &cmproto.CloudAccountInfo{
			Account:  &cloudAccounts[i],
			Clusters: clusterIDs,
		}
		accountList = append(accountList, cloudAccounts[i].AccountID)
		la.cloudAccountList = append(la.cloudAccountList, cloudAccountInfo)
	}

	// get accountList perm
	la.getAccountListPerm(accountList)

	return nil
}

func (la *ListAction) getAccountListPerm(accountList []string) {
	if len(accountList) == 0 {
		return
	}

	if la.req.ProjectID != "" && la.req.Operator != "" {
		v3Perm, err := GetProjectAccountsV3Perm(la.iam, actions.PermInfo{
			ProjectID: la.req.ProjectID,
			UserID:    la.req.Operator,
		}, accountList)
		if err != nil {
			blog.Errorf("listCluster GetProjectAccountsV3Perm failed: %v", err.Error())
		}
		la.resp.WebAnnotations = &cmproto.WebAnnotations{
			Perms: v3Perm,
		}
	}

	return
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

// shieldCloudSecret return secret by '***'
func shieldCloudSecret(account *cmproto.Account) (*cmproto.Account, error) {
	shield := func(key string) string {
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

	account.SecretKey = shield(account.SecretKey)
	account.ClientSecret = shield(account.ClientSecret)
	if account.ServiceAccountSecret != "" {
		sa := &api.GkeServiceAccount{}
		if err := json.Unmarshal([]byte(account.ServiceAccountSecret), sa); err != nil {
			return nil, err
		}
		shieldPrivateKey := shield(base64.StdEncoding.EncodeToString([]byte(sa.PrivateKey)))
		sa.PrivateKey = shieldPrivateKey
		shieldServiceAccountByte, err := json.Marshal(sa)
		if err != nil {
			return nil, err
		}
		account.ServiceAccountSecret = string(shieldServiceAccountByte)
	}
	return account, nil
}

// ListPermDataAction action for list permData account
type ListPermDataAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req              *cmproto.ListCloudAccountPermRequest
	resp             *cmproto.ListCloudAccountPermResponse
	cloudAccountList []*cmproto.CloudAccount
}

// NewListPermAction create list action for project account list
func NewListPermAction(model store.ClusterManagerModel) *ListPermDataAction {
	return &ListPermDataAction{
		model: model,
	}
}

func (la *ListPermDataAction) listCloudAccount() error {
	condEqual := make(operator.M)
	// ! we don't setting bson tag in proto file
	// ! all fields are in lowcase
	if len(la.req.ProjectID) != 0 {
		condEqual["projectid"] = la.req.ProjectID
	}
	if len(la.req.AccountName) != 0 {
		condEqual["accountname"] = la.req.AccountName
	}
	condE := operator.NewLeafCondition(operator.Eq, condEqual)

	condIN := make(operator.M)
	if len(la.req.AccountID) > 0 {
		condIN["accountid"] = la.req.AccountID
	}
	condI := operator.NewLeafCondition(operator.In, condIN)

	cond := operator.NewBranchCondition(operator.And, condE, condI)
	cloudAccounts, err := la.model.ListCloudAccount(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}

	// get cloud AccountInfo and relative cluster
	for i := range cloudAccounts {
		if !cloudAccounts[i].Enable {
			continue
		}
		cloudAccounts[i].Account.SecretKey = ""
		cloudAccounts[i].Account.SecretID = ""
		la.cloudAccountList = append(la.cloudAccountList, &cloudAccounts[i])
	}
	return nil
}

func (la *ListPermDataAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.cloudAccountList
}

func (la *ListPermDataAction) validate() error {
	err := la.req.Validate()
	if err != nil {
		return err
	}

	if la.req.ProjectID == "" && len(la.req.AccountID) == 0 && len(la.req.AccountName) == 0 {
		return fmt.Errorf("ListPermDataAction query parameter is empty")
	}
	return nil
}

// Handle handle list cluster account list
func (la *ListPermDataAction) Handle(
	ctx context.Context, req *cmproto.ListCloudAccountPermRequest, resp *cmproto.ListCloudAccountPermResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cloudAccount permData failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
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
