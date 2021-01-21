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

package template

import (
	"context"
	"errors"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// DeleteAction action for delete config template.
type DeleteAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.DeleteConfigTemplateReq
	resp *pb.DeleteConfigTemplateResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB
}

// NewDeleteAction create new DeleteAction
func NewDeleteAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.DeleteConfigTemplateReq, resp *pb.DeleteConfigTemplateResp) *DeleteAction {
	action := &DeleteAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return error
func (act *DeleteAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages
func (act *DeleteAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *DeleteAction) Output() error {
	// do nothing.
	return nil
}

func (act *DeleteAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("template_id", act.req.TemplateId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *DeleteAction) deleteConfigTemplate() (pbcommon.ErrCode, string) {
	if len(act.req.TemplateId) == 0 {
		return pbcommon.ErrCode_E_DM_PARAMS_INVALID, "can't delete resource without template_id"
	}

	exec := act.tx.
		Limit(1).
		Where(&database.ConfigTemplate{BizID: act.req.BizId, TemplateID: act.req.TemplateId}).
		Delete(&database.ConfigTemplate{})

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_ROW_AFFECTED_ERR,
			"delete template failed, there is no template fit in conditions"
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *DeleteAction) deleteConfigTemplateVersions() (pbcommon.ErrCode, string) {
	if len(act.req.TemplateId) == 0 {
		return pbcommon.ErrCode_E_DM_PARAMS_INVALID, "can't delete resource without template_id"
	}

	// NOTE: delete batch versions.
	exec := act.tx.
		Where(&database.ConfigTemplateVersion{BizID: act.req.BizId, TemplateID: act.req.TemplateId}).
		Delete(&database.ConfigTemplateVersion{})

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *DeleteAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// delete config template.
	if errCode, errMsg := act.deleteConfigTemplate(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// delete config template versions.
	if errCode, errMsg := act.deleteConfigTemplateVersions(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// commit tx.
	if err := act.tx.Commit().Error; err != nil {
		act.tx.Rollback()
		return act.Err(pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, err.Error())
	}
	return nil
}
