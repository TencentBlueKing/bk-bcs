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

	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/publish"
	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// ReloadAction reloads target release or multi release.
type ReloadAction struct {
	ctx        context.Context
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient
	publishMgr *publish.Manager

	req  *pb.ReloadReq
	resp *pb.ReloadResp

	release      *pbcommon.Release
	multiRelease *pbcommon.MultiRelease
}

// NewReloadAction creates new ReloadAction.
func NewReloadAction(ctx context.Context, viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	publishMgr *publish.Manager,
	req *pb.ReloadReq, resp *pb.ReloadResp) *ReloadAction {

	action := &ReloadAction{ctx: ctx, viper: viper, dataMgrCli: dataMgrCli,
		publishMgr: publishMgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ReloadAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReloadAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ReloadAction) Output() error {
	// do nothing.
	return nil
}

func (act *ReloadAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if len(act.req.ReleaseId) == 0 && len(act.req.MultiReleaseId) == 0 {
		return errors.New("invalid input data, release_id or multi_release_id is required")
	}
	if len(act.req.ReleaseId) != 0 && len(act.req.MultiReleaseId) != 0 {
		return errors.New("invalid input data, only support release_id or multi_release_id")
	}

	if err = common.ValidateString("release_id", act.req.ReleaseId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_release_id", act.req.MultiReleaseId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if act.req.ReloadSpec == nil || len(act.req.ReloadSpec.Info) == 0 {
		return errors.New("invalid input data, reload_spec is required")
	}
	if err = common.ValidateString("operator", act.req.Operator,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *ReloadAction) queryRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("Reload[%s]| request to datamanager QueryRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	act.release = resp.Data

	return resp.Code, resp.Message
}

func (act *ReloadAction) queryMultiRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseReq{
		Seq:            act.req.Seq,
		BizId:          act.req.BizId,
		MultiReleaseId: act.req.MultiReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("Reload[%s]| request to datamanager QueryMultiRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryMultiRelease, %+v", err)
	}
	act.multiRelease = resp.Data

	return resp.Code, resp.Message
}

func (act *ReloadAction) reload() (pbcommon.ErrCode, string) {
	publishing := &pbcommon.Publishing{ReloadSpec: act.req.ReloadSpec, Nice: act.req.Nice}

	if len(act.req.ReleaseId) != 0 {
		publishing.BizId = act.release.BizId
		publishing.AppId = act.release.AppId
		publishing.Strategies = act.release.Strategies
	} else {
		publishing.BizId = act.multiRelease.BizId
		publishing.AppId = act.multiRelease.AppId
		publishing.Strategies = act.multiRelease.Strategies
	}

	signalling := pbcommon.Signalling{
		Type:       pbcommon.SignallingType_ST_SignallingTypeReload,
		Publishing: publishing,
	}
	logger.V(2).Infof("Reload[%s]| send publishing message now, %+v", act.req.Seq, signalling)

	if err := act.publishMgr.Publish(&signalling); err != nil {
		return pbcommon.ErrCode_E_TS_PUBLISH_RELEASE_FAILED, fmt.Sprintf("send publishing message, %+v", err)
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *ReloadAction) Do() error {
	if len(act.req.ReleaseId) != 0 {
		// query target release.
		if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		if act.release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) &&
			act.release.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {

			return act.Err(pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN,
				"can't publish the release, it's not in published/rollbacked state.")
		}
	} else {
		// query multi release.
		if errCode, errMsg := act.queryMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		if act.multiRelease.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) &&
			act.multiRelease.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {

			return act.Err(pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN,
				"can't reload the multi release, it's not in published/rollbacked state")
		}
	}

	// publish release reload message.
	if errCode, errMsg := act.reload(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
