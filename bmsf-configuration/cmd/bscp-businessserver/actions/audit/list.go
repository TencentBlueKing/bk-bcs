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

package audit

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

// ListAction query audit list.
type ListAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryAuditListReq
	resp *pb.QueryAuditListResp

	audits []*pbcommon.Audit
}

// NewListAction creates new ListAction.
func NewListAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryAuditListReq, resp *pb.QueryAuditListResp) *ListAction {
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
	act.resp.Audits = act.audits
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

	if act.req.SourceType < int32(pbcommon.SourceType_ST_BUSINESS) ||
		act.req.SourceType >= int32(pbcommon.SourceType_ST_END) {
		return errors.New("invalid params, unknow source type")
	}
	if act.req.OpType < int32(pbcommon.SourceOpType_SOT_CREATE) ||
		act.req.OpType >= int32(pbcommon.SourceOpType_SOT_END) {
		return errors.New("invalid params, unknow source op type")
	}

	if len(act.req.Sourceid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, sourceid too long")
	}

	if len(act.req.Operator) > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too long")
	}
	return nil
}

func (act *ListAction) list() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAuditListReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		SourceType: act.req.SourceType,
		OpType:     act.req.OpType,
		Sourceid:   act.req.Sourceid,
		Operator:   act.req.Operator,
		Index:      act.req.Index,
		Limit:      act.req.Limit,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryAuditList[%d]| request to datamanager QueryAuditList, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryAuditList(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryAuditList, %+v", err)
	}
	act.audits = resp.Audits

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *ListAction) Do() error {
	// query audit list.
	if errCode, errMsg := act.list(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
