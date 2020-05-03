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

package rollback

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

// RollbackAction handles rollback release logic actions.
type RollbackAction struct {
	viper          *viper.Viper
	businessSvrCli pbbusinessserver.BusinessClient
	md             *structs.IntegrationMetadata

	req  *pb.IntegrateReq
	resp *pb.IntegrateResp

	business *pbcommon.Business
}

// NewRollbackAction creates new RollbackAction.
func NewRollbackAction(viper *viper.Viper, businessSvrCli pbbusinessserver.BusinessClient, md *structs.IntegrationMetadata,
	req *pb.IntegrateReq, resp *pb.IntegrateResp) *RollbackAction {
	action := &RollbackAction{viper: viper, businessSvrCli: businessSvrCli, md: md, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *RollbackAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *RollbackAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_ITG_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *RollbackAction) Output() error {
	// do nothing.
	return nil
}

func (act *RollbackAction) verify() error {
	length := len(act.md.Spec.BusinessName)
	if length == 0 {
		return errors.New("invalid params, businessName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, businessName too long")
	}

	length = len(act.md.Release.Releaseid)
	if length == 0 {
		return errors.New("invalid params, releaseid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, releaseid too long")
	}

	if len(act.md.Release.NewReleaseid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, newReleaseid too long")
	}

	return nil
}

func (act *RollbackAction) queryBusiness() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.QueryBusinessReq{
		Seq:  act.req.Seq,
		Name: act.md.Spec.BusinessName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Rollback[%d]| request to businessserver QueryBusiness, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.QueryBusiness(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver QueryBusiness, %+v", err)
	}
	act.business = resp.Business

	return resp.ErrCode, resp.ErrMsg
}

func (act *RollbackAction) rollback() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.RollbackReleaseReq{
		Seq:          act.req.Seq,
		Bid:          act.business.Bid,
		Releaseid:    act.md.Release.Releaseid,
		NewReleaseid: act.md.Release.NewReleaseid,
		Operator:     act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Rollback[%d]| request to businessserver RollbackRelease, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.RollbackRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver RollbackRelease, %+v", err)
	}
	// new re-publish release id when the newReleaseid not empty.
	// If newReleaseid is empty it would rollback last release state.
	act.resp.Releaseid = resp.Releaseid

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *RollbackAction) Do() error {
	// query busienss information used for publishing.
	if errCode, errMsg := act.queryBusiness(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// rollback target release now.
	if errCode, errMsg := act.rollback(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
