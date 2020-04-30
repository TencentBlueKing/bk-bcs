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

package cluster

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

// CreateAction creates a cluster object.
type CreateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateClusterReq
	resp *pb.CreateClusterResp

	newClusterid string
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateClusterReq, resp *pb.CreateClusterResp) *CreateAction {
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

	length = len(act.req.Creator)
	if length == 0 {
		return errors.New("invalid params, creator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, creator too long")
	}

	if len(act.req.RClusterid) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}

	if len(act.req.Labels) > database.BSCPCLUSTERLABELSLENLIMIT {
		return errors.New("invalid params, labels too long")
	}

	// TODO check cluster labels format.

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *CreateAction) genClusterID() error {
	id, err := common.GenClusterid()
	if err != nil {
		return err
	}
	act.newClusterid = id
	return nil
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateClusterReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Clusterid:  act.newClusterid,
		Name:       act.req.Name,
		Appid:      act.req.Appid,
		Labels:     act.req.Labels,
		RClusterid: act.req.RClusterid,
		Creator:    act.req.Creator,
		Memo:       act.req.Memo,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateCluster[%d]| request to datamanager CreateCluster, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateCluster(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateCluster, %+v", err)
	}
	// maybe already exist
	act.resp.Clusterid = resp.Clusterid

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on new cluster created.
	audit.Audit(int32(pbcommon.SourceType_ST_CLUSTER), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, act.resp.Clusterid, act.req.Creator, act.req.Memo)

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	if err := act.genClusterID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
	}

	// create cluster.
	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
