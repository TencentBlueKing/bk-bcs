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

package report

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
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/logger"
)

// ReportAction report sidecar release information.
type ReportAction struct {
	viper            *viper.Viper
	bcsControllerCli pbbcscontroller.BCSControllerClient

	req  *pb.ReportReq
	resp *pb.ReportResp
}

// NewReportAction creates new ReportAction.
func NewReportAction(viper *viper.Viper, bcsControllerCli pbbcscontroller.BCSControllerClient,
	req *pb.ReportReq, resp *pb.ReportResp) *ReportAction {
	action := &ReportAction{viper: viper, bcsControllerCli: bcsControllerCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ReportAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReportAction) Input() error {
	if act.req.Labels == "" {
		act.req.Labels = strategy.EmptySidecarLabels
	}
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CONNS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ReportAction) Output() error {
	// do nothing.
	return nil
}

func (act *ReportAction) verify() error {
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

	if act.req.Labels != strategy.EmptySidecarLabels {
		labels := strategy.SidecarLabels{}
		if err := json.Unmarshal([]byte(act.req.Labels), &labels); err != nil {
			return fmt.Errorf("invalid params, labels[%+v], %+v", act.req.Labels, err)
		}
	}

	if act.req.Infos == nil || len(act.req.Infos) == 0 {
		return errors.New("invalid params, infos missing")
	}
	return nil
}

func (act *ReportAction) report() (pbcommon.ErrCode, string) {
	r := &pbbcscontroller.ReportReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Appid:     act.req.Appid,
		Clusterid: act.req.Clusterid,
		Zoneid:    act.req.Zoneid,
		Dc:        act.req.Dc,
		IP:        act.req.IP,
		Labels:    act.req.Labels,
		Infos:     act.req.Infos,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("bcscontroller.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Report[%d]| request to bcs-controller Report, %+v", act.req.Seq, r)

	resp, err := act.bcsControllerCli.Report(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CONNS_SYSTEM_UNKONW, fmt.Sprintf("request to bcs-controller Report, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *ReportAction) Do() error {
	if errCode, errMsg := act.report(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
