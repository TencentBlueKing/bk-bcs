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
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// PublishAction is multi release publish action object.
type PublishAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.PublishMultiReleaseReq
	resp *pb.PublishMultiReleaseResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB

	multiRelease database.MultiRelease
	multiCommit  database.MultiCommit
}

// NewPublishAction creates new PublishAction.
func NewPublishAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.PublishMultiReleaseReq, resp *pb.PublishMultiReleaseResp) *PublishAction {
	action := &PublishAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PublishAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
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
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.MultiReleaseid)
	if length == 0 {
		return errors.New("invalid params, multi releaseid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, multi releaseid too long")
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

func (act *PublishAction) queryMultiRelease() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.MultiRelease{})

	err := act.tx.Where(&database.MultiRelease{Bid: act.req.Bid, MultiReleaseid: act.req.MultiReleaseid}).
		Last(&act.multiRelease).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "multi release non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) queryMultiCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.MultiCommit{})

	err := act.tx.Where(&database.MultiCommit{Bid: act.req.Bid, MultiCommitid: act.multiRelease.MultiCommitid, State: int32(pbcommon.CommitState_CS_CONFIRMED)}).
		Last(&act.multiCommit).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "target multi release confirmed multi commit non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) publishMultiRelease() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.MultiRelease{})

	ups := map[string]interface{}{
		"State":        int32(pbcommon.ReleaseState_RS_PUBLISHED),
		"LastModifyBy": act.req.Operator,
	}

	exec := act.tx.Model(&database.MultiRelease{}).
		Where(&database.MultiRelease{Bid: act.req.Bid, MultiReleaseid: act.req.MultiReleaseid}).
		Where("Fstate IN (?, ?)", pbcommon.ReleaseState_RS_INIT, pbcommon.ReleaseState_RS_PUBLISHED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_PUBLISH_RELEASE_FAILED, "publish the multi release failed, there is no release that fit the conditions."
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) updateMultiCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.MultiCommit{})

	ups := map[string]interface{}{
		"MultiReleaseid": act.req.MultiReleaseid,
		"LastModifyBy":   act.req.Operator,
	}

	exec := act.tx.Model(&database.MultiCommit{}).
		Where(&database.MultiCommit{Bid: act.req.Bid, MultiCommitid: act.multiCommit.MultiCommitid}).
		Where("Fstate = ?", pbcommon.CommitState_CS_CONFIRMED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "publish multi release and update the commit failed(commit no-exist or not in confirmed state)."
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PublishAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// query multi release.
	if errCode, errMsg := act.queryMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// query multi commit.
	if errCode, errMsg := act.queryMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// publish multi release.
	if errCode, errMsg := act.publishMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// update multi commit.
	if errCode, errMsg := act.updateMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}
	act.tx.Commit()

	return nil
}
