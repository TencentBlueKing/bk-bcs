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
	"bk-bscp/pkg/bkrepo"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// CreateWithContentAction creates a multi commit with content action object.
type CreateWithContentAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	authSvrCli pbauthserver.AuthClient
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.CreateMultiCommitWithContentReq
	resp *pb.CreateMultiCommitWithContentResp

	newMultiCommitID string
}

// NewCreateWithContentAction creates new CreateWithContentAction.
func NewCreateWithContentAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.CreateMultiCommitWithContentReq, resp *pb.CreateMultiCommitWithContentResp) *CreateWithContentAction {

	action := &CreateWithContentAction{
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

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateWithContentAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateWithContentAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Authorize checks the action authorization.
func (act *CreateWithContentAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *CreateWithContentAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateWithContentAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if len(act.req.Metadatas) == 0 {
		return errors.New("invalid input data, empty metadatas")
	}

	for _, metadata := range act.req.Metadatas {
		if err = common.ValidateString("cfg_id", metadata.CfgId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if len(metadata.Contents) == 0 {
			return errors.New("invalid input data, empty contents")
		}

		for _, content := range metadata.Contents {
			content.ContentId = strings.ToUpper(content.ContentId)
			if err = common.ValidateString("content_id", content.ContentId,
				database.BSCPCONTENTIDLENLIMIT, database.BSCPCONTENTIDLENLIMIT); err != nil {
				return err
			}
		}
	}

	if err = common.ValidateString("memo", act.req.Memo,
		database.BSCPEMPTY, database.BSCPLONGSTRLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *CreateWithContentAction) genMultiCommitID() error {
	id, err := common.GenMultiCommitID()
	if err != nil {
		return err
	}
	act.newMultiCommitID = id
	return nil
}

func (act *CreateWithContentAction) authorize() (pbcommon.ErrCode, string) {
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

func (act *CreateWithContentAction) queryApp() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryAppReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		AppId: act.req.AppId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateMultiCommitWithContent[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryApp, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *CreateWithContentAction) queryConfig(cfgID string) (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryConfigReq{
		Seq:   act.kit.Rid,
		BizId: act.req.BizId,
		AppId: act.req.AppId,
		CfgId: cfgID,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateMultiCommitWithContent[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryConfig(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryConfig, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateWithContentAction) createMultiCommitWithContent() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.CreateMultiCommitWithContentReq{
		Seq:           act.kit.Rid,
		BizId:         act.req.BizId,
		AppId:         act.req.AppId,
		MultiCommitId: act.newMultiCommitID,
		Metadatas:     act.req.Metadatas,
		Memo:          act.req.Memo,
		Operator:      act.kit.User,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("CreateMultiCommitWithContent[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.CreateMultiCommitWithContent(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to datamanager CreateMultiCommitWithContent, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.resp.Data = &pb.CreateMultiCommitWithContentResp_RespData{MultiCommitId: resp.Data.MultiCommitId}
	act.newMultiCommitID = resp.Data.MultiCommitId

	// audit here on new multi commit created.
	audit.Audit(int32(pbcommon.SourceType_ST_MULTI_COMMIT), int32(pbcommon.SourceOpType_SOT_CREATE),
		act.req.BizId, act.newMultiCommitID, act.kit.User, act.req.Memo)

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateWithContentAction) validateConfigs() (pbcommon.ErrCode, string) {
	for _, metadata := range act.req.Metadatas {
		newCommitID, err := common.GenCommitID()
		if err != nil {
			return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("validate config failed, %+v", err)
		}
		metadata.CommitId = newCommitID

		// query config.
		if errCode, errMsg := act.queryConfig(metadata.CfgId); errCode != pbcommon.ErrCode_E_OK {
			return errCode, fmt.Sprintf("validate config failed, %+v", errMsg)
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateWithContentAction) validateContent(contentID string) (pbcommon.ErrCode, string) {
	if !act.req.ValidateContent {
		return pbcommon.ErrCode_E_OK, "OK"
	}

	contentURL, err := bkrepo.GenContentURL(fmt.Sprintf("http://%s", act.viper.GetString("bkrepo.host")),
		act.req.BizId, contentID)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error()
	}

	auth := &bkrepo.Auth{Token: act.viper.GetString("bkrepo.token"), UID: act.kit.User}
	if err := bkrepo.ValidateContentExistence(contentURL, auth, act.viper.GetDuration("bkrepo.timeout")); err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error()
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *CreateWithContentAction) validateContents() (pbcommon.ErrCode, string) {
	for _, metadata := range act.req.Metadatas {
		for _, content := range metadata.Contents {
			if errCode, errMsg := act.validateContent(content.ContentId); errCode != pbcommon.ErrCode_E_OK {
				return errCode, fmt.Sprintf("validate content metadata failed, %+v", errMsg)
			}
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateWithContentAction) Do() error {
	if err := act.genMultiCommitID(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, err.Error())
	}

	// validate configs.
	if errCode, errMsg := act.validateConfigs(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// validate existence.
	if errCode, errMsg := act.validateContents(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create multi commit with content.
	if errCode, errMsg := act.createMultiCommitWithContent(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	return nil
}
