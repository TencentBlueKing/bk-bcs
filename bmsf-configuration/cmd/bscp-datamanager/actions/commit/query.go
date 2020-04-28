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
	"errors"

	"github.com/bluele/gcache"
	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-datamanager/modules/metrics"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/logger"
)

// QueryAction is commit query action object.
type QueryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	collector   *metrics.Collector
	commitCache gcache.Cache

	req  *pb.QueryCommitReq
	resp *pb.QueryCommitResp

	sd *dbsharding.ShardingDB

	commit database.Commit
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager, collector *metrics.Collector, commitCache gcache.Cache,
	req *pb.QueryCommitReq, resp *pb.QueryCommitResp) *QueryAction {
	action := &QueryAction{viper: viper, smgr: smgr, collector: collector, commitCache: commitCache, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
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
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
	}
	return nil
}

func (act *QueryAction) queryCommit() (pbcommon.ErrCode, string) {
	// query commit from cache.
	if cache, err := act.commitCache.Get(act.req.Commitid); err == nil && cache != nil {
		act.collector.StatCommitCache(true)
		commit := cache.(*pbcommon.Commit)

		logger.V(3).Infof("QueryCommit[%d]| query commit cache hit success[%s]", act.req.Seq, act.req.Commitid)
		finalCommit := *commit

		if act.req.Abstract {
			finalCommit.Configs = []byte{}
			finalCommit.Template = ""
			finalCommit.TemplateRule = ""
		}
		act.resp.Commit = &finalCommit
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatCommitCache(false)

	// query commit from db.
	act.sd.AutoMigrate(&database.Commit{})

	var st database.Commit
	err := act.sd.DB().
		Where(&database.Commit{Bid: act.req.Bid, Commitid: act.req.Commitid}).
		Last(&st).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "commit non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	commit := &pbcommon.Commit{
		Bid:           st.Bid,
		Commitid:      st.Commitid,
		Appid:         st.Appid,
		Cfgsetid:      st.Cfgsetid,
		Op:            st.Op,
		Operator:      st.Operator,
		Templateid:    st.Templateid,
		Template:      st.Template,
		TemplateRule:  st.TemplateRule,
		PrevConfigs:   st.PrevConfigs,
		Configs:       st.Configs,
		Changes:       st.Changes,
		Releaseid:     st.Releaseid,
		Memo:          st.Memo,
		State:         st.State,
		CreatedAt:     st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     st.UpdatedAt.Format("2006-01-02 15:04:05"),
		MultiCommitid: st.MultiCommitid,
	}
	newCache := *commit

	if commit.State == int32(pbcommon.CommitState_CS_CONFIRMED) ||
		commit.State == int32(pbcommon.CommitState_CS_CANCELED) {
		if err := act.commitCache.Set(act.req.Commitid, &newCache); err != nil {
			logger.Warn("QueryCommit[%d]| update local commit cache, %+v", act.req.Seq, err)
		}
	}

	if act.req.Abstract {
		commit.Configs = []byte{}
		commit.Template = ""
		commit.TemplateRule = ""
	}
	act.resp.Commit = commit

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
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
