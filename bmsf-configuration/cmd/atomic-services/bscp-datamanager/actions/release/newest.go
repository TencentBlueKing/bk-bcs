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
	"math"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// NewestAction is newest release query action object.
type NewestAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryNewestReleasesReq
	resp *pb.QueryNewestReleasesResp

	sd *dbsharding.ShardingDB
}

// NewNewestAction creates new NewestAction.
func NewNewestAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryNewestReleasesReq, resp *pb.QueryNewestReleasesResp) *NewestAction {
	action := &NewestAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *NewestAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *NewestAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *NewestAction) Output() error {
	// do nothing.
	return nil
}

func (act *NewestAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if err = common.ValidateString("cfg_id", act.req.CfgId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if act.req.Page == nil {
		return errors.New("invalid input data, page is required")
	}
	if err = common.ValidateInt32("page.start", act.req.Page.Start,
		database.BSCPEMPTY, math.MaxInt32); err != nil {
		return err
	}
	if err = common.ValidateInt32("page.limit", act.req.Page.Limit,
		database.BSCPNOTEMPTY, database.BSCPQUERYLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *NewestAction) queryNewestReleases() (pbcommon.ErrCode, string) {
	var sts []database.Release

	// release tree must be sequential according to the ID order, not the update time.
	err := act.sd.DB().
		Offset(int(act.req.Page.Start)).Limit(int(act.req.Page.Limit)).
		Order("Fid DESC").
		Where(&database.Release{
			BizID: act.req.BizId,
			CfgID: act.req.CfgId,
			State: int32(pbcommon.ReleaseState_RS_PUBLISHED),
		}).
		Find(&sts).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	var releases []*pbcommon.Release
	for _, st := range sts {
		release := &pbcommon.Release{
			Id:             st.ID,
			BizId:          st.BizID,
			ReleaseId:      st.ReleaseID,
			Name:           st.Name,
			AppId:          st.AppID,
			CfgId:          st.CfgID,
			CfgName:        st.CfgName,
			CfgFpath:       st.CfgFpath,
			User:           st.User,
			UserGroup:      st.UserGroup,
			FilePrivilege:  st.FilePrivilege,
			FileFormat:     st.FileFormat,
			FileMode:       st.FileMode,
			CommitId:       st.CommitID,
			MultiReleaseId: st.MultiReleaseID,
			StrategyId:     st.StrategyID,
			Strategies:     st.Strategies,
			Creator:        st.Creator,
			Memo:           st.Memo,
			State:          st.State,
			LastModifyBy:   st.LastModifyBy,
			CreatedAt:      st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		releases = append(releases, release)
	}
	act.resp.Data = &pb.QueryNewestReleasesResp_RespData{Info: releases}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *NewestAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query newest releases.
	if errCode, errMsg := act.queryNewestReleases(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
