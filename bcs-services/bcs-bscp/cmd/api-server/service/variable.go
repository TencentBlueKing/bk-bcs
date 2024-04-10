/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbapp "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app"
	pbrelease "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/release"
	pbtv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-variable"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

type variableService struct {
	cfgClient pbcs.ConfigClient
}

func newVariableService(cfgClient pbcs.ConfigClient) *variableService {
	s := &variableService{
		cfgClient: cfgClient,
	}
	return s
}

// ExportGlobalVariables exports global variables.
func (s *variableService) ExportGlobalVariables(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())
	sep := r.URL.Query().Get("sep")
	if sep == "" {
		sep = " "
	}

	vars, err := s.cfgClient.ListTemplateVariables(kt.RpcCtx(), &pbcs.ListTemplateVariablesReq{
		BizId: kt.BizID,
		All:   true,
	})
	if err != nil {
		logs.Errorf("list template variables failed, err: %s", err)
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	var vs []*pbtv.TemplateVariableSpec
	for _, v := range vars.Details {
		vs = append(vs, v.Spec)
	}
	buf := getVarBuffer(vs, sep)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=global_%d_variable.txt", kt.BizID))
	w.Header().Set("Content-Type", "text/plain")
	_, err = buf.WriteTo(w)
	if err != nil {
		logs.Errorf("write response failed, err: %s", err)
		_ = render.Render(w, r, rest.BadRequest(err))
	}
}

// ExportAppVariables exports app variables.
func (s *variableService) ExportAppVariables(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())
	sep := r.URL.Query().Get("sep")
	if sep == "" {
		sep = " "
	}

	vars, err := s.cfgClient.ListAppTmplVariables(kt.RpcCtx(), &pbcs.ListAppTmplVariablesReq{
		BizId: kt.BizID,
		AppId: kt.AppID,
	})
	if err != nil {
		logs.Errorf("list app template variables failed, err: %s", err)
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	buf := getVarBuffer(vars.Details, sep)

	var app *pbapp.App
	if app, err = s.cfgClient.GetApp(kt.RpcCtx(), &pbcs.GetAppReq{
		BizId: kt.BizID,
		AppId: kt.AppID,
	}); err != nil {
		logs.Errorf("get app failed, err: %s", err)
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d_%s_variable.txt",
		kt.BizID, app.Spec.Name))
	w.Header().Set("Content-Type", "text/plain")
	_, err = buf.WriteTo(w)
	if err != nil {
		logs.Errorf("write response failed, err: %s", err)
		_ = render.Render(w, r, rest.BadRequest(err))
	}
}

// ExportReleasedAppVariables exports released app variables.
func (s *variableService) ExportReleasedAppVariables(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())
	sep := r.URL.Query().Get("sep")
	if sep == "" {
		sep = " "
	}
	releaseIDStr := chi.URLParam(r, "release_id")
	releaseID, _ := strconv.Atoi(releaseIDStr)

	vars, err := s.cfgClient.ListReleasedAppTmplVariables(kt.RpcCtx(), &pbcs.ListReleasedAppTmplVariablesReq{
		BizId:     kt.BizID,
		AppId:     kt.AppID,
		ReleaseId: uint32(releaseID),
	})
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}
	buf := getVarBuffer(vars.Details, sep)

	var app *pbapp.App
	if app, err = s.cfgClient.GetApp(kt.RpcCtx(), &pbcs.GetAppReq{
		BizId: kt.BizID,
		AppId: kt.AppID,
	}); err != nil {
		logs.Errorf("get app failed, err: %s", err)
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	var rel *pbrelease.Release
	if rel, err = s.cfgClient.GetRelease(kt.RpcCtx(), &pbcs.GetReleaseReq{
		BizId:     kt.BizID,
		AppId:     kt.AppID,
		ReleaseId: uint32(releaseID),
	}); err != nil {
		logs.Errorf("get release failed, err: %s", err)
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d_%s_%s_variable.txt",
		kt.BizID, app.Spec.Name, rel.Spec.Name))
	w.Header().Set("Content-Type", "text/plain")
	_, err = buf.WriteTo(w)
	if err != nil {
		logs.Errorf("write response failed, err: %s", err)
		_ = render.Render(w, r, rest.BadRequest(err))
	}
}

// getVarBuffer get variable buffer to export
func getVarBuffer(vars []*pbtv.TemplateVariableSpec, sep string) bytes.Buffer {
	var buf bytes.Buffer
	for _, v := range vars {
		// 导出格式：每行一个变量，有四个字段，依次分别是变量名称、变量类型、变量值、变量描述（描述可为空），各字段以分隔符隔开
		buf.WriteString(v.Name)
		buf.WriteString(sep)
		buf.WriteString(v.Type)
		buf.WriteString(sep)
		buf.WriteString(v.DefaultVal)
		buf.WriteString(sep)
		buf.WriteString(v.Memo)
		buf.WriteString("\n")
	}
	return buf
}
