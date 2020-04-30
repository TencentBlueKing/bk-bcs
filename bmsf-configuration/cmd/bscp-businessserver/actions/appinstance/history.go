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

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/logger"
)

// HistoryAction query history app instance list.
type HistoryAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryHistoryAppInstancesReq
	resp *pb.QueryHistoryAppInstancesResp

	instances []*pbcommon.AppInstance
}

// NewHistoryAction creates new HistoryAction.
func NewHistoryAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryHistoryAppInstancesReq, resp *pb.QueryHistoryAppInstancesResp) *HistoryAction {
	action := &HistoryAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *HistoryAction) Output() error {
	act.resp.Instances = act.instances
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

	if len(act.req.Clusterid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, clusterid too long")
	}

	if len(act.req.Zoneid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, zoneid too long")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too big")
	}
	return nil
}

func (act *HistoryAction) history() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryHistoryAppInstancesReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Appid:     act.req.Appid,
		Clusterid: act.req.Clusterid,
		Zoneid:    act.req.Zoneid,
		QueryType: act.req.QueryType,
		Index:     act.req.Index,
		Limit:     act.req.Limit,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryHistoryAppInstances[%d]| request to datamanager QueryHistoryAppInstances, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryHistoryAppInstances(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryHistoryAppInstances, %+v", err)
	}
	act.instances = resp.Instances

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *HistoryAction) Do() error {
	// query history app instances.
	if errCode, errMsg := act.history(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
