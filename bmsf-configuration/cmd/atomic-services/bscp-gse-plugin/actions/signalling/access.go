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

package signalling

import (
	"context"
	"errors"
	"path/filepath"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/internal/safeviper"
	"bk-bscp/pkg/common"
)

// AccessAction schedules available gse plugin for bcs sidecar to access(only the plugin instance on host).
type AccessAction struct {
	ctx   context.Context
	viper *safeviper.SafeViper

	req  *pb.AccessReq
	resp *pb.AccessResp
}

// NewAccessAction creates new AccessAction.
func NewAccessAction(ctx context.Context, viper *safeviper.SafeViper,
	req *pb.AccessReq, resp *pb.AccessResp) *AccessAction {
	action := &AccessAction{ctx: ctx, viper: viper, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *AccessAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *AccessAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CONNS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *AccessAction) Output() error {
	// do nothing.
	return nil
}

func (act *AccessAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cloud_id", act.req.CloudId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("ip", act.req.Ip,
		database.BSCPNOTEMPTY, database.BSCPNORMALSTRLENLIMIT); err != nil {
		return err
	}
	act.req.Path = filepath.Clean(act.req.Path)
	if err = common.ValidateString("path", act.req.Path,
		database.BSCPNOTEMPTY, database.BSCPCFGFPATHLENLIMIT); err != nil {
		return err
	}
	return nil
}

// Do makes the workflows of this action base on input messages.
func (act *AccessAction) Do() error {
	endpoints := []*pbcommon.Endpoint{
		&pbcommon.Endpoint{
			Ip:   act.viper.GetString("server.endpoint.ip"),
			Port: act.viper.GetInt32("server.endpoint.port"),
		},
	}
	act.resp.Endpoints = endpoints
	return nil
}
