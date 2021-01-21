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
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
)

// CreateAction is multi release create action object.
type CreateAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.CreateMultiReleaseReq
	resp *pb.CreateMultiReleaseResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB

	commits []database.Commit
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateMultiReleaseReq, resp *pb.CreateMultiReleaseResp) *CreateAction {
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
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_release_id", act.req.MultiReleaseId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("name", act.req.Name,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_commit_id", act.req.MultiCommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("strategies", act.req.Strategies,
		database.BSCPEMPTY, database.BSCPSTRATEGYCONTENTSIZELIMIT); err != nil {
		return err
	}
	if len(act.req.Strategies) == 0 {
		act.req.Strategies = strategy.EmptyStrategy
	}
	if err = common.ValidateString("creator", act.req.Creator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("strategy_id", act.req.StrategyId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CreateAction) genReleaseID() (string, pbcommon.ErrCode, string) {
	id, err := common.GenReleaseID()
	if err != nil {
		return "", pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, fmt.Sprintf("gen new release_id failed, %+v", err)
	}
	return id, pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createMultiRelease() (pbcommon.ErrCode, string) {
	release := &database.MultiRelease{
		BizID:          act.req.BizId,
		MultiReleaseID: act.req.MultiReleaseId,
		Name:           act.req.Name,
		AppID:          act.req.AppId,
		StrategyID:     act.req.StrategyId,
		Strategies:     act.req.Strategies,
		MultiCommitID:  act.req.MultiCommitId,
		Creator:        act.req.Creator,
		LastModifyBy:   act.req.Creator,
		Memo:           act.req.Memo,
		State:          int32(pbcommon.ReleaseState_RS_INIT),
	}

	err := act.tx.
		Create(release).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Data = &pb.CreateMultiReleaseResp_RespData{MultiReleaseId: act.req.MultiReleaseId}

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createRelease(commitID, releaseID string, config *database.Config) (pbcommon.ErrCode, string) {
	release := &database.Release{
		BizID:          act.req.BizId,
		ReleaseID:      releaseID,
		Name:           act.req.Name,
		AppID:          config.AppID,
		CfgID:          config.CfgID,
		CfgName:        config.Name,
		CfgFpath:       config.Fpath,
		User:           config.User,
		UserGroup:      config.UserGroup,
		FilePrivilege:  config.FilePrivilege,
		FileFormat:     config.FileFormat,
		FileMode:       config.FileMode,
		StrategyID:     act.req.StrategyId,
		Strategies:     act.req.Strategies,
		CommitID:       commitID,
		Creator:        act.req.Creator,
		LastModifyBy:   act.req.Creator,
		Memo:           act.req.Memo,
		State:          int32(pbcommon.ReleaseState_RS_INIT),
		MultiReleaseID: act.req.MultiReleaseId,
	}

	err := act.tx.
		Create(release).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryConfig(cfgID string) (*database.Config, pbcommon.ErrCode, string) {
	config := database.Config{}

	err := act.tx.
		Where(&database.Config{
			BizID: act.req.BizId,
			AppID: act.req.AppId,
			CfgID: cfgID,
		}).
		Last(&config).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return nil, pbcommon.ErrCode_E_DM_NOT_FOUND, "config non-exist."
	}
	if err != nil {
		return nil, pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return &config, pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createReleases() (pbcommon.ErrCode, string) {
	for _, commit := range act.commits {
		cfg, errCode, errMsg := act.queryConfig(commit.CfgID)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}

		releaseID, errCode, errMsg := act.genReleaseID()
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
		if errCode, errMsg := act.createRelease(commit.CommitID, releaseID, cfg); errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) querySubCommits() (pbcommon.ErrCode, string) {
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
func (act *CreateAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// create multi release.
	if errCode, errMsg := act.createMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// query sub commits.
	if errCode, errMsg := act.querySubCommits(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// create releases.
	if errCode, errMsg := act.createReleases(); errCode != pbcommon.ErrCode_E_OK {
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
