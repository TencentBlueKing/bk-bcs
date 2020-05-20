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

	configsaction "bk-bscp/cmd/bscp-templateserver/actions/configs"
	templateaction "bk-bscp/cmd/bscp-templateserver/actions/template"
	templatebindingaction "bk-bscp/cmd/bscp-templateserver/actions/templatebinding"
	templatesetaction "bk-bscp/cmd/bscp-templateserver/actions/templateset"
	templateversionaction "bk-bscp/cmd/bscp-templateserver/actions/templateversion"
	variableaction "bk-bscp/cmd/bscp-templateserver/actions/variable"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/templateserver"
	"bk-bscp/pkg/logger"
)

// Render renders configs content base on target commit.
func (ts *TemplateServer) Render(ctx context.Context, req *pb.RenderReq) (*pb.RenderResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("Render[%d]| input[%+v]", req.Seq, req)
	response := &pb.RenderResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("Render", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("Render[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewRenderAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// PreviewRendering previews template rendering result.
func (ts *TemplateServer) PreviewRendering(ctx context.Context, req *pb.PreviewRenderingReq) (*pb.PreviewRenderingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PreviewRendering[%d]| input[%+v]", req.Seq, req)
	response := &pb.PreviewRenderingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("PreviewRendering", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PreviewRendering[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewPreviewAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// CreateConfigTemplateSet create config template set.
func (ts *TemplateServer) CreateConfigTemplateSet(ctx context.Context, req *pb.CreateConfigTemplateSetReq) (*pb.CreateConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigTemplateSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("CreateConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewCreateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// DeleteConfigTemplateSet delete config template set.
func (ts *TemplateServer) DeleteConfigTemplateSet(ctx context.Context, req *pb.DeleteConfigTemplateSetReq) (*pb.DeleteConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigTemplateSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("DeleteConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewDeleteAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// UpdateConfigTemplateSet update config template set.
func (ts *TemplateServer) UpdateConfigTemplateSet(ctx context.Context, req *pb.UpdateConfigTemplateSetReq) (*pb.UpdateConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigTemplateSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("UpdateConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewUpdateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateSet query config template set.
func (ts *TemplateServer) QueryConfigTemplateSet(ctx context.Context, req *pb.QueryConfigTemplateSetReq) (*pb.QueryConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewQueryAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateSetList query config template set list.
func (ts *TemplateServer) QueryConfigTemplateSetList(ctx context.Context, req *pb.QueryConfigTemplateSetListReq) (*pb.QueryConfigTemplateSetListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateSetList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateSetListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryConfigTemplateSetList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateSetList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewListAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// CreateConfigTemplate create config template.
func (ts *TemplateServer) CreateConfigTemplate(ctx context.Context, req *pb.CreateConfigTemplateReq) (*pb.CreateConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigTemplateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("CreateConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewCreateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// DeleteConfigTemplate delele config template.
func (ts *TemplateServer) DeleteConfigTemplate(ctx context.Context, req *pb.DeleteConfigTemplateReq) (*pb.DeleteConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigTemplateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("DeleteConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewDeleteAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// UpdateConfigTemplate update config template.
func (ts *TemplateServer) UpdateConfigTemplate(ctx context.Context, req *pb.UpdateConfigTemplateReq) (*pb.UpdateConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigTemplateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("UpdateConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewUpdateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplate query config template.
func (ts *TemplateServer) QueryConfigTemplate(ctx context.Context, req *pb.QueryConfigTemplateReq) (*pb.QueryConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewQueryAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateList query config template list.
func (ts *TemplateServer) QueryConfigTemplateList(ctx context.Context, req *pb.QueryConfigTemplateListReq) (*pb.QueryConfigTemplateListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryConfigTemplateList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewListAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// CreateTemplateVersion create config template version.
func (ts *TemplateServer) CreateTemplateVersion(ctx context.Context, req *pb.CreateTemplateVersionReq) (*pb.CreateTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateTemplateVersionResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("CreateTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewCreateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// DeleteTemplateVersion delete config template version.
func (ts *TemplateServer) DeleteTemplateVersion(ctx context.Context, req *pb.DeleteTemplateVersionReq) (*pb.DeleteTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteTemplateVersionResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("DeleteTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewDeleteAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// UpdateTemplateVersion update config template version.
func (ts *TemplateServer) UpdateTemplateVersion(ctx context.Context, req *pb.UpdateTemplateVersionReq) (*pb.UpdateTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateTemplateVersionResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("UpdateTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewUpdateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryTemplateVersion query config template version.
func (ts *TemplateServer) QueryTemplateVersion(ctx context.Context, req *pb.QueryTemplateVersionReq) (*pb.QueryTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryTemplateVersionResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewQueryAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryTemplateVersionList list config template version.
func (ts *TemplateServer) QueryTemplateVersionList(ctx context.Context, req *pb.QueryTemplateVersionListReq) (*pb.QueryTemplateVersionListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryTemplateVersionList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryTemplateVersionListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryTemplateVersionList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryTemplateVersionList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewListAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// CreateConfigTemplateBinding create config template binding.
func (ts *TemplateServer) CreateConfigTemplateBinding(ctx context.Context, req *pb.CreateConfigTemplateBindingReq) (*pb.CreateConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigTemplateBindingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("CreateConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewCreateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// DeleteConfigTemplateBinding delete config template binding.
func (ts *TemplateServer) DeleteConfigTemplateBinding(ctx context.Context, req *pb.DeleteConfigTemplateBindingReq) (*pb.DeleteConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigTemplateBindingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("DeleteConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewDeleteAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// SyncConfigTemplateBinding sync config template binding.
func (ts *TemplateServer) SyncConfigTemplateBinding(ctx context.Context, req *pb.SyncConfigTemplateBindingReq) (*pb.SyncConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("SyncConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.SyncConfigTemplateBindingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("SyncConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("SyncConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewUpdateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateBinding query config template binding.
func (ts *TemplateServer) QueryConfigTemplateBinding(ctx context.Context, req *pb.QueryConfigTemplateBindingReq) (*pb.QueryConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateBindingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewQueryAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateBindingList query config template binding list.
func (ts *TemplateServer) QueryConfigTemplateBindingList(ctx context.Context, req *pb.QueryConfigTemplateBindingListReq) (*pb.QueryConfigTemplateBindingListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateBindingList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateBindingListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryConfigTemplateBindingList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateBindingList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewListAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// CreateVariable create variable.
func (ts *TemplateServer) CreateVariable(ctx context.Context, req *pb.CreateVariableReq) (*pb.CreateVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateVariableResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("CreateVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewCreateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// DeleteVariable delete variable.
func (ts *TemplateServer) DeleteVariable(ctx context.Context, req *pb.DeleteVariableReq) (*pb.DeleteVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteVariableResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("DeleteVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewDeleteAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// UpdateVariable update variable.
func (ts *TemplateServer) UpdateVariable(ctx context.Context, req *pb.UpdateVariableReq) (*pb.UpdateVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateVariableResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("UpdateVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewUpdateAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryVariable query variable.
func (ts *TemplateServer) QueryVariable(ctx context.Context, req *pb.QueryVariableReq) (*pb.QueryVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryVariableResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewQueryAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}

// QueryVariableList query variable list.
func (ts *TemplateServer) QueryVariableList(ctx context.Context, req *pb.QueryVariableListReq) (*pb.QueryVariableListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryVariableList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryVariableListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryVariableList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryVariableList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewListAction(ts.viper, ts.dataMgrCli, req, response)
	ts.executor.Execute(action)

	return response, nil
}
