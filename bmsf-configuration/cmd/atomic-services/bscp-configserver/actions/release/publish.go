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
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// PublishAction publishes target release object.
type PublishAction struct {
	kit              kit.Kit
	viper            *viper.Viper
	authSvrCli       pbauthserver.AuthClient
	dataMgrCli       pbdatamanager.DataManagerClient
	gseControllerCli pbgsecontroller.GSEControllerClient

	req  *pb.PublishReleaseReq
	resp *pb.PublishReleaseResp

	release *pbcommon.Release
	app     *pbcommon.App
}

// NewPublishAction creates new PublishAction.
func NewPublishAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	gseControllerCli pbgsecontroller.GSEControllerClient,
	req *pb.PublishReleaseReq, resp *pb.PublishReleaseResp) *PublishAction {

	action := &PublishAction{
		kit:              kit,
		viper:            viper,
		authSvrCli:       authSvrCli,
		dataMgrCli:       dataMgrCli,
		gseControllerCli: gseControllerCli,
		req:              req,
		resp:             resp,
	}

	action.resp.Result = true
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PublishAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PublishAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Authorize checks the action authorization.
func (act *PublishAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *PublishAction) Output() error {
	// do nothing.
	return nil
}

func (act *PublishAction) verify() error {
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

func (act *PublishAction) authorize() (pbcommon.ErrCode, string) {
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

func (act *PublishAction) publishData() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.PublishReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
		Operator:  act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("PublishRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager PublishRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	// audit here on release published.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_PUBLISH),
		act.req.BizId, act.req.ReleaseId, act.kit.User, "")

	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) publish() (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.PublishReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
		Operator:  act.kit.User,
		Nice:      1,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("gsecontroller.callTimeout"))
	defer cancel()

	logger.V(4).Infof("PublishRelease[%s]| request to gse-controller, %+v", r.Seq, r)

	resp, err := act.gseControllerCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to gse-controller PublishRelease, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *PublishAction) queryRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("PublishRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	act.release = resp.Data
	return resp.Code, resp.Message
}

func (act *PublishAction) queryApp() (pbcommon.ErrCode, string) {
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

	logger.V(4).Infof("PublishRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	act.app = resp.Data

	return resp.Code, resp.Message
}

// Do makes the workflows of this action base on input messages.
func (act *PublishAction) Do() error {
	// query release.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// check current release state.
	if act.release.State != int32(pbcommon.ReleaseState_RS_INIT) &&
		act.release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) {
		return act.Err(pbcommon.ErrCode_E_CS_PUBLISHED_NOT_INIT_RELEASE,
			"can't publish the release which is not init/published state")
	}

	if act.release.AppId != act.req.AppId {
		return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, "can't publish release, inconsonant app_id")
	}

	// query app.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.release.State == int32(pbcommon.ReleaseState_RS_INIT) {
		// make release data published.
		if errCode, errMsg := act.publishData(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}

	// gsecontroller publish.
	if errCode, errMsg := act.publish(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	return nil
}
