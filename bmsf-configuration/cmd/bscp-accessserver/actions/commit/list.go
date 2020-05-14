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

package commit

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/accessserver"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/logger"
)

// ListAction query commit list.
type ListAction struct {
	viper    *viper.Viper
	buSvrCli pbbusinessserver.BusinessClient

	req  *pb.QueryHistoryCommitsReq
	resp *pb.QueryHistoryCommitsResp

	commits []*pbcommon.Commit
}

// NewListAction creates new ListAction.
func NewListAction(viper *viper.Viper, buSvrCli pbbusinessserver.BusinessClient,
	req *pb.QueryHistoryCommitsReq, resp *pb.QueryHistoryCommitsResp) *ListAction {
	action := &ListAction{viper: viper, buSvrCli: buSvrCli, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_AS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ListAction) Output() error {
	act.resp.Commits = act.commits
	return nil
}

func (act *ListAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	if len(act.req.Appid) == 0 && len(act.req.Cfgsetid) == 0 && len(act.req.Operator) == 0 {
		return errors.New("invalid params, appid or cfgsetid or operator missing")
	}

	if len(act.req.Appid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	if len(act.req.Cfgsetid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	if len(act.req.Operator) > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too big")
	}
	return nil
}

func (act *ListAction) list() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.QueryHistoryCommitsReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Appid:     act.req.Appid,
		Cfgsetid:  act.req.Cfgsetid,
		Operator:  act.req.Operator,
		QueryType: act.req.QueryType,
		Index:     act.req.Index,
		Limit:     act.req.Limit,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryHistoryCommits[%d]| request to businessserver QueryHistoryCommits, %+v", act.req.Seq, r)

	resp, err := act.buSvrCli.QueryHistoryCommits(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver QueryHistoryCommits, %+v", err)
	}
	act.commits = resp.Commits

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *ListAction) Do() error {
	// query commit list.
	if errCode, errMsg := act.list(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
