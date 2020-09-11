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

package business

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

// UpdateAction updates target business object.
type UpdateAction struct {
	viper    *viper.Viper
	buSvrCli pbbusinessserver.BusinessClient

	req  *pb.UpdateBusinessReq
	resp *pb.UpdateBusinessResp
}

// NewUpdateAction creates new UpdateAction.
func NewUpdateAction(viper *viper.Viper, buSvrCli pbbusinessserver.BusinessClient,
	req *pb.UpdateBusinessReq, resp *pb.UpdateBusinessResp) *UpdateAction {
	action := &UpdateAction{viper: viper, buSvrCli: buSvrCli, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_AS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *UpdateAction) Output() error {
	// do nothing.
	return nil
}

func (act *UpdateAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	length = len(act.req.Name)
	if length == 0 {
		return errors.New("invalid params, name missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, name too long")
	}

	if len(act.req.Depid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, depid too long")
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *UpdateAction) update() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.UpdateBusinessReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Name:     act.req.Name,
		Depid:    act.req.Depid,
		Memo:     act.req.Memo,
		State:    act.req.State,
		Operator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("UpdateBusiness[%d]| request to businessserver UpdateBusiness, %+v", act.req.Seq, r)

	resp, err := act.buSvrCli.UpdateBusiness(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver UpdateBusiness, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *UpdateAction) Do() error {
	// update business.
	if errCode, errMsg := act.update(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
