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

package metadata

import (
	"context"
	"errors"
	"fmt"

	"bk-bscp/cmd/atomic-services/bscp-gse-plugin/modules/tunnel"
	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	pbtunnelserver "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/internal/safeviper"
	"bk-bscp/internal/types"
	"bk-bscp/pkg/common"
)

// QueryAction query metadata for bcs sidecar instance.
type QueryAction struct {
	ctx       context.Context
	viper     *safeviper.SafeViper
	gseTunnel *tunnel.Tunnel

	req  *pb.QueryAppMetadataReq
	resp *pb.QueryAppMetadataResp
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(ctx context.Context, viper *safeviper.SafeViper, gseTunnel *tunnel.Tunnel,
	req *pb.QueryAppMetadataReq, resp *pb.QueryAppMetadataResp) *QueryAction {
	action := &QueryAction{ctx: ctx, viper: viper, gseTunnel: gseTunnel, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CONNS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryAction) Output() error {
	// do nothing.
	return nil
}

func (act *QueryAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *QueryAction) query() (pbcommon.ErrCode, string) {
	req := &pbtunnelserver.GTCMDQueryAppMetadataReq{
		Seq:   act.req.Seq,
		BizId: act.req.BizId,
		AppId: act.req.AppId,
	}

	messageID := common.SequenceNum()

	resp, err := act.gseTunnel.QueryAppMetadata(messageID, req)
	if err == types.ErrorTimeout {
		return pbcommon.ErrCode_E_TIMEOUT, "timeout"
	}
	if err != nil {
		return pbcommon.ErrCode_E_CONNS_SYSTEM_UNKNOWN, fmt.Sprintf("request to gse tunnel QueryAppMetadata, %+v", err)
	}

	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// query app metadata.
	if errCode, errMsg := act.query(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
