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

// CreateAction creates a new config template bind relation.
type CreateAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateTemplateBindReq
	resp *pb.CreateTemplateBindResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB

	configTemplate database.ConfigTemplate
}

// NewCreateAction create new CreateAction
func NewCreateAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateTemplateBindReq, resp *pb.CreateTemplateBindResp) *CreateAction {
	action := &CreateAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return error
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handle input message
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handle output message
func (act *CreateAction) Output() error {
	// do nothing
	return nil
}

func (act *CreateAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("template_id", act.req.TemplateId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cfg_id", act.req.CfgId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("creator", act.req.Creator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateInt32("state", act.req.State, 0, 1); err != nil {
		return err
	}
	return nil
}

func (act *CreateAction) queryConfigTemplate() (pbcommon.ErrCode, string) {
	err := act.tx.
		Where(&database.ConfigTemplate{BizID: act.req.BizId, TemplateID: act.req.TemplateId}).
		Last(&act.configTemplate).Error

	if err == database.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "config template not found"
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *CreateAction) createBindConfig() (pbcommon.ErrCode, string) {
	st := database.Config{
		BizID:         act.req.BizId,
		AppID:         act.req.AppId,
		CfgID:         act.req.CfgId,
		Name:          act.configTemplate.CfgName,
		Fpath:         act.configTemplate.CfgFpath,
		User:          act.configTemplate.User,
		UserGroup:     act.configTemplate.UserGroup,
		FilePrivilege: act.configTemplate.FilePrivilege,
		FileFormat:    act.configTemplate.FileFormat,
		FileMode:      act.configTemplate.FileMode,
		Creator:       act.req.Creator,
		Memo:          act.configTemplate.Memo,
		LastModifyBy:  act.req.Creator,
	}

	err := act.tx.
		Where(database.Config{
			BizID: act.req.BizId,
			AppID: act.req.AppId,
			Name:  act.configTemplate.CfgName,
			Fpath: act.configTemplate.CfgFpath,
		}).
		Assign(st).
		FirstOrCreate(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createTemplateBind() (pbcommon.ErrCode, string) {
	st := database.TemplateBind{
		TemplateID:   act.req.TemplateId,
		BizID:        act.req.BizId,
		CfgID:        act.req.CfgId,
		AppID:        act.req.AppId,
		Creator:      act.req.Creator,
		LastModifyBy: act.req.Creator,
		State:        act.req.State,
	}

	err := act.tx.
		Where(database.TemplateBind{
			BizID:      st.BizID,
			TemplateID: st.TemplateID,
			AppID:      st.AppID,
		}).
		Attrs(st).
		FirstOrCreate(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Data = &pb.CreateTemplateBindResp_RespData{CfgId: st.CfgID}

	if st.CfgID != act.req.CfgId || st.Creator != act.req.Creator ||
		st.LastModifyBy != act.req.Creator || st.State != act.req.State {
		return pbcommon.ErrCode_E_DM_ALREADY_EXISTS, "already bind"
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// query config template.
	if errCode, errMsg := act.queryConfigTemplate(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// create config template bind relation.
	if errCode, errMsg := act.createTemplateBind(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// create config.
	if errCode, errMsg := act.createBindConfig(); errCode != pbcommon.ErrCode_E_OK {
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
