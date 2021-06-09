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

// CancelAction cancels target release object.
type CancelAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	authSvrCli pbauthserver.AuthClient
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CancelReleaseReq
	resp *pb.CancelReleaseResp

	release *pbcommon.Release
}

// NewCancelAction creates new CancelAction.
func NewCancelAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CancelReleaseReq, resp *pb.CancelReleaseResp) *CancelAction {

	action := &CancelAction{
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
func (act *CancelAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CancelAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Authorize checks the action authorization.
func (act *CancelAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *CancelAction) Output() error {
	// do nothing.
	return nil
}

func (act *CancelAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("release_id", act.req.ReleaseId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CancelAction) authorize() (pbcommon.ErrCode, string) {
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

func (act *CancelAction) queryApp() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAppReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		AppId: act.req.AppId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CancelRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *CancelAction) query() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CancelAction[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.release = resp.Data

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CancelAction) cancel() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CancelReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
		Operator:  act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CancelRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.CancelRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager CancelRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	// audit here on release canceled.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_CANCEL),
		act.req.BizId, act.req.ReleaseId, act.kit.User, "")

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CancelAction) Do() error {
	// query release.
	if errCode, errMsg := act.query(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.release.State == int32(pbcommon.ReleaseState_RS_CANCELED) {
		// already canceled.
		return nil
	}

	if act.release.State == int32(pbcommon.ReleaseState_RS_PUBLISHED) {
		// already published.
		return act.Err(pbcommon.ErrCode_E_CS_RELEASE_ALREADY_PUBLISHED,
			"can't cancel the release, already published")
	}

	if act.release.State == int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
		// already rollbacked.
		return act.Err(pbcommon.ErrCode_E_CS_RELEASE_ALREADY_ROLLBACKED,
			"can't cancel the release, already rollbacked")
	}

	// cancel release.
	if errCode, errMsg := act.cancel(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
