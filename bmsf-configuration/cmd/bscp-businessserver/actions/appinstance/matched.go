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

// MatchedAction query app instance list which matched target release or strategy.
type MatchedAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryMatchedAppInstancesReq
	resp *pb.QueryMatchedAppInstancesResp

	instances []*pbcommon.AppInstance
	appid     string
}

// NewMatchedAction creates new MatchedAction.
func NewMatchedAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryMatchedAppInstancesReq, resp *pb.QueryMatchedAppInstancesResp) *MatchedAction {
	action := &MatchedAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *MatchedAction) Output() error {
	act.resp.Instances = act.instances
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

	if len(act.req.Releaseid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, releaseid too long")
	}
	if len(act.req.Strategyid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, strategyid too long")
	}

	if len(act.req.Releaseid) == 0 && len(act.req.Strategyid) == 0 {
		return errors.New("invalid params, releaseid and strategyid missing")
	}
	if len(act.req.Releaseid) != 0 && len(act.req.Strategyid) != 0 {
		return errors.New("invalid params, only releaseid or strategyid")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too big")
	}
	return nil
}

func (act *MatchedAction) queryRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.req.Releaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryMatchedAppInstances[%d]| request to datamanager QueryRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	act.appid = resp.Release.Appid

	// maybe empty(release without strategy).
	act.req.Strategyid = resp.Release.Strategyid

	return resp.ErrCode, resp.ErrMsg
}

func (act *MatchedAction) matched() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMatchedAppInstancesReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Appid:      act.appid,
		Strategyid: act.req.Strategyid,
		Index:      act.req.Index,
		Limit:      act.req.Limit,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryMatchedAppInstances[%d]| request to datamanager QueryMatchedAppInstances, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryMatchedAppInstances(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryMatchedAppInstances, %+v", err)
	}
	act.instances = resp.Instances

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *MatchedAction) Do() error {
	if len(act.req.Strategyid) == 0 {
		if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}

	// query matched instances.
	if errCode, errMsg := act.matched(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
