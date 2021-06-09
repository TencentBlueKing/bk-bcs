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

// ReloadAction reloads target release or multi release.
type ReloadAction struct {
	kit              kit.Kit
	viper            *viper.Viper
	authSvrCli       pbauthserver.AuthClient
	dataMgrCli       pbdatamanager.DataManagerClient
	gseControllerCli pbgsecontroller.GSEControllerClient

	req  *pb.ReloadReq
	resp *pb.ReloadResp

	app *pbcommon.App

	release *pbcommon.Release

	multiRelease *pbcommon.MultiRelease
	releaseIDs   []string
	metadatas    []*pbcommon.ReleaseMetadata

	reloadSpec *pbcommon.ReloadSpec
}

// NewReloadAction creates new ReloadAction.
func NewReloadAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	gseControllerCli pbgsecontroller.GSEControllerClient,
	req *pb.ReloadReq, resp *pb.ReloadResp) *ReloadAction {

	action := &ReloadAction{
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
func (act *ReloadAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReloadAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Authorize checks the action authorization.
func (act *ReloadAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *ReloadAction) Output() error {
	// do nothing.
	return nil
}

func (act *ReloadAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if len(act.req.ReleaseId) == 0 && len(act.req.MultiReleaseId) == 0 {
		return errors.New("invalid input data, release_id or multi_release_id is required")
	}
	if len(act.req.ReleaseId) != 0 && len(act.req.MultiReleaseId) != 0 {
		return errors.New("invalid params, only support release_id or multi_release_id")
	}

	if err = common.ValidateString("release_id", act.req.ReleaseId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_release_id", act.req.ReleaseId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *ReloadAction) authorize() (pbcommon.ErrCode, string) {
	// check authorize resource at first, it may be deleted.
	if errCode, errMsg := act.queryApp(act.req.AppId); errCode != pbcommon.ErrCode_E_OK {
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

func (act *ReloadAction) reload() (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.ReloadReq{
		Seq:            act.kit.Rid,
		BizId:          act.req.BizId,
		ReleaseId:      act.req.ReleaseId,
		MultiReleaseId: act.req.MultiReleaseId,
		Operator:       act.kit.User,
		ReloadSpec:     act.reloadSpec,
		Nice:           float64(len(act.releaseIDs)),
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("gsecontroller.callTimeout"))
	defer cancel()

	logger.V(4).Infof("Reload[%s]| request to gse-controller, %+v", r.Seq, r)

	resp, err := act.gseControllerCli.Reload(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to gse-controller Reload, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *ReloadAction) genReloadSpec() {
	if len(act.req.ReleaseId) != 0 {
		effectInfo := &pbcommon.EffectInfo{CfgId: act.release.CfgId, ReleaseId: act.release.ReleaseId}

		reloadSpec := &pbcommon.ReloadSpec{Rollback: act.req.Rollback, Info: []*pbcommon.EffectInfo{effectInfo}}
		act.reloadSpec = reloadSpec
	} else {
		info := []*pbcommon.EffectInfo{}
		for _, md := range act.metadatas {
			info = append(info, &pbcommon.EffectInfo{CfgId: md.CfgId, ReleaseId: md.ReleaseId})
		}

		reloadSpec := &pbcommon.ReloadSpec{
			Rollback:       act.req.Rollback,
			MultiReleaseId: act.req.MultiReleaseId,
			Info:           info,
		}
		act.reloadSpec = reloadSpec
	}
}

func (act *ReloadAction) queryRelease(releaseID string) (*pbcommon.Release, pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: releaseID,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("Reload[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	return resp.Data, resp.Code, resp.Message
}

func (act *ReloadAction) queryMultiRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseReq{
		Seq:            act.kit.Rid,
		BizId:          act.req.BizId,
		MultiReleaseId: act.req.MultiReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("Reload[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryMultiRelease, %+v", err)
	}
	act.multiRelease = resp.Data
	return resp.Code, resp.Message
}

func (act *ReloadAction) queryMetadatas() (pbcommon.ErrCode, string) {
	for _, releaseID := range act.releaseIDs {
		release, errCode, errMsg := act.queryRelease(releaseID)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}

		act.metadatas = append(act.metadatas, &pbcommon.ReleaseMetadata{
			CfgId:     release.CfgId,
			CommitId:  release.CommitId,
			ReleaseId: release.ReleaseId,
		})
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *ReloadAction) querySubReleaseList() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseSubListReq{
		Seq:            act.kit.Rid,
		BizId:          act.req.BizId,
		MultiReleaseId: act.req.MultiReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("Reload[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiReleaseSubList(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to datamanager QueryMultiReleaseSubList, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.releaseIDs = resp.Data.ReleaseIds

	return pbcommon.ErrCode_E_OK, ""
}

func (act *ReloadAction) queryApp(appID string) (pbcommon.ErrCode, string) {
	if act.app != nil {
		return pbcommon.ErrCode_E_OK, ""
	}

	r := &pbdatamanager.QueryAppReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		AppId: appID,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("Reload[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	act.app = resp.Data

	return resp.Code, resp.Message
}

// Do makes the workflows of this action base on input messages.
func (act *ReloadAction) Do() error {
	if len(act.req.ReleaseId) != 0 {
		// query release.
		release, errCode, errMsg := act.queryRelease(act.req.ReleaseId)
		if errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		if release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) &&
			release.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
			return act.Err(pbcommon.ErrCode_E_CS_RELOAD_UNPUBLISHED_RELEASE,
				"target release not in published/rollbacked state")
		}
		act.release = release

		if release.AppId != act.req.AppId {
			return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, "can't reload release, inconsonant app_id")
		}

		// query app.
		if errCode, errMsg := act.queryApp(act.release.AppId); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

	} else {
		// query multi release.
		if errCode, errMsg := act.queryMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		if act.multiRelease.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) &&
			act.multiRelease.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
			return act.Err(pbcommon.ErrCode_E_CS_RELOAD_UNPUBLISHED_RELEASE,
				"target multi release not in published/rollbacked state")
		}

		if act.multiRelease.AppId != act.req.AppId {
			return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, "can't reload multi release, inconsonant app_id")
		}

		// query multi release sub release list.
		if errCode, errMsg := act.querySubReleaseList(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// query sub release metadatas.
		if errCode, errMsg := act.queryMetadatas(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// query app.
		if errCode, errMsg := act.queryApp(act.multiRelease.AppId); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}

	// gen reload spec info.
	act.genReloadSpec()

	// gsecontroller publish.
	if errCode, errMsg := act.reload(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// audit here on release reload.
	if len(act.req.ReleaseId) != 0 {
		audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_RELOAD),
			act.req.BizId, act.req.ReleaseId, act.kit.User, "")
	} else {
		audit.Audit(int32(pbcommon.SourceType_ST_MULTI_RELEASE), int32(pbcommon.SourceOpType_SOT_RELOAD),
			act.req.BizId, act.req.MultiReleaseId, act.kit.User, "")
	}

	return nil
}
