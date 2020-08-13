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

package appinstance

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// HistoryAction is appinstance history list action object.
type HistoryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryHistoryAppInstancesReq
	resp *pb.QueryHistoryAppInstancesResp

	sd *dbsharding.ShardingDB

	appInstances []database.AppInstance
}

// NewHistoryAction creates new HistoryAction.
func NewHistoryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryHistoryAppInstancesReq, resp *pb.QueryHistoryAppInstancesResp) *HistoryAction {
	action := &HistoryAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *HistoryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *HistoryAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *HistoryAction) Output() error {
	instances := []*pbcommon.AppInstance{}
	for _, st := range act.appInstances {
		ins := &pbcommon.AppInstance{
			Instanceid: st.ID,
			Bid:        st.Bid,
			Appid:      st.Appid,
			Clusterid:  st.Clusterid,
			Zoneid:     st.Zoneid,
			Dc:         st.Dc,
			Labels:     st.Labels,
			IP:         st.IP,
			State:      st.State,
			CreatedAt:  st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		instances = append(instances, ins)
	}
	act.resp.Instances = instances
	return nil
}

func (act *HistoryAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Appid)
	if length == 0 {
		return errors.New("invalid params, appid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too big")
	}

	if len(act.req.Clusterid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, clusterid too long")
	}

	if len(act.req.Zoneid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, zoneid too long")
	}
	return nil
}

func (act *HistoryAction) queryAppInstances() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.AppInstance{})

	// query type, 0:All(default)  1:Online  2:Offline
	whereState := fmt.Sprintf("Fstate IN (%d, %d)",
		pbcommon.AppInstanceState_INSS_ONLINE, pbcommon.AppInstanceState_INSS_OFFLINE)

	if act.req.QueryType == 1 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.AppInstanceState_INSS_ONLINE)
	} else if act.req.QueryType == 2 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.AppInstanceState_INSS_OFFLINE)
	}

	err := act.sd.DB().
		Offset(act.req.Index).Limit(act.req.Limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.AppInstance{Bid: act.req.Bid, Appid: act.req.Appid, Clusterid: act.req.Clusterid, Zoneid: act.req.Zoneid}).
		Where(whereState).
		Find(&act.appInstances).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *HistoryAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query appinstance list.
	if errCode, errMsg := act.queryAppInstances(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
