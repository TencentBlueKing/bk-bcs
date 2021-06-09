/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package variable

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

// QueryAction query target variable object.
type QueryAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryVariableReq
	resp *pb.QueryVariableResp

	sd *dbsharding.ShardingDB

	variable database.Variable
}

// NewQueryAction create new QueryAction.
func NewQueryAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryVariableReq, resp *pb.QueryVariableResp) *QueryAction {
	action := &QueryAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

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

// Output handles the output messages
func (act *QueryAction) Output() error {
	variable := &pbcommon.Variable{
		BizId:        act.variable.BizID,
		VarId:        act.variable.VarID,
		VarGroupId:   act.variable.VarGroupID,
		Name:         act.variable.Name,
		Value:        act.variable.Value,
		Memo:         act.variable.Memo,
		Creator:      act.variable.Creator,
		LastModifyBy: act.variable.LastModifyBy,
		State:        act.variable.State,
		CreatedAt:    act.variable.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    act.variable.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	act.resp.Data = variable
	return nil
}

func (act *QueryAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if len(act.req.VarId) == 0 && (len(act.req.VarGroupId) == 0 || len(act.req.Name) == 0) {
		return errors.New("invalid input data, var_id or var_group_id + name is required")
	}

	if err = common.ValidateString("var_id", act.req.VarId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("var_group_id", act.req.VarGroupId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("name", act.req.Name,
		database.BSCPEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}

	return nil
}

func (act *QueryAction) queryVariable() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Where(&database.Variable{
			BizID:      act.req.BizId,
			VarID:      act.req.VarId,
			VarGroupID: act.req.VarGroupId,
			Name:       act.req.Name,
		}).
		Last(&act.variable).Error

	if err == database.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "variable not found"
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query variable.
	if errCode, errMsg := act.queryVariable(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
