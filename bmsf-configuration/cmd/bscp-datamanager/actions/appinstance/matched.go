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
	"encoding/json"
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
)

// MatchedAction is appinstance matched list action object.
type MatchedAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryMatchedAppInstancesReq
	resp *pb.QueryMatchedAppInstancesResp

	sd *dbsharding.ShardingDB

	strategy     database.Strategy
	appInstances []database.AppInstance
}

// NewMatchedAction creates new MatchedAction.
func NewMatchedAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryMatchedAppInstancesReq, resp *pb.QueryMatchedAppInstancesResp) *MatchedAction {
	action := &MatchedAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *MatchedAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *MatchedAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *MatchedAction) Output() error {
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

func (act *MatchedAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	if len(act.req.Appid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	if len(act.req.Strategyid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, strategyid too long")
	}

	// appid strategyid(release with strategy).
	// appid empty strategyid(release without strategy).
	// empty appid strategyid(target strategy).
	if len(act.req.Appid) == 0 && len(act.req.Strategyid) == 0 {
		return errors.New("invalid params, appid and strategyid missing")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too big")
	}
	return nil
}

func (act *MatchedAction) queryStrategy() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Strategy{})

	err := act.sd.DB().
		Where(&database.Strategy{Strategyid: act.req.Strategyid}).
		Where("Fstate = ?", pbcommon.StrategyState_SS_CREATED).
		Last(&act.strategy).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "strategt non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *MatchedAction) queryReachableAppInstances(appid string) (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.AppInstance{})

	err := act.sd.DB().
		Offset(act.req.Index).Limit(act.req.Limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.AppInstance{Bid: act.req.Bid, Appid: appid, State: int32(pbcommon.AppInstanceState_INSS_ONLINE)}).
		Find(&act.appInstances).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *MatchedAction) match() (pbcommon.ErrCode, string) {
	if act.strategy.Content == strategy.EmptyStrategy {
		return pbcommon.ErrCode_E_OK, ""
	}

	strategies := strategy.Strategy{}
	if err := json.Unmarshal([]byte(act.strategy.Content), &strategies); err != nil {
		return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, err.Error()
	}

	// matched app instance list.
	matchedAppInstances := []database.AppInstance{}

	strategyHandler := strategy.NewHandler(nil)

	for _, instance := range act.appInstances {
		matcher := strategyHandler.Matcher()

		ins := &pbcommon.AppInstance{
			Appid:     instance.Appid,
			Clusterid: instance.Clusterid,
			Zoneid:    instance.Zoneid,
			Dc:        instance.Dc,
			IP:        instance.IP,
			Labels:    instance.Labels,
		}

		if matcher(&strategies, ins) {
			matchedAppInstances = append(matchedAppInstances, instance)
		}
	}

	// rebuild final app instance list.
	act.appInstances = matchedAppInstances

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *MatchedAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	if len(act.req.Strategyid) == 0 {
		// query appinstance reachable list by appid.
		if errCode, errMsg := act.queryReachableAppInstances(act.req.Appid); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else {
		// query strategy content.
		if errCode, errMsg := act.queryStrategy(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// query appinstance reachable list by appid.
		if errCode, errMsg := act.queryReachableAppInstances(act.strategy.Appid); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// match strategy on reachable list.
		if errCode, errMsg := act.match(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}

	return nil
}
