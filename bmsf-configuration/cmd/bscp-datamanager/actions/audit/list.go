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

package audit

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// ListAction is audit list action object.
type ListAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryAuditListReq
	resp *pb.QueryAuditListResp

	sd *dbsharding.ShardingDB

	audits []database.Audit
}

// NewListAction creates new ListAction.
func NewListAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryAuditListReq, resp *pb.QueryAuditListResp) *ListAction {
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
	audits := []*pbcommon.Audit{}
	for _, st := range act.audits {
		audit := &pbcommon.Audit{
			ID:         st.ID,
			SourceType: st.SourceType,
			OpType:     st.OpType,
			Sourceid:   st.Sourceid,
			Bid:        st.Bid,
			Operator:   st.Operator,
			Memo:       st.Memo,
			State:      st.State,
			CreatedAt:  st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		audits = append(audits, audit)
	}
	act.resp.Audits = audits
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

	if act.req.SourceType < int32(pbcommon.SourceType_ST_BUSINESS) ||
		act.req.SourceType >= int32(pbcommon.SourceType_ST_END) {
		return errors.New("invalid params, unknow source type")
	}
	if act.req.OpType < int32(pbcommon.SourceOpType_SOT_CREATE) ||
		act.req.OpType >= int32(pbcommon.SourceOpType_SOT_END) {
		return errors.New("invalid params, unknow source op type")
	}

	if len(act.req.Sourceid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, sourceid too long")
	}

	if len(act.req.Operator) > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too long")
	}
	return nil
}

func (act *ListAction) queryAuditList() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Audit{})

	err := act.sd.DB().
		Offset(act.req.Index).Limit(act.req.Limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.Audit{Bid: act.req.Bid, Sourceid: act.req.Sourceid, Operator: act.req.Operator}).
		Where("Fsource_type = ?", act.req.SourceType).
		Where("Fop_type = ?", act.req.OpType).
		Find(&act.audits).Error

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

	// query audit list.
	if errCode, errMsg := act.queryAuditList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
