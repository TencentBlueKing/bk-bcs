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

package multirelease

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

// CreateAction creates a multi release object.
type CreateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateMultiReleaseReq
	resp *pb.CreateMultiReleaseResp

	multiCommit *pbcommon.MultiCommit

	strategies        string
	newMultiReleaseid string
	commitids         []string
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateMultiReleaseReq, resp *pb.CreateMultiReleaseResp) *CreateAction {
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

	length = len(act.req.MultiCommitid)
	if length == 0 {
		return errors.New("invalid params, multi commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, multi commitid too long")
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

func (act *CreateAction) genMultiReleaseID() error {
	id, err := common.GenMultiReleaseid()
	if err != nil {
		return err
	}
	act.newMultiReleaseid = id
	return nil
}

func (act *CreateAction) queryCommit(commitid string) (*pbcommon.Commit, pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryCommitReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Commitid: commitid,
		Abstract: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateMultiRelease[%d]| request to datamanager QueryCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryCommit(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryCommit, %+v", err)
	}

	return resp.Commit, resp.ErrCode, resp.ErrMsg
}

func (act *CreateAction) queryConfigSet(cfgsetid string) (*pbcommon.ConfigSet, pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryConfigSetReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Appid:    act.req.Appid,
		Cfgsetid: cfgsetid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateMultiRelease[%d]| request to datamanager QueryConfigSet, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryConfigSet(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryConfigSet, %+v", err)
	}

	return resp.ConfigSet, resp.ErrCode, resp.ErrMsg
}

func (act *CreateAction) createRelease(releaseid, commitid string) (pbcommon.ErrCode, string) {
	// query commit.
	commit, errCode, errMsg := act.queryCommit(commitid)
	if errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}

	// query configset.
	configSet, errCode, errMsg := act.queryConfigSet(commit.Cfgsetid)
	if errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}

	r := &pbdatamanager.CreateReleaseReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		Releaseid:      releaseid,
		Name:           act.req.Name,
		Appid:          act.multiCommit.Appid,
		Cfgsetid:       commit.Cfgsetid,
		CfgsetName:     configSet.Name,
		CfgsetFpath:    configSet.Fpath,
		Strategyid:     act.req.Strategyid,
		Strategies:     act.strategies,
		Commitid:       commit.Commitid,
		Memo:           act.req.Memo,
		Creator:        act.req.Creator,
		State:          int32(pbcommon.ReleaseState_RS_INIT),
		MultiReleaseid: act.newMultiReleaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateMultiRelease[%d]| request to datamanager CreateRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateRelease, %+v", err)
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on release created.
	audit.Audit(int32(pbcommon.SourceType_ST_RELEASE), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, releaseid, act.req.Creator, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) querySubCommitList() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiCommitSubListReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		MultiCommitid: act.req.MultiCommitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateMultiRelease[%d]| request to datamanager QueryMultiCommitSubList, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiCommitSubList(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryMultiCommitSubList, %+v", err)
	}
	act.commitids = resp.Commitids

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	return resp.ErrCode, resp.ErrMsg
}

func (act *CreateAction) createReleases() (pbcommon.ErrCode, string) {
	for _, commitid := range act.commitids {
		newReleaseid, err := common.GenReleaseid()
		if err != nil {
			return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("can't create sub release, %+v", err)
		}

		// create new sub release.
		if errCode, errMsg := act.createRelease(newReleaseid, commitid); errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createMultiRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateMultiReleaseReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		MultiReleaseid: act.newMultiReleaseid,
		Name:           act.req.Name,
		Appid:          act.multiCommit.Appid,
		Strategyid:     act.req.Strategyid,
		Strategies:     act.strategies,
		MultiCommitid:  act.req.MultiCommitid,
		Memo:           act.req.Memo,
		Creator:        act.req.Creator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateMultiRelease[%d]| request to datamanager CreateMultiRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateMultiRelease, %+v", err)
	}
	act.resp.MultiReleaseid = resp.MultiReleaseid

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on release created.
	audit.Audit(int32(pbcommon.SourceType_ST_MULTI_RELEASE), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, act.resp.MultiReleaseid, act.req.Creator, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryMultiCommit() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiCommitReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		MultiCommitid: act.req.MultiCommitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateMultiRelease[%d]| request to datamanager QueryMultiCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryMultiCommit, %+v", err)
	}
	act.multiCommit = resp.MultiCommit

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

	logger.V(2).Infof("CreateMultiRelease[%d]| request to datamanager QueryStrategy, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryStrategy(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryStrategy, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}
	if resp.Strategy.Appid != act.req.Appid {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "can't create multi release with strategy which is not under target app"
	}
	act.strategies = resp.Strategy.Content

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// query multi commit.
	if errCode, errMsg := act.queryMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// check multi commit state.
	if act.multiCommit.State != int32(pbcommon.CommitState_CS_CONFIRMED) {
		return act.Err(pbcommon.ErrCode_E_BS_CREATE_RELEASE_WITH_UNCONFIRMED_COMMIT,
			"can't create multi release with the unconfirmed multi commit.")
	}

	// query strategies.
	if errCode, errMsg := act.queryStrategy(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query multi commit sub list.
	if errCode, errMsg := act.querySubCommitList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if err := act.genMultiReleaseID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
	}

	// create multi release.
	if errCode, errMsg := act.createMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create sub releases.
	if errCode, errMsg := act.createReleases(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	return nil
}
