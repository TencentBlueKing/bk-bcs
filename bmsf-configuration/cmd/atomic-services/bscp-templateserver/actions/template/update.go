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

	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/internal/audit"
	"bk-bscp/internal/authorization"
	"bk-bscp/internal/database"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// UpdateAction update target config template object.
type UpdateAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	authSvrCli pbauthserver.AuthClient
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.UpdateConfigTemplateReq
	resp *pb.UpdateConfigTemplateResp
}

// NewUpdateAction creates new UpdateAction
func NewUpdateAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.UpdateConfigTemplateReq, resp *pb.UpdateConfigTemplateResp) *UpdateAction {

	action := &UpdateAction{
		kit:        kit,
		viper:      viper,
		authSvrCli: authSvrCli,
		dataMgrCli: dataMgrCli,
		req:        req,
		resp:       resp,
	}

	action.resp.Result = true
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *UpdateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *UpdateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Authorize checks the action authorization.
func (act *UpdateAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
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
	if err = common.ValidateString("cfg_fpath", act.req.CfgFpath, 0, database.BSCPCFGFPATHLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("user", act.req.User, 0, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("user_group", act.req.UserGroup, 0, database.BSCPNAMELENLIMIT); err != nil {
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

	if act.req.EngineType != int32(pbcommon.TemplateEngineType_TET_NONE) {
		if act.req.FileMode != int32(pbcommon.ConfigFileMode_CFM_TEMPLATE) {
			return fmt.Errorf("invalid input data, file mode type must be %+v in template render mode",
				pbcommon.ConfigFileMode_CFM_TEMPLATE)
		}
	} else {
		if act.req.FileMode == int32(pbcommon.ConfigFileMode_CFM_TEMPLATE) {
			return fmt.Errorf("invalid input data, file mode type must not be %+v in no-render mode",
				pbcommon.ConfigFileMode_CFM_TEMPLATE)
		}
	}

	if err = common.ValidateString("memo", act.req.Memo, 0, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}

	return nil
}

func (act *UpdateAction) authorize() (pbcommon.ErrCode, string) {
	// check authorize resource at first, it may be deleted.
	if errCode, errMsg := act.queryConfigTemplate(); errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}

	// check resource authorization.
	isAuthorized, err := authorization.Authorize(act.kit, act.req.TemplateId, auth.LocalAuthAction,
		act.authSvrCli, act.viper.GetDuration("authserver.callTimeout"))
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, fmt.Sprintf("authorize failed, %+v", err)
	}

	if !isAuthorized {
		return pbcommon.ErrCode_E_NOT_AUTHORIZED, "not authorized"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *UpdateAction) queryConfigTemplate() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryConfigTemplateReq{
		Seq:        act.kit.Rid,
		BizId:      act.req.BizId,
		TemplateId: act.req.TemplateId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("UpdateConfigTemplate[%s]| request to DataManager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryConfigTemplate(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to DataManager QueryConfigTemplate, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *UpdateAction) updateConfigTemplate() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.UpdateConfigTemplateReq{
		Seq:           act.kit.Rid,
		BizId:         act.req.BizId,
		TemplateId:    act.req.TemplateId,
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
		Operator:      act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("UpdateConfigTemplate[%s]| request to DataManager, %+v", req.Seq, req)

	resp, err := act.dataMgrCli.UpdateConfigTemplate(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to DataManager UpdateConfigTemplate, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	// audit here on template updated.
	audit.Audit(int32(pbcommon.SourceType_ST_TEMPLATE), int32(pbcommon.SourceOpType_SOT_UPDATE),
		act.req.BizId, act.req.TemplateId, act.kit.User, "")

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *UpdateAction) Do() error {
	// update config template.
	if errCode, errMsg := act.updateConfigTemplate(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
