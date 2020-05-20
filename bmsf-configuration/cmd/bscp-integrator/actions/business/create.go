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

package business

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/integrator"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/logger"
)

// CreateAction creates a business object.
type CreateAction struct {
	viper          *viper.Viper
	businessSvrCli pbbusinessserver.BusinessClient
	md             *structs.IntegrationMetadata

	req  *pb.IntegrateReq
	resp *pb.IntegrateResp
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, businessSvrCli pbbusinessserver.BusinessClient, md *structs.IntegrationMetadata,
	req *pb.IntegrateReq, resp *pb.IntegrateResp) *CreateAction {
	action := &CreateAction{viper: viper, businessSvrCli: businessSvrCli, md: md, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_ITG_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
	length := len(act.md.Spec.BusinessName)
	if length == 0 {
		return errors.New("invalid params, businessName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, businessName too long")
	}

	length = len(act.md.Spec.Depid)
	if length == 0 {
		return errors.New("invalid params, depid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, depid too long")
	}

	length = len(act.md.Spec.Dbid)
	if length == 0 {
		return errors.New("invalid params, dbid missing")
	}
	if length > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, dbid too long")
	}

	length = len(act.md.Spec.Dbname)
	if length == 0 {
		return errors.New("invalid params, dbname missing")
	}
	if length > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, dbname too long")
	}

	if len(act.md.Spec.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.CreateBusinessReq{
		Seq:     act.req.Seq,
		Name:    act.md.Spec.BusinessName,
		Depid:   act.md.Spec.Depid,
		Dbid:    act.md.Spec.Dbid,
		Dbname:  act.md.Spec.Dbname,
		Memo:    act.md.Spec.Memo,
		Creator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateBusiness[%d]| request to businessserver CreateBusiness, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CreateBusiness(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver CreateBusiness, %+v", err)
	}
	act.resp.Bid = resp.Bid

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// create business.
	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
