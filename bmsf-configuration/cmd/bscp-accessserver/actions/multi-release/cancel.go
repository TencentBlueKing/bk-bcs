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

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/accessserver"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/logger"
)

// CancelAction cancels target release object.
type CancelAction struct {
	viper    *viper.Viper
	buSvrCli pbbusinessserver.BusinessClient

	req  *pb.CancelMultiReleaseReq
	resp *pb.CancelMultiReleaseResp
}

// NewCancelAction creates new CancelAction.
func NewCancelAction(viper *viper.Viper, buSvrCli pbbusinessserver.BusinessClient,
	req *pb.CancelMultiReleaseReq, resp *pb.CancelMultiReleaseResp) *CancelAction {
	action := &CancelAction{viper: viper, buSvrCli: buSvrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CancelAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CancelAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_AS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CancelAction) Output() error {
	// do nothing.
	return nil
}

func (act *CancelAction) verify() error {
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

func (act *CancelAction) cancel() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.CancelMultiReleaseReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		MultiReleaseid: act.req.MultiReleaseid,
		Operator:       act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeoutLT"))
	defer cancel()

	logger.V(2).Infof("CancelMultiRelease[%d]| request to businessserver CancelMultiRelease, %+v", act.req.Seq, r)

	resp, err := act.buSvrCli.CancelMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver CancelMultiRelease, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *CancelAction) Do() error {
	// cancel release.
	if errCode, errMsg := act.cancel(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
