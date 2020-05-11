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

// CreateAction action to create config template
type CreateAction struct {
	viper *viper.Viper

	smgr *dbsharding.ShardingManager
	sd   *dbsharding.ShardingDB

	req  *pb.CreateVariableReq
	resp *pb.CreateVariableResp
}

// NewCreateAction create new CreateAction
func NewCreateAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.CreateVariableReq, resp *pb.CreateVariableResp) *CreateAction {
	action := &CreateAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return error
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handle input message
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handle output message
func (act *CreateAction) Output() error {
	// do nothing
	return nil
}

func (act *CreateAction) verify() error {
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

	if err := common.VerifyNormalName(act.req.Creator, "creator"); err != nil {
		return err
	}

	return nil
}

func (act *CreateAction) reCreateVariable() (pbcommon.ErrCode, string) {

	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_GLOBAL:
		ups := map[string]interface{}{
			"Vid":          act.req.Vid,
			"Value":        act.req.Value,
			"Memo":         act.req.Memo,
			"State":        int32(pbcommon.VariableState_VS_CREATED),
			"Creator":      act.req.Creator,
			"LastModifyBy": act.req.Creator,
		}

		exec := act.sd.DB().
			Model(&database.VariableGlobal{}).
			Where(database.VariableGlobal{
				Bid:   act.req.Bid,
				Key:   act.req.Key,
				State: int32(pbcommon.VariableState_VS_DELETED),
			}).
			Updates(ups)
		if err := exec.Error; err != nil {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
		if exec.RowsAffected == 0 {
			return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "recreate global variable failed, no eligible global variable"
		}
	case pbcommon.VariableType_VT_CLUSTER:
		ups := map[string]interface{}{
			"Vid":           act.req.Vid,
			"Value":         act.req.Value,
			"Cluster":       act.req.Cluster,
			"ClusterLabels": act.req.ClusterLabels,
			"Memo":          act.req.Memo,
			"State":         int32(pbcommon.VariableState_VS_CREATED),
			"Creator":       act.req.Creator,
			"LastModifyBy":  act.req.Creator,
		}

		exec := act.sd.DB().
			Model(&database.VariableCluster{}).
			Where(database.VariableCluster{
				Bid:           act.req.Bid,
				Key:           act.req.Key,
				Cluster:       act.req.Cluster,
				ClusterLabels: act.req.ClusterLabels,
				State:         int32(pbcommon.VariableState_VS_DELETED),
			}).
			Updates(ups)
		if err := exec.Error; err != nil {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
		if exec.RowsAffected == 0 {
			return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "recreate cluster variable failed, no eligible cluster variable"
		}
	case pbcommon.VariableType_VT_ZONE:
		ups := map[string]interface{}{
			"Vid":           act.req.Vid,
			"Value":         act.req.Value,
			"Cluster":       act.req.Cluster,
			"ClusterLabels": act.req.ClusterLabels,
			"Zone":          act.req.Zone,
			"Memo":          act.req.Memo,
			"State":         int32(pbcommon.VariableState_VS_CREATED),
			"Creator":       act.req.Creator,
			"LastModifyBy":  act.req.Creator,
		}

		exec := act.sd.DB().
			Model(&database.VariableZone{}).
			Where(database.VariableZone{
				Bid:           act.req.Bid,
				Key:           act.req.Key,
				Cluster:       act.req.Cluster,
				ClusterLabels: act.req.ClusterLabels,
				Zone:          act.req.Zone,
				State:         int32(pbcommon.VariableState_VS_DELETED),
			}).
			Updates(ups)
		if err := exec.Error; err != nil {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
		if exec.RowsAffected == 0 {
			return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "recreate zone variable failed, no eligible zone variable"
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createVariable() (pbcommon.ErrCode, string) {
	var err error
	var vid string
	var state int32
	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_GLOBAL:
		act.sd.AutoMigrate(&database.VariableGlobal{})
		st := database.VariableGlobal{
			Vid:          act.req.Vid,
			Bid:          act.req.Bid,
			Key:          act.req.Key,
			Value:        act.req.Value,
			Memo:         act.req.Memo,
			State:        int32(pbcommon.VariableState_VS_CREATED),
			Creator:      act.req.Creator,
			LastModifyBy: act.req.Creator,
		}
		err = act.sd.DB().
			Where(database.VariableGlobal{
				Bid: st.Bid,
				Key: st.Key,
			}).
			Attrs(st).
			FirstOrCreate(&st).Error
		vid = st.Vid
		state = st.State

	case pbcommon.VariableType_VT_CLUSTER:
		act.sd.AutoMigrate(&database.VariableCluster{})
		st := database.VariableCluster{
			Vid:           act.req.Vid,
			Bid:           act.req.Bid,
			Cluster:       act.req.Cluster,
			ClusterLabels: act.req.ClusterLabels,
			Key:           act.req.Key,
			Value:         act.req.Value,
			Memo:          act.req.Memo,
			State:         int32(pbcommon.VariableState_VS_CREATED),
			Creator:       act.req.Creator,
			LastModifyBy:  act.req.Creator,
		}
		err = act.sd.DB().
			Where(database.VariableCluster{
				Bid:           st.Bid,
				Cluster:       st.Cluster,
				ClusterLabels: st.ClusterLabels,
				Key:           st.Key,
			}).
			Attrs(st).
			FirstOrCreate(&st).Error
		vid = st.Vid
		state = st.State

	case pbcommon.VariableType_VT_ZONE:
		act.sd.AutoMigrate(&database.VariableZone{})
		st := database.VariableZone{
			Vid:           act.req.Vid,
			Bid:           act.req.Bid,
			Cluster:       act.req.Cluster,
			ClusterLabels: act.req.ClusterLabels,
			Zone:          act.req.Zone,
			Key:           act.req.Key,
			Value:         act.req.Value,
			Memo:          act.req.Memo,
			State:         int32(pbcommon.VariableState_VS_CREATED),
			Creator:       act.req.Creator,
			LastModifyBy:  act.req.Creator,
		}
		err = act.sd.DB().
			Where(database.VariableZone{
				Bid:           st.Bid,
				Cluster:       st.Cluster,
				ClusterLabels: st.ClusterLabels,
				Zone:          st.Zone,
				Key:           st.Key,
			}).
			Attrs(st).
			FirstOrCreate(&st).Error
		vid = st.Vid
		state = st.State
	}

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.resp.Vid = vid

	if vid != act.req.Vid {
		if state == int32(pbcommon.VariableState_VS_CREATED) {
			return pbcommon.ErrCode_E_DM_ALREADY_EXISTS, "the variable with target key already exist."
		}
		if errCode, errMsg := act.reCreateVariable(); errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
	}

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *CreateAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// create variable
	if errCode, errMsg := act.createVariable(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
