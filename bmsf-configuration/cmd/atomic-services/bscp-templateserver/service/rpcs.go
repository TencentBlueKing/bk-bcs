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

package service

import (
	"context"
	"time"

	healthzaction "bk-bscp/cmd/atomic-services/bscp-templateserver/actions/healthz"
	templateaction "bk-bscp/cmd/atomic-services/bscp-templateserver/actions/template"
	templatebindaction "bk-bscp/cmd/atomic-services/bscp-templateserver/actions/template-bind"
	templateversionaction "bk-bscp/cmd/atomic-services/bscp-templateserver/actions/template-version"
	variableaction "bk-bscp/cmd/atomic-services/bscp-templateserver/actions/variable"
	variablegroupaction "bk-bscp/cmd/atomic-services/bscp-templateserver/actions/variable-group"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// CreateTemplateBind create config template bind relation.
func (ts *TemplateServer) CreateTemplateBind(ctx context.Context,
	req *pb.CreateTemplateBindReq) (*pb.CreateTemplateBindResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateTemplateBindResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templatebindaction.NewCreateAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryTemplateBind query target config template bind relation.
func (ts *TemplateServer) QueryTemplateBind(ctx context.Context,
	req *pb.QueryTemplateBindReq) (*pb.QueryTemplateBindResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryTemplateBindResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templatebindaction.NewQueryAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryTemplateBindList query config template bind relation list
func (ts *TemplateServer) QueryTemplateBindList(ctx context.Context,
	req *pb.QueryTemplateBindListReq) (*pb.QueryTemplateBindListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryTemplateBindListResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templatebindaction.NewListAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// DeleteTemplateBind delete target config template bind relation.
func (ts *TemplateServer) DeleteTemplateBind(ctx context.Context,
	req *pb.DeleteTemplateBindReq) (*pb.DeleteTemplateBindResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.DeleteTemplateBindResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templatebindaction.NewDeleteAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateConfigTemplate create config template.
func (ts *TemplateServer) CreateConfigTemplate(ctx context.Context,
	req *pb.CreateConfigTemplateReq) (*pb.CreateConfigTemplateResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateConfigTemplateResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateaction.NewCreateAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryConfigTemplate query config template.
func (ts *TemplateServer) QueryConfigTemplate(ctx context.Context,
	req *pb.QueryConfigTemplateReq) (*pb.QueryConfigTemplateResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryConfigTemplateResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateaction.NewQueryAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryConfigTemplateList query config template list.
func (ts *TemplateServer) QueryConfigTemplateList(ctx context.Context,
	req *pb.QueryConfigTemplateListReq) (*pb.QueryConfigTemplateListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryConfigTemplateListResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateaction.NewListAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateConfigTemplate update config template.
func (ts *TemplateServer) UpdateConfigTemplate(ctx context.Context,
	req *pb.UpdateConfigTemplateReq) (*pb.UpdateConfigTemplateResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateConfigTemplateResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateaction.NewUpdateAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// DeleteConfigTemplate delele config template.
func (ts *TemplateServer) DeleteConfigTemplate(ctx context.Context,
	req *pb.DeleteConfigTemplateReq) (*pb.DeleteConfigTemplateResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.DeleteConfigTemplateResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateaction.NewDeleteAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// RenderConfigTemplate render target config template version.
func (ts *TemplateServer) RenderConfigTemplate(ctx context.Context,
	req *pb.RenderConfigTemplateReq) (*pb.RenderConfigTemplateResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.RenderConfigTemplateResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateaction.NewRenderAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateConfigTemplateVersion create config template version.
func (ts *TemplateServer) CreateConfigTemplateVersion(ctx context.Context,
	req *pb.CreateConfigTemplateVersionReq) (*pb.CreateConfigTemplateVersionResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateConfigTemplateVersionResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateversionaction.NewCreateAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryConfigTemplateVersion query config template version.
func (ts *TemplateServer) QueryConfigTemplateVersion(ctx context.Context,
	req *pb.QueryConfigTemplateVersionReq) (*pb.QueryConfigTemplateVersionResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryConfigTemplateVersionResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateversionaction.NewQueryAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryConfigTemplateVersionList list config template version.
func (ts *TemplateServer) QueryConfigTemplateVersionList(ctx context.Context,
	req *pb.QueryConfigTemplateVersionListReq) (*pb.QueryConfigTemplateVersionListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryConfigTemplateVersionListResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateversionaction.NewListAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateConfigTemplateVersion update config template version.
func (ts *TemplateServer) UpdateConfigTemplateVersion(ctx context.Context,
	req *pb.UpdateConfigTemplateVersionReq) (*pb.UpdateConfigTemplateVersionResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateConfigTemplateVersionResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateversionaction.NewUpdateAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// DeleteConfigTemplateVersion delete config template version.
func (ts *TemplateServer) DeleteConfigTemplateVersion(ctx context.Context,
	req *pb.DeleteConfigTemplateVersionReq) (*pb.DeleteConfigTemplateVersionResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.DeleteConfigTemplateVersionResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := templateversionaction.NewDeleteAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateVariableGroup create variable group.
func (ts *TemplateServer) CreateVariableGroup(ctx context.Context,
	req *pb.CreateVariableGroupReq) (*pb.CreateVariableGroupResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateVariableGroupResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := variablegroupaction.NewCreateAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryVariableGroup query variable group.
func (ts *TemplateServer) QueryVariableGroup(ctx context.Context,
	req *pb.QueryVariableGroupReq) (*pb.QueryVariableGroupResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryVariableGroupResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := variablegroupaction.NewQueryAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryVariableGroupList query variable group list.
func (ts *TemplateServer) QueryVariableGroupList(ctx context.Context,
	req *pb.QueryVariableGroupListReq) (*pb.QueryVariableGroupListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryVariableGroupListResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := variablegroupaction.NewListAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// DeleteVariableGroup delele variable group.
func (ts *TemplateServer) DeleteVariableGroup(ctx context.Context,
	req *pb.DeleteVariableGroupReq) (*pb.DeleteVariableGroupResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.DeleteVariableGroupResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := variablegroupaction.NewDeleteAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateVariable create variable.
func (ts *TemplateServer) CreateVariable(ctx context.Context,
	req *pb.CreateVariableReq) (*pb.CreateVariableResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateVariableResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := variableaction.NewCreateAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryVariable query variable.
func (ts *TemplateServer) QueryVariable(ctx context.Context,
	req *pb.QueryVariableReq) (*pb.QueryVariableResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryVariableResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := variableaction.NewQueryAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryVariableList query variable list.
func (ts *TemplateServer) QueryVariableList(ctx context.Context,
	req *pb.QueryVariableListReq) (*pb.QueryVariableListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryVariableListResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := variableaction.NewListAction(kit, ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateVariable update variable.
func (ts *TemplateServer) UpdateVariable(ctx context.Context,
	req *pb.UpdateVariableReq) (*pb.UpdateVariableResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateVariableResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := variableaction.NewUpdateAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// DeleteVariable delele variable.
func (ts *TemplateServer) DeleteVariable(ctx context.Context,
	req *pb.DeleteVariableReq) (*pb.DeleteVariableResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.DeleteVariableResp)

	defer func() {
		cost := ts.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := variableaction.NewDeleteAction(kit, ts.viper, ts.authSvrCli, ts.dataMgrCli, req, response)
	if err := ts.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// Healthz returns server health informations.
func (ts *TemplateServer) Healthz(ctx context.Context, req *pb.HealthzReq) (*pb.HealthzResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.HealthzResp)

	defer func() {
		cost := ts.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := healthzaction.NewAction(ctx, ts.viper, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}
