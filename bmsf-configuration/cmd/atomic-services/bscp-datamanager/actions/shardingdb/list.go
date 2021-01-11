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

package shardingdb

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

// ListAction is shardingdb list action object.
type ListAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryShardingDBListReq
	resp *pb.QueryShardingDBListResp

	totalCount int64
	dbs        []*pbcommon.ShardingDB
}

// NewListAction creates new ListAction.
func NewListAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryShardingDBListReq, resp *pb.QueryShardingDBListResp) *ListAction {
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
	act.resp.Data = &pb.QueryShardingDBListResp_RespData{TotalCount: uint32(act.totalCount), Info: act.dbs}
	return nil
}

func (act *ListAction) verify() error {
	var err error

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

func (act *ListAction) queryShardingDBCount() (pbcommon.ErrCode, string) {
	if !act.req.Page.ReturnTotal {
		return pbcommon.ErrCode_E_OK, "OK"
	}

	totalCount, err := act.smgr.QueryShardingDBCount()
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.totalCount = totalCount

	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *ListAction) queryShardingDBList() (pbcommon.ErrCode, string) {
	dbs, err := act.smgr.QueryShardingDBList()
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.dbs = dbs

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *ListAction) Do() error {
	// query shardingdb count.
	if errCode, errMsg := act.queryShardingDBCount(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query shardingdb list.
	if errCode, errMsg := act.queryShardingDBList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
