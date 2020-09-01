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

package reload

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-businessserver/modules/audit"
	"bk-bscp/internal/database"
	pbbcscontroller "bk-bscp/internal/protocol/bcs-controller"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	"bk-bscp/pkg/logger"
)

// ReloadAction reloads target release or multi release.
type ReloadAction struct {
	viper            *viper.Viper
	dataMgrCli       pbdatamanager.DataManagerClient
	bcsControllerCli pbbcscontroller.BCSControllerClient
	gseControllerCli pbgsecontroller.GSEControllerClient

	req  *pb.ReloadReq
	resp *pb.ReloadResp

	app *pbcommon.App

	release *pbcommon.Release

	multiRelease *pbcommon.MultiRelease
	releaseids   []string
	metadatas    []*pbcommon.ReleaseMetadata

	reloadSpec *pbcommon.ReloadSpec
}

// NewReloadAction creates new ReloadAction.
func NewReloadAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient, bcsControllerCli pbbcscontroller.BCSControllerClient,
	gseControllerCli pbgsecontroller.GSEControllerClient, req *pb.ReloadReq, resp *pb.ReloadResp) *ReloadAction {
	action := &ReloadAction{viper: viper, dataMgrCli: dataMgrCli, bcsControllerCli: bcsControllerCli, gseControllerCli: gseControllerCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ReloadAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReloadAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ReloadAction) Output() error {
	// do nothing.
	return nil
}

func (act *ReloadAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	if len(act.req.Releaseid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, releaseid too long")
	}

	if len(act.req.MultiReleaseid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, multiReleaseid too long")
	}

	if len(act.req.Releaseid) == 0 && len(act.req.MultiReleaseid) == 0 {
		return errors.New("invalid params, releaseid and multiReleaseid both missing")
	}
	if len(act.req.Releaseid) != 0 && len(act.req.MultiReleaseid) != 0 {
		return errors.New("invalid params, only support releaseid or multiReleaseid")
	}

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}
	return nil
}

func (act *ReloadAction) reloadBCSMode() (pbcommon.ErrCode, string) {
	r := &pbbcscontroller.ReloadReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		Releaseid:      act.req.Releaseid,
		MultiReleaseid: act.req.MultiReleaseid,
		Operator:       act.req.Operator,
		ReloadSpec:     act.reloadSpec,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("bcscontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Reload[%d]| request to bcs-controller Reload, %+v", act.req.Seq, r)

	resp, err := act.bcsControllerCli.Reload(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to bcs-controller Reload, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *ReloadAction) reloadGSEPluginMode() (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.ReloadReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		Releaseid:      act.req.Releaseid,
		MultiReleaseid: act.req.MultiReleaseid,
		Operator:       act.req.Operator,
		ReloadSpec:     act.reloadSpec,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("gsecontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Reload[%d]| request to gse-controller Reload, %+v", act.req.Seq, r)

	resp, err := act.gseControllerCli.Reload(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to gse-controller Reload, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *ReloadAction) genReloadSpec() {
	if len(act.req.Releaseid) != 0 {
		effectInfo := &pbcommon.EffectInfo{Cfgsetid: act.release.Cfgsetid, Releaseid: act.release.Releaseid}

		reloadSpec := &pbcommon.ReloadSpec{Rollback: act.req.Rollback, Info: []*pbcommon.EffectInfo{effectInfo}}
		act.reloadSpec = reloadSpec
	} else {
		info := []*pbcommon.EffectInfo{}
		for _, md := range act.metadatas {
			info = append(info, &pbcommon.EffectInfo{Cfgsetid: md.Cfgsetid, Releaseid: md.Releaseid})
		}

		reloadSpec := &pbcommon.ReloadSpec{Rollback: act.req.Rollback, MultiReleaseid: act.req.MultiReleaseid, Info: info}
		act.reloadSpec = reloadSpec
	}
}

func (act *ReloadAction) queryRelease(releaseid string) (*pbcommon.Release, pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: releaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Reload[%d]| request to datamanager QueryRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	return resp.Release, resp.ErrCode, resp.ErrMsg
}

func (act *ReloadAction) queryMultiRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		MultiReleaseid: act.req.MultiReleaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Reload[%d]| request to datamanager QueryMultiRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryMultiRelease, %+v", err)
	}
	act.multiRelease = resp.MultiRelease

	return resp.ErrCode, resp.ErrMsg
}

func (act *ReloadAction) queryMetadatas() (pbcommon.ErrCode, string) {
	for _, releaseid := range act.releaseids {
		release, errCode, errMsg := act.queryRelease(releaseid)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}

		act.metadatas = append(act.metadatas, &pbcommon.ReleaseMetadata{
			Cfgsetid:  release.Cfgsetid,
			Commitid:  release.Commitid,
			Releaseid: release.Releaseid,
		})
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *ReloadAction) querySubReleaseList() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseSubListReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		MultiReleaseid: act.req.MultiReleaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Reload[%d]| request to datamanager QueryMultiReleaseSubList, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiReleaseSubList(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryMultiReleaseSubList, %+v", err)
	}
	act.releaseids = resp.Releaseids

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	return resp.ErrCode, resp.ErrMsg
}

func (act *ReloadAction) queryApp(appid string) (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAppReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: appid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Reload[%d]| request to datamanager QueryApp, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	act.app = resp.App

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *ReloadAction) Do() error {
	if len(act.req.Releaseid) != 0 {
		// query release.
		release, errCode, errMsg := act.queryRelease(act.req.Releaseid)
		if errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		if release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) && release.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
			return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "target release not in published/rollbacked state")
		}
		act.release = release

		// query app.
		if errCode, errMsg := act.queryApp(act.release.Appid); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

	} else {
		// query multi release.
		if errCode, errMsg := act.queryMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		if act.multiRelease.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) && act.multiRelease.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
			return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "target multi release not in published/rollbacked state")
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
		if errCode, errMsg := act.queryApp(act.multiRelease.Appid); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}

	// gen reload spec info.
	act.genReloadSpec()

	// deploy publish.
	if act.app.DeployType == int32(pbcommon.DeployType_DT_BCS) {
		// bcs connserver mode publish.

		// bcscontroller publish.
		if errCode, errMsg := act.reloadBCSMode(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else if act.app.DeployType == int32(pbcommon.DeployType_DT_GSE_PLUGIN) ||
		act.app.DeployType == int32(pbcommon.DeployType_DT_GSE) {
		// gse plugin sidecar mode.

		// gsecontroller publish.
		if errCode, errMsg := act.reloadGSEPluginMode(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "unknow deploy type")
	}

	// audit here on release reload.
	if len(act.req.Releaseid) != 0 {
		audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_RELOAD),
			act.req.Bid, act.req.Releaseid, act.req.Operator, "")
	} else {
		audit.Audit(int32(pbcommon.SourceType_ST_MULTI_RELEASE), int32(pbcommon.SourceOpType_SOT_RELOAD),
			act.req.Bid, act.req.MultiReleaseid, act.req.Operator, "")
	}

	return nil
}
