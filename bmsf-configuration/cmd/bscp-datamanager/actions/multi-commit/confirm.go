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

package multicommit

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// ConfirmAction is multi commit confirm action object.
type ConfirmAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.ConfirmMultiCommitReq
	resp *pb.ConfirmMultiCommitResp

	sd *dbsharding.ShardingDB
}

// NewConfirmAction creates new ConfirmAction.
func NewConfirmAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.ConfirmMultiCommitReq, resp *pb.ConfirmMultiCommitResp) *ConfirmAction {
	action := &ConfirmAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ConfirmAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ConfirmAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ConfirmAction) Output() error {
	// do nothing.
	return nil
}

func (act *ConfirmAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.MultiCommitid)
	if length == 0 {
		return errors.New("invalid params, multi commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, multi commitid too long")
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

func (act *ConfirmAction) confirmMultiCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.MultiCommit{})

	ups := map[string]interface{}{
		"State":    pbcommon.CommitState_CS_CONFIRMED,
		"Operator": act.req.Operator,
	}

	exec := act.sd.DB().
		Model(&database.MultiCommit{}).
		Where(&database.MultiCommit{Bid: act.req.Bid, MultiCommitid: act.req.MultiCommitid}).
		Where("Fstate = ?", pbcommon.CommitState_CS_INIT).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "confirm the multi commit failed(multi commit no-exist or not in init state)."
	}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *ConfirmAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// confirm multi commit.
	if errCode, errMsg := act.confirmMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
