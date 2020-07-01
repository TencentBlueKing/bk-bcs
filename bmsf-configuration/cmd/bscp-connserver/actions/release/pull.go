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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbbcscontroller "bk-bscp/internal/protocol/bcs-controller"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/logger"
)

// PullAction pull release (target or newest) for bcs sidecar.
type PullAction struct {
	viper            *viper.Viper
	bcsControllerCli pbbcscontroller.BCSControllerClient
	dataMgrCli       pbdatamanager.DataManagerClient

	req  *pb.PullReleaseReq
	resp *pb.PullReleaseResp

	release *pbcommon.Release
}

// NewPullAction creates new PullAction.
func NewPullAction(viper *viper.Viper, bcsControllerCli pbbcscontroller.BCSControllerClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.PullReleaseReq, resp *pb.PullReleaseResp) *PullAction {
	action := &PullAction{viper: viper, bcsControllerCli: bcsControllerCli, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PullAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PullAction) Input() error {
	if act.req.Labels == "" {
		act.req.Labels = strategy.EmptySidecarLabels
	}
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CONNS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PullAction) Output() error {
	// do nothing.
	return nil
}

func (act *PullAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Appid)
	if length == 0 {
		return errors.New("invalid params, appid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	length = len(act.req.Clusterid)
	if length == 0 {
		return errors.New("invalid params, clusterid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, clusterid too long")
	}

	length = len(act.req.Zoneid)
	if length == 0 {
		return errors.New("invalid params, zoneid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, zoneid too long")
	}

	length = len(act.req.Dc)
	if length == 0 {
		return errors.New("invalid params, dc missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, dc too long")
	}

	length = len(act.req.IP)
	if length == 0 {
		return errors.New("invalid params, ip missing")
	}
	if length > database.BSCPNORMALSTRLENLIMIT {
		return errors.New("invalid params, ip too long")
	}

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	if act.req.Labels != strategy.EmptySidecarLabels {
		labels := strategy.SidecarLabels{}
		if err := json.Unmarshal([]byte(act.req.Labels), &labels); err != nil {
			return fmt.Errorf("invalid params, labels[%+v], %+v", act.req.Labels, err)
		}
	}
	return nil
}

func (act *PullAction) pullRelease() (pbcommon.ErrCode, string) {
	r := &pbbcscontroller.PullReleaseReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		Appid:          act.req.Appid,
		Clusterid:      act.req.Clusterid,
		Zoneid:         act.req.Zoneid,
		Dc:             act.req.Dc,
		IP:             act.req.IP,
		Labels:         act.req.Labels,
		Cfgsetid:       act.req.Cfgsetid,
		LocalReleaseid: act.req.LocalReleaseid,
		Releaseid:      act.req.Releaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("bcscontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PullRelease[%d]| request to bcs-controller PullRelease, %+v", act.req.Seq, r)

	resp, err := act.bcsControllerCli.PullRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CONNS_SYSTEM_UNKONW, fmt.Sprintf("request to bcs-controller PullRelease, %+v", err)
	}
	logger.V(2).Infof("PullRelease[%d]| release checking, %+v", act.req.Seq, resp)
	act.release = resp.Release

	return resp.ErrCode, resp.ErrMsg
}

func (act *PullAction) queryReleaseConfigs() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseConfigsReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Appid:     act.req.Appid,
		Clusterid: act.req.Clusterid,
		Zoneid:    act.req.Zoneid,
		Cfgsetid:  act.req.Cfgsetid,
		Commitid:  act.release.Commitid,
		Abstract:  true,
		Index:     act.req.IP,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PullRelease[%d]| request to datamanager QueryReleaseConfigs, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryReleaseConfigs(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CONNS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryReleaseConfigs, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	act.resp.Cid = resp.Configs.Cid
	act.resp.CfgLink = resp.Configs.CfgLink

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PullAction) Do() error {
	if errCode, errMsg := act.pullRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.release == nil {
		// target release is not need to effect or there
		// is no newest release.
		return nil
	}

	act.resp.NeedEffect = true
	act.resp.Release = act.release

	// query release configs.
	if errCode, errMsg := act.queryReleaseConfigs(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
