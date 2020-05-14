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

package release

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

// CreateAction creates a release object.
type CreateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateReleaseReq
	resp *pb.CreateReleaseResp

	commit    *pbcommon.Commit
	configSet *pbcommon.ConfigSet

	strategies   string
	newReleaseid string
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateReleaseReq, resp *pb.CreateReleaseResp) *CreateAction {
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

	length = len(act.req.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
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

	if len(act.req.Strategyid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, strategyid too long")
	}
	return nil
}

func (act *CreateAction) genReleaseID() error {
	id, err := common.GenReleaseid()
	if err != nil {
		return err
	}
	act.newReleaseid = id
	return nil
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateReleaseReq{
		Seq:         act.req.Seq,
		Bid:         act.req.Bid,
		Releaseid:   act.newReleaseid,
		Name:        act.req.Name,
		Appid:       act.commit.Appid,
		Cfgsetid:    act.configSet.Cfgsetid,
		CfgsetName:  act.configSet.Name,
		CfgsetFpath: act.configSet.Fpath,
		Strategyid:  act.req.Strategyid,
		Strategies:  act.strategies,
		Commitid:    act.req.Commitid,
		Memo:        act.req.Memo,
		Creator:     act.req.Creator,
		State:       int32(pbcommon.ReleaseState_RS_INIT),
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateRelease[%d]| request to datamanager CreateRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateRelease, %+v", err)
	}
	act.resp.Releaseid = act.newReleaseid

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on release created.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, act.resp.Releaseid, act.req.Creator, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryCommit() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryCommitReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Commitid: act.req.Commitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateRelease[%d]| request to datamanager QueryCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryCommit, %+v", err)
	}
	act.commit = resp.Commit

	return resp.ErrCode, resp.ErrMsg
}

func (act *CreateAction) queryConfigSet() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryConfigSetReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Cfgsetid: act.commit.Cfgsetid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateRelease[%d]| request to datamanager QueryConfigSet, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryConfigSet(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryConfigSet, %+v", err)
	}
	act.configSet = resp.ConfigSet

	return resp.ErrCode, resp.ErrMsg
}

func (act *CreateAction) queryStrategy() (pbcommon.ErrCode, string) {
	if len(act.req.Strategyid) == 0 {
		return pbcommon.ErrCode_E_OK, ""
	}

	r := &pbdatamanager.QueryStrategyReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Strategyid: act.req.Strategyid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateRelease[%d]| request to datamanager QueryStrategy, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryStrategy(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryStrategy, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}
	if resp.Strategy.Appid != act.commit.Appid {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "can't create release with strategy which is not under target app"
	}
	act.strategies = resp.Strategy.Content

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// query commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// check commit state.
	if act.commit.State != int32(pbcommon.CommitState_CS_CONFIRMED) {
		return act.Err(pbcommon.ErrCode_E_BS_CREATE_RELEASE_WITH_UNCONFIRMED_COMMIT, "can't create release with the unconfirmed commit.")
	}

	// query config set.
	if errCode, errMsg := act.queryConfigSet(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query strategies.
	if errCode, errMsg := act.queryStrategy(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if err := act.genReleaseID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
	}

	// create release.
	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
