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

package commit

import (
	"errors"

	"github.com/bluele/gcache"
	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// ConfirmAction is commit confirm action object.
type ConfirmAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	commitCache gcache.Cache

	req  *pb.ConfirmCommitReq
	resp *pb.ConfirmCommitResp

	sd *dbsharding.ShardingDB

	commit     database.Commit
	lastCommit database.Commit
}

// NewConfirmAction creates new ConfirmAction.
func NewConfirmAction(viper *viper.Viper, smgr *dbsharding.ShardingManager, commitCache gcache.Cache,
	req *pb.ConfirmCommitReq, resp *pb.ConfirmCommitResp) *ConfirmAction {
	action := &ConfirmAction{viper: viper, smgr: smgr, commitCache: commitCache, req: req, resp: resp}

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

	length = len(act.req.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
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

func (act *ConfirmAction) queryCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Commit{})

	err := act.sd.DB().
		Where(&database.Commit{Bid: act.req.Bid, Appid: act.req.Appid, Commitid: act.req.Commitid}).
		Last(&act.commit).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "commit non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *ConfirmAction) queryLastCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Commit{})

	err := act.sd.DB().
		Where(&database.Commit{Bid: act.req.Bid, Appid: act.req.Appid, Cfgsetid: act.req.Cfgsetid, State: int32(pbcommon.CommitState_CS_CONFIRMED)}).
		Where("Freleaseid != ''").
		Where("Fid < ?", act.commit.ID).
		Last(&act.lastCommit).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		// whatever last commt.
		return pbcommon.ErrCode_E_OK, ""
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *ConfirmAction) confirmCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Commit{})

	prevConfigs := act.lastCommit.Configs
	if prevConfigs == nil {
		prevConfigs = []byte{}
	}

	ups := map[string]interface{}{
		"PrevConfigs": prevConfigs,
		"State":       pbcommon.CommitState_CS_CONFIRMED,
		"Operator":    act.req.Operator,
	}

	exec := act.sd.DB().
		Model(&database.Commit{}).
		Where(&database.Commit{Bid: act.req.Bid, Commitid: act.req.Commitid}).
		Where("Fstate = ?", pbcommon.CommitState_CS_INIT).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "confirm the commit failed(commit no-exist or not in init state)."
	}
	act.commitCache.Remove(act.req.Commitid)

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

	// query commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query last commit.
	if errCode, errMsg := act.queryLastCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// confirm commit.
	if errCode, errMsg := act.confirmCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
