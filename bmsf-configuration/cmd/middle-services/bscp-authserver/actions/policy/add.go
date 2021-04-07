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

package policy

import (
	"context"
	"errors"

	"github.com/spf13/viper"

	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth/bkiam"
	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth/local"
	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/common"
)

// AddAction is policy add action.
type AddAction struct {
	ctx   context.Context
	viper *viper.Viper

	authMode string

	localAuthController *local.Controller
	bkiamAuthController *bkiam.Controller

	req  *pb.AddPolicyReq
	resp *pb.AddPolicyResp
}

// NewAddAction creates a new AddAction.
func NewAddAction(ctx context.Context, viper *viper.Viper, authMode string,
	localAuthController *local.Controller, bkiamAuthController *bkiam.Controller,
	req *pb.AddPolicyReq, resp *pb.AddPolicyResp) *AddAction {

	action := &AddAction{
		ctx:                 ctx,
		viper:               viper,
		authMode:            authMode,
		localAuthController: localAuthController,
		bkiamAuthController: bkiamAuthController,
		req:                 req,
		resp:                resp,
	}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *AddAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *AddAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_AUTH_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *AddAction) Output() error {
	// do nothing.
	return nil
}

func (act *AddAction) verify() error {
	var err error

	if act.req.Metadata == nil {
		return errors.New("invalid input data, metadata is required")
	}

	// TODO bkiam auth mode.

	if err = common.ValidateString("subject", act.req.Metadata.V0,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}

	if err = common.ValidateString("object", act.req.Metadata.V1,
		database.BSCPEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}

	if err = common.ValidateString("action", act.req.Metadata.V2,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}

	return nil
}

func (act *AddAction) addPolicy() (pbcommon.ErrCode, string) {
	if act.authMode == auth.AuthModeLocal {
		if _, err := act.localAuthController.AddPolicy(act.req.Metadata.V0,
			act.req.Metadata.V1, act.req.Metadata.V2); err != nil {
			return pbcommon.ErrCode_E_AUTH_LOCAL_ADD_POLICY_FAILED, err.Error()
		}
	} else {
		// TODO bkiam auth mode.
		return pbcommon.ErrCode_E_AUTH_BKIAM_ADD_POLICY_FAILED, "bkiam mode not implemented"
	}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *AddAction) Do() error {
	// add new policy.
	if errCode, errMsg := act.addPolicy(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
