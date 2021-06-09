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
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// CreateAction creates a new config template object.
type CreateAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	authSvrCli pbauthserver.AuthClient

	req  *pb.CreateConfigTemplateReq
	resp *pb.CreateConfigTemplateResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB
}

// NewCreateAction create new CreateAction
func NewCreateAction(ctx context.Context, viper *viper.Viper,
	smgr *dbsharding.ShardingManager, authSvrCli pbauthserver.AuthClient,
	req *pb.CreateConfigTemplateReq, resp *pb.CreateConfigTemplateResp) *CreateAction {

	action := &CreateAction{ctx: ctx, viper: viper, smgr: smgr, authSvrCli: authSvrCli, req: req, resp: resp}

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
	if err = common.ValidateString("user", act.req.User, 0, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("user_group", act.req.UserGroup, 0, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("file_privilege", act.req.FilePrivilege,
		0, database.BSCPNORMALSTRLENLIMIT); err != nil {
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
	if err = common.ValidateString("memo", act.req.Memo, 0, database.BSCPLONGSTRLENLIMIT); err != nil {
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

func (act *CreateAction) createConfigTemplate() (pbcommon.ErrCode, string) {
	st := database.ConfigTemplate{
		TemplateID:    act.req.TemplateId,
		BizID:         act.req.BizId,
		Name:          act.req.Name,
		CfgName:       act.req.CfgName,
		CfgFpath:      act.req.CfgFpath,
		User:          act.req.User,
		UserGroup:     act.req.UserGroup,
		FilePrivilege: act.req.FilePrivilege,
		FileFormat:    act.req.FileFormat,
		FileMode:      act.req.FileMode,
		EngineType:    act.req.EngineType,
		Memo:          act.req.Memo,
		Creator:       act.req.Creator,
		LastModifyBy:  act.req.Creator,
		State:         act.req.State,
	}

	err := act.tx.
		Where(database.ConfigTemplate{BizID: st.BizID, Name: st.Name}).
		Attrs(st).
		FirstOrCreate(&st).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Data = &pb.CreateConfigTemplateResp_RespData{TemplateId: st.TemplateID}

	if st.TemplateID != act.req.TemplateId {
		return pbcommon.ErrCode_E_DM_ALREADY_EXISTS, "the config template with target name already exist"
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *CreateAction) createAuthPolicy() (pbcommon.ErrCode, string) {
	r := &pbauthserver.AddPolicyReq{
		Seq:      act.req.Seq,
		Metadata: &pbauthserver.AuthMetadata{V0: act.req.Creator, V1: act.req.TemplateId, V2: auth.LocalAuthAction},
	}

	ctx, cancel := context.WithTimeout(act.ctx, act.viper.GetDuration("authserver.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateConfigTemplate[%s]| request to authserver, %+v", r.Seq, r)

	resp, err := act.authSvrCli.AddPolicy(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, fmt.Sprintf("request to AuthServer AddPolicy, %+v", err)
	}
	return resp.Code, resp.Message
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

	// create config template.
	if errCode, errMsg := act.createConfigTemplate(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// add new auth policy.
	if errCode, errMsg := act.createAuthPolicy(); errCode != pbcommon.ErrCode_E_OK {
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
