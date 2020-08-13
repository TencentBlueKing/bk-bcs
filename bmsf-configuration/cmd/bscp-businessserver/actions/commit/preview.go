/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commit

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbtemplateserver "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// PreviewAction preview target commit object render result
type PreviewAction struct {
	viper          *viper.Viper
	dataMgrCli     pbdatamanager.DataManagerClient
	templateSvrCli pbtemplateserver.TemplateClient

	req  *pb.PreviewCommitReq
	resp *pb.PreviewCommitResp

	commit *pbcommon.Commit
}

// NewPreviewAction create new PreviewAction
func NewPreviewAction(viper *viper.Viper, templateSvrCli pbtemplateserver.TemplateClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.PreviewCommitReq, resp *pb.PreviewCommitResp) *PreviewAction {
	action := &PreviewAction{viper: viper, dataMgrCli: dataMgrCli, templateSvrCli: templateSvrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PreviewAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input message
func (act *PreviewAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PreviewAction) Output() error {
	// do nothing.
	return nil
}

func (act *PreviewAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
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

func (act *PreviewAction) query() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryCommitReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Commitid: act.req.Commitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryCommit[%d]| request to datamanager QueryCommit, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryCommit, %+v", err)
	}
	act.commit = resp.Commit

	return resp.ErrCode, resp.ErrMsg
}

func (act *PreviewAction) renderWithoutTemplate() (pbcommon.ErrCode, string) {
	act.resp.Cfgslist = append(act.resp.Cfgslist, &pbcommon.Configs{
		Bid:      act.commit.Bid,
		Cfgsetid: act.commit.Cfgsetid,
		Commitid: act.commit.Commitid,
		Appid:    act.commit.Appid,
		Creator:  act.commit.Operator,
		Cid:      common.SHA256(string(act.commit.Configs)),
		Content:  act.commit.Configs,
	})
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *PreviewAction) render() (pbcommon.ErrCode, string) {
	req := &pbtemplateserver.PreviewRenderingReq{
		Seq:      act.req.Seq,
		Bid:      act.req.Bid,
		Appid:    act.commit.Appid,
		Commitid: act.req.Commitid,
		Operator: act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PreviewRendering[%d]| request to datamanager PreviewRendering, %+v", act.req.Seq, req)

	resp, err := act.templateSvrCli.PreviewRendering(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager PreviewRendering, %+v", err)
	}
	if resp.ErrCode == pbcommon.ErrCode_E_OK {
		act.resp.Cfgslist = resp.Cfgslist
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *PreviewAction) listConfigs() (pbcommon.ErrCode, string) {
	cfgsList := []*pbcommon.Configs{}

	index := 0
	for {
		r := &pbdatamanager.QueryConfigsListReq{
			Seq:      act.req.Seq,
			Bid:      act.req.Bid,
			Cfgsetid: act.commit.Cfgsetid,
			Commitid: act.commit.Commitid,
			Index:    int32(index),
			Limit:    database.BSCPTEMPLATEBINDINGNUMLIMIT,
		}

		ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
		defer cancel()

		logger.V(2).Infof("QueryConfigsList[%d]| request to datmanager QueryConfigsList, %+v", r.Seq, r)

		resp, err := act.dataMgrCli.QueryConfigsList(ctx, r)
		if err != nil {
			return pbcommon.ErrCode_E_BS_SYSTEM_UNKONW, fmt.Sprintf("request to datmanager QueryConfigsList, %+v", err)
		}
		if resp.ErrCode != pbcommon.ErrCode_E_OK {
			return resp.ErrCode, resp.ErrMsg
		}
		cfgsList = append(cfgsList, resp.Cfgslist...)

		if len(resp.Cfgslist) < database.BSCPTEMPLATEBINDINGNUMLIMIT {
			break
		}

		index += len(resp.Cfgslist)
	}
	act.resp.Cfgslist = cfgsList

	return pbcommon.ErrCode_E_OK, ""
}

// Do do action
func (act *PreviewAction) Do() error {
	// query targey commit.
	if errCode, errMsg := act.query(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// already confirmed.
	if act.commit.State == int32(pbcommon.CommitState_CS_CONFIRMED) {
		if errCode, errMsg := act.listConfigs(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
		return nil
	}

	// already canceled.
	if act.commit.State == int32(pbcommon.CommitState_CS_CANCELED) {
		return act.Err(pbcommon.ErrCode_E_BS_COMMIT_ALREADY_CANCELED, "can't preview the commit which is already canceled.")
	}

	// pre rendering configs
	if len(act.commit.Template) != 0 || len(act.commit.Templateid) != 0 {
		if errCode, errMsg := act.render(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else {
		if errCode, errMsg := act.renderWithoutTemplate(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}

	return nil
}
