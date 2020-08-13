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

package procattr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-businessserver/modules/audit"
	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/logger"
)

// UpdateAction updates target ProcAttr object.
type UpdateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.UpdateProcAttrReq
	resp *pb.UpdateProcAttrResp
}

// NewUpdateAction creates new UpdateAction.
func NewUpdateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.UpdateProcAttrReq, resp *pb.UpdateProcAttrResp) *UpdateAction {
	action := &UpdateAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *UpdateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *UpdateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *UpdateAction) Output() error {
	// do nothing.
	return nil
}

func (act *UpdateAction) verify() error {
	length := len(act.req.Cloudid)
	if length == 0 {
		return errors.New("invalid params, cloudid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cloudid too long")
	}

	length = len(act.req.IP)
	if length == 0 {
		return errors.New("invalid params, ip missing")
	}
	if length > database.BSCPNORMALSTRLENLIMIT {
		return errors.New("invalid params, ip too long")
	}

	length = len(act.req.Bid)
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

	length = len(act.req.Path)
	if length == 0 {
		return errors.New("invalid params, path missing")
	}
	if length > database.BSCPCFGSETFPATHLENLIMIT {
		return errors.New("invalid params, path too long")
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
		act.req.Dc = act.req.Cloudid
	} else {
		if length > database.BSCPLONGSTRLENLIMIT {
			return errors.New("invalid params, dc tag too long")
		}
	}

	if len(act.req.Labels) == 0 {
		act.req.Labels = strategy.EmptySidecarLabels
	}
	if len(act.req.Labels) > database.BSCPLABELSSIZELIMIT {
		return errors.New("invalid params, labels too large")
	}

	if act.req.Labels != strategy.EmptySidecarLabels {
		labels := strategy.SidecarLabels{}
		if err := json.Unmarshal([]byte(act.req.Labels), &labels); err != nil {
			return fmt.Errorf("invalid params, labels[%+v], %+v", act.req.Labels, err)
		}
	}

	if len(act.req.Memo) > database.BSCPCFGSETFPATHLENLIMIT {
		return errors.New("invalid params, memo too long")
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

func (act *UpdateAction) update() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.UpdateProcAttrReq{
		Seq:       act.req.Seq,
		Cloudid:   act.req.Cloudid,
		IP:        act.req.IP,
		Bid:       act.req.Bid,
		Appid:     act.req.Appid,
		Clusterid: act.req.Clusterid,
		Zoneid:    act.req.Zoneid,
		Dc:        act.req.Dc,
		Labels:    act.req.Labels,
		Path:      act.req.Path,
		Memo:      act.req.Memo,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("UpdateProcAttr[%d]| request to datamanager UpdateProcAttr, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.UpdateProcAttr(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager UpdateProcAttr, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on procattr updated.
	audit.Audit(int32(pbcommon.SourceType_ST_PROC_ATTR), int32(pbcommon.SourceOpType_SOT_UPDATE),
		act.req.Bid, act.req.Appid, act.req.Operator, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *UpdateAction) Do() error {
	// update procattr.
	if errCode, errMsg := act.update(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
