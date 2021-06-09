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

	appinstanceaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/appinstance"
	appaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/application"
	auditaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/audit"
	commitaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/commit"
	configaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/config"
	contentaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/content"
	healthzaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/healthz"
	metadataaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/metadata"
	multicommitaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/multi-commit"
	multireleaseaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/multi-release"
	procattraction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/procattr"
	releaseaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/release"
	shardingaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/sharding"
	shardingdbaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/shardingdb"
	strategyaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/strategy"
	templateaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/template"
	templatebindaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/template-bind"
	templateversionaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/template-version"
	variableaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/variable"
	variablegroupaction "bk-bscp/cmd/atomic-services/bscp-datamanager/actions/variable-group"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// QueryAppMetadata returns metadata informations of target app.
func (dm *DataManager) QueryAppMetadata(ctx context.Context,
	req *pb.QueryAppMetadataReq) (*pb.QueryAppMetadataResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryAppMetadataResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := metadataaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateApp creates new app.
func (dm *DataManager) CreateApp(ctx context.Context, req *pb.CreateAppReq) (*pb.CreateAppResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateAppResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appaction.NewCreateAction(ctx, dm.viper, dm.smgr, dm.authSvrCli, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryApp returns target app.
func (dm *DataManager) QueryApp(ctx context.Context, req *pb.QueryAppReq) (*pb.QueryAppResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryAppResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryAppList returns all apps.
func (dm *DataManager) QueryAppList(ctx context.Context, req *pb.QueryAppListReq) (*pb.QueryAppListResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryAppListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateApp updates target app.
func (dm *DataManager) UpdateApp(ctx context.Context, req *pb.UpdateAppReq) (*pb.UpdateAppResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateAppResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// DeleteApp deletes target app.
func (dm *DataManager) DeleteApp(ctx context.Context, req *pb.DeleteAppReq) (*pb.DeleteAppResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.DeleteAppResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appaction.NewDeleteAction(ctx, dm.viper, dm.smgr, dm.authSvrCli, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateTemplateBind create config template bind relation.
func (dm *DataManager) CreateTemplateBind(ctx context.Context,
	req *pb.CreateTemplateBindReq) (*pb.CreateTemplateBindResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateTemplateBindResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templatebindaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryTemplateBind query target config template bind relation.
func (dm *DataManager) QueryTemplateBind(ctx context.Context,
	req *pb.QueryTemplateBindReq) (*pb.QueryTemplateBindResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryTemplateBindResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templatebindaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryTemplateBindList query config template bind relation list
func (dm *DataManager) QueryTemplateBindList(ctx context.Context,
	req *pb.QueryTemplateBindListReq) (*pb.QueryTemplateBindListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryTemplateBindListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templatebindaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// DeleteTemplateBind delete target config template bind relation.
func (dm *DataManager) DeleteTemplateBind(ctx context.Context,
	req *pb.DeleteTemplateBindReq) (*pb.DeleteTemplateBindResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.DeleteTemplateBindResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templatebindaction.NewDeleteAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateConfigTemplate create config template.
func (dm *DataManager) CreateConfigTemplate(ctx context.Context,
	req *pb.CreateConfigTemplateReq) (*pb.CreateConfigTemplateResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateConfigTemplateResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateaction.NewCreateAction(ctx, dm.viper, dm.smgr, dm.authSvrCli, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryConfigTemplate query config template
func (dm *DataManager) QueryConfigTemplate(ctx context.Context,
	req *pb.QueryConfigTemplateReq) (*pb.QueryConfigTemplateResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryConfigTemplateResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryConfigTemplateList query config template list.
func (dm *DataManager) QueryConfigTemplateList(ctx context.Context,
	req *pb.QueryConfigTemplateListReq) (*pb.QueryConfigTemplateListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryConfigTemplateListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateConfigTemplate update target config template.
func (dm *DataManager) UpdateConfigTemplate(ctx context.Context,
	req *pb.UpdateConfigTemplateReq) (*pb.UpdateConfigTemplateResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateConfigTemplateResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// DeleteConfigTemplate delete target config template.
func (dm *DataManager) DeleteConfigTemplate(ctx context.Context,
	req *pb.DeleteConfigTemplateReq) (*pb.DeleteConfigTemplateResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.DeleteConfigTemplateResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateaction.NewDeleteAction(ctx, dm.viper, dm.smgr, dm.authSvrCli, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateConfigTemplateVersion create config template version.
func (dm *DataManager) CreateConfigTemplateVersion(ctx context.Context,
	req *pb.CreateConfigTemplateVersionReq) (*pb.CreateConfigTemplateVersionResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateConfigTemplateVersionResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateversionaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryConfigTemplateVersion query target config template version.
func (dm *DataManager) QueryConfigTemplateVersion(ctx context.Context,
	req *pb.QueryConfigTemplateVersionReq) (*pb.QueryConfigTemplateVersionResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryConfigTemplateVersionResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateversionaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryConfigTemplateVersionList query config template version list.
func (dm *DataManager) QueryConfigTemplateVersionList(ctx context.Context,
	req *pb.QueryConfigTemplateVersionListReq) (*pb.QueryConfigTemplateVersionListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryConfigTemplateVersionListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateversionaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateConfigTemplateVersion update target config template version.
func (dm *DataManager) UpdateConfigTemplateVersion(ctx context.Context,
	req *pb.UpdateConfigTemplateVersionReq) (*pb.UpdateConfigTemplateVersionResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateConfigTemplateVersionResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateversionaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// DeleteConfigTemplateVersion delete target config template version.
func (dm *DataManager) DeleteConfigTemplateVersion(ctx context.Context,
	req *pb.DeleteConfigTemplateVersionReq) (*pb.DeleteConfigTemplateVersionResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.DeleteConfigTemplateVersionResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := templateversionaction.NewDeleteAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateConfig creates new config.
func (dm *DataManager) CreateConfig(ctx context.Context, req *pb.CreateConfigReq) (*pb.CreateConfigResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateConfigResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := configaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryConfig returns target config.
func (dm *DataManager) QueryConfig(ctx context.Context, req *pb.QueryConfigReq) (*pb.QueryConfigResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryConfigResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := configaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryConfigList returns all configs.
func (dm *DataManager) QueryConfigList(ctx context.Context,
	req *pb.QueryConfigListReq) (*pb.QueryConfigListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryConfigListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := configaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateConfig updates target config.
func (dm *DataManager) UpdateConfig(ctx context.Context, req *pb.UpdateConfigReq) (*pb.UpdateConfigResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateConfigResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := configaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// DeleteConfig deletes target config.
func (dm *DataManager) DeleteConfig(ctx context.Context, req *pb.DeleteConfigReq) (*pb.DeleteConfigResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.DeleteConfigResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := configaction.NewDeleteAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateConfigContent creates new config content.
func (dm *DataManager) CreateConfigContent(ctx context.Context,
	req *pb.CreateConfigContentReq) (*pb.CreateConfigContentResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateConfigContentResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := contentaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryConfigContent returns target config content.
func (dm *DataManager) QueryConfigContent(ctx context.Context,
	req *pb.QueryConfigContentReq) (*pb.QueryConfigContentResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryConfigContentResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := contentaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryConfigContentList returns config content list.
func (dm *DataManager) QueryConfigContentList(ctx context.Context,
	req *pb.QueryConfigContentListReq) (*pb.QueryConfigContentListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryConfigContentListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := contentaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryReleaseConfigContent returns configs of target relase.
func (dm *DataManager) QueryReleaseConfigContent(ctx context.Context,
	req *pb.QueryReleaseConfigContentReq) (*pb.QueryReleaseConfigContentResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryReleaseConfigContentResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())

		if response.Data != nil {
			copied := *response.Data
			copied.Index = common.EmptyStr()
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v content:%+v]",
				method, req.Seq, cost, response.Code, response.Message, copied)
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := contentaction.NewReleaseConfigContentAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateCommit creates new commit.
func (dm *DataManager) CreateCommit(ctx context.Context, req *pb.CreateCommitReq) (*pb.CreateCommitResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := commitaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryCommit returns target commit.
func (dm *DataManager) QueryCommit(ctx context.Context, req *pb.QueryCommitReq) (*pb.QueryCommitResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := commitaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryHistoryCommits returns history commits of target configset.
func (dm *DataManager) QueryHistoryCommits(ctx context.Context,
	req *pb.QueryHistoryCommitsReq) (*pb.QueryHistoryCommitsResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryHistoryCommitsResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := commitaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateCommit updates target commit.
func (dm *DataManager) UpdateCommit(ctx context.Context, req *pb.UpdateCommitReq) (*pb.UpdateCommitResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := commitaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CancelCommit cancels target commit.
func (dm *DataManager) CancelCommit(ctx context.Context, req *pb.CancelCommitReq) (*pb.CancelCommitResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CancelCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := commitaction.NewCancelAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// ConfirmCommit confirms target commit.
func (dm *DataManager) ConfirmCommit(ctx context.Context, req *pb.ConfirmCommitReq) (*pb.ConfirmCommitResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.ConfirmCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := commitaction.NewConfirmAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateMultiCommitWithContent creates new multi commit with content.
func (dm *DataManager) CreateMultiCommitWithContent(ctx context.Context,
	req *pb.CreateMultiCommitWithContentReq) (*pb.CreateMultiCommitWithContentResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateMultiCommitWithContentResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multicommitaction.NewCreateWithContentAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateMultiCommit creates new multi commit.
func (dm *DataManager) CreateMultiCommit(ctx context.Context,
	req *pb.CreateMultiCommitReq) (*pb.CreateMultiCommitResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateMultiCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multicommitaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryMultiCommit returns target multi commit.
func (dm *DataManager) QueryMultiCommit(ctx context.Context,
	req *pb.QueryMultiCommitReq) (*pb.QueryMultiCommitResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryMultiCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multicommitaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryHistoryMultiCommits returns history multi commits of target configset.
func (dm *DataManager) QueryHistoryMultiCommits(ctx context.Context,
	req *pb.QueryHistoryMultiCommitsReq) (*pb.QueryHistoryMultiCommitsResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryHistoryMultiCommitsResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multicommitaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryMultiCommitSubList returns multi commit sub list.
func (dm *DataManager) QueryMultiCommitSubList(ctx context.Context,
	req *pb.QueryMultiCommitSubListReq) (*pb.QueryMultiCommitSubListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryMultiCommitSubListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multicommitaction.NewSubListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateMultiCommit updates target multi commit.
func (dm *DataManager) UpdateMultiCommit(ctx context.Context,
	req *pb.UpdateMultiCommitReq) (*pb.UpdateMultiCommitResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateMultiCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multicommitaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CancelMultiCommit cancels target multi commit.
func (dm *DataManager) CancelMultiCommit(ctx context.Context,
	req *pb.CancelMultiCommitReq) (*pb.CancelMultiCommitResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CancelMultiCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multicommitaction.NewCancelAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// ConfirmMultiCommit confirms target multi commit.
func (dm *DataManager) ConfirmMultiCommit(ctx context.Context,
	req *pb.ConfirmMultiCommitReq) (*pb.ConfirmMultiCommitResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.ConfirmMultiCommitResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multicommitaction.NewConfirmAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateRelease creates new release.
func (dm *DataManager) CreateRelease(ctx context.Context, req *pb.CreateReleaseReq) (*pb.CreateReleaseResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryRelease returns target release.
func (dm *DataManager) QueryRelease(ctx context.Context, req *pb.QueryReleaseReq) (*pb.QueryReleaseResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())

		if response.Data != nil {
			copied := *response.Data
			copied.Strategies = common.EmptyStr()
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v release:%+v]",
				method, req.Seq, cost, response.Code, response.Message, copied)
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := releaseaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryNewestReleases returns newest releases.
func (dm *DataManager) QueryNewestReleases(ctx context.Context,
	req *pb.QueryNewestReleasesReq) (*pb.QueryNewestReleasesResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryNewestReleasesResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(4).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)

		if response.Data != nil {
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v newest_num:%d]",
				method, req.Seq, cost, response.Code, response.Message, len(response.Data.Info))
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := releaseaction.NewNewestAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryHistoryReleases returns history releases.
func (dm *DataManager) QueryHistoryReleases(ctx context.Context,
	req *pb.QueryHistoryReleasesReq) (*pb.QueryHistoryReleasesResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryHistoryReleasesResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())

		if response.Data != nil {
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v total_count:%d info_num:%d]",
				method, req.Seq, cost, response.Code, response.Message, response.Data.TotalCount,
				len(response.Data.Info))
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := releaseaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateRelease updates target release.
func (dm *DataManager) UpdateRelease(ctx context.Context, req *pb.UpdateReleaseReq) (*pb.UpdateReleaseResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CancelRelease cancels target release.
func (dm *DataManager) CancelRelease(ctx context.Context, req *pb.CancelReleaseReq) (*pb.CancelReleaseResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CancelReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewCancelAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// RollbackRelease rollbacks target release.
func (dm *DataManager) RollbackRelease(ctx context.Context,
	req *pb.RollbackReleaseReq) (*pb.RollbackReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.RollbackReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewRollbackAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// PublishRelease publishs target release.
func (dm *DataManager) PublishRelease(ctx context.Context, req *pb.PublishReleaseReq) (*pb.PublishReleaseResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PublishReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewPublishAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateMultiRelease creates new multi release.
func (dm *DataManager) CreateMultiRelease(ctx context.Context,
	req *pb.CreateMultiReleaseReq) (*pb.CreateMultiReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateMultiReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multireleaseaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryMultiRelease returns target multi release.
func (dm *DataManager) QueryMultiRelease(ctx context.Context,
	req *pb.QueryMultiReleaseReq) (*pb.QueryMultiReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryMultiReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())

		if response.Data != nil {
			copied := *response.Data
			copied.Strategies = common.EmptyStr()
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v multi_release:%+v",
				method, req.Seq, cost, response.Code, response.Message, copied)
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := multireleaseaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryMultiReleaseSubList returns target multi release sub list.
func (dm *DataManager) QueryMultiReleaseSubList(ctx context.Context,
	req *pb.QueryMultiReleaseSubListReq) (*pb.QueryMultiReleaseSubListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryMultiReleaseSubListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multireleaseaction.NewSubListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryHistoryMultiReleases returns history multi releases.
func (dm *DataManager) QueryHistoryMultiReleases(ctx context.Context,
	req *pb.QueryHistoryMultiReleasesReq) (*pb.QueryHistoryMultiReleasesResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryHistoryMultiReleasesResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())

		if response.Data != nil {
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v total_count:%d info_num:%d]",
				method, req.Seq, cost, response.Code, response.Message, response.Data.TotalCount,
				len(response.Data.Info))
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := multireleaseaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateMultiRelease updates target multi release.
func (dm *DataManager) UpdateMultiRelease(ctx context.Context,
	req *pb.UpdateMultiReleaseReq) (*pb.UpdateMultiReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateMultiReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multireleaseaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CancelMultiRelease cancels target multi release.
func (dm *DataManager) CancelMultiRelease(ctx context.Context,
	req *pb.CancelMultiReleaseReq) (*pb.CancelMultiReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CancelMultiReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multireleaseaction.NewCancelAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// RollbackMultiRelease rollbacks target multi release.
func (dm *DataManager) RollbackMultiRelease(ctx context.Context,
	req *pb.RollbackMultiReleaseReq) (*pb.RollbackMultiReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.RollbackMultiReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multireleaseaction.NewRollbackAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// PublishMultiRelease publishs target multi release.
func (dm *DataManager) PublishMultiRelease(ctx context.Context,
	req *pb.PublishMultiReleaseReq) (*pb.PublishMultiReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PublishMultiReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := multireleaseaction.NewPublishAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateAppInstance creates new app instance.
func (dm *DataManager) CreateAppInstance(ctx context.Context,
	req *pb.CreateAppInstanceReq) (*pb.CreateAppInstanceResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateAppInstanceResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appinstanceaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryReachableAppInstances returns reachable app instances.
func (dm *DataManager) QueryReachableAppInstances(ctx context.Context,
	req *pb.QueryReachableAppInstancesReq) (*pb.QueryReachableAppInstancesResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryReachableAppInstancesResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appinstanceaction.NewReachableAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateAppInstance updates target app instances.
func (dm *DataManager) UpdateAppInstance(ctx context.Context,
	req *pb.UpdateAppInstanceReq) (*pb.UpdateAppInstanceResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateAppInstanceResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appinstanceaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryMatchedAppInstances returns app instances which matched target strategy.
func (dm *DataManager) QueryMatchedAppInstances(ctx context.Context,
	req *pb.QueryMatchedAppInstancesReq) (*pb.QueryMatchedAppInstancesResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryMatchedAppInstancesResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appinstanceaction.NewMatchedAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryEffectedAppInstances returns app instances which effected target release.
func (dm *DataManager) QueryEffectedAppInstances(ctx context.Context,
	req *pb.QueryEffectedAppInstancesReq) (*pb.QueryEffectedAppInstancesResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryEffectedAppInstancesResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appinstanceaction.NewEffectedAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateAppInstanceRelease creates new app instance release.
func (dm *DataManager) CreateAppInstanceRelease(ctx context.Context,
	req *pb.CreateAppInstanceReleaseReq) (*pb.CreateAppInstanceReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateAppInstanceReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appinstanceaction.NewCreateReleaseAction(ctx, dm.viper, dm.smgr, dm.collector, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryAppInstanceRelease returns release of target app instance.
func (dm *DataManager) QueryAppInstanceRelease(ctx context.Context,
	req *pb.QueryAppInstanceReleaseReq) (*pb.QueryAppInstanceReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryAppInstanceReleaseResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := appinstanceaction.NewQueryReleaseAction(ctx, dm.viper, dm.smgr, dm.collector, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateStrategy creates new strategy.
func (dm *DataManager) CreateStrategy(ctx context.Context, req *pb.CreateStrategyReq) (*pb.CreateStrategyResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateStrategyResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := strategyaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryStrategy returns target strategy.
func (dm *DataManager) QueryStrategy(ctx context.Context, req *pb.QueryStrategyReq) (*pb.QueryStrategyResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryStrategyResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())

		if response.Data != nil {
			copied := *response.Data
			copied.Content = common.EmptyStr()
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v strategy:%+v",
				method, req.Seq, cost, response.Code, response.Message, copied)
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := strategyaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryStrategyList returns all strategies of target app.
func (dm *DataManager) QueryStrategyList(ctx context.Context,
	req *pb.QueryStrategyListReq) (*pb.QueryStrategyListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryStrategyListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())

		if response.Data != nil {
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v total_count:%d info_num:%d]",
				method, req.Seq, cost, response.Code, response.Message, response.Data.TotalCount,
				len(response.Data.Info))
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := strategyaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// DeleteStrategy deletes target strategy.
func (dm *DataManager) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyReq) (*pb.DeleteStrategyResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.DeleteStrategyResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := strategyaction.NewDeleteAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateProcAttr creates new ProcAttr.
func (dm *DataManager) CreateProcAttr(ctx context.Context, req *pb.CreateProcAttrReq) (*pb.CreateProcAttrResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateProcAttrResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := procattraction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryHostProcAttr returns ProcAttr of target app on the host.
func (dm *DataManager) QueryHostProcAttr(ctx context.Context,
	req *pb.QueryHostProcAttrReq) (*pb.QueryHostProcAttrResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryHostProcAttrResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := procattraction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryHostProcAttrList returns ProcAttr list on target host.
func (dm *DataManager) QueryHostProcAttrList(ctx context.Context,
	req *pb.QueryHostProcAttrListReq) (*pb.QueryHostProcAttrListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryHostProcAttrListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := procattraction.NewHostListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryAppProcAttrList returns ProcAttr list of target app.
func (dm *DataManager) QueryAppProcAttrList(ctx context.Context,
	req *pb.QueryAppProcAttrListReq) (*pb.QueryAppProcAttrListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryAppProcAttrListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := procattraction.NewAppListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateProcAttr updates target app ProcAttr on the host.
func (dm *DataManager) UpdateProcAttr(ctx context.Context, req *pb.UpdateProcAttrReq) (*pb.UpdateProcAttrResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateProcAttrResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := procattraction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// DeleteProcAttr deletes target app ProcAttr on the host.
func (dm *DataManager) DeleteProcAttr(ctx context.Context, req *pb.DeleteProcAttrReq) (*pb.DeleteProcAttrResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.DeleteProcAttrResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := procattraction.NewDeleteAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// InitShardingDB registers system default sharding database instance.
func (dm *DataManager) InitShardingDB(ctx context.Context,
	req *pb.InitShardingDBReq) (*pb.InitShardingDBResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.InitShardingDBResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := shardingdbaction.NewInitAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateShardingDB registers new sharding database instance.
func (dm *DataManager) CreateShardingDB(ctx context.Context,
	req *pb.CreateShardingDBReq) (*pb.CreateShardingDBResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateShardingDBResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := shardingdbaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryShardingDB returns target sharding database information.
func (dm *DataManager) QueryShardingDB(ctx context.Context,
	req *pb.QueryShardingDBReq) (*pb.QueryShardingDBResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryShardingDBResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := shardingdbaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryShardingDBList returns all sharding databases.
func (dm *DataManager) QueryShardingDBList(ctx context.Context,
	req *pb.QueryShardingDBListReq) (*pb.QueryShardingDBListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryShardingDBListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := shardingdbaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateShardingDB updates target sharding database.
func (dm *DataManager) UpdateShardingDB(ctx context.Context,
	req *pb.UpdateShardingDBReq) (*pb.UpdateShardingDBResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateShardingDBResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := shardingdbaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateSharding registers new sharding relation.
func (dm *DataManager) CreateSharding(ctx context.Context, req *pb.CreateShardingReq) (*pb.CreateShardingResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateShardingResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := shardingaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QuerySharding returns target sharding relation.
func (dm *DataManager) QuerySharding(ctx context.Context, req *pb.QueryShardingReq) (*pb.QueryShardingResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryShardingResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := shardingaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryShardingList returns sharding relation list.
func (dm *DataManager) QueryShardingList(ctx context.Context, req *pb.QueryShardingListReq) (*pb.QueryShardingListResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryShardingListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := shardingaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateSharding updates target sharding relation.
func (dm *DataManager) UpdateSharding(ctx context.Context, req *pb.UpdateShardingReq) (*pb.UpdateShardingResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateShardingResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := shardingaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateAudit creates new audit.
func (dm *DataManager) CreateAudit(ctx context.Context, req *pb.CreateAuditReq) (*pb.CreateAuditResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateAuditResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := auditaction.NewCreateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryAuditList returns history audits.
func (dm *DataManager) QueryAuditList(ctx context.Context, req *pb.QueryAuditListReq) (*pb.QueryAuditListResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryAuditListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := auditaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateVariableGroup create variable group.
func (dm *DataManager) CreateVariableGroup(ctx context.Context,
	req *pb.CreateVariableGroupReq) (*pb.CreateVariableGroupResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateVariableGroupResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := variablegroupaction.NewCreateAction(ctx, dm.viper, dm.smgr, dm.authSvrCli, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryVariableGroup query variable group.
func (dm *DataManager) QueryVariableGroup(ctx context.Context,
	req *pb.QueryVariableGroupReq) (*pb.QueryVariableGroupResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryVariableGroupResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := variablegroupaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryVariableGroupList query variable group list.
func (dm *DataManager) QueryVariableGroupList(ctx context.Context,
	req *pb.QueryVariableGroupListReq) (*pb.QueryVariableGroupListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryVariableGroupListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := variablegroupaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// DeleteVariableGroup delete target variable group.
func (dm *DataManager) DeleteVariableGroup(ctx context.Context,
	req *pb.DeleteVariableGroupReq) (*pb.DeleteVariableGroupResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.DeleteVariableGroupResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := variablegroupaction.NewDeleteAction(ctx, dm.viper, dm.smgr, dm.authSvrCli, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// CreateVariable create variable.
func (dm *DataManager) CreateVariable(ctx context.Context,
	req *pb.CreateVariableReq) (*pb.CreateVariableResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.CreateVariableResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := variableaction.NewCreateAction(ctx, dm.viper, dm.smgr, dm.authSvrCli, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryVariable query variable.
func (dm *DataManager) QueryVariable(ctx context.Context,
	req *pb.QueryVariableReq) (*pb.QueryVariableResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryVariableResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := variableaction.NewQueryAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// QueryVariableList query variable list.
func (dm *DataManager) QueryVariableList(ctx context.Context,
	req *pb.QueryVariableListReq) (*pb.QueryVariableListResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryVariableListResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := variableaction.NewListAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// UpdateVariable update target variable.
func (dm *DataManager) UpdateVariable(ctx context.Context,
	req *pb.UpdateVariableReq) (*pb.UpdateVariableResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.UpdateVariableResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := variableaction.NewUpdateAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// DeleteVariable delete target variable.
func (dm *DataManager) DeleteVariable(ctx context.Context,
	req *pb.DeleteVariableReq) (*pb.DeleteVariableResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.DeleteVariableResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := variableaction.NewDeleteAction(ctx, dm.viper, dm.smgr, dm.authSvrCli, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// Healthz returns server health informations.
func (dm *DataManager) Healthz(ctx context.Context, req *pb.HealthzReq) (*pb.HealthzResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.HealthzResp)

	defer func() {
		cost := dm.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := healthzaction.NewAction(ctx, dm.viper, dm.smgr, req, response)
	if err := dm.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}
