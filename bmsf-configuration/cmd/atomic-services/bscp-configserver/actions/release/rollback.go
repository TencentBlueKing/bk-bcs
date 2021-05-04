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

// RollbackAction rollbacks target release.
type RollbackAction struct {
	kit              kit.Kit
	viper            *viper.Viper
	authSvrCli       pbauthserver.AuthClient
	dataMgrCli       pbdatamanager.DataManagerClient
	gseControllerCli pbgsecontroller.GSEControllerClient

	req  *pb.RollbackReleaseReq
	resp *pb.RollbackReleaseResp

	// current release which wanted to be rollbacked.
	currentRelease *pbcommon.Release

	// app informations.
	app *pbcommon.App

	// re-publish target release(newReleaseid), and newRePubReleaseid is the
	// new release id. It's empty if there is no newReleaseid in request.
	newRePubReleaseID string

	// release which wanted to be re-published.
	newRePubRelease *pbcommon.Release

	// re-publish flag.
	isReleaseRePublished bool
}

// NewRollbackAction creates new RollbackAction.
func NewRollbackAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	gseControllerCli pbgsecontroller.GSEControllerClient,
	req *pb.RollbackReleaseReq, resp *pb.RollbackReleaseResp) *RollbackAction {

	action := &RollbackAction{
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
func (act *RollbackAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *RollbackAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Authorize checks the action authorization.
func (act *RollbackAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *RollbackAction) Output() error {
	// do nothing.
	return nil
}

func (act *RollbackAction) verify() error {
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
	if err = common.ValidateString("new_release_id", act.req.NewReleaseId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *RollbackAction) genReleaseID() error {
	id, err := common.GenReleaseID()
	if err != nil {
		return err
	}
	act.newRePubReleaseID = id
	return nil
}

func (act *RollbackAction) authorize() (pbcommon.ErrCode, string) {
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

func (act *RollbackAction) createRelease() (pbcommon.ErrCode, string) {
	newReleaseName := fmt.Sprintf("Rollback-%s", act.newRePubRelease.Name)

	if act.currentRelease.AppId != act.newRePubRelease.AppId ||
		act.currentRelease.CfgId != act.newRePubRelease.CfgId {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, "not the same app-config"
	}

	r := &pbdatamanager.CreateReleaseReq{
		Seq:           act.kit.Rid,
		BizId:         act.req.BizId,
		ReleaseId:     act.newRePubReleaseID,
		Name:          newReleaseName,
		AppId:         act.newRePubRelease.AppId,
		CfgId:         act.newRePubRelease.CfgId,
		CfgName:       act.newRePubRelease.CfgName,
		CfgFpath:      act.newRePubRelease.CfgFpath,
		User:          act.newRePubRelease.User,
		UserGroup:     act.newRePubRelease.UserGroup,
		FilePrivilege: act.newRePubRelease.FilePrivilege,
		FileFormat:    act.newRePubRelease.FileFormat,
		FileMode:      act.newRePubRelease.FileMode,
		StrategyId:    act.newRePubRelease.StrategyId,
		Strategies:    act.newRePubRelease.Strategies,
		CommitId:      act.newRePubRelease.CommitId,
		Memo:          act.newRePubRelease.Memo,
		Creator:       act.kit.User,
		State:         int32(pbcommon.ReleaseState_RS_INIT),
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("RollbackRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.CreateRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager CreateRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.resp.Data.ReleaseId = act.newRePubReleaseID

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) queryRelease(releaseID string) (*pbcommon.Release, pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: releaseID,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("RollbackRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	return resp.Data, resp.Code, resp.Message
}

func (act *RollbackAction) publishData() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.PublishReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.newRePubReleaseID,
		Operator:  act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("RollbackRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager PublishRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	// audit here on release published.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_PUBLISH),
		act.req.BizId, act.newRePubReleaseID, act.kit.User, "ROLLBACK-REPUB")

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) publish() (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.PublishReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.newRePubReleaseID,
		Operator:  act.kit.User,
		Nice:      1,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("gsecontroller.callTimeout"))
	defer cancel()

	logger.V(4).Infof("RollbackRelease[%s]| request to gse-controller, %+v", r.Seq, r)

	resp, err := act.gseControllerCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to gse-controller PublishRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) rollbackData() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.RollbackReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
		Operator:  act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("RollbackRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.RollbackRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager RollbackRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	// audit here on release rollbacked.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_ROLLBACK),
		act.req.BizId, act.req.ReleaseId, act.kit.User, "")

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) rollback() (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.RollbackReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
		Operator:  act.kit.User,
		Nice:      1,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("gsecontroller.callTimeout"))
	defer cancel()

	logger.V(4).Infof("RollbackRelease[%s]| request to gse-controller, %+v", r.Seq, r)

	resp, err := act.gseControllerCli.RollbackRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to gse-controller RollbackRelease, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *RollbackAction) queryApp() (pbcommon.ErrCode, string) {
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

	logger.V(4).Infof("RollbackRelease[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	act.app = resp.Data

	return resp.Code, resp.Message
}

// Do makes the workflows of this action base on input messages.
func (act *RollbackAction) Do() error {
	// query current release.
	currentRelease, errCode, errMsg := act.queryRelease(act.req.ReleaseId)
	if errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	act.currentRelease = currentRelease

	// check current release state.
	if act.currentRelease.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) &&
		act.currentRelease.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
		return act.Err(pbcommon.ErrCode_E_CS_ROLLBACK_UNPUBLISHED_RELEASE, "can't rollback the unpublished release.")
	}
	if act.currentRelease.AppId != act.req.AppId {
		return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, "can't rollback release, inconsonant app_id")
	}

	if act.currentRelease.State == int32(pbcommon.ReleaseState_RS_PUBLISHED) {
		// rollback current release, mark ROLLBACKED in data level.
		// sidecar would re-pull last releases, and ignore this release.
		if errCode, errMsg := act.rollbackData(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}

	// query app.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// gsecontroller pub rololback msg.
	if errCode, errMsg := act.rollback(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// need re-publish target release, not only rollback last release state mode,
	// create new release base on target release.
	if len(act.req.NewReleaseId) != 0 {
		// TODO support reentry, do not create release everytime.

		// gen new releaseid for re-publish.
		if err := act.genReleaseID(); err != nil {
			return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error())
		}

		// query re-publish release.
		newRePubRelease, errCode, errMsg := act.queryRelease(act.req.NewReleaseId)
		if errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
		act.newRePubRelease = newRePubRelease

		if errCode, errMsg := act.createRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// make release data published.
		if errCode, errMsg := act.publishData(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// gsecontroller publish.
		if errCode, errMsg := act.publish(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}
	return nil
}
