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

// PublishAction is release publish action object.
type PublishAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.PublishReleaseReq
	resp *pb.PublishReleaseResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB

	release database.Release
	commit  database.Commit
}

// NewPublishAction creates new PublishAction.
func NewPublishAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.PublishReleaseReq, resp *pb.PublishReleaseResp) *PublishAction {

	action := &PublishAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PublishAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PublishAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PublishAction) Output() error {
	// do nothing.
	return nil
}

func (act *PublishAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("release_id", act.req.ReleaseId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *PublishAction) queryRelease() (pbcommon.ErrCode, string) {
	err := act.tx.
		Where(&database.Release{BizID: act.req.BizId, ReleaseID: act.req.ReleaseId}).
		Last(&act.release).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "release non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) queryCommit() (pbcommon.ErrCode, string) {
	err := act.tx.
		Where(&database.Commit{
			BizID:    act.req.BizId,
			CommitID: act.release.CommitID,
			State:    int32(pbcommon.CommitState_CS_CONFIRMED),
		}).
		Last(&act.commit).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "target release confirmed commit non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) publishRelease() (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"State":        int32(pbcommon.ReleaseState_RS_PUBLISHED),
		"LastModifyBy": act.req.Operator,
	}

	exec := act.tx.Model(&database.Release{}).
		Where(&database.Release{BizID: act.req.BizId, ReleaseID: act.req.ReleaseId}).
		Where("Fstate IN (?, ?)", pbcommon.ReleaseState_RS_INIT, pbcommon.ReleaseState_RS_PUBLISHED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_PUBLISH_RELEASE_FAILED, "no update for the release"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) updateCommit() (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"ReleaseID": act.req.ReleaseId,
	}

	exec := act.tx.Model(&database.Commit{}).
		Where(&database.Commit{BizID: act.req.BizId, CommitID: act.commit.CommitID}).
		Where("Fstate = ?", pbcommon.CommitState_CS_CONFIRMED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "no update for the commit"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) updateConfig() (pbcommon.ErrCode, string) {
	ups := map[string]interface{}{
		"State": int32(pbcommon.ConfigState_CS_RELEASED),
	}

	exec := act.tx.Model(&database.Config{}).
		Where(&database.Config{BizID: act.req.BizId, CfgID: act.commit.CfgID}).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "no update for the config"
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PublishAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// query release.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// query commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// publish release.
	if errCode, errMsg := act.publishRelease(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// update commit.
	if errCode, errMsg := act.updateCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// update config.
	if errCode, errMsg := act.updateConfig(); errCode != pbcommon.ErrCode_E_OK {
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
