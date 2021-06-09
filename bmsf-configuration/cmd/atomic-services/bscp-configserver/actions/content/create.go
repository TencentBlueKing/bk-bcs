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

package content

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/internal/audit"
	"bk-bscp/internal/authorization"
	"bk-bscp/internal/database"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/bkrepo"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// CreateAction creates a config content object.
type CreateAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	authSvrCli pbauthserver.AuthClient
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateConfigContentReq
	resp *pb.CreateConfigContentResp

	commit *pbcommon.Commit

	labelsOr  []map[string]string
	labelsAnd []map[string]string
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateConfigContentReq, resp *pb.CreateConfigContentResp) *CreateAction {

	action := &CreateAction{
		kit:        kit,
		viper:      viper,
		authSvrCli: authSvrCli,
		dataMgrCli: dataMgrCli,
		req:        req,
		resp:       resp,
	}

	action.resp.Result = true
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	action.labelsOr = []map[string]string{}
	action.labelsAnd = []map[string]string{}

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}

	for _, labelsOr := range act.req.LabelsOr {
		if err := strategy.ValidateLabels(labelsOr.Labels); err != nil {
			return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID,
				fmt.Sprintf("invalid content labels_or index formats, %+v", err))
		}
		if len(labelsOr.Labels) != 0 {
			act.labelsOr = append(act.labelsOr, labelsOr.Labels)
		}
	}

	for _, labelsAnd := range act.req.LabelsAnd {
		if err := strategy.ValidateLabels(labelsAnd.Labels); err != nil {
			return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID,
				fmt.Sprintf("invalid content labels_and index formats, %+v", err))
		}
		if len(labelsAnd.Labels) != 0 {
			act.labelsAnd = append(act.labelsAnd, labelsAnd.Labels)
		}
	}

	return nil
}

// Authorize checks the action authorization.
func (act *CreateAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("commit_id", act.req.CommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	act.req.ContentId = strings.ToUpper(act.req.ContentId)
	if err = common.ValidateString("content_id", act.req.ContentId,
		database.BSCPCONTENTIDLENLIMIT, database.BSCPCONTENTIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CreateAction) authorize() (pbcommon.ErrCode, string) {
	// check authorize resource at first, it may be deleted.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}

	// check resource authorization.
	isAuthorized, err := authorization.Authorize(act.kit, act.req.AppId, auth.LocalAuthAction,
		act.authSvrCli, act.viper.GetDuration("authserver.callTimeout"))
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("authorize failed, %+v", err)
	}

	if !isAuthorized {
		return pbcommon.ErrCode_E_NOT_AUTHORIZED, "not authorized"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) queryApp() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAppReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		AppId: act.req.AppId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateConfigContent[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *CreateAction) queryCommit() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryCommitReq{
		Seq:      act.kit.Rid,
		BizId:    act.req.BizId,
		CommitId: act.req.CommitId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateConfigContent[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryCommit, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.commit = resp.Data

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateAction) validateContent() (pbcommon.ErrCode, string) {
	if !act.req.ValidateContent {
		return pbcommon.ErrCode_E_OK, "OK"
	}

	contentURL, err := bkrepo.GenContentURL(fmt.Sprintf("http://%s", act.viper.GetString("bkrepo.host")),
		act.req.BizId, act.req.ContentId)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error()
	}

	auth := &bkrepo.Auth{Token: act.viper.GetString("bkrepo.token"), UID: act.kit.User}
	if err := bkrepo.ValidateContentExistence(contentURL, auth, act.viper.GetDuration("bkrepo.timeout")); err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	contentIndex := &strategy.ContentIndex{LabelsOr: act.labelsOr, LabelsAnd: act.labelsAnd}
	index, err := json.Marshal(contentIndex)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error()
	}

	r := &pbdatamanager.CreateConfigContentReq{
		Seq:         act.kit.Rid,
		BizId:       act.req.BizId,
		AppId:       act.commit.AppId,
		CfgId:       act.commit.CfgId,
		CommitId:    act.req.CommitId,
		ContentId:   act.req.ContentId,
		ContentSize: act.req.ContentSize,
		Index:       string(index),
		Creator:     act.kit.User,
		Memo:        act.req.Memo,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateConfigContent[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.CreateConfigContent(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager CreateConfigContent, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	// audit here on new config content created.
	audit.Audit(int32(pbcommon.SourceType_ST_CONTENT), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.BizId, act.req.CommitId, act.kit.User, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// query commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// already confirmed.
	if act.commit.State == int32(pbcommon.CommitState_CS_CONFIRMED) {
		return act.Err(pbcommon.ErrCode_E_CS_COMMIT_ALREADY_CONFIRMED,
			"can't create config content, the commit is already confirmed.")
	}

	// already canceled.
	if act.commit.State == int32(pbcommon.CommitState_CS_CANCELED) {
		return act.Err(pbcommon.ErrCode_E_CS_COMMIT_ALREADY_CANCELED,
			"can't create config content, the commit is already canceled.")
	}

	// validate existence.
	if errCode, errMsg := act.validateContent(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create config content.
	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
