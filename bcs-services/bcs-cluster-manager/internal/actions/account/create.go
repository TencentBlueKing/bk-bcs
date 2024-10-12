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

// Package account xxx
package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateAction action for create cloud account
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	iam   iam.PermClient

	cloud *cmproto.Cloud
	req   *cmproto.CreateCloudAccountRequest
	resp  *cmproto.CreateCloudAccountResponse
}

// NewCreateAction create cloudVPC action
func NewCreateAction(model store.ClusterManagerModel, iam iam.PermClient) *CreateAction {
	return &CreateAction{
		model: model,
		iam:   iam,
	}
}

func (ca *CreateAction) createCloudAccount() error {
	timeStr := time.Now().Format(time.RFC3339)
	accountID := generateAccountID(ca.cloud)
	cloudAccount := &cmproto.CloudAccount{
		CloudID:     ca.req.CloudID,
		ProjectID:   ca.req.ProjectID,
		AccountID:   accountID,
		AccountName: ca.req.AccountName,
		Desc:        ca.req.Desc,
		Account:     ca.req.Account,
		Enable:      ca.req.Enable.GetValue(),
		Creator:     ca.req.Creator,
		Updater:     ca.req.Creator,
		CreatTime:   timeStr,
		UpdateTime:  timeStr,
	}
	err := ca.model.CreateCloudAccount(ca.ctx, cloudAccount)
	if err != nil {
		return err
	}

	err = cloudaccount.NewBCSAccountPermClient(ca.iam).AuthorizeResourceCreatorPerm(ca.req.Creator, authutils.ResourceInfo{
		Type: string(cloudaccount.SysCloudAccount),
		ID:   accountID,
		Name: ca.req.AccountName,
	})
	if err != nil {
		blog.Errorf("createCloudAccount AuthorizeResourceCreatorPerm[%s:%s] failed: %v", accountID,
			ca.req.AccountName, err)
		return err
	}

	ca.resp.Data = cloudAccount
	ca.resp.Data.Account = nil

	return nil
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create cloud account request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateCloudAccountRequest, resp *cmproto.CreateCloudAccountResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create cloudAccount failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.createCloudAccount(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// create operationLog
	err := ca.model.CreateOperationLog(ca.ctx, &cmproto.OperationLog{
		ResourceType: common.Account.String(),
		ResourceID:   ca.resp.Data.AccountID,
		TaskID:       "",
		Message:      fmt.Sprintf("创建云账号信息[%s]", ca.resp.Data.AccountID),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ResourceName: ca.req.AccountName,
	})
	if err != nil {
		blog.Errorf("CreateCloudAccount[%s] CreateOperationLog failed: %v", ca.resp.Data.AccountID, err)
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func (ca *CreateAction) validate() error {
	err := ca.req.Validate()
	if err != nil {
		return err
	}
	err = ca.checkCloudAccountName()
	if err != nil {
		return err
	}

	ca.cloud, err = actions.GetCloudByCloudID(ca.model, ca.req.CloudID)
	if err != nil {
		return err
	}

	vm, err := cloudprovider.GetCloudValidateMgr(ca.cloud.CloudProvider)
	if err != nil {
		return err
	}
	err = vm.ImportCloudAccountValidate(ca.req.Account)
	if err != nil {
		return err
	}

	return nil
}

func (ca *CreateAction) checkCloudAccountName() error {
	accounts, err := getAllAccountsByCloudID(ca.ctx, ca.model, ca.req.CloudID)
	if err != nil {
		return err
	}

	for i := range accounts {
		if ca.req.AccountName == accounts[i].AccountName {
			return fmt.Errorf("cloud[%s] Account[%s] duplicate", ca.req.CloudID, ca.req.AccountName)
		}
	}

	return nil
}

// generate random accountID
func generateAccountID(cloud *cmproto.Cloud) string {
	randomStr := utils.RandomString(8)

	return fmt.Sprintf("BCS-%s-%s", cloud.CloudID, randomStr)
}
