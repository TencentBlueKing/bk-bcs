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

package reload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/bcs-controller"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/logger"
	mq "bk-bscp/pkg/natsmq"
)

// ReloadAction reloads target release or multi release.
type ReloadAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient
	publisher  *mq.Publisher
	pubTopic   string

	req  *pb.ReloadReq
	resp *pb.ReloadResp

	release      *pbcommon.Release
	multiRelease *pbcommon.MultiRelease
}

// NewReloadAction creates new ReloadAction.
func NewReloadAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient, publisher *mq.Publisher, pubTopic string,
	req *pb.ReloadReq, resp *pb.ReloadResp) *ReloadAction {
	action := &ReloadAction{viper: viper, dataMgrCli: dataMgrCli, publisher: publisher, pubTopic: pubTopic, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ReloadAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReloadAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BCS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ReloadAction) Output() error {
	// do nothing.
	return nil
}

func (act *ReloadAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	if len(act.req.Releaseid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, releaseid too long")
	}

	if len(act.req.MultiReleaseid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, multiReleaseid too long")
	}

	if len(act.req.Releaseid) == 0 && len(act.req.MultiReleaseid) == 0 {
		return errors.New("invalid params, releaseid and multiReleaseid both missing")
	}
	if len(act.req.Releaseid) != 0 && len(act.req.MultiReleaseid) != 0 {
		return errors.New("invalid params, only support releaseid or multiReleaseid")
	}

	if act.req.ReloadSpec == nil || len(act.req.ReloadSpec.Info) == 0 {
		return errors.New("invalid params, reloadSpec missing")
	}

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}
	return nil
}

func (act *ReloadAction) queryRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.req.Releaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Reload[%d]| request to datamanager QueryRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	act.release = resp.Release

	return resp.ErrCode, resp.ErrMsg
}

func (act *ReloadAction) queryMultiRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiReleaseReq{
		Seq:            act.req.Seq,
		Bid:            act.req.Bid,
		MultiReleaseid: act.req.MultiReleaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Reload[%d]| request to datamanager QueryMultiRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryMultiRelease, %+v", err)
	}
	act.multiRelease = resp.MultiRelease

	return resp.ErrCode, resp.ErrMsg
}

func (act *ReloadAction) reload() (pbcommon.ErrCode, string) {
	publishing := structs.Publishing{}

	if len(act.req.Releaseid) != 0 {
		publishing.Bid = act.release.Bid
		publishing.Appid = act.release.Appid
		publishing.Strategies = act.release.Strategies
	} else {
		publishing.Bid = act.multiRelease.Bid
		publishing.Appid = act.multiRelease.Appid
		publishing.Strategies = act.multiRelease.Strategies
	}

	reloadSpec := structs.ReloadSpec{Rollback: act.req.ReloadSpec.Rollback, Info: []structs.EffectInfo{}}

	if len(act.req.ReloadSpec.MultiReleaseid) != 0 {
		reloadSpec.MultiReleaseid = act.req.ReloadSpec.MultiReleaseid
	}

	for _, eInfo := range act.req.ReloadSpec.Info {
		reloadSpec.Info = append(reloadSpec.Info, structs.EffectInfo{Cfgsetid: eInfo.Cfgsetid, Releaseid: eInfo.Releaseid})
	}
	publishing.ReloadSpec = reloadSpec

	signalling := &structs.Signalling{
		Type:       structs.SignallingTypeReload,
		Publishing: publishing,
	}

	msg, err := json.Marshal(signalling)
	if err != nil {
		return pbcommon.ErrCode_E_BCS_MARSHAL_PUBLISHING_FAILED, fmt.Sprintf("marshal to send publishing message, %+v", err)
	}
	logger.V(2).Infof("Reload[%d]| send publishing message now, %+v", act.req.Seq, signalling)

	if err := act.publisher.Publish(act.pubTopic, msg); err != nil {
		return pbcommon.ErrCode_E_BCS_PUBLISH_RELEASE_FAILED, fmt.Sprintf("send publishing message, %+v", err)
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *ReloadAction) Do() error {
	if len(act.req.Releaseid) != 0 {
		// query target release.
		if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		if act.release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) && act.release.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
			return act.Err(pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't reload the release, it's not in published/rollbacked state.")
		}
	} else {
		// query multi release.
		if errCode, errMsg := act.queryMultiRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		if act.multiRelease.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) && act.multiRelease.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
			return act.Err(pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't reload the multi release, it's not in published/rollbacked state")
		}
	}

	// publish release reload message.
	if errCode, errMsg := act.reload(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
