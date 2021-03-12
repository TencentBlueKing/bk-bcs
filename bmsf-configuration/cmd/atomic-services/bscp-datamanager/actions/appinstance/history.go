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

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
)

// HistoryAction is appinstance history list action object.
type HistoryAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryHistoryAppInstancesReq
	resp *pb.QueryHistoryAppInstancesResp

	sd *dbsharding.ShardingDB

	totalCount   int64
	appInstances []database.AppInstance

	labelsOr  []map[string]string
	labelsAnd []map[string]string
}

// NewHistoryAction creates new HistoryAction.
func NewHistoryAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryHistoryAppInstancesReq, resp *pb.QueryHistoryAppInstancesResp) *HistoryAction {
	action := &HistoryAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	action.labelsOr = []map[string]string{}
	action.labelsAnd = []map[string]string{}

	return action
}

// Err setup error code message in response and return the error.
func (act *HistoryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *HistoryAction) Input() error {
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
func (act *HistoryAction) Output() error {
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
	act.resp.Data = &pb.QueryHistoryAppInstancesResp_RespData{TotalCount: uint32(act.totalCount), Info: instances}
	return nil
}

func (act *HistoryAction) verify() error {
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
		database.BSCPNOTEMPTY, database.BSCPQUERYLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *HistoryAction) queryAppInstances() ([]database.AppInstance, pbcommon.ErrCode, string) {
	// query type, 0:All(default)  1:Online  2:Offline
	whereState := fmt.Sprintf("Fstate IN (%d, %d)",
		pbcommon.AppInstanceState_INSS_ONLINE, pbcommon.AppInstanceState_INSS_OFFLINE)

	if act.req.QueryType == 1 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.AppInstanceState_INSS_ONLINE)
	} else if act.req.QueryType == 2 {
		whereState = fmt.Sprintf("Fstate = %d", pbcommon.AppInstanceState_INSS_OFFLINE)
	}

	appInstances := []database.AppInstance{}

	index := 0
	for {
		instances := []database.AppInstance{}

		err := act.sd.DB().
			Offset(index).Limit(database.BSCPQUERYLIMITLB).
			Order("Fupdate_time DESC, Fid DESC").
			Where(&database.AppInstance{BizID: act.req.BizId, AppID: act.req.AppId}).
			Where(whereState).
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

func (act *HistoryAction) matchAppInstances(instances []database.AppInstance) (pbcommon.ErrCode, string) {
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
func (act *HistoryAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query appinstance list.
	instances, errCode, errMsg := act.queryAppInstances()
	if errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// match appinstance list.
	if errCode, errMsg := act.matchAppInstances(instances); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
