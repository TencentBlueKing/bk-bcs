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

// ListAction is audit list action object.
type ListAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryAuditListReq
	resp *pb.QueryAuditListResp

	sd *dbsharding.ShardingDB

	totalCount int64
	audits     []database.Audit
}

// NewListAction creates new ListAction.
func NewListAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryAuditListReq, resp *pb.QueryAuditListResp) *ListAction {
	action := &ListAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
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
			Id:         st.ID,
			SourceType: st.SourceType,
			OpType:     st.OpType,
			SourceId:   st.SourceID,
			BizId:      st.BizID,
			Operator:   st.Operator,
			Memo:       st.Memo,
			State:      st.State,
			CreatedAt:  st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		audits = append(audits, audit)
	}
	act.resp.Data = &pb.QueryAuditListResp_RespData{TotalCount: uint32(act.totalCount), Info: audits}
	return nil
}

func (act *ListAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateInt32("source_type", act.req.SourceType,
		int32(pbcommon.SourceType_ST_APP), int32(pbcommon.SourceType_ST_END)); err != nil {
		return err
	}
	if err = common.ValidateInt32("op_type", act.req.OpType,
		int32(pbcommon.SourceOpType_SOT_CREATE), int32(pbcommon.SourceOpType_SOT_END)); err != nil {
		return err
	}
	if err = common.ValidateString("source_id", act.req.SourceId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPEMPTY, database.BSCPNAMELENLIMIT); err != nil {
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

func (act *ListAction) queryAuditCount() (pbcommon.ErrCode, string) {
	if !act.req.Page.ReturnTotal {
		return pbcommon.ErrCode_E_OK, ""
	}

	err := act.sd.DB().
		Model(&database.Audit{}).
		Where(&database.Audit{
			BizID:    act.req.BizId,
			SourceID: act.req.SourceId,
			Operator: act.req.Operator,
		}).
		Where("Fsource_type = ?", act.req.SourceType).
		Where("Fop_type = ?", act.req.OpType).
		Count(&act.totalCount).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *ListAction) queryAuditList() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Offset(int(act.req.Page.Start)).Limit(int(act.req.Page.Limit)).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.Audit{
			BizID:    act.req.BizId,
			SourceID: act.req.SourceId,
			Operator: act.req.Operator,
		}).
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
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query audit count.
	if errCode, errMsg := act.queryAuditCount(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query audit list.
	if errCode, errMsg := act.queryAuditList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
