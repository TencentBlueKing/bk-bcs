/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package templatebinding

import (
	"errors"

	"bk-bscp/internal/dbsharding"
	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// UpdateAction action for update config template binding
type UpdateAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager
	sd    *dbsharding.ShardingDB

	req  *pb.UpdateConfigTemplateBindingReq
	resp *pb.UpdateConfigTemplateBindingResp
}

// NewUpdateAction create new UpdateAction
func NewUpdateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.UpdateConfigTemplateBindingReq, resp *pb.UpdateConfigTemplateBindingResp) *UpdateAction {
	action := &UpdateAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return error.
func (act *UpdateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *UpdateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *UpdateAction) Output() error {
	// do nothing.
	return nil
}

func (act *UpdateAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Templateid, "templateid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Appid, "appid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Versionid, "versionid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Commitid, "commitid"); err != nil {
		return err
	}

	if err := common.VerifyTemplateBindingParams(act.req.BindingParams); err != nil {
		return nil
	}

	if err := common.VerifyNormalName(act.req.Operator, "operator"); err != nil {
		return nil
	}

	return nil
}

func (act *UpdateAction) updateTemplateBinding() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigTemplateBinding{})

	ups := map[string]interface{}{
		"Versionid":     act.req.Versionid,
		"Commitid":      act.req.Commitid,
		"BindingParams": act.req.BindingParams,
		"State":         act.req.State,
		"LastModifyBy":  act.req.Operator,
	}

	exec := act.sd.DB().
		Model(&database.ConfigTemplateBinding{}).
		Where(&database.ConfigTemplateBinding{
			Bid:        act.req.Bid,
			Templateid: act.req.Templateid,
			Appid:      act.req.Appid,
		}).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "update config template binding failed, there is no config template binding fit in conditions"
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *UpdateAction) Do() error {
	// business sharding db
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// update config template binding
	if errCode, errMsg := act.updateTemplateBinding(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
