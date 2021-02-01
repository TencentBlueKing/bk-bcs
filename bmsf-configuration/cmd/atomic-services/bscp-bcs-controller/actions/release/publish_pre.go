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

package release

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/bcs-controller"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// PublishPreAction publish-pre target release object.
type PublishPreAction struct {
	ctx        context.Context
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.PublishReleasePreReq
	resp *pb.PublishReleasePreResp

	release *pbcommon.Release
	commit  *pbcommon.Commit
}

// NewPublishPreAction creates new PublishPreAction.
func NewPublishPreAction(ctx context.Context, viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.PublishReleasePreReq, resp *pb.PublishReleasePreResp) *PublishPreAction {
	action := &PublishPreAction{ctx: ctx, viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PublishPreAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PublishPreAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BCS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PublishPreAction) Output() error {
	// do nothing.
	return nil
}

func (act *PublishPreAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("release_id", act.req.ReleaseId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *PublishPreAction) queryRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("PublishReleasePre[%s]| request to datamanager, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	act.release = resp.Data

	return resp.Code, resp.Message
}

func (act *PublishPreAction) queryCommit() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryCommitReq{
		Seq:      act.req.Seq,
		BizId:    act.req.BizId,
		CommitId: act.release.CommitId,
	}

	ctx, cancel := context.WithTimeout(act.ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("PublishReleasePre[%s]| request to datamanager, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryCommit, %+v", err)
	}
	act.commit = resp.Data

	return resp.Code, resp.Message
}

// Do makes the workflows of this action base on input messages.
func (act *PublishPreAction) Do() error {
	// query target release.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.release.State == int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
		return act.Err(pbcommon.ErrCode_E_BCS_ALREADY_ROLLBACKED, "can't publish the release, it's already rollbacked.")
	}

	if act.release.State == int32(pbcommon.ReleaseState_RS_CANCELED) {
		return act.Err(pbcommon.ErrCode_E_BCS_ALREADY_CANCELED, "can't publish the release, it's already canceled.")
	}

	// query target commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.commit.State != int32(pbcommon.CommitState_CS_CONFIRMED) {
		return act.Err(pbcommon.ErrCode_E_BCS_SYSTEM_UNKNOWN, "can't publish the release with a commit not confirmed.")
	}

	return nil
}
