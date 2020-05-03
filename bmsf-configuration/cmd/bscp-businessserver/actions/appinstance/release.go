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

package appinstance

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/logger"
)

// ReleaseAction query release of target app instance.
type ReleaseAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryAppInstanceReleaseReq
	resp *pb.QueryAppInstanceReleaseResp
}

// NewReleaseAction creates new ReleaseAction.
func NewReleaseAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryAppInstanceReleaseReq, resp *pb.QueryAppInstanceReleaseResp) *ReleaseAction {
	action := &ReleaseAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ReleaseAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReleaseAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ReleaseAction) Output() error {
	// do nothing.
	return nil
}

func (act *ReleaseAction) verify() error {
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
	return nil
}

func (act *ReleaseAction) release() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAppInstanceReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Appid:     act.req.Appid,
		Clusterid: act.req.Clusterid,
		Zoneid:    act.req.Zoneid,
		Dc:        act.req.Dc,
		IP:        act.req.IP,
		Cfgsetid:  act.req.Cfgsetid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryAppInstanceRelease[%d]| request to datamanager QueryAppInstanceRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryAppInstanceRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryAppInstanceRelease, %+v", err)
	}

	act.resp.Releaseid = resp.Releaseid
	act.resp.Commitid = resp.Commitid
	act.resp.Cid = resp.Cid
	act.resp.CfgLink = resp.CfgLink
	act.resp.Content = resp.Content

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *ReleaseAction) Do() error {
	// query app instance release.
	if errCode, errMsg := act.release(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
