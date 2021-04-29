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
	"bk-bscp/pkg/logger"
)

// UpdateAction action for update config template.
type UpdateAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.UpdateConfigTemplateReq
	resp *pb.UpdateConfigTemplateResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB

	configTemplate   database.ConfigTemplate
	templateBindList []database.TemplateBind
}

// NewUpdateAction create new UpdateAction.
func NewUpdateAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.UpdateConfigTemplateReq, resp *pb.UpdateConfigTemplateResp) *UpdateAction {
	action := &UpdateAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return error.
func (act *UpdateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
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
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("template_id", act.req.TemplateId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("name", act.req.Name,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cfg_name", act.req.CfgName,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	act.req.CfgFpath = common.ParseFpath(act.req.CfgFpath)
	if err = common.ValidateString("cfg_fpath", act.req.CfgFpath, 0,
		database.BSCPCFGFPATHLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("user", act.req.User, 0,
		database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("user_group", act.req.UserGroup, 0,
		database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("file_privilege", act.req.FilePrivilege, 0,
		database.BSCPNORMALSTRLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateStrings("file_format", act.req.FileFormat, "", "unix", "windows"); err != nil {
		return err
	}
	if err = common.ValidateInt32("file_mode", int32(act.req.FileMode),
		int32(pbcommon.ConfigFileMode_CFM_TEXT), int32(pbcommon.ConfigFileMode_CFM_TEMPLATE)); err != nil {
		return err
	}
	if err = common.ValidateInt32("engine_type", int32(act.req.EngineType),
		int32(pbcommon.TemplateEngineType_TET_NONE), int32(pbcommon.TemplateEngineType_TET_EXTERNAL)); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo, 0,
		database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateInt32("state", act.req.State, 0, 1); err != nil {
		return err
	}
	return nil
}

func (act *UpdateAction) queryTemplateBindList(start, limit int) ([]database.TemplateBind, pbcommon.ErrCode, string) {
	templateBinds := []database.TemplateBind{}

	err := act.tx.
		Offset(start).Limit(limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.TemplateBind{BizID: act.req.BizId, TemplateID: act.req.TemplateId}).
		Find(&templateBinds).Error

	if err != nil {
		return nil, pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return templateBinds, pbcommon.ErrCode_E_OK, "OK"
}

func (act *UpdateAction) updateConfig(appID, cfgID string) (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"Name":          act.req.CfgName,
		"Fpath":         act.req.CfgFpath,
		"User":          act.req.User,
		"UserGroup":     act.req.UserGroup,
		"FilePrivilege": act.req.FilePrivilege,
		"FileFormat":    act.req.FileFormat,
		"FileMode":      act.req.FileMode,
		"Memo":          act.req.Memo,
		"LastModifyBy":  act.req.Operator,
	}

	exec := act.tx.
		Model(&database.Config{}).
		Where(&database.Config{
			BizID: act.req.BizId,
			AppID: appID,
			CfgID: cfgID,
		}).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "no update for the config"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *UpdateAction) updateBindConfigs() (pbcommon.ErrCode, string) {
	if !act.needSyncBindConfigs() {
		return pbcommon.ErrCode_E_OK, "OK"
	}

	start := 0
	limit := database.BSCPQUERYLIMITMB

	for {
		templateBinds, errCode, errMsg := act.queryTemplateBindList(start, limit)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}

		for _, bind := range templateBinds {
			errCode, errMsg = act.updateConfig(bind.AppID, bind.CfgID)
			if errCode != pbcommon.ErrCode_E_OK {
				logger.Errorf("UpdateConfigTemplate[%s]| update bind config[%+v] failed, %+v, %s",
					act.req.Seq, bind, errCode, errMsg)
				return errCode, errMsg
			}
		}

		if len(templateBinds) < limit {
			break
		}
		start += len(templateBinds)
	}

	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *UpdateAction) queryConfigTemplate() (pbcommon.ErrCode, string) {
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

func (act *UpdateAction) needSyncBindConfigs() bool {
	if act.req.CfgName != act.configTemplate.CfgName {
		return true
	}
	if act.req.CfgFpath != act.configTemplate.CfgFpath {
		return true
	}
	if act.req.User != act.configTemplate.User {
		return true
	}
	if act.req.UserGroup != act.configTemplate.UserGroup {
		return true
	}
	if act.req.FilePrivilege != act.configTemplate.FilePrivilege {
		return true
	}
	if act.req.FileFormat != act.configTemplate.FileFormat {
		return true
	}
	if act.req.FileMode != act.configTemplate.FileMode {
		return true
	}
	return false
}

func (act *UpdateAction) updateConfigTemplate() (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"Name":          act.req.Name,
		"CfgName":       act.req.CfgName,
		"CfgFpath":      act.req.CfgFpath,
		"User":          act.req.User,
		"UserGroup":     act.req.UserGroup,
		"FilePrivilege": act.req.FilePrivilege,
		"FileFormat":    act.req.FileFormat,
		"FileMode":      act.req.FileMode,
		"EngineType":    act.req.EngineType,
		"Memo":          act.req.Memo,
		"State":         act.req.State,
		"LastModifyBy":  act.req.Operator,
	}

	exec := act.tx.
		Model(&database.ConfigTemplate{}).
		Where(&database.ConfigTemplate{BizID: act.req.BizId, TemplateID: act.req.TemplateId}).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "no update for the config template"
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *UpdateAction) Do() error {
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

	// update config template.
	if errCode, errMsg := act.updateConfigTemplate(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// update bind configs.
	if errCode, errMsg := act.updateBindConfigs(); errCode != pbcommon.ErrCode_E_OK {
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
