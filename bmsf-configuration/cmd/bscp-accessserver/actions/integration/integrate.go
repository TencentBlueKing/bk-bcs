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

package integration

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/accessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbintegrator "bk-bscp/internal/protocol/integrator"
	"bk-bscp/pkg/logger"
)

// IntegrateAction handles logic integrate actions.
type IntegrateAction struct {
	viper  *viper.Viper
	itgCli pbintegrator.IntegratorClient

	req  *pb.IntegrateReq
	resp *pb.IntegrateResp
}

// NewIntegrateAction creates new IntegrateAction.
func NewIntegrateAction(viper *viper.Viper, itgCli pbintegrator.IntegratorClient,
	req *pb.IntegrateReq, resp *pb.IntegrateResp) *IntegrateAction {
	action := &IntegrateAction{viper: viper, itgCli: itgCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *IntegrateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *IntegrateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_AS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *IntegrateAction) Output() error {
	// do nothing.
	return nil
}

func (act *IntegrateAction) verify() error {
	length := len(act.req.Metadata)
	if length == 0 {
		return errors.New("invalid params, metadata missing")
	}
	if length > database.BSCPITGTPLSIZELIMIT {
		return errors.New("invalid params, metadata too large")
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

func (act *IntegrateAction) integrate() (pbcommon.ErrCode, string) {
	r := &pbintegrator.IntegrateReq{
		Seq:      act.req.Seq,
		Metadata: act.req.Metadata,
		Operator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("integrator.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Integrate[%d]| request to integrator IntegrateReq, %+v", act.req.Seq, r)

	resp, err := act.itgCli.Integrate(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to integrator Integrate, %+v", err)
	}

	act.resp.Bid = resp.Bid
	act.resp.Appid = resp.Appid
	act.resp.Cfgsetid = resp.Cfgsetid
	act.resp.Commitid = resp.Commitid
	act.resp.Strategyid = resp.Strategyid
	act.resp.Releaseid = resp.Releaseid

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *IntegrateAction) Do() error {
	// integrate.
	if errCode, errMsg := act.integrate(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
