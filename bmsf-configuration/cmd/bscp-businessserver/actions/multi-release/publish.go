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

package multirelease

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

// PublishAction publishes target multi release object.
type PublishAction struct {
	viper            *viper.Viper
	dataMgrCli       pbdatamanager.DataManagerClient
	bcsControllerCli pbbcscontroller.BCSControllerClient
	gseControllerCli pbgsecontroller.GSEControllerClient

	req  *pb.PublishMultiReleaseReq
	resp *pb.PublishMultiReleaseResp

	multiRelease *pbcommon.MultiRelease
	app          *pbcommon.App
	isPublished  bool
	releaseids   []string
}

// NewPublishAction creates new PublishAction.
func NewPublishAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient, bcsControllerCli pbbcscontroller.BCSControllerClient,
	gseControllerCli pbgsecontroller.GSEControllerClient, req *pb.PublishMultiReleaseReq, resp *pb.PublishMultiReleaseResp) *PublishAction {
	action := &PublishAction{viper: viper, dataMgrCli: dataMgrCli, bcsControllerCli: bcsControllerCli, gseControllerCli: gseControllerCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PublishAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PublishAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PublishAction) Output() error {
	// do nothing.
	return nil
}

func (act *PublishAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.MultiReleaseid)
	if length == 0 {
		return errors.New("invalid params, multi releaseid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, multi releaseid too long")
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

func (act *PublishAction) queryApp() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAppReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.multiRelease.Appid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%d]| request to datamanager QueryApp, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	act.app = resp.App

	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) querySubReleaseList() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseSubListReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		MultiReleaseid: act.req.MultiReleaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%d]| request to datamanager QueryMultiReleaseSubList, %+v", act.req.Seq, r)

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

func (act *PublishAction) queryMultiRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		MultiReleaseid: act.req.MultiReleaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%d]| request to datamanager QueryMultiRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryMultiRelease, %+v", err)
	}
	act.multiRelease = resp.MultiRelease

	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) publishPreBCSMode(releaseid string) (pbcommon.ErrCode, string) {
	r := &pbbcscontroller.PublishReleasePreReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: releaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("bcscontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%d]| request to bcs-controller PublishReleasePre, %+v", act.req.Seq, r)

	resp, err := act.bcsControllerCli.PublishReleasePre(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to bcs-controller PublishReleasePre, %+v", err)
	}

	if resp.ErrCode == pbcommon.ErrCode_E_BCS_ALREADY_PUBLISHED {
		act.isPublished = true
		return pbcommon.ErrCode_E_OK, ""
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) publishPreGSEPluginMode(releaseid string) (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.PublishReleasePreReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: releaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("gsecontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%d]| request to gse-controller PublishReleasePre, %+v", act.req.Seq, r)

	resp, err := act.gseControllerCli.PublishReleasePre(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to gse-controller PublishReleasePre, %+v", err)
	}

	if resp.ErrCode == pbcommon.ErrCode_E_BCS_ALREADY_PUBLISHED {
		act.isPublished = true
		return pbcommon.ErrCode_E_OK, ""
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) publishData(releaseid string) (pbcommon.ErrCode, string) {
	r := &pbdatamanager.PublishReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: releaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%d]| request to datamanager PublishRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager PublishRelease, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on release published.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_PUBLISH),
		act.req.Bid, releaseid, act.req.Operator, "")

	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) publishBCSMode(releaseid string) (pbcommon.ErrCode, string) {
	r := &pbbcscontroller.PublishReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: releaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("bcscontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%d]| request to bcs-controller PublishRelease, %+v", act.req.Seq, r)

	resp, err := act.bcsControllerCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to bcs-controller PublishRelease, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) publishGSEPluginMode(releaseid string) (pbcommon.ErrCode, string) {
	r := &pbgsecontroller.PublishReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: releaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("gsecontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%d]| request to gse-controller PublishRelease, %+v", act.req.Seq, r)

	resp, err := act.gseControllerCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to gse-controller PublishRelease, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) publishMultiReleaseData() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.PublishMultiReleaseReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		MultiReleaseid: act.req.MultiReleaseid,
		Operator:       act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PublishMultiRelease[%d]| request to datamanager PublishMultiRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.PublishMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager PublishMultiRelease, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on release published.
	audit.Audit(int32(pbcommon.SourceType_ST_MULTI_RELEASE), int32(pbcommon.SourceOpType_SOT_PUBLISH),
		act.req.Bid, act.req.MultiReleaseid, act.req.Operator, "")

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PublishAction) Do() error {
	// query multi release.
	if errCode, errMsg := act.queryMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.multiRelease.State == int32(pbcommon.ReleaseState_RS_PUBLISHED) {
		// already published.
		return nil
	}
	if act.multiRelease.State != int32(pbcommon.ReleaseState_RS_INIT) {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW,
			"can't publish the multi release which is not init state")
	}

	// query app.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query multi release sub list.
	if errCode, errMsg := act.querySubReleaseList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	for _, releaseid := range act.releaseids {
		// deploy publish.
		if act.app.DeployType == int32(pbcommon.DeployType_DT_BCS) {
			// bcs connserver mode publish.
			act.isPublished = false

			// bcscontroller publish pre.
			if errCode, errMsg := act.publishPreBCSMode(releaseid); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}

			// already published.
			if act.isPublished {
				continue
			}

			// make release data published.
			if errCode, errMsg := act.publishData(releaseid); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}

			// bcscontroller publish.
			if errCode, errMsg := act.publishBCSMode(releaseid); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}
		} else if act.app.DeployType == int32(pbcommon.DeployType_DT_GSE_PLUGIN) ||
			act.app.DeployType == int32(pbcommon.DeployType_DT_GSE) {
			// gse plugin sidecar mode.
			act.isPublished = false

			// gsecontroller publish pre.
			if errCode, errMsg := act.publishPreGSEPluginMode(releaseid); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}

			// already published.
			if act.isPublished {
				continue
			}

			// make release data published.
			if errCode, errMsg := act.publishData(releaseid); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}

			// gsecontroller publish.
			if errCode, errMsg := act.publishGSEPluginMode(releaseid); errCode != pbcommon.ErrCode_E_OK {
				return act.Err(errCode, errMsg)
			}
		} else {
			return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "unknow deploy type")
		}
	}

	// make multi release data published.
	if errCode, errMsg := act.publishMultiReleaseData(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
