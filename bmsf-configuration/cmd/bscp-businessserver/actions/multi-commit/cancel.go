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

package multicommit

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
	"bk-bscp/pkg/logger"
)

// CancelAction cancels target multi commit object.
type CancelAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CancelMultiCommitReq
	resp *pb.CancelMultiCommitResp

	commitids []string
}

// NewCancelAction creates new CancelAction.
func NewCancelAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CancelMultiCommitReq, resp *pb.CancelMultiCommitResp) *CancelAction {
	action := &CancelAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CancelAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CancelAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CancelAction) Output() error {
	// do nothing.
	return nil
}

func (act *CancelAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.MultiCommitid)
	if length == 0 {
		return errors.New("invalid params, multi commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, multi commitid too long")
	}

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}
	return nil
}

func (act *CancelAction) cancel() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CancelMultiCommitReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		MultiCommitid: act.req.MultiCommitid,
		Operator:      act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CancelMultiCommit[%d]| request to datamanager CancelMultiCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CancelMultiCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CancelMultiCommit, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on commit canceled.
	audit.Audit(int32(pbcommon.SourceType_ST_MULTI_COMMIT), int32(pbcommon.SourceOpType_SOT_CANCEL),
		act.req.Bid, act.req.MultiCommitid, act.req.Operator, "")

	return resp.ErrCode, resp.ErrMsg
}

func (act *CancelAction) querySubCommitList() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiCommitSubListReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		MultiCommitid: act.req.MultiCommitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CancelMultiCommit[%d]| request to datamanager QueryMultiCommitSubList, %+v", act.req.Seq, r)

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

func (act *CancelAction) cancelCommit(commitid string) (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CancelCommitReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Commitid: commitid,
		Operator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CancelMultiCommit[%d]| request to datamanager CancelCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CancelCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CancelCommit, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on commit canceled.
	audit.Audit(int32(pbcommon.SourceType_ST_COMMIT), int32(pbcommon.SourceOpType_SOT_CANCEL),
		act.req.Bid, commitid, act.req.Operator, "")

	return resp.ErrCode, resp.ErrMsg
}

func (act *CancelAction) cancelSubCommits() (pbcommon.ErrCode, string) {
	for _, commitid := range act.commitids {
		if errCode, errMsg := act.cancelCommit(commitid); errCode != pbcommon.ErrCode_E_OK {
			logger.Warnf("CancelMultiCommit[%d]| cancel sub commit[%s], %+v", act.req.Seq, commitid, errMsg)
			continue
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CancelAction) Do() error {
	// query multi commit sub commit list.
	if errCode, errMsg := act.querySubCommitList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// cancel multi commit.
	if errCode, errMsg := act.cancel(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// cancel sub commits.
	if errCode, errMsg := act.cancelSubCommits(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
