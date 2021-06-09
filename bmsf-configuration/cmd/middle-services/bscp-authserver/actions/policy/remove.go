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
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/common"
)

// RemoveAction is policy remove action.
type RemoveAction struct {
	ctx   context.Context
	viper *viper.Viper

	authMode string

	localAuthController *local.Controller
	bkiamAuthController *bkiam.Controller

	req  *pb.RemovePolicyReq
	resp *pb.RemovePolicyResp
}

// NewRemoveAction creates a new RemoveAction.
func NewRemoveAction(ctx context.Context, viper *viper.Viper, authMode string,
	localAuthController *local.Controller, bkiamAuthController *bkiam.Controller,
	req *pb.RemovePolicyReq, resp *pb.RemovePolicyResp) *RemoveAction {

	action := &RemoveAction{
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
func (act *RemoveAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *RemoveAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_AUTH_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *RemoveAction) Output() error {
	// do nothing.
	return nil
}

func (act *RemoveAction) verify() error {
	var err error

	if err = common.ValidateInt32("mode", act.req.Mode,
		int32(pbauthserver.RemovePolicyMode_RPM_SINGLE), int32(pbauthserver.RemovePolicyMode_RPM_MULTI)); err != nil {
		return err
	}

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

func (act *RemoveAction) removePolicyMultiMode() (pbcommon.ErrCode, string) {
	if _, err := act.localAuthController.RemovePolicy("", act.req.Metadata.V1); err != nil {
		return pbcommon.ErrCode_E_AUTH_LOCAL_REM_POLICY_FAILED, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *RemoveAction) removePolicySingleMode() (pbcommon.ErrCode, string) {
	if _, err := act.localAuthController.RemovePolicy(act.req.Metadata.V0,
		act.req.Metadata.V1, act.req.Metadata.V2); err != nil {
		return pbcommon.ErrCode_E_AUTH_LOCAL_REM_POLICY_FAILED, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *RemoveAction) removePolicy() (pbcommon.ErrCode, string) {
	if act.authMode == auth.AuthModeLocal {
		if act.req.Mode == int32(pbauthserver.RemovePolicyMode_RPM_MULTI) {
			return act.removePolicyMultiMode()
		}
		return act.removePolicySingleMode()
	}

	// TODO bkiam auth mode.
	return pbcommon.ErrCode_E_AUTH_BKIAM_REM_POLICY_FAILED, "bkiam mode not implemented"
}

// Do makes the workflows of this action base on input messages.
func (act *RemoveAction) Do() error {
	// remove policy.
	if errCode, errMsg := act.removePolicy(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
