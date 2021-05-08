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

package multirelease

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

// RollbackAction is multi release rollback action object.
type RollbackAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.RollbackMultiReleaseReq
	resp *pb.RollbackMultiReleaseResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB

	releases []database.Release
}

// NewRollbackAction creates new RollbackAction.
func NewRollbackAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.RollbackMultiReleaseReq, resp *pb.RollbackMultiReleaseResp) *RollbackAction {
	action := &RollbackAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *RollbackAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
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
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_release_id", act.req.MultiReleaseId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *RollbackAction) rollbackMultiRelease() (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"State":        pbcommon.ReleaseState_RS_ROLLBACKED,
		"LastModifyBy": act.req.Operator,
	}

	exec := act.tx.
		Model(&database.MultiRelease{}).
		Where(&database.MultiRelease{BizID: act.req.BizId, MultiReleaseID: act.req.MultiReleaseId}).
		Where("Fstate IN (?, ?)", pbcommon.ReleaseState_RS_PUBLISHED, pbcommon.ReleaseState_RS_ROLLBACKED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "no update for the multi release"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) rollbackRelease(releaseID string) (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"State":        pbcommon.ReleaseState_RS_ROLLBACKED,
		"LastModifyBy": act.req.Operator,
	}

	exec := act.tx.
		Model(&database.Release{}).
		Where(&database.Release{BizID: act.req.BizId, ReleaseID: releaseID}).
		Where("Fstate IN (?, ?)", pbcommon.ReleaseState_RS_PUBLISHED, pbcommon.ReleaseState_RS_ROLLBACKED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "no update for the release"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) rollbackReleases() (pbcommon.ErrCode, string) {
	for _, release := range act.releases {
		errCode, errMsg := act.rollbackRelease(release.ReleaseID)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *RollbackAction) querySubReleases() (pbcommon.ErrCode, string) {
	err := act.tx.
		Model(&database.Release{}).
		Where(&database.Release{BizID: act.req.BizId, MultiReleaseID: act.req.MultiReleaseId}).
		Find(&act.releases).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *RollbackAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// rollback multi release.
	if errCode, errMsg := act.rollbackMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query sub releases.
	if errCode, errMsg := act.querySubReleases(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// rollback releases.
	if errCode, errMsg := act.rollbackReleases(); errCode != pbcommon.ErrCode_E_OK {
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
