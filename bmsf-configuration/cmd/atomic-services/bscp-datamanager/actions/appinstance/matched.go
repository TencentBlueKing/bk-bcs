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
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/internal/types"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
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

	procAttrs          []database.ProcAttr
	reachableInstances []database.AppInstance

	totalCount   int64
	appInstances []database.AppInstance

	offlineExpirationSec int64
}

// NewMatchedAction creates new MatchedAction.
func NewMatchedAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryMatchedAppInstancesReq, resp *pb.QueryMatchedAppInstancesResp) *MatchedAction {

	action := &MatchedAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	// max app instance offline timeout.
	action.offlineExpirationSec = int64(types.AppInstanceOfflineMaxTimeout / time.Second)

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

	// app_id strategy_id(release with strategy).
	// app_id empty strategy_id(release without strategy).
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
		database.BSCPNOTEMPTY, database.BSCPQUERYLIMITMB); err != nil {
		return err
	}
	return nil
}

func (act *MatchedAction) queryStrategy() (pbcommon.ErrCode, string) {
	if len(act.req.StrategyId) == 0 {
		// empty strategy to match.
		return pbcommon.ErrCode_E_OK, ""
	}

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

	// save the app_id from strategy.
	act.req.AppId = act.strategy.AppID

	return pbcommon.ErrCode_E_OK, ""
}

func (act *MatchedAction) queryProcAttrList() (pbcommon.ErrCode, string) {
	if len(act.req.AppId) == 0 {
		return pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, "can't match procattrs with empty app_id"
	}

	// use bscpdb to list procattrs.
	sd, err := act.smgr.ShardingDB(dbsharding.BSCPDBKEY)
	if err != nil {
		return pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error()
	}

	index := 0
	limit := database.BSCPQUERYLIMITMB

	startTime := time.Now()

	for {
		procAttrs := []database.ProcAttr{}

		err := sd.DB().
			Offset(index).Limit(limit).
			Order("Fupdate_time DESC, Fid DESC").
			Where(&database.ProcAttr{BizID: act.req.BizId, AppID: act.req.AppId}).
			Find(&procAttrs).Error

		if err != nil {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
		act.procAttrs = append(act.procAttrs, procAttrs...)

		if len(procAttrs) < limit {
			break
		}
		index += len(procAttrs)
	}

	logger.V(2).Infof("QueryMatchedAppInstances[%s]| query procattr list done, count[%d], cost: %+v",
		act.req.Seq, len(act.procAttrs), time.Since(startTime))

	return pbcommon.ErrCode_E_OK, ""
}

func (act *MatchedAction) queryReachableAppInstanceList() (pbcommon.ErrCode, string) {
	offlineInstances := []database.AppInstance{}

	index := 0
	limit := database.BSCPQUERYLIMITMB

	startTime := time.Now()

	for {
		instances := []database.AppInstance{}

		err := act.sd.DB().
			Offset(index).Limit(limit).
			Order("Fcreate_time DESC, Fid DESC").
			Where(&database.AppInstance{
				BizID: act.req.BizId,
				AppID: act.req.AppId,
				State: int32(pbcommon.AppInstanceState_INSS_ONLINE),
			}).
			Find(&instances).Error

		if err != nil {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}

		// handle long time no update offline instance.
		for _, inst := range instances {
			// TODO: add flush session time.
			updateInterval := time.Now().Unix() - inst.UpdatedAt.Unix()

			if updateInterval > act.offlineExpirationSec {
				offlineInstances = append(offlineInstances, inst)
			} else {
				act.reachableInstances = append(act.reachableInstances, inst)
			}
		}

		// still use instances from database to check index.
		if len(instances) < limit {
			break
		}
		index += len(instances)
	}
	logger.V(2).Infof("QueryMatchedAppInstances[%s]| query reachable list done, timeout instances count[%d], cost: %+v",
		act.req.Seq, len(offlineInstances), time.Since(startTime))

	// handle long time no update offline instance.
	for _, instance := range offlineInstances {
		// update offline state.
		ups := map[string]interface{}{"State": int32(pbcommon.AppInstanceState_INSS_OFFLINE)}

		err := act.sd.DB().
			Model(&database.AppInstance{}).
			Where(&database.AppInstance{ID: instance.ID}).
			Updates(ups).Error

		if err != nil {
			logger.Warnf("QueryMatchedAppInstances[%s]| update long time instance offline state failed, %+v, %+v",
				act.req.Seq, instance, err)
		}
	}
	logger.V(2).Infof("QueryMatchedAppInstances[%s]| query reachable list done, instances count[%d], cost: %+v",
		act.req.Seq, len(act.reachableInstances), time.Since(startTime))

	return pbcommon.ErrCode_E_OK, ""
}

func (act *MatchedAction) instanceKey(cloudID, ip, path string) string {
	metadata := cloudID + ":" + ip + ":" + filepath.Clean(path)

	// gen sha1 as an instance key.
	key := common.SHA1(metadata)
	if len(key) != 0 {
		return key
	}
	return metadata
}

func (act *MatchedAction) match() (pbcommon.ErrCode, string) {
	// unmarshal strategy content.
	strategies := strategy.Strategy{}

	if len(act.req.StrategyId) != 0 {
		if err := json.Unmarshal([]byte(act.strategy.Content), &strategies); err != nil {
			return pbcommon.ErrCode_E_DM_SYSTEM_UNKNOWN, err.Error()
		}
	}

	// all app instances.
	instances := []database.AppInstance{}

	// reachable app instances.
	reachableInstancesMap := make(map[string]*database.AppInstance, 0)

	for _, instance := range act.reachableInstances {
		key := act.instanceKey(instance.CloudID, instance.IP, instance.Path)
		reachableInstancesMap[key] = &instance
		instances = append(instances, instance)
	}

	// append procattr list for process app.
	for _, procAttr := range act.procAttrs {
		key := act.instanceKey(procAttr.CloudID, procAttr.IP, procAttr.Path)

		if _, isExist := reachableInstancesMap[key]; !isExist {
			sidecarLabels := &strategy.SidecarLabels{}

			if err := json.Unmarshal([]byte(procAttr.Labels), &sidecarLabels.Labels); err != nil {
				logger.Warnf("QueryMatchedAppInstances[%s]| unmarshal procattr lables failed, %+v, %+v",
					act.req.Seq, procAttr.Labels, err)
				continue
			}

			labels, err := json.Marshal(sidecarLabels)
			if err != nil {
				logger.Warnf("QueryMatchedAppInstances[%s]| marshal procattr lables failed, %+v, %+v",
					act.req.Seq, sidecarLabels, err)
				continue
			}

			instances = append(instances, database.AppInstance{
				BizID:     procAttr.BizID,
				AppID:     procAttr.AppID,
				CloudID:   procAttr.CloudID,
				IP:        procAttr.IP,
				Path:      procAttr.Path,
				Labels:    string(labels),
				State:     int32(pbcommon.AppInstanceState_INSS_OFFLINE),
				CreatedAt: procAttr.CreatedAt,
				UpdatedAt: procAttr.UpdatedAt,
			})
		}
	}

	// now the instances is all app instance list include process app procattr.

	// matched app instance list.
	matchedAppInstances := []database.AppInstance{}

	strategyHandler := strategy.NewHandler(nil)

	startTime := time.Now()

	for _, instance := range instances {
		if len(act.req.StrategyId) == 0 || act.strategy.Content == strategy.EmptyStrategy {
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
	logger.V(2).Infof("QueryMatchedAppInstances[%s]| matched done, instances count[%d], matched count[%d], cost: %+v",
		act.req.Seq, len(instances), len(matchedAppInstances), time.Since(startTime))

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

	// query strategy content.
	if errCode, errMsg := act.queryStrategy(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query app procattrs(it's empty for container app) base on app_id from request or strategy.
	if errCode, errMsg := act.queryProcAttrList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query appinstance reachable list by app_id.
	if errCode, errMsg := act.queryReachableAppInstanceList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// match strategy on reachable list and procattr list.
	if errCode, errMsg := act.match(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	return nil
}
