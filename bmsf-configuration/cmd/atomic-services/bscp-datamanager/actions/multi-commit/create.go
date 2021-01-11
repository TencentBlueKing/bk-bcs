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

// CreateAction is multi commit create action object.
type CreateAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateMultiCommitReq
	resp *pb.CreateMultiCommitResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateMultiCommitReq, resp *pb.CreateMultiCommitResp) *CreateAction {
	action := &CreateAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_commit_id", act.req.MultiCommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if len(act.req.Metadatas) == 0 {
		return errors.New("invalid input data, empty metadatas")
	}

	for _, metadata := range act.req.Metadatas {
		if err = common.ValidateString("cfg_id", metadata.CfgId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateInt32("commit_mode", metadata.CommitMode,
			int32(pbcommon.CommitMode_CM_CONFIGS), int32(pbcommon.CommitMode_CM_TEMPLATE)); err != nil {
			return err
		}
	}

	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CreateAction) createCommitMultiMode(cfgID, commitID string, commitMode int32) (pbcommon.ErrCode, string) {
	commit := database.Commit{
		BizID:         act.req.BizId,
		AppID:         act.req.AppId,
		CommitID:      commitID,
		CfgID:         cfgID,
		CommitMode:    commitMode,
		Operator:      act.req.Operator,
		Memo:          act.req.Memo,
		State:         int32(pbcommon.CommitState_CS_INIT),
		MultiCommitID: act.req.MultiCommitId,
	}

	err := act.tx.
		Where(database.Commit{
			BizID:         act.req.BizId,
			AppID:         act.req.AppId,
			CfgID:         cfgID,
			MultiCommitID: act.req.MultiCommitId,
		}).
		Assign(commit).
		FirstOrCreate(&commit).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createCommits() (pbcommon.ErrCode, string) {
	for _, metadata := range act.req.Metadatas {
		errCode, errMsg := act.createCommitMultiMode(metadata.CfgId, metadata.CommitId, metadata.CommitMode)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createMultiCommit() (pbcommon.ErrCode, string) {
	commit := &database.MultiCommit{
		BizID:         act.req.BizId,
		AppID:         act.req.AppId,
		MultiCommitID: act.req.MultiCommitId,
		Operator:      act.req.Operator,
		Memo:          act.req.Memo,
		State:         int32(pbcommon.CommitState_CS_INIT),
	}

	err := act.tx.
		Create(commit).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Data = &pb.CreateMultiCommitResp_RespData{MultiCommitId: act.req.MultiCommitId}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// create multi commit.
	if errCode, errMsg := act.createMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// create commits.
	if errCode, errMsg := act.createCommits(); errCode != pbcommon.ErrCode_E_OK {
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
