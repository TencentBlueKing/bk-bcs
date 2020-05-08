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

package templatebinding

import (
	"bk-bscp/pkg/common"
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
)

// DeleteAction delete a binding object
type DeleteAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.DeleteConfigTemplateBindingReq
	resp *pb.DeleteConfigTemplateBindingResp

	newVid          string
	templateBinding *pbcommon.ConfigTemplateBinding
}

// NewDeleteAction creates new DeleteAction
func NewDeleteAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.DeleteConfigTemplateBindingReq, resp *pb.DeleteConfigTemplateBindingResp) *DeleteAction {
	action := &DeleteAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *DeleteAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *DeleteAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *DeleteAction) Output() error {

	return nil
}

func (act *DeleteAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}
	if err := common.VerifyID(act.req.Templateid, "templateid"); err != nil {
		return err
	}
	if err := common.VerifyID(act.req.Appid, "appid"); err != nil {
		return err
	}
	return nil
}

func (act *DeleteAction) queryBinding() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryConfigTemplateBindingReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Templateid: act.req.Templateid,
		Appid:      act.req.Appid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryConfigTemplateBinding(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	act.templateBinding = resp.ConfigTemplateBinding

	return pbcommon.ErrCode_E_OK, ""
}

func (act *DeleteAction) deleteConfigSet() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.DeleteConfigSetReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Cfgsetid: act.templateBinding.Cfgsetid,
		Operator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.DeleteConfigSet(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	return pbcommon.ErrCode_E_OK, ""
}

func (act *DeleteAction) deleteBinding() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.DeleteConfigTemplateBindingReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Templateid: act.req.Templateid,
		Appid:      act.req.Appid,
		Operator:   act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.DeleteConfigTemplateBinding(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	return pbcommon.ErrCode_E_OK, ""
}

// Do do action
func (act *DeleteAction) Do() error {

	if errCode, errMsg := act.queryBinding(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, fmt.Sprintf("query binding failed when delete binding, %s", errMsg))
	}

	if errCode, errMsg := act.deleteConfigSet(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, fmt.Sprintf("delete configset failed, when delete binding %s", errMsg))
	}

	if errCode, errMsg := act.deleteBinding(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	return nil
}
