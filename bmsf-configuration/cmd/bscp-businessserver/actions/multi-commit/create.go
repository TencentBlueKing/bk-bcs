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

package multicommit

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-businessserver/modules/audit"
	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// CreateAction creates a multi commit object.
type CreateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateMultiCommitReq
	resp *pb.CreateMultiCommitResp

	newMultiCommitid string
	succCfgsets      []*pbcommon.CommitResult
	failCfgsets      []*pbcommon.CommitResult

	multiCommit *pbcommon.MultiCommit
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateMultiCommitReq, resp *pb.CreateMultiCommitResp) *CreateAction {
	action := &CreateAction{viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	act.resp.SuccCfgsets = act.succCfgsets
	act.resp.FailCfgsets = act.failCfgsets
	return nil
}

func (act *CreateAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Appid)
	if length == 0 {
		return errors.New("invalid params, appid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	if len(act.req.ReuseCommitid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, reuse commitid too long")
	}

	if len(act.req.Metadatas) == 0 {
		return errors.New("invalid params, invalid metadatas")
	}

	for _, metadata := range act.req.Metadatas {
		length := len(metadata.Cfgsetid)
		if length == 0 {
			return errors.New("invalid params, cfgsetid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, cfgsetid too long")
		}

		if metadata.Configs == nil {
			metadata.Configs = []byte{}
		}

		if len(metadata.Configs) > database.BSCPCONFIGSSIZELIMIT {
			return errors.New("invalid params, configs content too big")
		}
		if len(metadata.Changes) > database.BSCPCHANGESSIZELIMIT {
			return errors.New("invalid params, configs changes too big")
		}

		if len(metadata.Templateid) > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, templateid too long")
		}
		if len(metadata.Template) > database.BSCPTPLSIZELIMIT {
			return errors.New("invalid params, template size too big")
		}
		if len(metadata.TemplateRule) > database.BSCPTPLRULESSIZELIMIT {
			return errors.New("invalid params, template rules too long")
		}

		if len(metadata.Configs) != 0 && len(metadata.Template) != 0 {
			return errors.New("invalid params, configs and template concurrence")
		}
		if len(metadata.Configs) != 0 && len(metadata.Templateid) != 0 {
			return errors.New("invalid params, configs and templateid concurrence")
		}
		if len(metadata.Template) != 0 && len(metadata.Templateid) != 0 {
			return errors.New("invalid params, template and templateid concurrence")
		}
		if len(metadata.Template) != 0 && len(metadata.TemplateRule) == 0 {
			return errors.New("invalid params, empty template rules")
		}
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}

	return nil
}

func (act *CreateAction) genMultiCommitID() error {
	id, err := common.GenMultiCommitid()
	if err != nil {
		return err
	}
	act.newMultiCommitid = id
	return nil
}

func (act *CreateAction) createMultiCommit() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateMultiCommitReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		MultiCommitid: act.newMultiCommitid,
		Appid:         act.req.Appid,
		Memo:          act.req.Memo,
		Operator:      act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateMultiCommit[%d]| request to datamanager CreateMultiCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateMultiCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateMultiCommit, %+v", err)
	}
	act.resp.MultiCommitid = resp.MultiCommitid

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on new multi commit created.
	audit.Audit(int32(pbcommon.SourceType_ST_MULTI_COMMIT), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, act.resp.MultiCommitid, act.req.Operator, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createCommit(commitid string, metadata *pbcommon.CommitMetadata) (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateCommitReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		Commitid:      commitid,
		Appid:         act.req.Appid,
		Operator:      act.req.Operator,
		Cfgsetid:      metadata.Cfgsetid,
		Templateid:    metadata.Templateid,
		Template:      metadata.Template,
		TemplateRule:  metadata.TemplateRule,
		Configs:       metadata.Configs,
		Changes:       metadata.Changes,
		Memo:          act.req.Memo,
		MultiCommitid: act.newMultiCommitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateMultiCommit[%d]| request to datamanager CreateCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.CreateCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateCommit, %+v", err)
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	// audit here on new commit created.
	audit.Audit(int32(pbcommon.SourceType_ST_COMMIT), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.Bid, commitid, act.req.Operator, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) createCommits() (pbcommon.ErrCode, string) {
	for _, metadata := range act.req.Metadatas {
		newCommitid, err := common.GenCommitid()
		if err != nil {
			act.failCfgsets = append(act.failCfgsets, &pbcommon.CommitResult{Cfgsetid: metadata.Cfgsetid, Result: err.Error()})
			continue
		}

		// create a new commit.
		common.DelayRandomMS(50)
		if errCode, errMsg := act.createCommit(newCommitid, metadata); errCode != pbcommon.ErrCode_E_OK {
			act.failCfgsets = append(act.failCfgsets, &pbcommon.CommitResult{Cfgsetid: metadata.Cfgsetid, Result: errMsg})
			continue
		}

		// create success.
		act.succCfgsets = append(act.succCfgsets, &pbcommon.CommitResult{Cfgsetid: metadata.Cfgsetid, Result: "OK", Commitid: newCommitid})
	}

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryMultiCommit() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiCommitReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		MultiCommitid: act.req.ReuseCommitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("CreateMultiCommit[%d]| request to datamanager QueryMultiCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryMultiCommit, %+v", err)
	}
	act.multiCommit = resp.MultiCommit

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	if len(act.req.ReuseCommitid) != 0 {
		// query multi commit.
		if errCode, errMsg := act.queryMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}

		// check multi commit state, and reuse it.
		if act.multiCommit.State != int32(pbcommon.CommitState_CS_INIT) {
			return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, "can't reuse the multi commit not in init state")
		}

		// reuse the multi commit.
		act.newMultiCommitid = act.req.ReuseCommitid
		act.resp.MultiCommitid = act.req.ReuseCommitid
	} else {
		if err := act.genMultiCommitID(); err != nil {
			return act.Err(pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, err.Error())
		}

		// create multi commit.
		if errCode, errMsg := act.createMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}

	// create sub normal commits.
	if errCode, errMsg := act.createCommits(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
