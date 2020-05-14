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

package templateset

import (
	"errors"

	"bk-bscp/internal/dbsharding"
	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// UpdateAction action for update config template set
type UpdateAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager
	sd    *dbsharding.ShardingDB

	req  *pb.UpdateConfigTemplateSetReq
	resp *pb.UpdateConfigTemplateSetResp
}

// NewUpdateAction create new UpdateAction
func NewUpdateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.UpdateConfigTemplateSetReq, resp *pb.UpdateConfigTemplateSetResp) *UpdateAction {
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

	if err := common.VerifyID(act.req.Setid, "setid"); err != nil {
		return err
	}

	if err := common.VerifyNormalName(act.req.Name, "name"); err != nil {
		return err
	}

	if err := common.VerifyMemo(act.req.Memo); err != nil {
		return err
	}

	if err := common.VerifyNormalName(act.req.Operator, "operator"); err != nil {
		return err
	}

	return nil
}

func (act *UpdateAction) updateConfigTemplateSet() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigTemplateSet{})

	ups := map[string]interface{}{
		"Name":         act.req.Name,
		"Memo":         act.req.Memo,
		"LastModifyBy": act.req.Operator,
	}

	exec := act.sd.DB().
		Model(&database.ConfigTemplateSet{}).
		Where(&database.ConfigTemplateSet{
			Bid:   act.req.Bid,
			Setid: act.req.Setid,
			State: int32(pbcommon.ConfigTemplateSetState_CTSS_CREATED),
		}).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "update config template set failed, there is no config template set fit in conditions"
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

	// update config template set
	if errCode, errMsg := act.updateConfigTemplateSet(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
