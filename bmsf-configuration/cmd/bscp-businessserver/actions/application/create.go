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

package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-businessserver/modules/audit"
	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// CreateAction creates an application object.
type CreateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateAppReq
	resp *pb.CreateAppResp

	newAppid            string
	newDefaultClusterid string
	newDefaultZoneid    string
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateAppReq, resp *pb.CreateAppResp) *CreateAction {
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

	length = len(act.req.Name)
	if length == 0 {
		return errors.New("invalid params, name missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, name too long")
	}

	if act.req.DeployType != int32(pbcommon.DeployType_DT_BCS) &&
		act.req.DeployType != int32(pbcommon.DeployType_DT_GSE) &&
		act.req.DeployType != int32(pbcommon.DeployType_DT_GSE_PLUGIN) {
		return errors.New("invalid params, unknow deployType, 0:bcs  1:gse 2:gse plugin")
	}

	length = len(act.req.Creator)
	if length == 0 {
		return errors.New("invalid params, creator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, creator too long")
	}
	return nil
}

func (act *CreateAction) genAppID() error {
	id, err := common.GenAppid()
	if err != nil {
		return err
	}
	act.newAppid = id
	return nil
}

func (act *CreateAction) genClusterID() error {
	id, err := common.GenClusterid()
	if err != nil {
		return err
	}
	act.newDefaultClusterid = id
	return nil
}

func (act *CreateAction) genZoneID() error {
	id, err := common.GenZoneid()
	if err != nil {
		return err
	}
	act.newDefaultZoneid = id
	return nil
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateAppReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Appid:      act.newAppid,
		Name:       act.req.Name,
		DeployType: act.req.DeployType,
		Creator:    act.req.Creator,
		Memo:       act.req.Memo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateApp[%d]| request to datamanager CreateApp, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateApp, %+v", err)
	}
	act.resp.Appid = resp.Appid
	act.newAppid = resp.Appid

	if resp.ErrCode == pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return pbcommon.ErrCode_E_OK, ""
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on new app created.
	audit.Audit(int32(pbcommon.SourceType_ST_APP), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, act.newAppid, act.req.Creator, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createDefaultCluster() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateClusterReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Clusterid: act.newDefaultClusterid,
		Name:      "default",
		Appid:     act.newAppid,
		Creator:   act.req.Creator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateApp[%d]| request to datamanager CreateCluster, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateCluster(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateCluster, %+v", err)
	}
	act.newDefaultClusterid = resp.Clusterid

	if resp.ErrCode == pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return pbcommon.ErrCode_E_OK, ""
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on new cluster created.
	audit.Audit(int32(pbcommon.SourceType_ST_CLUSTER), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, act.newDefaultClusterid, act.req.Creator, "default")

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createDefaultZone() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateZoneReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Appid:     act.newAppid,
		Clusterid: act.newDefaultClusterid,
		Zoneid:    act.newDefaultZoneid,
		Name:      "default",
		Creator:   act.req.Creator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateApp[%d]| request to datamanager CreateZone, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateZone(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateZone, %+v", err)
	}
	act.newDefaultZoneid = resp.Zoneid

	if resp.ErrCode == pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return pbcommon.ErrCode_E_OK, ""
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on new zone created.
	audit.Audit(int32(pbcommon.SourceType_ST_ZONE), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, act.newDefaultZoneid, act.req.Creator, "default")

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	if err := act.genAppID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
	}

	// create app.
	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create default cluster/zone.
	if err := act.genClusterID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
	}
	if errCode, errMsg := act.createDefaultCluster(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if err := act.genZoneID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
	}
	if errCode, errMsg := act.createDefaultZone(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
