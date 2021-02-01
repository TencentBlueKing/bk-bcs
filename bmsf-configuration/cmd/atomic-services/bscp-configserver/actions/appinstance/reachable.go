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
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// ReachableAction query reachable app instance list.
type ReachableAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryReachableAppInstancesReq
	resp *pb.QueryReachableAppInstancesResp

	labelsOr  []map[string]string
	labelsAnd []map[string]string
}

// NewReachableAction creates new ReachableAction.
func NewReachableAction(kit kit.Kit, viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryReachableAppInstancesReq, resp *pb.QueryReachableAppInstancesResp) *ReachableAction {
	action := &ReachableAction{kit: kit, viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Result = true
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	action.labelsOr = []map[string]string{}
	action.labelsAnd = []map[string]string{}

	return action
}

// Err setup error code message in response and return the error.
func (act *ReachableAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReachableAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}

	for _, labelsOr := range act.req.LabelsOr {
		if err := strategy.ValidateLabels(labelsOr.Labels); err != nil {
			return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, fmt.Sprintf("invalid labels_or formats, %+v", err))
		}
		if len(labelsOr.Labels) != 0 {
			act.labelsOr = append(act.labelsOr, labelsOr.Labels)
		}
	}

	for _, labelsAnd := range act.req.LabelsAnd {
		if err := strategy.ValidateLabels(labelsAnd.Labels); err != nil {
			return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, fmt.Sprintf("invalid labels_and formats, %+v", err))
		}
		if len(labelsAnd.Labels) != 0 {
			act.labelsAnd = append(act.labelsAnd, labelsAnd.Labels)
		}
	}

	return nil
}

// Output handles the output messages.
func (act *ReachableAction) Output() error {
	// do nothing.
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
		database.BSCPNOTEMPTY, database.BSCPQUERYLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *ReachableAction) reachable() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReachableAppInstancesReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		AppId:     act.req.AppId,
		LabelsOr:  act.req.LabelsOr,
		LabelsAnd: act.req.LabelsAnd,
		Page:      act.req.Page,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("QueryReachableAppInstances[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryReachableAppInstances(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to datamanager QueryReachableAppInstances, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.resp.Data = &pb.QueryReachableAppInstancesResp_RespData{TotalCount: resp.Data.TotalCount, Info: resp.Data.Info}

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *ReachableAction) Do() error {
	// query reachable app instances.
	if errCode, errMsg := act.reachable(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
