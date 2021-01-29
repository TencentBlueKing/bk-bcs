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

	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/bcs-controller"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
	mq "bk-bscp/pkg/natsmq"
)

// RollbackAction rollbacks target release object.
type RollbackAction struct {
	ctx        context.Context
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient
	publisher  *mq.Publisher
	pubTopic   string

	req  *pb.RollbackReleaseReq
	resp *pb.RollbackReleaseResp

	release *pbcommon.Release
}

// NewRollbackAction creates new RollbackAction.
func NewRollbackAction(ctx context.Context, viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	publisher *mq.Publisher, pubTopic string,
	req *pb.RollbackReleaseReq, resp *pb.RollbackReleaseResp) *RollbackAction {

	action := &RollbackAction{ctx: ctx, viper: viper, dataMgrCli: dataMgrCli,
		publisher: publisher, pubTopic: pubTopic, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *RollbackAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *RollbackAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BCS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *RollbackAction) Output() error {
	// do nothing.
	return nil
}

func (act *RollbackAction) verify() error {
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

func (act *RollbackAction) queryRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(2).Infof("RollbackRelease[%s]| request to datamanager, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	act.release = resp.Data

	return resp.Code, resp.Message
}

func (act *RollbackAction) rollback() (pbcommon.ErrCode, string) {
	publishing := &pbcommon.Publishing{
		BizId:      act.release.BizId,
		AppId:      act.release.AppId,
		CfgId:      act.release.CfgId,
		CfgName:    act.release.CfgName,
		CfgFpath:   act.release.CfgFpath,
		Serialno:   act.release.Id,
		ReleaseId:  act.release.ReleaseId,
		Strategies: act.release.Strategies,
	}

	signalling := pbcommon.Signalling{
		Type:       pbcommon.SignallingType_ST_SignallingTypeRollback,
		Publishing: publishing,
	}

	msg, err := proto.Marshal(&signalling)
	if err != nil {
		return pbcommon.ErrCode_E_BCS_MARSHAL_PUBLISHING_FAILED,
			fmt.Sprintf("marshal to send rollback publishing message, %+v", err)
	}
	logger.V(2).Infof("RollbackRelease[%s]| send rollback publishing message now, %+v", act.req.Seq, signalling)

	if err := act.publisher.Publish(act.pubTopic, msg); err != nil {
		return pbcommon.ErrCode_E_BCS_PUBLISH_RELEASE_FAILED, fmt.Sprintf("send rollback publishing message, %+v", err)
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *RollbackAction) Do() error {
	// query target release.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.release.State != int32(pbcommon.ReleaseState_RS_ROLLBACKED) {
		return act.Err(pbcommon.ErrCode_E_BCS_SYSTEM_UNKNOWN,
			"can't rollback the release, it's not in rollbacked state.")
	}

	// publish rollback message.
	if errCode, errMsg := act.rollback(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	return nil
}
