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

package strategy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-businessserver/modules/audit"
	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// CreateAction creates a strategy object.
type CreateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateStrategyReq
	resp *pb.CreateStrategyResp

	strategies    *strategy.Strategy
	newStrategyid string
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateStrategyReq, resp *pb.CreateStrategyResp) *CreateAction {
	action := &CreateAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}

	if err := strategy.ValidateLabels(act.req.Labels); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, fmt.Sprintf("invalid labels formats, %+v", err))
	}

	if err := strategy.ValidateLabels(act.req.LabelsAnd); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, fmt.Sprintf("invalid labelsAnd formats, %+v", err))
	}

	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
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

	length = len(act.req.Name)
	if length == 0 {
		return errors.New("invalid params, name missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, name too long")
	}

	if act.req.Clusterids == nil {
		act.req.Clusterids = []string{}
	}
	if len(act.req.Clusterids) > database.BSCPBATCHLIMIT {
		return errors.New("invalid params, clusterids list too long")
	}

	if act.req.Zoneids == nil {
		act.req.Zoneids = []string{}
	}
	if len(act.req.Zoneids) > database.BSCPBATCHLIMIT {
		return errors.New("invalid params, zoneids list too long")
	}

	if act.req.Dcs == nil {
		act.req.Dcs = []string{}
	}
	if len(act.req.Dcs) > database.BSCPBATCHLIMIT {
		return errors.New("invalid params, dcs list too long")
	}

	if act.req.IPs == nil {
		act.req.IPs = []string{}
	}
	if len(act.req.IPs) > database.BSCPBATCHLIMIT {
		return errors.New("invalid params, ips list too long")
	}

	if act.req.Labels == nil {
		act.req.Labels = make(map[string]string)
	}
	if len(act.req.Labels) > database.BSCPBATCHLIMIT {
		return errors.New("invalid params, labels set too large")
	}

	if act.req.LabelsAnd == nil {
		act.req.LabelsAnd = make(map[string]string)
	}
	if len(act.req.LabelsAnd) > database.BSCPBATCHLIMIT {
		return errors.New("invalid params, labelsAnd set too large")
	}

	length = len(act.req.Creator)
	if length == 0 {
		return errors.New("invalid params, creator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, creator too long")
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *CreateAction) genStrategyID() error {
	id, err := common.GenStrategyid()
	if err != nil {
		return err
	}
	act.newStrategyid = id
	return nil
}

func (act *CreateAction) queryCluster(clusterid string) (*pbcommon.Cluster, error) {
	r := &pbdatamanager.QueryClusterReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Clusterid: clusterid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateStrategy[%d]| request to datamanager QueryCluster, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryCluster(ctx, r)
	if err != nil {
		return nil, err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return nil, errors.New(resp.ErrMsg)
	}
	return resp.Cluster, nil
}

func (act *CreateAction) queryZone(zoneid string) (*pbcommon.Zone, error) {
	r := &pbdatamanager.QueryZoneReq{
		Seq:    act.req.Seq,
		Bid:    act.req.Bid,
		Zoneid: zoneid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateStrategy[%d]| request to datamanager QueryZone, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryZone(ctx, r)
	if err != nil {
		return nil, err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return nil, errors.New(resp.ErrMsg)
	}
	return resp.Zone, nil
}

func (act *CreateAction) query() (pbcommon.ErrCode, string) {
	for _, clusterid := range act.req.Clusterids {
		cluster, err := act.queryCluster(clusterid)
		if err != nil {
			return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("can't query cluster[%+v] information to create strategy", clusterid)
		}

		if cluster.Appid != act.req.Appid {
			return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("cluster[%+v] not under target app[%+v]", cluster, act.req.Appid)
		}
	}

	for _, zoneid := range act.req.Zoneids {
		zone, err := act.queryZone(zoneid)
		if err != nil {
			return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("can't query zone[%+v] information to create strategy", zoneid)
		}

		if zone.Appid != act.req.Appid {
			return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("zone[%+v] not under target app[%+v]", zone, act.req.Appid)
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryStrategy() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryStrategyReq{
		Seq:   act.req.Seq,
		Bid:   act.req.Bid,
		Appid: act.req.Appid,
		Name:  act.req.Name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateStrategy[%d]| request to datamanager QueryStrategy, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryStrategy(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryStrategy, %+v", err)
	}

	if resp.ErrCode == pbcommon.ErrCode_E_OK {
		act.resp.Strategyid = resp.Strategy.Strategyid
		return pbcommon.ErrCode_E_BS_ALREADY_EXISTS, fmt.Sprintf("strategy with name[%+v] already exist", act.req.Name)
	}

	if resp.ErrCode != pbcommon.ErrCode_E_DM_NOT_FOUND {
		return resp.ErrCode, resp.ErrMsg
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	act.strategies = &strategy.Strategy{
		Appid:      act.req.Appid,
		Clusterids: act.req.Clusterids,
		Zoneids:    act.req.Zoneids,
		Dcs:        act.req.Dcs,
		IPs:        act.req.IPs,
		Labels:     act.req.Labels,
		LabelsAnd:  act.req.LabelsAnd,
	}

	content, err := json.Marshal(act.strategies)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("can't marshal strategy content, %+v", err)
	}

	r := &pbdatamanager.CreateStrategyReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Appid:      act.req.Appid,
		Strategyid: act.newStrategyid,
		Name:       act.req.Name,
		Content:    string(content),
		Memo:       act.req.Memo,
		Creator:    act.req.Creator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateStrategy[%d]| request to datamanager CreateStrategy, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateStrategy(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateStrategy, %+v", err)
	}
	act.resp.Strategyid = resp.Strategyid

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on new strategy created.
	audit.Audit(int32(pbcommon.SourceType_ST_STRATEGY), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, act.resp.Strategyid, act.req.Creator, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	if err := act.genStrategyID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
	}

	if errCode, errMsg := act.queryStrategy(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if errCode, errMsg := act.query(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
