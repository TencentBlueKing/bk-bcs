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

package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
)

// UpdateAction update action for cloud account
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateCloudAccountRequest
	resp  *cmproto.UpdateCloudAccountResponse
}

// NewUpdateAction create update action for cloud account
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) updateCloudAccount(destCloudAccount *cmproto.CloudAccount) error {
	timeStr := time.Now().UTC().Format(time.RFC3339)
	destCloudAccount.UpdateTime = timeStr
	destCloudAccount.Updater = ua.req.Updater

	if len(ua.req.AccountName) != 0 {
		destCloudAccount.AccountName = ua.req.AccountName
	}
	if len(ua.req.Desc) != 0 {
		destCloudAccount.Desc = ua.req.Desc
	}
	if ua.req.Enable != nil {
		destCloudAccount.Enable = ua.req.Enable.GetValue()
	}
	if len(ua.req.ProjectID) != 0 {
		destCloudAccount.ProjectID = ua.req.ProjectID
	}
	if ua.req.Account != nil {
		destCloudAccount.Account = ua.req.Account
	}

	return ua.model.UpdateCloudAccount(ua.ctx, destCloudAccount, false)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle update cloud account
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateCloudAccountRequest, resp *cmproto.UpdateCloudAccountResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update cloudAccount failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	destCloudAccount, err := ua.model.GetCloudAccount(ua.ctx, req.CloudID, req.AccountID, false)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find cloudAccount %s failed when pre-update checking, err %s", req.AccountID, err.Error())
		return
	}
	if err := ua.updateCloudAccount(destCloudAccount); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

// MigrateAction migrate action for cloud account
type MigrateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.MigrateCloudAccountRequest
	resp  *cmproto.MigrateCloudAccountResponse

	accountIDs []string
}

// NewMigrateAction create migrate action for cloud account
func NewMigrateAction(model store.ClusterManagerModel) *MigrateAction {
	return &MigrateAction{
		model: model,
	}
}

func (ma *MigrateAction) validate() error {
	err := ma.req.Validate()
	if err != nil {
		return err
	}

	if ma.req.GetAccountIDs() == "" && !ma.req.GetAll() {
		return fmt.Errorf("encry data is empty, please input correct data")
	}

	return nil
}

func (ma *MigrateAction) migrateCloudAccount() error {
	if ma.req.GetAccountIDs() != "" {
		ma.accountIDs = strings.Split(ma.req.GetAccountIDs(), ",")
		// handle special accountIDs

		for i := range ma.accountIDs {
			err := ma.migrateSingleAccountByID(ma.req.CloudID, ma.accountIDs[i])
			if err != nil {
				blog.Errorf("migrateCloudAccount[%s:%s] failed: %v", ma.req.CloudID, ma.accountIDs[i], err)
			}
		}

		return nil
	}

	if ma.req.GetAll() {
		condM := make(operator.M)
		condM["cloudid"] = ma.req.CloudID
		cond := operator.NewLeafCondition(operator.Eq, condM)

		cloudAccounts, err := ma.model.ListCloudAccount(ma.ctx, cond, &storeopt.ListOption{
			SkipDecrypt: true,
		})
		if err != nil {
			blog.Errorf("migrateCloudAccount ListCloudAccount failed: %v", err)
			return err
		}
		for i := range cloudAccounts {
			err = ma.migrateSingleAccountByID(ma.req.CloudID, cloudAccounts[i].AccountID)
			if err != nil {
				blog.Errorf("migrateCloudAccount[%s:%s] failed: %v", ma.req.CloudID, cloudAccounts[i].AccountID, err)
			}
		}
	}

	return nil
}

func (ma *MigrateAction) migrateSingleAccountByID(cloudID, accountID string) error {
	account, err := ma.model.GetCloudAccount(context.Background(), cloudID, accountID, true)
	if err != nil {
		return err
	}

	// 数据库原始明文数据转换为当前环境加密数据
	if ma.req.GetEncrypt() == nil {
		return ma.model.UpdateCloudAccount(context.Background(), account, false)
	}

	// 数据库原始数据解密后转换为当前环境加密数据
	encryptCli, err := encrypt.InitEncryptClient(ma.req.GetEncrypt())
	if err != nil {
		return err
	}

	err = util.DecryptCloudAccountData(encryptCli, account.Account)
	if err != nil {
		return err
	}

	// 当前环境加密
	return ma.model.UpdateCloudAccount(context.Background(), account, false)
}

func (ma *MigrateAction) setResp(code uint32, msg string) {
	ma.resp.Code = code
	ma.resp.Message = msg
	ma.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle migrate cloud account
func (ma *MigrateAction) Handle(
	ctx context.Context, req *cmproto.MigrateCloudAccountRequest, resp *cmproto.MigrateCloudAccountResponse) {

	if req == nil || resp == nil {
		blog.Errorf("migrate cloudAccount failed, req or resp is empty")
		return
	}
	ma.ctx = ctx
	ma.req = req
	ma.resp = resp

	if err := ma.validate(); err != nil {
		ma.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ma.migrateCloudAccount(); err != nil {
		ma.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// create operationLog
	accountIDs := ma.req.AccountIDs
	if ma.req.GetAll() {
		accountIDs = "all"
	}
	err := ma.model.CreateOperationLog(ma.ctx, &cmproto.OperationLog{
		ResourceType: common.Account.String(),
		ResourceID:   "",
		TaskID:       "",
		Message:      fmt.Sprintf("迁移云[%s]账号[%s]", ma.req.CloudID, accountIDs),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		blog.Errorf("MigrateCloudAccount[%s] CreateOperationLog failed: %v", ma.req.CloudID, err)
	}

	ma.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
