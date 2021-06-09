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

package template

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/cmd/atomic-services/bscp-templateserver/plugin"
	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/internal/authorization"
	"bk-bscp/internal/database"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/bkrepo"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// RenderAction render target config template version.
type RenderAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	authSvrCli pbauthserver.AuthClient
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.RenderConfigTemplateReq
	resp *pb.RenderConfigTemplateResp

	template        *pbcommon.ConfigTemplate
	templateVersion *pbcommon.ConfigTemplateVersion

	innerVariables []*pbcommon.Variable

	enginePluginName  string
	templateContent   string
	renderedContent   string
	renderedContentID string
}

// NewRenderAction creates new RenderAction
func NewRenderAction(kit kit.Kit, viper *viper.Viper,
	authSvrCli pbauthserver.AuthClient, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.RenderConfigTemplateReq, resp *pb.RenderConfigTemplateResp) *RenderAction {

	action := &RenderAction{
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
func (act *RenderAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *RenderAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_TPL_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Authorize checks the action authorization.
func (act *RenderAction) Authorize() error {
	if errCode, errMsg := act.authorize(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}

// Output handles the output messages.
func (act *RenderAction) Output() error {
	// do nothing.
	return nil
}

func (act *RenderAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("template_id", act.req.TemplateId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("version_id", act.req.VersionId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}

	if len(act.req.Variables) == 0 && len(act.req.VarGroupId) == 0 {
		return errors.New("invalid input data, variables or var_group_id is required")
	}

	if err = common.ValidateString("variables", act.req.Variables,
		database.BSCPEMPTY, database.BSCPTEMPLATEVARSLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("var_group_id", act.req.VarGroupId,
		database.BSCPEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *RenderAction) authorize() (pbcommon.ErrCode, string) {
	// check authorize resource at first, it may be deleted.
	if errCode, errMsg := act.queryConfigTemplate(); errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}

	// check resource authorization.
	isAuthorized, err := authorization.Authorize(act.kit, act.req.TemplateId, auth.LocalAuthAction,
		act.authSvrCli, act.viper.GetDuration("authserver.callTimeout"))
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, fmt.Sprintf("authorize failed, %+v", err)
	}

	if !isAuthorized {
		return pbcommon.ErrCode_E_NOT_AUTHORIZED, "template not authorized"
	}

	if len(act.req.VarGroupId) == 0 {
		// not need to check variable group authorization.
		return pbcommon.ErrCode_E_OK, ""
	}

	// check authorize resource at first, it may be deleted.
	if errCode, errMsg := act.queryVariableGroup(); errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}

	// check resource authorization.
	isAuthorized, err = authorization.Authorize(act.kit, act.req.VarGroupId, auth.LocalAuthAction,
		act.authSvrCli, act.viper.GetDuration("authserver.callTimeout"))
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, fmt.Sprintf("authorize failed, %+v", err)
	}

	if !isAuthorized {
		return pbcommon.ErrCode_E_NOT_AUTHORIZED, "variable group not authorized"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *RenderAction) queryConfigTemplate() (pbcommon.ErrCode, string) {
	if act.template != nil {
		return pbcommon.ErrCode_E_OK, ""
	}

	r := &pbdatamanager.QueryConfigTemplateReq{
		Seq:        act.kit.Rid,
		BizId:      act.req.BizId,
		TemplateId: act.req.TemplateId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("RenderConfigTemplate[%s]| request to DataManager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryConfigTemplate(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to DataManager QueryConfigTemplate, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.template = resp.Data

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RenderAction) queryVariableGroup() (pbcommon.ErrCode, string) {
	if len(act.req.VarGroupId) == 0 {
		return pbcommon.ErrCode_E_OK, ""
	}

	r := &pbdatamanager.QueryVariableGroupReq{
		Seq:        act.kit.Rid,
		BizId:      act.req.BizId,
		VarGroupId: act.req.VarGroupId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("RenderConfigTemplate[%s]| request to DataManager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryVariableGroup(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to DataManager QueryVariableGroup, %+v", err)
	}
	return resp.Code, resp.Message
}

func (act *RenderAction) queryConfigTemplateVersion() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryConfigTemplateVersionReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		VersionId: act.req.VersionId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("RenderConfigTemplate[%s]| request to DataManager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryConfigTemplateVersion(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to DataManager QueryConfigTemplateVersion, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.templateVersion = resp.Data

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RenderAction) queryVariables(index, limit int) ([]*pbcommon.Variable, pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryVariableListReq{
		Seq:        act.kit.Rid,
		BizId:      act.req.BizId,
		VarGroupId: act.req.VarGroupId,
		Page:       &pbcommon.Page{Start: int32(index), Limit: int32(limit)},
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	resp, err := act.dataMgrCli.QueryVariableList(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to DataManager QueryVariableList, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return nil, resp.Code, resp.Message
	}
	return resp.Data.Info, pbcommon.ErrCode_E_OK, ""
}

func (act *RenderAction) queryInnerVariables() (pbcommon.ErrCode, string) {
	if len(act.req.VarGroupId) == 0 {
		// render base on user variables, not bscp group variables.
		return pbcommon.ErrCode_E_OK, ""
	}

	// query bscp group variables of the group.
	index := 0
	limit := database.BSCPQUERYLIMITMB

	for {
		variables, errCode, errMsg := act.queryVariables(index, limit)
		if errCode != pbcommon.ErrCode_E_OK {
			logger.Errorf("RenderConfigTemplate[%s]| query inner variables failed, %s", act.kit.Rid, errMsg)
			return errCode, errMsg
		}
		act.innerVariables = append(act.innerVariables, variables...)

		if len(variables) < limit {
			break
		}
		index += len(variables)
	}

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RenderAction) queryVersionContent() (pbcommon.ErrCode, string) {
	contentURL, err := bkrepo.GenContentURL(fmt.Sprintf("http://%s", act.viper.GetString("bkrepo.host")),
		act.req.BizId, act.templateVersion.ContentId)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, err.Error()
	}

	option := &bkrepo.DownloadContentOption{URL: contentURL, ContentID: act.templateVersion.ContentId}
	auth := &bkrepo.Auth{Token: act.viper.GetString("bkrepo.token"), UID: act.kit.User}

	content, cost, err := bkrepo.DownloadContentInMemory(option, auth, act.viper.GetDuration("bkrepo.timeout"))
	if err != nil {
		logger.Errorf("RenderConfigTemplate[%s]| request to bkrepo download version content[%s] failed, %+v",
			act.kit.Rid, act.templateVersion.ContentId, err)

		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, err.Error()
	}

	logger.V(4).Infof("RenderConfigTemplate[%s]| request to bkrepo download content[%s] success, length[%d] cost[%+v]",
		act.kit.Rid, act.templateVersion.ContentId, len(content), cost)

	act.templateContent = content

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RenderAction) uploadRenderedContent() (pbcommon.ErrCode, string) {
	contentURL, err := bkrepo.GenContentURL(fmt.Sprintf("http://%s", act.viper.GetString("bkrepo.host")),
		act.req.BizId, act.renderedContentID)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, err.Error()
	}

	option := &bkrepo.UploadContentOption{URL: contentURL, ContentID: act.renderedContentID}
	auth := &bkrepo.Auth{Token: act.viper.GetString("bkrepo.token"), UID: act.kit.User}

	cost, err := bkrepo.UploadContentInMemory(option, auth, act.renderedContent, act.viper.GetDuration("bkrepo.timeout"))
	if err != nil {
		logger.Errorf("RenderConfigTemplate[%s]| request to bkrepo upload rendered content[%s] failed, length[%d], %+v",
			act.kit.Rid, act.renderedContentID, len(act.renderedContent), err)

		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, err.Error()
	}

	logger.V(4).Infof("RenderConfigTemplate[%s]| request to bkrepo upload content[%s] success, length[%d] cost[%+v]",
		act.kit.Rid, act.renderedContentID, len(act.renderedContent), cost)

	act.resp.Data = &pb.RenderConfigTemplateResp_RespData{ContentId: act.renderedContentID}

	return pbcommon.ErrCode_E_OK, ""
}

func (act *RenderAction) renderPre() (pbcommon.ErrCode, string) {
	switch pbcommon.TemplateEngineType(act.template.EngineType) {
	case pbcommon.TemplateEngineType_TET_NONE:
		return pbcommon.ErrCode_E_TPL_NONEED_RENDER, "not a template, no need to render"

	case pbcommon.TemplateEngineType_TET_GOLANG:
		act.enginePluginName = plugin.EnginePluginGolang
		return pbcommon.ErrCode_E_OK, "OK"

	case pbcommon.TemplateEngineType_TET_PYTHONMAKO:
		act.enginePluginName = plugin.EnginePluginMako
		return pbcommon.ErrCode_E_OK, "OK"

	case pbcommon.TemplateEngineType_TET_EXTERNAL:
		return pbcommon.ErrCode_E_TPL_NONEED_RENDER, "external template, no need to render"

	default:
		return pbcommon.ErrCode_E_TPL_UNKNOWN_ENGINE_TYPE, "unknown engine type"
	}
}

func (act *RenderAction) render() (pbcommon.ErrCode, string) {
	// create render engine.
	engine, err := plugin.NewEngine(act.viper.GetString("templateplugin.binDir"))
	if err != nil {
		return pbcommon.ErrCode_E_TPL_SYSTEM_UNKNOWN, fmt.Sprintf("create render engine failed, %+v", err)
	}

	// validate plugin.
	if err := engine.ValidatePlugin(act.enginePluginName); err != nil {
		return pbcommon.ErrCode_E_TPL_RENDER_PLUGIN_CHECK_FAILED, fmt.Sprintf("validate plugin failed, %+v", err)
	}

	renderInConf := &plugin.RenderInConf{Template: act.templateContent}
	if len(act.req.Variables) != 0 {
		// render by user variables.
		if err := json.Unmarshal([]byte(act.req.Variables), &renderInConf.Vars); err != nil {
			return pbcommon.ErrCode_E_TPL_RENDER_PLUGIN_CHECK_FAILED,
				fmt.Sprintf("invalid input template vars, %+v", err)
		}
	} else {
		// render by inner variables.
		variables := make(map[string]interface{}, 0)
		for _, variable := range act.innerVariables {
			variables[variable.Name] = variable.Value
		}
		renderInConf.Vars = variables
	}

	// do render.
	out, err := engine.Execute(renderInConf, act.enginePluginName)
	if err != nil {
		return pbcommon.ErrCode_E_TPL_RENDER_FAILED, fmt.Sprintf("execute render failed, %+v", err)
	}
	if out.Code != plugin.ErrCodeOK {
		return pbcommon.ErrCode_E_TPL_RENDER_FAILED,
			fmt.Sprintf("execute render failed, code %d, msg %s", out.Code, out.Message)
	}
	act.renderedContent = out.Content
	act.renderedContentID = common.SHA256(out.Content)

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *RenderAction) Do() error {
	// query config template.
	if errCode, errMsg := act.queryConfigTemplate(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query config template version.
	if errCode, errMsg := act.queryConfigTemplateVersion(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.templateVersion.TemplateId != act.req.TemplateId {
		return act.Err(pbcommon.ErrCode_E_TPL_PARAMS_INVALID, "inconformable template_id and version_id")
	}

	// rander template pre.
	if errCode, errMsg := act.renderPre(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// download version content.
	if errCode, errMsg := act.queryVersionContent(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query inner variables of target group.
	if errCode, errMsg := act.queryInnerVariables(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// rander template.
	if errCode, errMsg := act.render(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// upload render content.
	if errCode, errMsg := act.uploadRenderedContent(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
