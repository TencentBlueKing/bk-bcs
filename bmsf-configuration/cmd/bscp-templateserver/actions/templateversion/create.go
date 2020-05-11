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

package templateversion

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// CreateAction create a template version object
type CreateAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateTemplateVersionReq
	resp *pb.CreateTemplateVersionResp

	newVersionid string
}

// NewCreateAction creates new CreateAction
func NewCreateAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateTemplateVersionReq, resp *pb.CreateTemplateVersionResp) *CreateAction {
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
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
	if err := common.VerifyID(act.req.Bid, "bid"); err != nil {
		return err
	}

	if err := common.VerifyID(act.req.Templateid, "templateid"); err != nil {
		return err
	}

	if err := common.VerifyNormalName(act.req.VersionName, "versionName"); err != nil {
		return err
	}

	if err := common.VerifyTemplateContent(act.req.Content); err != nil {
		return err
	}

	if err := common.VerifyMemo(act.req.Memo); err != nil {
		return err
	}

	return nil
}

func (act *CreateAction) genVersionID() error {
	id, err := common.GenTemplateVersionid()
	if err != nil {
		return err
	}
	act.newVersionid = id
	return nil
}

func (act *CreateAction) queryTemplate() (pbcommon.ErrCode, string) {
	req := &pbdatamanager.QueryConfigTemplateReq{
		Seq:        act.req.Seq,
		Bid:        act.req.Bid,
		Templateid: act.req.Templateid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("QueryConfigTemplate[%d]| request to datamanager QueryConfigTemplate, %+v", req.Seq, req)

	resp, err := act.dataMgrCli.QueryConfigTemplate(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryConfigTemplate, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

func (act *CreateAction) createVersion() (pbcommon.ErrCode, string) {
	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	req := &pbdatamanager.CreateTemplateVersionReq{
		Seq:         act.req.Seq,
		Bid:         act.req.Bid,
		Templateid:  act.req.Templateid,
		Versionid:   act.newVersionid,
		VersionName: act.req.VersionName,
		Content:     act.req.Content,
		Memo:        act.req.Memo,
		Creator:     act.req.Creator,
	}

	logger.V(2).Infof("CreateTemplateVersion[%d]| request to datamanager CreateTemplateVersion, %+v", req.Seq, req)

	resp, err := act.dataMgrCli.CreateTemplateVersion(ctx, req)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager CreateTemplateVersion, %+v", err)
	}

	act.resp.Versionid = resp.Versionid

	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do do action
func (act *CreateAction) Do() error {

	// query config template
	if errCode, errMsg := act.queryTemplate(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// generate version id
	if err := act.genVersionID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_SYSTEM_UNKONW, err.Error())
	}

	// create config template version
	if errCode, errMsg := act.createVersion(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
