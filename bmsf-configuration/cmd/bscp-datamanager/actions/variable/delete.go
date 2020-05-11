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
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// DeleteAction action for deleting variable
type DeleteAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.DeleteVariableReq
	resp *pb.DeleteVariableResp

	sd *dbsharding.ShardingDB
}

// NewDeleteAction create new DeleteAction
func NewDeleteAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.DeleteVariableReq, resp *pb.DeleteVariableResp) *DeleteAction {
	action := &DeleteAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return error
func (act *DeleteAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages
func (act *DeleteAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *DeleteAction) Output() error {
	// do nothing.
	return nil
}

func (act *DeleteAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Vid, "vid"); err != nil {
		return err
	}

	if err := common.VerifyNormalName(act.req.Operator, "operator"); err != nil {
		return err
	}

	return nil
}

func (act *DeleteAction) deleteVariable() (pbcommon.ErrCode, string) {
	var err error
	ups := map[string]interface{}{
		"State":        pbcommon.VariableState_VS_DELETED,
		"LastModifyBy": act.req.Operator,
	}
	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_GLOBAL:
		act.sd.AutoMigrate(&database.VariableGlobal{})
		err = act.sd.DB().
			Model(&database.VariableGlobal{}).
			Where(&database.VariableGlobal{
				Bid: act.req.Bid,
				Vid: act.req.Vid,
			}).
			Updates(ups).Error
	case pbcommon.VariableType_VT_CLUSTER:
		act.sd.AutoMigrate(&database.VariableCluster{})
		err = act.sd.DB().
			Model(&database.VariableCluster{}).
			Where(&database.VariableCluster{
				Bid: act.req.Bid,
				Vid: act.req.Vid,
			}).
			Updates(ups).Error
	case pbcommon.VariableType_VT_ZONE:
		act.sd.AutoMigrate(&database.VariableZone{})
		err = act.sd.DB().
			Model(&database.VariableZone{}).
			Where(&database.VariableZone{
				Bid: act.req.Bid,
				Vid: act.req.Vid,
			}).
			Updates(ups).Error
	default:
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, "invalid variable type"
	}

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *DeleteAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// delete variable
	if errCode, errMsg := act.deleteVariable(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
