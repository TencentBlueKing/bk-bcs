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

package templatebind

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

// DeleteAction action for delete config template bind.
type DeleteAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.DeleteTemplateBindReq
	resp *pb.DeleteTemplateBindResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB

	templateBind database.TemplateBind
}

// NewDeleteAction create new DeleteAction
func NewDeleteAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.DeleteTemplateBindReq, resp *pb.DeleteTemplateBindResp) *DeleteAction {
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
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}

	if (len(act.req.TemplateId) == 0 || len(act.req.AppId) == 0) && len(act.req.CfgId) == 0 {
		return errors.New("can't delete resource without (template_id/app_id) or (cfg_id)")
	}

	if len(act.req.TemplateId) != 0 {
		if err = common.ValidateString("template_id", act.req.TemplateId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
	}
	if len(act.req.AppId) != 0 {
		if err = common.ValidateString("app_id", act.req.AppId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
	}
	if len(act.req.CfgId) != 0 {
		if err = common.ValidateString("cfg_id", act.req.CfgId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
	}

	return nil
}

func (act *DeleteAction) queryTemplateBind() (pbcommon.ErrCode, string) {
	err := act.tx.
		Where(&database.TemplateBind{
			BizID:      act.req.BizId,
			TemplateID: act.req.TemplateId,
			AppID:      act.req.AppId,
			CfgID:      act.req.CfgId,
		}).
		Last(&act.templateBind).Error

	if err == database.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "config template not found"
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *DeleteAction) deleteBindConfig() (pbcommon.ErrCode, string) {
	if len(act.templateBind.CfgID) == 0 {
		return pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN,
			"can't delete bind config, empty cfg_id from template bind"
	}

	exec := act.tx.
		Limit(1).
		Where(&database.Config{BizID: act.req.BizId, CfgID: act.templateBind.CfgID}).
		Delete(&database.Config{})

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *DeleteAction) deleteTemplateBind() (pbcommon.ErrCode, string) {
	if (len(act.req.TemplateId) == 0 || len(act.req.AppId) == 0) && len(act.req.CfgId) == 0 {
		return pbcommon.ErrCode_E_DM_PARAMS_INVALID,
			"can't delete resource without (template_id/app_id) or (cfg_id)"
	}

	// delete by cfg_id.
	if len(act.req.CfgId) != 0 {
		exec := act.tx.
			Limit(1).
			Where(&database.TemplateBind{BizID: act.req.BizId, CfgID: act.req.CfgId}).
			Delete(&database.TemplateBind{})

		if err := exec.Error; err != nil {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
	} else {
		// delete by template_id/app_id.
		exec := act.tx.
			Limit(1).
			Where(&database.TemplateBind{
				BizID:      act.req.BizId,
				TemplateID: act.req.TemplateId,
				AppID:      act.req.AppId,
			}).
			Delete(&database.TemplateBind{})

		if err := exec.Error; err != nil {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
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

	// query template bind.
	if errCode, errMsg := act.queryTemplateBind(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// delete bind config.
	if errCode, errMsg := act.deleteBindConfig(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// delete config template bind.
	if errCode, errMsg := act.deleteTemplateBind(); errCode != pbcommon.ErrCode_E_OK {
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
