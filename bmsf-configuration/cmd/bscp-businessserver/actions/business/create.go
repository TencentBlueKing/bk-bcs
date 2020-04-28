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

	"bk-bscp/cmd/bscp-businessserver/modules/audit"
	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// CreateAction creates a business object.
type CreateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateBusinessReq
	resp *pb.CreateBusinessResp

	newBid string
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateBusinessReq, resp *pb.CreateBusinessResp) *CreateAction {
	action := &CreateAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
	length := len(act.req.Name)
	if length == 0 {
		return errors.New("invalid params, name missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, name too long")
	}

	length = len(act.req.Depid)
	if length == 0 {
		return errors.New("invalid params, depid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, depid too long")
	}

	length = len(act.req.Dbid)
	if length == 0 {
		return errors.New("invalid params, dbid missing")
	}
	if length > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, dbid too long")
	}

	length = len(act.req.Dbname)
	if length == 0 {
		return errors.New("invalid params, dbname missing")
	}
	if length > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, dbname too long")
	}

	length = len(act.req.Creator)
	if length == 0 {
		return errors.New("invalid params, creator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, creator too long")
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}

	if len(act.req.Auth) > database.BSCPAUTHLENLIMIT {
		return errors.New("invalid params, auth too long")
	}
	return nil
}

func (act *CreateAction) genBusinessID() error {
	id, err := common.GenBid()
	if err != nil {
		return err
	}
	act.newBid = id
	return nil
}

func (act *CreateAction) createSharding() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateShardingReq{
		Seq:    act.req.Seq,
		Key:    act.newBid,
		Dbid:   act.req.Dbid,
		Dbname: act.req.Dbname,
		Memo:   act.req.Memo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateBusiness[%d]| request to datamanager CreateSharding, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateSharding(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateSharding, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *CreateAction) createBusiness() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateBusinessReq{
		Seq:     act.req.Seq,
		Bid:     act.newBid,
		Name:    act.req.Name,
		Depid:   act.req.Depid,
		Creator: act.req.Creator,
		Memo:    act.req.Memo,
		Auth:    act.req.Auth,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateBusiness[%d]| request to datamanager CreateBusiness, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateBusiness(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateBusiness, %+v", err)
	}
	act.resp.Bid = resp.Bid

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on new business created.
	audit.Audit(int32(pbcommon.SourceType_ST_BUSINESS), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.resp.Bid, act.resp.Bid, act.req.Creator, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	if err := act.genBusinessID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
	}

	// create sharding for the new business.
	if errCode, errMsg := act.createSharding(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create a new business.
	if errCode, errMsg := act.createBusiness(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
