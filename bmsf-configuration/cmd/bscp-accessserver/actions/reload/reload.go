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

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/accessserver"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/logger"
)

// ReloadAction creates a reload action object.
type ReloadAction struct {
	viper    *viper.Viper
	buSvrCli pbbusinessserver.BusinessClient

	req  *pb.ReloadReq
	resp *pb.ReloadResp
}

// NewReloadAction creates new ReloadAction.
func NewReloadAction(viper *viper.Viper, buSvrCli pbbusinessserver.BusinessClient,
	req *pb.ReloadReq, resp *pb.ReloadResp) *ReloadAction {
	action := &ReloadAction{viper: viper, buSvrCli: buSvrCli, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_AS_PARAMS_INVALID, err.Error())
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

func (act *ReloadAction) reload() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.ReloadReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		Releaseid:      act.req.Releaseid,
		MultiReleaseid: act.req.MultiReleaseid,
		Operator:       act.req.Operator,
		Rollback:       act.req.Rollback,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeoutLT"))
	defer cancel()

	logger.V(2).Infof("Reload[%d]| request to businessserver Reload, %+v", act.req.Seq, r)

	resp, err := act.buSvrCli.Reload(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver Reload, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *ReloadAction) Do() error {
	// reload release/multi release.
	if errCode, errMsg := act.reload(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
