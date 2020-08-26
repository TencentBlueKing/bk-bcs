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

// ListAction action for query config template version
type ListAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager
	sd    *dbsharding.ShardingDB

	req  *pb.QueryVariableListReq
	resp *pb.QueryVariableListResp

	variablesGlobal  []database.VariableGlobal
	variablesCluster []database.VariableCluster
	variablesZone    []database.VariableZone
}

// NewListAction create new ListAction
func NewListAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryVariableListReq, resp *pb.QueryVariableListResp) *ListAction {
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
	var variables []*pbcommon.Variable
	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_GLOBAL:
		for _, variableGlobal := range act.variablesGlobal {
			variable := &pbcommon.Variable{
				Bid:           variableGlobal.Bid,
				Vid:           variableGlobal.Vid,
				Cluster:       "",
				ClusterLabels: "",
				Zone:          "",
				Type:          int32(pbcommon.VariableType_VT_GLOBAL),
				Key:           variableGlobal.Key,
				Value:         variableGlobal.Value,
				Memo:          variableGlobal.Memo,
				Creator:       variableGlobal.Creator,
				LastModifyBy:  variableGlobal.LastModifyBy,
				State:         variableGlobal.State,
				CreatedAt:     variableGlobal.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:     variableGlobal.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
			variables = append(variables, variable)
		}

	case pbcommon.VariableType_VT_CLUSTER:
		for _, variableCluster := range act.variablesCluster {
			variable := &pbcommon.Variable{
				Bid:           variableCluster.Bid,
				Vid:           variableCluster.Vid,
				Cluster:       variableCluster.Cluster,
				ClusterLabels: variableCluster.ClusterLabels,
				Zone:          "",
				Type:          int32(pbcommon.VariableType_VT_CLUSTER),
				Key:           variableCluster.Key,
				Value:         variableCluster.Value,
				Memo:          variableCluster.Memo,
				Creator:       variableCluster.Creator,
				LastModifyBy:  variableCluster.LastModifyBy,
				State:         variableCluster.State,
				CreatedAt:     variableCluster.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:     variableCluster.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
			variables = append(variables, variable)
		}
	case pbcommon.VariableType_VT_ZONE:
		for _, variableZone := range act.variablesZone {
			variable := &pbcommon.Variable{
				Bid:           variableZone.Bid,
				Vid:           variableZone.Vid,
				Cluster:       variableZone.Cluster,
				ClusterLabels: variableZone.ClusterLabels,
				Zone:          variableZone.Zone,
				Type:          int32(pbcommon.VariableType_VT_ZONE),
				Key:           variableZone.Key,
				Value:         variableZone.Value,
				Memo:          variableZone.Memo,
				Creator:       variableZone.Creator,
				LastModifyBy:  variableZone.LastModifyBy,
				State:         variableZone.State,
				CreatedAt:     variableZone.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:     variableZone.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
			variables = append(variables, variable)
		}
	}
	act.resp.Vars = variables
	return nil
}

func (act *ListAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_CLUSTER:
		if err := common.VerifyNormalName(act.req.Cluster, "cluster"); err != nil {
			return err
		}
		if err := common.VerifyClusterLabels(act.req.ClusterLabels); err != nil {
			return err
		}
	case pbcommon.VariableType_VT_ZONE:
		if err := common.VerifyNormalName(act.req.Cluster, "cluster"); err != nil {
			return err
		}
		if err := common.VerifyClusterLabels(act.req.ClusterLabels); err != nil {
			return err
		}
		if err := common.VerifyNormalName(act.req.Zone, "zone"); err != nil {
			return err
		}
	}

	return nil
}

func (act *ListAction) queryVariables() (pbcommon.ErrCode, string) {

	var err error
	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_GLOBAL:
		act.sd.AutoMigrate(&database.VariableGlobal{})
		err = act.sd.DB().
			Offset(act.req.Index).Limit(act.req.Limit).
			Order("Fupdate_time DESC, Fid DESC").
			Where(map[string]interface{}{
				"Fbid":   act.req.Bid,
				"Fstate": int32(pbcommon.VariableState_VS_CREATED),
			}).
			Find(&act.variablesGlobal).Error

	case pbcommon.VariableType_VT_CLUSTER:
		act.sd.AutoMigrate(&database.VariableCluster{})
		err = act.sd.DB().
			Offset(act.req.Index).Limit(act.req.Limit).
			Order("Fupdate_time DESC, Fid DESC").
			Where(map[string]interface{}{
				"Fbid":            act.req.Bid,
				"Fcluster":        act.req.Cluster,
				"Fcluster_labels": act.req.ClusterLabels,
				"Fstate":          int32(pbcommon.VariableState_VS_CREATED),
			}).
			Find(&act.variablesCluster).Error

	case pbcommon.VariableType_VT_ZONE:
		act.sd.AutoMigrate(&database.VariableZone{})
		err = act.sd.DB().
			Offset(act.req.Index).Limit(act.req.Limit).
			Order("Fupdate_time DESC, Fid DESC").
			Where(map[string]interface{}{
				"Fbid":            act.req.Bid,
				"Fcluster":        act.req.Cluster,
				"Fcluster_labels": act.req.ClusterLabels,
				"Fzone":           act.req.Zone,
				"Fstate":          int32(pbcommon.VariableState_VS_CREATED),
			}).
			Find(&act.variablesZone).Error
	}

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *ListAction) Do() error {
	// business sharding db
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query config variables.
	if errCode, errMsg := act.queryVariables(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
