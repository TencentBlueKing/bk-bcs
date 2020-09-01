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

// ListAction is cluster list action object.
type ListAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryClusterListReq
	resp *pb.QueryClusterListResp

	sd *dbsharding.ShardingDB

	clusters []database.Cluster
}

// NewListAction creates new ListAction.
func NewListAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryClusterListReq, resp *pb.QueryClusterListResp) *ListAction {
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
	clusters := []*pbcommon.Cluster{}
	for _, st := range act.clusters {
		cluster := &pbcommon.Cluster{
			Bid:          st.Bid,
			Appid:        st.Appid,
			Clusterid:    st.Clusterid,
			Name:         st.Name,
			Labels:       st.Labels,
			RClusterid:   st.RClusterid,
			Creator:      st.Creator,
			LastModifyBy: st.LastModifyBy,
			Memo:         st.Memo,
			State:        st.State,
			CreatedAt:    st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		clusters = append(clusters, cluster)
	}
	act.resp.Clusters = clusters
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

	if len(act.req.Appid) == 0 && len(act.req.AppName) == 0 {
		return errors.New("invalid params, appid or appname missing")
	}

	if len(act.req.Appid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	if len(act.req.AppName) > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, appname too long")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too big")
	}
	return nil
}

func (act *ListAction) queryApp() (pbcommon.ErrCode, string) {
	if len(act.req.Appid) != 0 {
		return pbcommon.ErrCode_E_OK, ""
	}
	act.sd.AutoMigrate(&database.App{})

	var st database.App
	err := act.sd.DB().
		Where(&database.App{Bid: act.req.Bid, Name: act.req.AppName}).
		Where("Fstate = ?", pbcommon.AppState_AS_CREATED).
		Last(&st).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "app with target name non-exist, can't query cluster list under it."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	act.req.Appid = st.Appid

	return pbcommon.ErrCode_E_OK, ""
}

func (act *ListAction) queryClusterList() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Cluster{})

	err := act.sd.DB().
		Offset(act.req.Index).Limit(act.req.Limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.Cluster{Bid: act.req.Bid, Appid: act.req.Appid}).
		Where("Fstate = ?", pbcommon.ClusterState_CS_CREATED).
		Find(&act.clusters).Error

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

	// query app.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query cluster list.
	if errCode, errMsg := act.queryClusterList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
