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

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// UpdateAction action to update variable
type UpdateAction struct {
	viper *viper.Viper

	smgr *dbsharding.ShardingManager
	sd   *dbsharding.ShardingDB

	req  *pb.UpdateVariableReq
	resp *pb.UpdateVariableResp
}

// NewUpdateAction create new UpdateAction
func NewUpdateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.UpdateVariableReq, resp *pb.UpdateVariableResp) *UpdateAction {
	action := &UpdateAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return error
func (act *UpdateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handle input message
func (act *UpdateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handle output message
func (act *UpdateAction) Output() error {
	// do nothing
	return nil
}

func (act *UpdateAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Vid, "vid"); err != nil {
		return err
	}

	if err := common.VerifyVarKey(act.req.Key); err != nil {
		return err
	}

	if err := common.VerifyVarValue(act.req.Value); err != nil {
		return err
	}

	if err := common.VerifyMemo(act.req.Memo); err != nil {
		return err
	}

	if err := common.VerifyNormalName(act.req.Operator, "operator"); err != nil {
		return err
	}

	return nil
}

func (act *UpdateAction) updateVariable() (pbcommon.ErrCode, string) {
	var exec *gorm.DB
	ups := map[string]interface{}{
		"Key":          act.req.Key,
		"Value":        act.req.Value,
		"Memo":         act.req.Memo,
		"LastModifyBy": act.req.Operator,
	}
	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_GLOBAL:
		act.sd.AutoMigrate(&database.VariableGlobal{})
		exec = act.sd.DB().
			Model(&database.VariableGlobal{}).
			Where(&database.VariableGlobal{
				Bid:   act.req.Bid,
				Vid:   act.req.Vid,
				State: int32(pbcommon.VariableState_VS_CREATED),
			}).
			Updates(ups)

	case pbcommon.VariableType_VT_CLUSTER:
		act.sd.AutoMigrate(&database.VariableCluster{})
		exec = act.sd.DB().
			Model(&database.VariableCluster{}).
			Where(&database.VariableCluster{
				Bid:   act.req.Bid,
				Vid:   act.req.Vid,
				State: int32(pbcommon.VariableState_VS_CREATED),
			}).
			Updates(ups)

	case pbcommon.VariableType_VT_ZONE:
		act.sd.AutoMigrate(&database.VariableZone{})
		exec = act.sd.DB().
			Model(&database.VariableZone{}).
			Where(&database.VariableZone{
				Bid:   act.req.Bid,
				Vid:   act.req.Vid,
				State: int32(pbcommon.VariableState_VS_CREATED),
			}).
			Updates(ups)
	}

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "update variable, failed, there is no variable fit in conditions"
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *UpdateAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// update variable
	if errCode, errMsg := act.updateVariable(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
