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
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/accessserver"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/logger"
)

// DeleteAction deletes target ProcAttr object.
type DeleteAction struct {
	viper    *viper.Viper
	buSvrCli pbbusinessserver.BusinessClient

	req  *pb.DeleteProcAttrReq
	resp *pb.DeleteProcAttrResp
}

// NewDeleteAction creates new DeleteAction.
func NewDeleteAction(viper *viper.Viper, buSvrCli pbbusinessserver.BusinessClient,
	req *pb.DeleteProcAttrReq, resp *pb.DeleteProcAttrResp) *DeleteAction {
	action := &DeleteAction{viper: viper, buSvrCli: buSvrCli, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_AS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *DeleteAction) Output() error {
	// do nothing.
	return nil
}

func (act *DeleteAction) verify() error {
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

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}
	return nil
}

func (act *DeleteAction) delete() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.DeleteProcAttrReq{
		Seq:      act.req.Seq,
		Cloudid:  act.req.Cloudid,
		IP:       act.req.IP,
		Bid:      act.req.Bid,
		Appid:    act.req.Appid,
		Path:     act.req.Path,
		Operator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("DeleteProcAttr[%d]| request to businessserver DeleteProcAttr, %+v", act.req.Seq, r)

	resp, err := act.buSvrCli.DeleteProcAttr(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver DeleteProcAttr, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *DeleteAction) Do() error {
	// delete procattr.
	if errCode, errMsg := act.delete(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
