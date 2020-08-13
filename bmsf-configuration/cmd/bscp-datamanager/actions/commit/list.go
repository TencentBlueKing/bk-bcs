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
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// ListAction is commit list action object.
type ListAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryHistoryCommitsReq
	resp *pb.QueryHistoryCommitsResp

	sd *dbsharding.ShardingDB

	commits []database.Commit
}

// NewListAction creates new ListAction.
func NewListAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryHistoryCommitsReq, resp *pb.QueryHistoryCommitsResp) *ListAction {
	action := &ListAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ListAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ListAction) Output() error {
	commits := []*pbcommon.Commit{}
	for _, st := range act.commits {
		commit := &pbcommon.Commit{
			Bid:           st.Bid,
			Commitid:      st.Commitid,
			Appid:         st.Appid,
			Cfgsetid:      st.Cfgsetid,
			Op:            st.Op,
			Operator:      st.Operator,
			Templateid:    st.Templateid,
			Releaseid:     st.Releaseid,
			MultiCommitid: st.MultiCommitid,
			Memo:          st.Memo,
			State:         st.State,
			CreatedAt:     st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		commits = append(commits, commit)
	}
	act.resp.Commits = commits
	return nil
}

func (act *ListAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	if len(act.req.Appid) == 0 && len(act.req.Cfgsetid) == 0 && len(act.req.Operator) == 0 {
		return errors.New("invalid params, appid or cfgsetid or operator missing")
	}

	if len(act.req.Appid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	if len(act.req.Cfgsetid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	if len(act.req.Operator) > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too big")
	}
	return nil
}

func (act *ListAction) queryHistoryCommits() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Commit{})

	// query type, 0:All(default)  1:Init  2:Confirmed  3:Canceled
	whereState := fmt.Sprintf("Fstate IN (%d, %d, %d)",
		pbcommon.CommitState_CS_INIT, pbcommon.CommitState_CS_CONFIRMED, pbcommon.CommitState_CS_CANCELED)

	if act.req.QueryType == 1 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.CommitState_CS_INIT)
	} else if act.req.QueryType == 2 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.CommitState_CS_CONFIRMED)
	} else if act.req.QueryType == 3 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.CommitState_CS_CANCELED)
	}

	// selected fields.
	fields := "Fid, Fcommitid, Fbid, Fappid, Fcfgsetid, Ftemplateid, Fop, " +
		"Foperator, Freleaseid, Fmemo, Fstate, Fcreate_time, Fupdate_time, Fmulti_commitid"

	err := act.sd.DB().
		Select(fields).
		Offset(act.req.Index).Limit(act.req.Limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.Commit{Bid: act.req.Bid, Appid: act.req.Appid, Cfgsetid: act.req.Cfgsetid, Operator: act.req.Operator}).
		Where(whereState).
		Find(&act.commits).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *ListAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query commits list.
	if errCode, errMsg := act.queryHistoryCommits(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
