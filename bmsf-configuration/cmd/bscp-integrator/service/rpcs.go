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
	"fmt"
	"time"

	"gopkg.in/yaml.v2"

	businessaction "bk-bscp/cmd/bscp-integrator/actions/business"
	commitaction "bk-bscp/cmd/bscp-integrator/actions/commit"
	constructionaction "bk-bscp/cmd/bscp-integrator/actions/construction"
	pubaction "bk-bscp/cmd/bscp-integrator/actions/publish"
	reloadaction "bk-bscp/cmd/bscp-integrator/actions/reload"
	rollbackaction "bk-bscp/cmd/bscp-integrator/actions/rollback"
	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/integrator"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/logger"
)

// Integrate handles logic integrations.
func (itg *Integrator) Integrate(ctx context.Context, req *pb.IntegrateReq) (*pb.IntegrateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("Integrate[%d]| input[%+v]", req.Seq, req)
	response := &pb.IntegrateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := itg.collector.StatRequest("Integrate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("Integrate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	length := len(req.Metadata)
	if length == 0 {
		response.ErrCode = pbcommon.ErrCode_E_ITG_PARAMS_INVALID
		response.ErrMsg = "invalid params, metadata missing"
		return response, nil
	}
	if length > database.BSCPITGTPLSIZELIMIT {
		response.ErrCode = pbcommon.ErrCode_E_ITG_PARAMS_INVALID
		response.ErrMsg = "invalid params, metadata too large"
		return response, nil
	}

	length = len(req.Operator)
	if length == 0 {
		response.ErrCode = pbcommon.ErrCode_E_ITG_PARAMS_INVALID
		response.ErrMsg = "invalid params, operator missing"
		return response, nil
	}
	if length > database.BSCPNAMELENLIMIT {
		response.ErrCode = pbcommon.ErrCode_E_ITG_PARAMS_INVALID
		response.ErrMsg = "invalid params, operator too long"
		return response, nil
	}

	// integration metadata.
	md := &structs.IntegrationMetadata{}
	if err := yaml.Unmarshal([]byte(req.Metadata), md); err != nil {
		response.ErrCode = pbcommon.ErrCode_E_ITG_METADATA_INVALID
		response.ErrMsg = fmt.Sprintf("unmarshal metadata failed, %+v", err)
		return response, nil
	}

	switch md.Kind {
	case structs.IntegrationMetadataKindBusiness:
		itg.handleKindBusiness(md, req, response)

	case structs.IntegrationMetadataKindConstruction:
		itg.handleKindConstruction(md, req, response)

	case structs.IntegrationMetadataKindCommit:
		itg.handleKindCommit(md, req, response)

	case structs.IntegrationMetadataKindPublish:
		itg.handleKindPublish(md, req, response)

	case structs.IntegrationMetadataKindReload:
		itg.handleKindReload(md, req, response)

	default:
		response.ErrCode = pbcommon.ErrCode_E_ITG_UNKNOW_METADATA_KIND
		response.ErrMsg = fmt.Sprintf("unknow metadata type[%+v]", md.Kind)
	}
	return response, nil
}

func (itg *Integrator) handleKindBusiness(md *structs.IntegrationMetadata, req *pb.IntegrateReq, resp *pb.IntegrateResp) {
	switch md.Op {
	case structs.IntegrationMetadataOpCreate:
		action := businessaction.NewCreateAction(itg.viper, itg.businessSvrCli, md, req, resp)
		itg.executor.Execute(action)
		return

	default:
		resp.ErrCode = pbcommon.ErrCode_E_ITG_UNKNOW_METADATA_OP
		resp.ErrMsg = fmt.Sprintf("unknow business kind metadata op[%+v]", md.Op)
		return
	}
}

func (itg *Integrator) handleKindConstruction(md *structs.IntegrationMetadata, req *pb.IntegrateReq, resp *pb.IntegrateResp) {
	switch md.Op {
	case structs.IntegrationMetadataOpCreate:
		action := constructionaction.NewConstructAction(itg.viper, itg.businessSvrCli, md, req, resp)
		itg.executor.Execute(action)
		return

	default:
		resp.ErrCode = pbcommon.ErrCode_E_ITG_UNKNOW_METADATA_OP
		resp.ErrMsg = fmt.Sprintf("unknow construction kind metadata op[%+v]", md.Op)
		return
	}
}

func (itg *Integrator) handleKindCommit(md *structs.IntegrationMetadata, req *pb.IntegrateReq, resp *pb.IntegrateResp) {
	switch md.Op {
	case structs.IntegrationMetadataOpCommit:
		action := commitaction.NewCommitAction(itg.viper, itg.businessSvrCli, md, req, resp)
		itg.executor.Execute(action)
		return

	default:
		resp.ErrCode = pbcommon.ErrCode_E_ITG_UNKNOW_METADATA_OP
		resp.ErrMsg = fmt.Sprintf("unknow commit kind metadata op[%+v]", md.Op)
		return
	}
}

func (itg *Integrator) handleKindPublish(md *structs.IntegrationMetadata, req *pb.IntegrateReq, resp *pb.IntegrateResp) {
	switch md.Op {
	case structs.IntegrationMetadataOpPub:
		action := pubaction.NewPublishAction(itg.viper, itg.businessSvrCli, md, req, resp)
		itg.executor.Execute(action)
		return

	case structs.IntegrationMetadataOpRollback:
		action := rollbackaction.NewRollbackAction(itg.viper, itg.businessSvrCli, md, req, resp)
		itg.executor.Execute(action)
		return

	default:
		resp.ErrCode = pbcommon.ErrCode_E_ITG_UNKNOW_METADATA_OP
		resp.ErrMsg = fmt.Sprintf("unknow publish kind metadata op[%+v]", md.Op)
		return
	}
}

func (itg *Integrator) handleKindReload(md *structs.IntegrationMetadata, req *pb.IntegrateReq, resp *pb.IntegrateResp) {
	switch md.Op {
	case structs.IntegrationMetadataOpReload:
		action := reloadaction.NewReloadAction(itg.viper, itg.businessSvrCli, md, req, resp)
		itg.executor.Execute(action)
		return

	default:
		resp.ErrCode = pbcommon.ErrCode_E_ITG_UNKNOW_METADATA_OP
		resp.ErrMsg = fmt.Sprintf("unknow reload kind metadata op[%+v]", md.Op)
		return
	}
}
