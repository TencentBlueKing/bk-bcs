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
	"context"
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// CreateAction is commit create action object.
type CreateAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateCommitReq
	resp *pb.CreateCommitResp

	sd *dbsharding.ShardingDB
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateCommitReq, resp *pb.CreateCommitResp) *CreateAction {
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
	if err = common.ValidateString("commit_id", act.req.CommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cfg_id", act.req.CfgId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateInt32("commit_mode", act.req.CommitMode,
		int32(pbcommon.CommitMode_CM_CONFIGS), int32(pbcommon.CommitMode_CM_TEMPLATE)); err != nil {
		return err
	}
	if err = common.ValidateString("multi_commit_id", act.req.MultiCommitId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CreateAction) createCommit() (pbcommon.ErrCode, string) {
	commit := &database.Commit{
		BizID:      act.req.BizId,
		CommitID:   act.req.CommitId,
		AppID:      act.req.AppId,
		CfgID:      act.req.CfgId,
		CommitMode: act.req.CommitMode,
		Operator:   act.req.Operator,
		Memo:       act.req.Memo,
		State:      int32(pbcommon.CommitState_CS_INIT),

		// normale mode, multi commitid is current commitid(unique index).
		MultiCommitID: act.req.CommitId,
	}

	err := act.sd.DB().
		Create(commit).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Data = &pb.CreateCommitResp_RespData{CommitId: act.req.CommitId}

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createCommitMultiMode() (pbcommon.ErrCode, string) {
	commit := database.Commit{
		BizID:         act.req.BizId,
		CommitID:      act.req.CommitId,
		AppID:         act.req.AppId,
		CfgID:         act.req.CfgId,
		CommitMode:    act.req.CommitMode,
		Operator:      act.req.Operator,
		Memo:          act.req.Memo,
		State:         int32(pbcommon.CommitState_CS_INIT),
		MultiCommitID: act.req.MultiCommitId,
	}

	err := act.sd.DB().
		Where(database.Commit{
			BizID:         act.req.BizId,
			AppID:         act.req.AppId,
			CfgID:         act.req.CfgId,
			MultiCommitID: act.req.MultiCommitId,
		}).
		Assign(commit).
		FirstOrCreate(&commit).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Data = &pb.CreateCommitResp_RespData{CommitId: act.req.CommitId}

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

	if len(act.req.MultiCommitId) == 0 {
		// create commit.
		if errCode, errMsg := act.createCommit(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else {
		// create commit in multi mode.
		if errCode, errMsg := act.createCommitMultiMode(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}
	return nil
}
