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
	"context"
	"encoding/json"
	"errors"
	"math"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
)

// MatchedAction is appinstance matched list action object.
type MatchedAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryMatchedAppInstancesReq
	resp *pb.QueryMatchedAppInstancesResp

	sd *dbsharding.ShardingDB

	strategy database.Strategy

	totalCount   int64
	appInstances []database.AppInstance
}

// NewMatchedAction creates new MatchedAction.
func NewMatchedAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryMatchedAppInstancesReq, resp *pb.QueryMatchedAppInstancesResp) *MatchedAction {
	action := &MatchedAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *MatchedAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
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
			Id:        st.ID,
			BizId:     st.BizID,
			AppId:     st.AppID,
			CloudId:   st.CloudID,
			Ip:        st.IP,
			Path:      st.Path,
			Labels:    st.Labels,
			State:     st.State,
			CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		instances = append(instances, ins)
	}
	act.resp.Data = &pb.QueryMatchedAppInstancesResp_RespData{TotalCount: uint32(act.totalCount), Info: instances}
	return nil
}

func (act *MatchedAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	// appid strategyid(release with strategy).
	// appid empty strategyid(release without strategy).
	if len(act.req.AppId) == 0 && len(act.req.StrategyId) == 0 {
		return errors.New("invalid input data, app_id or strategy_id is required")
	}

	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("strategy_id", act.req.StrategyId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

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

func (act *MatchedAction) queryStrategy() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Where(&database.Strategy{StrategyID: act.req.StrategyId}).
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

func (act *MatchedAction) queryReachableCount(appID string) (pbcommon.ErrCode, string) {
	if !act.req.Page.ReturnTotal {
		return pbcommon.ErrCode_E_OK, ""
	}

	err := act.sd.DB().
		Model(&database.AppInstance{}).
		Where(&database.AppInstance{
			BizID: act.req.BizId,
			AppID: appID,
			State: int32(pbcommon.AppInstanceState_INSS_ONLINE),
		}).
		Count(&act.totalCount).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *MatchedAction) queryReachableAppInstances(appID string) (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Offset(int(act.req.Page.Start)).Limit(int(act.req.Page.Limit)).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.AppInstance{
			BizID: act.req.BizId,
			AppID: appID,
			State: int32(pbcommon.AppInstanceState_INSS_ONLINE),
		}).
		Find(&act.appInstances).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *MatchedAction) queryReachableAppInstanceList(appID string) ([]database.AppInstance,
	pbcommon.ErrCode, string) {

	appInstances := []database.AppInstance{}

	index := 0
	for {
		instances := []database.AppInstance{}

		err := act.sd.DB().
			Offset(index).Limit(database.BSCPQUERYLIMITLB).
			Order("Fupdate_time DESC, Fid DESC").
			Where(&database.AppInstance{
				BizID: act.req.BizId,
				AppID: appID,
				State: int32(pbcommon.AppInstanceState_INSS_ONLINE),
			}).
			Find(&instances).Error

		if err != nil {
			return nil, pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
		appInstances = append(appInstances, instances...)

		if len(instances) < database.BSCPQUERYLIMITLB {
			break
		}
		index += len(instances)
	}

	return appInstances, pbcommon.ErrCode_E_OK, ""
}

func (act *MatchedAction) match(instances []database.AppInstance) (pbcommon.ErrCode, string) {
	strategies := strategy.Strategy{}
	if err := json.Unmarshal([]byte(act.strategy.Content), &strategies); err != nil {
		return pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, err.Error()
	}

	// matched app instance list.
	matchedAppInstances := []database.AppInstance{}

	strategyHandler := strategy.NewHandler(nil)

	for _, instance := range instances {
		if act.strategy.Content == strategy.EmptyStrategy {
			matchedAppInstances = append(matchedAppInstances, instance)
			continue
		}

		matcher := strategyHandler.Matcher()

		ins := &pbcommon.AppInstance{
			AppId:   instance.AppID,
			CloudId: instance.CloudID,
			Ip:      instance.IP,
			Path:    instance.Path,
			Labels:  instance.Labels,
		}

		if matcher(&strategies, ins) {
			matchedAppInstances = append(matchedAppInstances, instance)
		}
	}

	if act.req.Page.ReturnTotal {
		act.totalCount = int64(len(matchedAppInstances))
	}

	// rebuild final app instance list and split pages in memory.
	start := act.req.Page.Start
	end := start + act.req.Page.Limit

	if int(start) >= len(matchedAppInstances) {
		act.appInstances = []database.AppInstance{}
		return pbcommon.ErrCode_E_OK, ""
	}

	if int(end) >= len(matchedAppInstances) {
		act.appInstances = matchedAppInstances[start:]
		return pbcommon.ErrCode_E_OK, ""
	}

	act.appInstances = matchedAppInstances[start:end]
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *MatchedAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	if len(act.req.StrategyId) == 0 {
		// query appinstance reachable count by appid.
		if errCode, errMsg := act.queryReachableCount(act.req.AppId); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// query appinstance reachable list by appid.
		if errCode, errMsg := act.queryReachableAppInstances(act.req.AppId); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else {
		// query strategy content.
		if errCode, errMsg := act.queryStrategy(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// query appinstance reachable list by appid.
		instances, errCode, errMsg := act.queryReachableAppInstanceList(act.strategy.AppID)
		if errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// match strategy on reachable list.
		if errCode, errMsg := act.match(instances); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}

	return nil
}
