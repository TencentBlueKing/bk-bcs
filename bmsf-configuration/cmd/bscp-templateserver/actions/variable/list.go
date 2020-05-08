/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package variable

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// ListAction query variable object list
type ListAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryVariableListReq
	resp *pb.QueryVariableListResp

	vars []*pbcommon.Variable
}

// NewListAction creates new ListAction
func NewListAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryVariableListReq, resp *pb.QueryVariableListResp) *ListAction {
	action := &ListAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ListAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ListAction) Output() error {
	act.resp.Vars = act.vars
	return nil
}

func (act *ListAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	switch pbcommon.VariableType(act.req.Type) {
	case pbcommon.VariableType_VT_CLUSTER:
		if err := common.VerifyID(act.req.Cluster, "cluster"); err != nil {
			return err
		}
		if err := common.VerifyClusterLabels(act.req.ClusterLabels); err != nil {
			return err
		}
	case pbcommon.VariableType_VT_ZONE:
		if err := common.VerifyID(act.req.Cluster, "cluster"); err != nil {
			return err
		}
		if err := common.VerifyClusterLabels(act.req.ClusterLabels); err != nil {
			return err
		}
		if err := common.VerifyID(act.req.Zone, "zone"); err != nil {
			return err
		}
	}

	if err := common.VerifyQueryLimit(act.req.Limit); err != nil {
		return err
	}

	return nil
}

func (act *ListAction) queryVariableList() (pbcommon.ErrCode, string) {
	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	req := &pbdatamanager.QueryVariableListReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		Cluster:       act.req.Cluster,
		ClusterLabels: act.req.ClusterLabels,
		Zone:          act.req.Zone,
		Type:          act.req.Type,
		Index:         act.req.Index,
		Limit:         act.req.Limit,
	}

	logger.V(2).Infof("QueryVariableList[%d]| request to datamanager QueryVariableList, %+v", req.Seq, req)

	resp, err := act.dataMgrCli.QueryVariableList(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryVariableList, %+v", err)
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	act.vars = resp.Vars

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *ListAction) Do() error {
	// query variable
	if errCode, errMsg := act.queryVariableList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
