/* Tencent is pleased to support the open source community by making Blueking Container Service available.
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
	"context"
	"errors"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// CancelAction is multi commit cancel action object.
type CancelAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CancelMultiCommitReq
	resp *pb.CancelMultiCommitResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB

	commits []database.Commit
}

// NewCancelAction creates new CancelAction.
func NewCancelAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CancelMultiCommitReq, resp *pb.CancelMultiCommitResp) *CancelAction {

	action := &CancelAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CancelAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CancelAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CancelAction) Output() error {
	// do nothing.
	return nil
}

func (act *CancelAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_commit_id", act.req.MultiCommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CancelAction) cancelMultiCommit() (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"State":    pbcommon.CommitState_CS_CANCELED,
		"Operator": act.req.Operator,
	}

	exec := act.tx.
		Model(&database.MultiCommit{}).
		Where(&database.MultiCommit{BizID: act.req.BizId, MultiCommitID: act.req.MultiCommitId}).
		Where("Fstate != ?", pbcommon.CommitState_CS_CONFIRMED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "no update for the multi commit"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CancelAction) cancelCommit(commitID string) (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"State":    pbcommon.CommitState_CS_CANCELED,
		"Operator": act.req.Operator,
	}

	exec := act.tx.
		Model(&database.Commit{}).
		Where(&database.Commit{BizID: act.req.BizId, CommitID: commitID}).
		Where("Fstate != ?", pbcommon.CommitState_CS_CONFIRMED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "no update for the commit"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CancelAction) cancelCommits() (pbcommon.ErrCode, string) {
	for _, commit := range act.commits {
		errCode, errMsg := act.cancelCommit(commit.CommitID)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CancelAction) querySubCommits() (pbcommon.ErrCode, string) {
	err := act.tx.
		Model(&database.Commit{}).
		Where(&database.Commit{BizID: act.req.BizId, MultiCommitID: act.req.MultiCommitId}).
		Find(&act.commits).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CancelAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// cancel multi commit.
	if errCode, errMsg := act.cancelMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// query sub commits.
	if errCode, errMsg := act.querySubCommits(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// cancel commits.
	if errCode, errMsg := act.cancelCommits(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// commit tx.
	if err := act.tx.Commit().Error; err != nil {
		act.tx.Rollback()
		return act.Err(pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, err.Error())
	}

	return nil
}
