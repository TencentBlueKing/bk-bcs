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

	"github.com/bluele/gcache"
	"github.com/spf13/viper"

	"bk-bscp/cmd/atomic-services/bscp-datamanager/modules/metrics"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// QueryAction is commit query action object.
type QueryAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	collector   *metrics.Collector
	commitCache gcache.Cache

	req  *pb.QueryCommitReq
	resp *pb.QueryCommitResp

	sd *dbsharding.ShardingDB
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	collector *metrics.Collector, commitCache gcache.Cache,
	req *pb.QueryCommitReq, resp *pb.QueryCommitResp) *QueryAction {

	action := &QueryAction{ctx: ctx, viper: viper, smgr: smgr, collector: collector,
		commitCache: commitCache, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryAction) Output() error {
	// do nothing.
	return nil
}

func (act *QueryAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("commit_id", act.req.CommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *QueryAction) queryCommit() (pbcommon.ErrCode, string) {
	// query commit from cache.
	if cache, err := act.commitCache.Get(act.req.CommitId); err == nil && cache != nil {
		act.collector.StatCommitCache(true)
		logger.V(3).Infof("QueryCommit[%s]| query commit cache hit success[%s]", act.req.Seq, act.req.CommitId)

		commit := cache.(*pbcommon.Commit)
		act.resp.Data = commit
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatCommitCache(false)

	// query commit from db.
	var st database.Commit
	err := act.sd.DB().
		Where(&database.Commit{BizID: act.req.BizId, CommitID: act.req.CommitId}).
		Last(&st).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "commit non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	commit := &pbcommon.Commit{
		BizId:         st.BizID,
		CommitId:      st.CommitID,
		AppId:         st.AppID,
		CfgId:         st.CfgID,
		Operator:      st.Operator,
		CommitMode:    st.CommitMode,
		ReleaseId:     st.ReleaseID,
		Memo:          st.Memo,
		State:         st.State,
		CreatedAt:     st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     st.UpdatedAt.Format("2006-01-02 15:04:05"),
		MultiCommitId: st.MultiCommitID,
	}

	if commit.State == int32(pbcommon.CommitState_CS_CONFIRMED) ||
		commit.State == int32(pbcommon.CommitState_CS_CANCELED) {
		if err := act.commitCache.Set(act.req.CommitId, commit); err != nil {
			logger.Warn("QueryCommit[%s]| update local commit cache, %+v", act.req.Seq, err)
		}
	}
	act.resp.Data = commit

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
