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

package sharding

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// UpdateAction is starding update action object.
type UpdateAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.UpdateShardingReq
	resp *pb.UpdateShardingResp
}

// NewUpdateAction creates new UpdateAction.
func NewUpdateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.UpdateShardingReq, resp *pb.UpdateShardingResp) *UpdateAction {
	action := &UpdateAction{viper: viper, smgr: smgr, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *UpdateAction) Output() error {
	// do nothing.
	return nil
}

func (act *UpdateAction) verify() error {
	length := len(act.req.Key)
	if length == 0 {
		return errors.New("invalid params, key missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, key too long")
	}

	if len(act.req.Dbid) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, dbid too long")
	}

	if len(act.req.Dbname) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, dbname too long")
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *UpdateAction) updateSharding() (pbcommon.ErrCode, string) {
	sharding := &pbcommon.Sharding{
		Key:    act.req.Key,
		Dbid:   act.req.Dbid,
		Dbname: act.req.Dbname,
		Memo:   act.req.Memo,
		State:  act.req.State,
	}

	if err := act.smgr.UpdateSharding(sharding); err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *UpdateAction) Do() error {
	// update sharding.
	if errCode, errMsg := act.updateSharding(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
