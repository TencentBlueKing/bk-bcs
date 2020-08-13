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

package cluster

import (
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// QueryAction is cluster query action object.
type QueryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryClusterReq
	resp *pb.QueryClusterResp

	sd *dbsharding.ShardingDB

	cluster database.Cluster
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryClusterReq, resp *pb.QueryClusterResp) *QueryAction {
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
	cluster := &pbcommon.Cluster{
		Bid:          act.cluster.Bid,
		Appid:        act.cluster.Appid,
		Clusterid:    act.cluster.Clusterid,
		Name:         act.cluster.Name,
		RClusterid:   act.cluster.RClusterid,
		Creator:      act.cluster.Creator,
		LastModifyBy: act.cluster.LastModifyBy,
		Memo:         act.cluster.Memo,
		State:        act.cluster.State,
		Labels:       act.cluster.Labels,
		CreatedAt:    act.cluster.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    act.cluster.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	act.resp.Cluster = cluster
	return nil
}

func (act *QueryAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	if len(act.req.Clusterid) == 0 && (len(act.req.Appid) == 0 || len(act.req.Name) == 0) {
		return errors.New("invalid params, clusterid or appid-name missing")
	}

	if len(act.req.Clusterid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, clusterid too long")
	}

	if len(act.req.Appid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	if len(act.req.Name) > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, name too long")
	}

	if len(act.req.Labels) > database.BSCPCLUSTERLABELSLENLIMIT {
		return errors.New("invalid params, labels too long")
	}

	// TODO check cluster labels format.
	return nil
}

func (act *QueryAction) queryCluster() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Cluster{})

	var err error

	if len(act.req.Clusterid) != 0 {
		err = act.sd.DB().
			Where(&database.Cluster{Bid: act.req.Bid, Clusterid: act.req.Clusterid}).
			Where("Fstate = ?", pbcommon.ClusterState_CS_CREATED).
			Last(&act.cluster).Error
	} else {
		err = act.sd.DB().
			Where(&database.Cluster{Bid: act.req.Bid, Appid: act.req.Appid, Name: act.req.Name}).
			Where("Flabels = ?", act.req.Labels).
			Where("Fstate = ?", pbcommon.ClusterState_CS_CREATED).
			Last(&act.cluster).Error
	}

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "cluster non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query cluster.
	if errCode, errMsg := act.queryCluster(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
