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

// QueryAction action for query config template version
type QueryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager
	sd    *dbsharding.ShardingDB

	req  *pb.QueryVariableReq
	resp *pb.QueryVariableResp

	variableGlobal  database.VariableGlobal
	variableCluster database.VariableCluster
	variableZone    database.VariableZone
}

// NewQueryAction create new QueryAction
func NewQueryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryVariableReq, resp *pb.QueryVariableResp) *QueryAction {
	action := &QueryAction{viper: viper, smgr: smgr, req: req, resp: resp}

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
	var variable *pbcommon.Variable
	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_GLOBAL:
		variable = &pbcommon.Variable{
			Bid:           act.variableGlobal.Bid,
			Vid:           act.variableGlobal.Vid,
			Cluster:       "",
			ClusterLabels: "",
			Zone:          "",
			Type:          int32(pbcommon.VariableType_VT_GLOBAL),
			Key:           act.variableGlobal.Key,
			Value:         act.variableGlobal.Value,
			Memo:          act.variableGlobal.Memo,
			Creator:       act.variableGlobal.Creator,
			LastModifyBy:  act.variableGlobal.LastModifyBy,
			State:         act.variableGlobal.State,
			CreatedAt:     act.variableGlobal.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     act.variableGlobal.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	case pbcommon.VariableType_VT_CLUSTER:
		variable = &pbcommon.Variable{
			Bid:           act.variableCluster.Bid,
			Vid:           act.variableCluster.Vid,
			Cluster:       act.variableCluster.Cluster,
			ClusterLabels: act.variableCluster.ClusterLabels,
			Zone:          "",
			Type:          int32(pbcommon.VariableType_VT_CLUSTER),
			Key:           act.variableCluster.Key,
			Value:         act.variableCluster.Value,
			Memo:          act.variableCluster.Memo,
			Creator:       act.variableCluster.Creator,
			LastModifyBy:  act.variableCluster.LastModifyBy,
			State:         act.variableCluster.State,
			CreatedAt:     act.variableCluster.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     act.variableCluster.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	case pbcommon.VariableType_VT_ZONE:
		variable = &pbcommon.Variable{
			Bid:           act.variableZone.Bid,
			Vid:           act.variableZone.Vid,
			Cluster:       act.variableZone.Cluster,
			ClusterLabels: act.variableZone.ClusterLabels,
			Zone:          act.variableZone.Zone,
			Type:          int32(pbcommon.VariableType_VT_ZONE),
			Key:           act.variableZone.Key,
			Value:         act.variableZone.Value,
			Memo:          act.variableZone.Memo,
			Creator:       act.variableZone.Creator,
			LastModifyBy:  act.variableZone.LastModifyBy,
			State:         act.variableZone.State,
			CreatedAt:     act.variableZone.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     act.variableZone.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	act.resp.Var = variable
	return nil
}

func (act *QueryAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Vid, "vid"); err != nil {
		return err
	}

	if err := common.VerifyVariableType(act.req.Type); err != nil {
		return err
	}

	return nil
}

func (act *QueryAction) queryVariable() (pbcommon.ErrCode, string) {

	var err error
	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_GLOBAL:
		act.sd.AutoMigrate(&database.VariableGlobal{})
		err = act.sd.DB().
			Where(map[string]interface{}{
				"Fbid":   act.req.Bid,
				"Fvid":   act.req.Vid,
				"Fstate": int32(pbcommon.VariableState_VS_CREATED),
			}).
			Last(&act.variableGlobal).Error
	case pbcommon.VariableType_VT_CLUSTER:
		act.sd.AutoMigrate(&database.VariableCluster{})
		err = act.sd.DB().
			Where(map[string]interface{}{
				"Fbid":   act.req.Bid,
				"Fvid":   act.req.Vid,
				"Fstate": int32(pbcommon.VariableState_VS_CREATED),
			}).
			Last(&act.variableCluster).Error
	case pbcommon.VariableType_VT_ZONE:
		act.sd.AutoMigrate(&database.VariableZone{})
		err = act.sd.DB().
			Where(map[string]interface{}{
				"Fbid":   act.req.Bid,
				"Fvid":   act.req.Vid,
				"Fstate": int32(pbcommon.VariableState_VS_CREATED),
			}).
			Last(&act.variableZone).Error
	}

	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "variable no found"
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *QueryAction) Do() error {
	// business sharding db
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query config variable.
	if errCode, errMsg := act.queryVariable(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
