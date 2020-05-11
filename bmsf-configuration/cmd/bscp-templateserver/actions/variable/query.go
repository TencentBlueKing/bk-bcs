/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package variable

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// QueryAction query variable object
type QueryAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryVariableReq
	resp *pb.QueryVariableResp

	variable *pbcommon.Variable
}

// NewQueryAction creates new QueryAction
func NewQueryAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryVariableReq, resp *pb.QueryVariableResp) *QueryAction {
	action := &QueryAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryAction) Output() error {
	act.resp.Var = act.variable
	return nil
}

func (act *QueryAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Vid, "vid"); err != nil {
		return err
	}

	if err := common.VerifyVariableType(act.req.Type); err != nil {
		return err
	}

	return nil
}

func (act *QueryAction) queryVariable() (pbcommon.ErrCode, string) {
	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	req := &pbdatamanager.QueryVariableReq{
		Seq:  act.req.Seq,
		Bid:  act.req.Bid,
		Vid:  act.req.Vid,
		Type: act.req.Type,
	}

	logger.V(2).Infof("QueryVariable[%d]| request to datamanager QueryVariable, %+v", req.Seq, req)

	resp, err := act.dataMgrCli.QueryVariable(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryVariable, %+v", err)
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}
	act.variable = resp.Var

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *QueryAction) Do() error {
	// query variable
	if errCode, errMsg := act.queryVariable(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
