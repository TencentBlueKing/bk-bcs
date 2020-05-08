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
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
)

// ListAction list template binding object
type ListAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryConfigTemplateBindingListReq
	resp *pb.QueryConfigTemplateBindingListResp
}

// NewListAction creates new ListAction
func NewListAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryConfigTemplateBindingListReq, resp *pb.QueryConfigTemplateBindingListResp) *ListAction {
	action := &ListAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ListAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ListAction) Output() error {
	return nil
}

func (act *ListAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if len(act.req.Templateid) != 0 && len(act.req.Appid) != 0 {
		return fmt.Errorf("invalid params, templateid or appid is required")
	}
	if len(act.req.Templateid) != 0 {
		if len(act.req.Templateid) > database.BSCPIDLENLIMIT {
			return fmt.Errorf("inavlid params, templateid too long")
		}
	}
	if len(act.req.Appid) != 0 {
		if len(act.req.Appid) > database.BSCPIDLENLIMIT {
			return fmt.Errorf("inavlid params, appid too long")
		}
	}

	if err := common.VerifyQueryLimit(act.req.Limit); err != nil {
		return err
	}

	return nil
}

func (act *ListAction) listBindings() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryConfigTemplateBindingListReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Templateid: act.req.Templateid,
		Appid:      act.req.Appid,
		Index:      act.req.Index,
		Limit:      act.req.Limit,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryConfigTemplateBindingList(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error()
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}
	act.resp.ConfigTemplateBindings = resp.ConfigTemplateBindings
	return pbcommon.ErrCode_E_OK, ""
}

// Do do action
func (act *ListAction) Do() error {
	// query config template binding list.
	if errCode, errMsg := act.listBindings(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
