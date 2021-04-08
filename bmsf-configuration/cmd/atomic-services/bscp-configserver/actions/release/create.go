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

package release

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
	pb "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// CreateAction creates a release object.
type CreateAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	authSvrCli pbauthserver.AuthClient
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateReleaseReq
	resp *pb.CreateReleaseResp

	commit *pbcommon.Commit
	config *pbcommon.Config

	strategies   string
	newReleaseID string
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateReleaseReq, resp *pb.CreateReleaseResp) *CreateAction {

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
	if err = common.ValidateString("commit_id", act.req.CommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("strategy_id", act.req.StrategyId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CreateAction) genReleaseID() error {
	id, err := common.GenReleaseID()
	if err != nil {
		return err
	}
	act.newReleaseID = id
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
	r := &pbdatamanager.QueryAppReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		AppId: act.req.AppId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateReleaseReq{
		Seq:           act.kit.Rid,
		BizId:         act.req.BizId,
		ReleaseId:     act.newReleaseID,
		Name:          act.req.Name,
		AppId:         act.commit.AppId,
		CfgId:         act.config.CfgId,
		CfgName:       act.config.Name,
		CfgFpath:      act.config.Fpath,
		User:          act.config.User,
		UserGroup:     act.config.UserGroup,
		FilePrivilege: act.config.FilePrivilege,
		FileFormat:    act.config.FileFormat,
		FileMode:      act.config.FileMode,
		StrategyId:    act.req.StrategyId,
		Strategies:    act.strategies,
		CommitId:      act.req.CommitId,
		Memo:          act.req.Memo,
		Creator:       act.kit.User,
		State:         int32(pbcommon.ReleaseState_RS_INIT),
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.CreateRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager CreateRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.resp.Data = &pb.CreateReleaseResp_RespData{ReleaseId: resp.Data.ReleaseId}
	act.newReleaseID = resp.Data.ReleaseId

	// audit here on release created.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.BizId, act.newReleaseID, act.kit.User, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryCommit() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryCommitReq{
		Seq:      act.kit.Rid,
		BizId:    act.req.BizId,
		CommitId: act.req.CommitId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryCommit, %+v", err)
	}
	act.commit = resp.Data
	return resp.Code, resp.Message
}

func (act *CreateAction) queryConfig() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryConfigReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		CfgId: act.commit.CfgId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryConfig(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryConfig, %+v", err)
	}
	act.config = resp.Data
	return resp.Code, resp.Message
}

func (act *CreateAction) queryStrategy() (pbcommon.ErrCode, string) {
	if len(act.req.StrategyId) == 0 {
		return pbcommon.ErrCode_E_OK, ""
	}

	r := &pbdatamanager.QueryStrategyReq{
		Seq:        act.kit.Rid,
		BizId:      act.req.BizId,
		StrategyId: act.req.StrategyId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryStrategy(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryStrategy, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	if resp.Data.AppId != act.commit.AppId {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, "can't create release with strategy which is not under target app"
	}
	act.strategies = resp.Data.Content

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// query commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// check commit state.
	if act.commit.State != int32(pbcommon.CommitState_CS_CONFIRMED) {
		return act.Err(pbcommon.ErrCode_E_CS_CREATE_RELEASE_WITH_UNCONFIRMED_COMMIT,
			"can't create release with the unconfirmed commit.")
	}

	// query config.
	if errCode, errMsg := act.queryConfig(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query strategies.
	if errCode, errMsg := act.queryStrategy(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if err := act.genReleaseID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error())
	}

	// create release.
	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
