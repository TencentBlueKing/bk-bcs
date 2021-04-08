/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package strategy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/internal/audit"
	"bk-bscp/internal/authorization"
	"bk-bscp/internal/database"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// CreateAction creates a strategy object.
type CreateAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	authSvrCli pbauthserver.AuthClient
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateStrategyReq
	resp *pb.CreateStrategyResp

	app *pbcommon.App

	strategies    *strategy.Strategy
	newStrategyID string

	labelsOr  []map[string]string
	labelsAnd []map[string]string
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateStrategyReq, resp *pb.CreateStrategyResp) *CreateAction {

	action := &CreateAction{
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

	action.labelsOr = []map[string]string{}
	action.labelsAnd = []map[string]string{}

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}

	for _, labelsOr := range act.req.LabelsOr {
		if err := strategy.ValidateLabels(labelsOr.Labels); err != nil {
			return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, fmt.Sprintf("invalid labels_or formats, %+v", err))
		}
		if len(labelsOr.Labels) != 0 {
			act.labelsOr = append(act.labelsOr, labelsOr.Labels)
		}
	}

	for _, labelsAnd := range act.req.LabelsAnd {
		if err := strategy.ValidateLabels(labelsAnd.Labels); err != nil {
			return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, fmt.Sprintf("invalid labels_and formats, %+v", err))
		}
		if len(labelsAnd.Labels) != 0 {
			act.labelsAnd = append(act.labelsAnd, labelsAnd.Labels)
		}
	}

	return nil
}

// Authorize checks the action authorization.
func (act *CreateAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("name", act.req.Name,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CreateAction) genStrategyID() error {
	id, err := common.GenStrategyID()
	if err != nil {
		return err
	}
	act.newStrategyID = id
	return nil
}

func (act *CreateAction) authorize() (pbcommon.ErrCode, string) {
	// check authorize resource at first, it may be deleted.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}

	// check resource authorization.
	isAuthorized, err := authorization.Authorize(act.kit, act.req.AppId, auth.LocalAuthAction,
		act.authSvrCli, act.viper.GetDuration("authserver.callTimeout"))
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("authorize failed, %+v", err)
	}

	if !isAuthorized {
		return pbcommon.ErrCode_E_NOT_AUTHORIZED, "not authorized"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryApp() (pbcommon.ErrCode, string) {
	if act.app != nil {
		return pbcommon.ErrCode_E_OK, ""
	}

	r := &pbdatamanager.QueryAppReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		AppId: act.req.AppId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateStrategy[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	act.app = resp.Data

	return resp.Code, resp.Message
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	act.strategies = &strategy.Strategy{
		LabelsOr:  act.labelsOr,
		LabelsAnd: act.labelsAnd,
	}

	content, err := json.Marshal(act.strategies)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("can't marshal strategy content, %+v", err)
	}

	r := &pbdatamanager.CreateStrategyReq{
		Seq:        act.kit.Rid,
		BizId:      act.req.BizId,
		AppId:      act.req.AppId,
		StrategyId: act.newStrategyID,
		Name:       act.req.Name,
		Content:    string(content),
		Memo:       act.req.Memo,
		Creator:    act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateStrategy[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.CreateStrategy(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager CreateStrategy, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK && resp.Code != pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return resp.Code, resp.Message
	}
	act.resp.Data = &pb.CreateStrategyResp_RespData{StrategyId: resp.Data.StrategyId}
	act.newStrategyID = resp.Data.StrategyId

	if resp.Code == pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return pbcommon.ErrCode_E_CS_ALREADY_EXISTS, resp.Message
	}

	// audit here on new strategy created.
	audit.Audit(int32(pbcommon.SourceType_ST_STRATEGY), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.BizId, act.newStrategyID, act.kit.User, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	if err := act.genStrategyID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error())
	}

	// query app.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create strategy.
	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
