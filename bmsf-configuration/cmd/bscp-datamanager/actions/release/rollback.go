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

package release

import (
	"errors"

	"github.com/bluele/gcache"
	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// RollbackAction is release rollback action object.
type RollbackAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	releaseCache gcache.Cache

	req  *pb.RollbackReleaseReq
	resp *pb.RollbackReleaseResp

	sd *dbsharding.ShardingDB
}

// NewRollbackAction creates new RollbackAction.
func NewRollbackAction(viper *viper.Viper, smgr *dbsharding.ShardingManager, releaseCache gcache.Cache,
	req *pb.RollbackReleaseReq, resp *pb.RollbackReleaseResp) *RollbackAction {
	action := &RollbackAction{viper: viper, smgr: smgr, releaseCache: releaseCache, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *RollbackAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *RollbackAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *RollbackAction) Output() error {
	// do nothing.
	return nil
}

func (act *RollbackAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Releaseid)
	if length == 0 {
		return errors.New("invalid params, releaseid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, releaseid too long")
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

func (act *RollbackAction) rollbackRelease() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Release{})

	ups := map[string]interface{}{
		"State":        pbcommon.ReleaseState_RS_ROLLBACKED,
		"LastModifyBy": act.req.Operator,
	}

	exec := act.sd.DB().
		Model(&database.Release{}).
		Where(&database.Release{Bid: act.req.Bid, Releaseid: act.req.Releaseid}).
		Where("Fstate IN (?, ?)", pbcommon.ReleaseState_RS_PUBLISHED, pbcommon.ReleaseState_RS_ROLLBACKED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "rollback the release failed(release no-exist or not published)."
	}
	act.releaseCache.Remove(act.req.Releaseid)

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *RollbackAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// rollback release.
	if errCode, errMsg := act.rollbackRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
