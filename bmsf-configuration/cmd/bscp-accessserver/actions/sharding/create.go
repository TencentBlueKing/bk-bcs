/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package sharding

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

// CreateAction creates a sharding object.
type CreateAction struct {
	viper    *viper.Viper
	buSvrCli pbbusinessserver.BusinessClient

	req  *pb.CreateShardingReq
	resp *pb.CreateShardingResp
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, buSvrCli pbbusinessserver.BusinessClient,
	req *pb.CreateShardingReq, resp *pb.CreateShardingResp) *CreateAction {
	action := &CreateAction{viper: viper, buSvrCli: buSvrCli, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_AS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
	length := len(act.req.Key)
	if length == 0 {
		return errors.New("invalid params, key missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, key too long")
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

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.CreateShardingReq{
		Seq:    act.req.Seq,
		Key:    act.req.Key,
		Dbid:   act.req.Dbid,
		Dbname: act.req.Dbname,
		Memo:   act.req.Memo,
		State:  act.req.State,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateSharding[%d]| request to businessserver CreateSharding, %+v", act.req.Seq, r)

	resp, err := act.buSvrCli.CreateSharding(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver CreateSharding, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// create sharding.
	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
