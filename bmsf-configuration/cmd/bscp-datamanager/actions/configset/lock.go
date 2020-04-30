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

package configset

import (
	"errors"
	"time"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// LockAction is configset lock action object.
type LockAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.LockConfigSetReq
	resp *pb.LockConfigSetResp

	sd *dbsharding.ShardingDB
}

// NewLockAction creates new LockAction.
func NewLockAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.LockConfigSetReq, resp *pb.LockConfigSetResp) *LockAction {
	action := &LockAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *LockAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *LockAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *LockAction) Output() error {
	// do nothing.
	return nil
}

func (act *LockAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *LockAction) lockConfigSet() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ConfigSetLock{})

	var lock database.ConfigSetLock

	err := act.sd.DB().
		Where(database.ConfigSetLock{Cfgsetid: act.req.Cfgsetid}).
		Attrs(database.ConfigSetLock{Operator: ""}).
		FirstOrCreate(&lock).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	if lock.Operator == act.req.Operator {
		// already locked by the operator.
		return pbcommon.ErrCode_E_OK, ""
	}

	if len(lock.Operator) != 0 {
		act.resp.Locker = lock.Operator
		act.resp.LockTime = lock.LockTime.Format("2006-01-02 15:04:05")
		return pbcommon.ErrCode_E_DM_CFGSET_LOCK_FAILED, "can't lock the configset now, already locked by other operator."
	}

	ups := map[string]interface{}{
		"Operator": act.req.Operator,
		"LockTime": time.Now(),
		"Memo":     act.req.Memo,
	}

	exec := act.sd.DB().
		Model(&database.ConfigSetLock{}).
		Where(&database.ConfigSetLock{Cfgsetid: act.req.Cfgsetid}).
		Where("Foperator = ''").
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_CFGSET_LOCK_FAILED, "can't lock the configset now."
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *LockAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// lock configSet.
	if errCode, errMsg := act.lockConfigSet(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
