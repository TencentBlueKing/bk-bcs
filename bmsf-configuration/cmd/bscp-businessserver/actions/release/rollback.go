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

	"bk-bscp/cmd/bscp-businessserver/modules/audit"
	"bk-bscp/internal/database"
	pbbcscontroller "bk-bscp/internal/protocol/bcs-controller"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// RollbackAction rollbacks target release.
type RollbackAction struct {
	viper            *viper.Viper
	dataMgrCli       pbdatamanager.DataManagerClient
	bcsControllerCli pbbcscontroller.BCSControllerClient
	gseControllerCli pbgsecontroller.GSEControllerClient

	req  *pb.RollbackReleaseReq
	resp *pb.RollbackReleaseResp

	// current release which wanted to be rollbacked.
	currentRelease *pbcommon.Release

	// app informations.
	app *pbcommon.App

	// re-publish target release(newReleaseid), and newRePubReleaseid is the
	// new release id. It's empty if there is no newReleaseid in request.
	newRePubReleaseid string

	// release which wanted to be re-published.
	newRePubRelease *pbcommon.Release

	// re-publish flag.
	isReleaseRePublished bool
}

// NewRollbackAction creates new RollbackAction.
func NewRollbackAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient, bcsControllerCli pbbcscontroller.BCSControllerClient,
	gseControllerCli pbgsecontroller.GSEControllerClient, req *pb.RollbackReleaseReq, resp *pb.RollbackReleaseResp) *RollbackAction {
	action := &RollbackAction{viper: viper, dataMgrCli: dataMgrCli, bcsControllerCli: bcsControllerCli, gseControllerCli: gseControllerCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *RollbackAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *RollbackAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *RollbackAction) Output() error {
	// do nothing.
	return nil
}

func (act *RollbackAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Releaseid)
	if length == 0 {
		return errors.New("invalid params, releaseid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, releaseid too long")
	}

	if len(act.req.NewReleaseid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, newReleaseid too long")
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

func (act *RollbackAction) genReleaseID() error {
	id, err := common.GenReleaseid()
	if err != nil {
		return err
	}
	act.newRePubReleaseid = id
	return nil
}

func (act *RollbackAction) createRelease() (pbcommon.ErrCode, string) {
	newReleaseName := fmt.Sprintf("Rollback-%s", act.newRePubRelease.Name)

	if act.currentRelease.Appid != act.newRePubRelease.Appid ||
		act.currentRelease.Cfgsetid != act.newRePubRelease.Cfgsetid {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "not the same app-configset"
	}

	r := &pbdatamanager.CreateReleaseReq{
		Seq:         act.req.Seq,
		Bid:         act.req.Bid,
		Releaseid:   act.newRePubReleaseid,
		Name:        newReleaseName,
		Appid:       act.newRePubRelease.Appid,
		Cfgsetid:    act.newRePubRelease.Cfgsetid,
		CfgsetName:  act.newRePubRelease.CfgsetName,
		CfgsetFpath: act.newRePubRelease.CfgsetFpath,
		Strategyid:  act.newRePubRelease.Strategyid,
		Strategies:  act.newRePubRelease.Strategies,
		Commitid:    act.newRePubRelease.Commitid,
		Memo:        act.newRePubRelease.Memo,
		Creator:     act.req.Operator,
		State:       int32(pbcommon.ReleaseState_RS_INIT),
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to datamanager CreateRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateRelease, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}
	act.resp.Releaseid = act.newRePubReleaseid

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) queryRelease(releaseid string) (*pbcommon.Release, pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: releaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to datamanager QueryRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	return resp.Release, resp.ErrCode, resp.ErrMsg
}

func (act *RollbackAction) publishPreBCSMode() (pbcommon.ErrCode, string) {
	r := &pbbcscontroller.PublishReleasePreReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.newRePubReleaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("bcscontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to bcs-controller PublishReleasePre, %+v", act.req.Seq, r)

	resp, err := act.bcsControllerCli.PublishReleasePre(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to bcs-controller PublishReleasePre, %+v", err)
	}

	if resp.ErrCode == pbcommon.ErrCode_E_BCS_ALREADY_PUBLISHED {
		return pbcommon.ErrCode_E_OK, ""
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *RollbackAction) publishPreGSEPluginMode() (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.PublishReleasePreReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.newRePubReleaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("gsecontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to gse-controller PublishReleasePre, %+v", act.req.Seq, r)

	resp, err := act.gseControllerCli.PublishReleasePre(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to gse-controller PublishReleasePre, %+v", err)
	}

	if resp.ErrCode == pbcommon.ErrCode_E_BCS_ALREADY_PUBLISHED {
		return pbcommon.ErrCode_E_OK, ""
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *RollbackAction) publishData() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.PublishReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.newRePubReleaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to datamanager PublishRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager PublishRelease, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on release published.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_PUBLISH),
		act.req.Bid, act.newRePubReleaseid, act.req.Operator, "ROLLBACK-REPUB")

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) publishBCSMode() error {
	r := &pbbcscontroller.PublishReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.newRePubReleaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("bcscontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to bcs-controller PublishRelease, %+v", act.req.Seq, r)

	resp, err := act.bcsControllerCli.PublishRelease(ctx, r)
	if err != nil {
		return err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

func (act *RollbackAction) publishGSEPluginMode() error {
	r := &pbgsecontroller.PublishReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.newRePubReleaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("gsecontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to gse-controller PublishRelease, %+v", act.req.Seq, r)

	resp, err := act.gseControllerCli.PublishRelease(ctx, r)
	if err != nil {
		return err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

func (act *RollbackAction) rollbackData() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.RollbackReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.req.Releaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to datamanager RollbackRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.RollbackRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager RollbackRelease, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on release rollbacked.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_ROLLBACK),
		act.req.Bid, act.req.Releaseid, act.req.Operator, "")

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) rollbackBCSMode() (pbcommon.ErrCode, string) {
	r := &pbbcscontroller.RollbackReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.req.Releaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("bcscontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to bcs-controller RollbackRelease, %+v", act.req.Seq, r)

	resp, err := act.bcsControllerCli.RollbackRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to bcs-controller RollbackRelease, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *RollbackAction) rollbackGSEPluginMode() (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.RollbackReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.req.Releaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("gsecontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to gse-controller RollbackRelease, %+v", act.req.Seq, r)

	resp, err := act.gseControllerCli.RollbackRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to gse-controller RollbackRelease, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *RollbackAction) queryApp() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAppReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.currentRelease.Appid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%d]| request to datamanager QueryApp, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	act.app = resp.App

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *RollbackAction) Do() error {
	// query current release.
	currentRelease, errCode, errMsg := act.queryRelease(act.req.Releaseid)
	if errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	act.currentRelease = currentRelease

	// check current release state.
	if act.currentRelease.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) &&
		act.currentRelease.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
		return act.Err(pbcommon.ErrCode_E_BS_ROLLBACK_UNPUBLISHED_RELEASE, "can't rollback the unpublished release.")
	}

	// rollback current release, mark ROLLBACKED in data level.
	// sidecar would re-pull last releases, and ignore this release.
	if errCode, errMsg := act.rollbackData(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query app.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// deploy publish.
	if act.app.DeployType == int32(pbcommon.DeployType_DT_BCS) {
		// bcscontroller pub rololback msg.
		if errCode, errMsg := act.rollbackBCSMode(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else if act.app.DeployType == int32(pbcommon.DeployType_DT_GSE_PLUGIN) ||
		act.app.DeployType == int32(pbcommon.DeployType_DT_GSE) {
		// gsecontroller pub rololback msg.
		if errCode, errMsg := act.rollbackGSEPluginMode(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "unknow deploy type")
	}

	// need re-publish target release, not only rollback last release state mode,
	// create new release base on target release.
	if len(act.req.NewReleaseid) != 0 {
		// TODO support reentry, do not create release everytime.

		// gen new releaseid for re-publish.
		if err := act.genReleaseID(); err != nil {
			return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
		}

		// query re-publish release.
		newRePubRelease, errCode, errMsg := act.queryRelease(act.req.NewReleaseid)
		if errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
		act.newRePubRelease = newRePubRelease

		if errCode, errMsg := act.createRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// deploy publish.
		if act.app.DeployType == int32(pbcommon.DeployType_DT_BCS) {
			// bcs connserver mode publish.

			// bcscontroller publish pre.
			if errCode, errMsg := act.publishPreBCSMode(); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}

			// make release data published.
			if errCode, errMsg := act.publishData(); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}

			// bcscontroller publish.
			if err := act.publishBCSMode(); err != nil {
				logger.Warnf("RollbackRelease[%d]| re-publish releae send bcscontroller msg, %+v", act.req.Seq, err)
				// do not return errors to client.
				return nil
			}
		} else if act.app.DeployType == int32(pbcommon.DeployType_DT_GSE_PLUGIN) ||
			act.app.DeployType == int32(pbcommon.DeployType_DT_GSE) {
			// gse plugin sidecar mode.

			// gsecontroller publish pre.
			if errCode, errMsg := act.publishPreGSEPluginMode(); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}

			// make release data published.
			if errCode, errMsg := act.publishData(); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}

			// gsecontroller publish.
			if err := act.publishGSEPluginMode(); err != nil {
				logger.Warnf("RollbackRelease[%d]| re-publish releae send gsecontroller msg, %+v", act.req.Seq, err)
				// do not return errors to client.
				return nil
			}
		} else {
			return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "unknow deploy type")
		}
	}
	return nil
}
