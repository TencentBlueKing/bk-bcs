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
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// ReachableAction is appinstance reachable list action object.
type ReachableAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryReachableAppInstancesReq
	resp *pb.QueryReachableAppInstancesResp

	sd *dbsharding.ShardingDB

	totalCount   int64
	appInstances []database.AppInstance

	labelsOr  []map[string]string
	labelsAnd []map[string]string

	offlineExpirationSec int64
}

// NewReachableAction creates new ReachableAction.
func NewReachableAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryReachableAppInstancesReq, resp *pb.QueryReachableAppInstancesResp) *ReachableAction {

	action := &ReachableAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	action.labelsOr = []map[string]string{}
	action.labelsAnd = []map[string]string{}

	action.offlineExpirationSec = int64(20 * time.Minute / time.Second)

	return action
}

// Err setup error code message in response and return the error.
func (act *ReachableAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReachableAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}

	for _, labelsOr := range act.req.LabelsOr {
		if err := strategy.ValidateLabels(labelsOr.Labels); err != nil {
			return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, fmt.Sprintf("invalid labels_or formats, %+v", err))
		}
		if len(labelsOr.Labels) != 0 {
			act.labelsOr = append(act.labelsOr, labelsOr.Labels)
		}
	}

	for _, labelsAnd := range act.req.LabelsAnd {
		if err := strategy.ValidateLabels(labelsAnd.Labels); err != nil {
			return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, fmt.Sprintf("invalid labels_and formats, %+v", err))
		}
		if len(labelsAnd.Labels) != 0 {
			act.labelsAnd = append(act.labelsAnd, labelsAnd.Labels)
		}
	}

	return nil
}

// Output handles the output messages.
func (act *ReachableAction) Output() error {
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
	act.resp.Data = &pb.QueryReachableAppInstancesResp_RespData{TotalCount: uint32(act.totalCount), Info: instances}
	return nil
}

func (act *ReachableAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
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

func (act *ReachableAction) queryAppInstances() ([]database.AppInstance, pbcommon.ErrCode, string) {
	appInstances := []database.AppInstance{}
	offlineInstances := []database.AppInstance{}

	index := 0
	limit := database.BSCPQUERYLIMITMB

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
			return nil, pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}

		// handle long time no update offline instance.
		for _, inst := range instances {
			updateInterval := time.Now().Unix() - inst.UpdatedAt.Unix()

			if updateInterval > act.offlineExpirationSec {
				offlineInstances = append(offlineInstances, inst)
			} else {
				appInstances = append(appInstances, inst)
			}
		}

		// still use instances from database to check index.
		if len(instances) < limit {
			break
		}
		index += len(instances)
	}

	// handle long time no update offline instance.
	for _, instance := range offlineInstances {
		// update offline state.
		ups := map[string]interface{}{"State": int32(pbcommon.AppInstanceState_INSS_OFFLINE)}

		err := act.sd.DB().
			Model(&database.AppInstance{}).
			Where(&database.AppInstance{ID: instance.ID}).
			Updates(ups).Error

		if err != nil {
			logger.Warnf("QueryReachableAppInstances[%s]| update long time instance offline state failed, %+v, %+v",
				act.req.Seq, instance, err)
		}
	}

	return appInstances, pbcommon.ErrCode_E_OK, ""
}

func (act *ReachableAction) match(instances []database.AppInstance) (pbcommon.ErrCode, string) {
	strategies := &strategy.Strategy{LabelsOr: act.labelsOr, LabelsAnd: act.labelsAnd}
	matcher := strategy.NewHandler(nil).Matcher()

	// matched app instance list.
	matchedAppInstances := []database.AppInstance{}

	for _, instance := range instances {
		ins := &pbcommon.AppInstance{
			AppId:   instance.AppID,
			CloudId: instance.CloudID,
			Ip:      instance.IP,
			Path:    instance.Path,
			Labels:  instance.Labels,
		}

		if matcher(strategies, ins) {
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
func (act *ReachableAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query appinstance reachable list.
	instances, errCode, errMsg := act.queryAppInstances()
	if errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// match reachable list.
	if errCode, errMsg := act.match(instances); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
