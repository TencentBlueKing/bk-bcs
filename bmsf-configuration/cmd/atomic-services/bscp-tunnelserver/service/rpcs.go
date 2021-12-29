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

	appinsaction "bk-bscp/cmd/atomic-services/bscp-tunnelserver/actions/appinstance"
	configaction "bk-bscp/cmd/atomic-services/bscp-tunnelserver/actions/config"
	effectaction "bk-bscp/cmd/atomic-services/bscp-tunnelserver/actions/effect"
	healthzaction "bk-bscp/cmd/atomic-services/bscp-tunnelserver/actions/healthz"
	metadataaction "bk-bscp/cmd/atomic-services/bscp-tunnelserver/actions/metadata"
	procattraction "bk-bscp/cmd/atomic-services/bscp-tunnelserver/actions/procattr"
	releaseaction "bk-bscp/cmd/atomic-services/bscp-tunnelserver/actions/release"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// QueryAppMetadata returns metadata informations of target app.
func (ts *TunnelServer) QueryAppMetadata(msgSeqID uint64, agent *modules.AgentInformation,
	req *pb.GTCMDQueryAppMetadataReq) (*pb.GTCMDQueryAppMetadataResp, error) {

	tunnelInfo := fmt.Sprintf("%d|%d|%s", msgSeqID, agent.CloudID, agent.HostIP)

	rtime := time.Now()
	logger.V(2).Infof("QueryAppMetadata[%s][%s]| input[%+v]", tunnelInfo, req.Seq, req)
	response := &pb.GTCMDQueryAppMetadataResp{Seq: req.Seq, Code: pbcommon.ErrCode_E_OK, Message: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryAppMetadata", response.Code, rtime, time.Now())
		logger.V(2).Infof("QueryAppMetadata[%s][%s]| output[%dms][%+v]", tunnelInfo, req.Seq, cost, response)
	}()

	action := metadataaction.NewQueryAction(ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("QueryAppMetadata[%s][%s]| %+v", tunnelInfo, req.Seq, err)
	}

	return response, nil
}

// QueryHostProcAttrList returns ProcAttr list on target host.
func (ts *TunnelServer) QueryHostProcAttrList(msgSeqID uint64, agent *modules.AgentInformation,
	req *pb.GTCMDQueryHostProcAttrListReq) (*pb.GTCMDQueryHostProcAttrListResp, error) {

	tunnelInfo := fmt.Sprintf("%d|%d|%s", msgSeqID, agent.CloudID, agent.HostIP)

	rtime := time.Now()
	logger.V(2).Infof("QueryHostProcAttrList[%s][%s]| input[%+v]", tunnelInfo, req.Seq, req)
	response := &pb.GTCMDQueryHostProcAttrListResp{Seq: req.Seq, Code: pbcommon.ErrCode_E_OK, Message: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("QueryHostProcAttrList", response.Code, rtime, time.Now())
		logger.V(2).Infof("QueryHostProcAttrList[%s][%s]| output[%dms][%+v]", tunnelInfo, req.Seq, cost, response)
	}()

	action := procattraction.NewQueryAction(ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("QueryHostProcAttrList[%s][%s]| %+v", tunnelInfo, req.Seq, err)
	}

	return response, nil
}

// CreateAppInstance creates app instances info.
func (ts *TunnelServer) CreateAppInstance(instance *pbcommon.AppInstance) error {
	rtime := time.Now()
	seq := common.Sequence()

	logger.V(2).Infof("CreateAppInstance[%s]| input[%+v]", seq, instance)
	errCode := pbcommon.ErrCode_E_OK

	defer func() {
		cost := ts.collector.StatRequest("CreateAppInstance", errCode, rtime, time.Now())
		logger.V(2).Infof("CreateAppInstance[%s]| output[%dms][%+v]", seq, cost, errCode)
	}()

	action := appinsaction.NewCreateAction(ts.viper, ts.dataMgrCli, instance)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("CreateAppInstance[%s]| %+v", seq, err)
		errCode = pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN
		return err
	}
	return nil
}

// UpdateAppInstance updates app instances info.
func (ts *TunnelServer) UpdateAppInstance(instance *pbcommon.AppInstance) error {
	rtime := time.Now()
	seq := common.Sequence()

	logger.V(2).Infof("UpdateAppInstance[%s]| input[%+v]", seq, instance)
	errCode := pbcommon.ErrCode_E_OK

	defer func() {
		cost := ts.collector.StatRequest("UpdateAppInstance", errCode, rtime, time.Now())
		logger.V(2).Infof("UpdateAppInstance[%s]| output[%dms][%+v]", seq, cost, errCode)
	}()

	action := appinsaction.NewUpdateAction(ts.viper, ts.dataMgrCli, instance)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("UpdateAppInstance[%s]| %+v", seq, err)
		errCode = pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN
		return err
	}
	return nil
}

// Report recvices local config informations reported by sidecar instances.
func (ts *TunnelServer) Report(msgSeqID uint64, agent *modules.AgentInformation, req *pb.GTCMDEffectReport) error {

	tunnelInfo := fmt.Sprintf("%d|%d|%s", msgSeqID, agent.CloudID, agent.HostIP)

	rtime := time.Now()
	logger.V(2).Infof("Report[%s][%s]| input[%+v]", tunnelInfo, req.Seq, req)
	errCode := pbcommon.ErrCode_E_OK

	defer func() {
		cost := ts.collector.StatRequest("Report", errCode, rtime, time.Now())
		logger.V(2).Infof("Report[%s][%s]| output[%dms][%+v]", tunnelInfo, req.Seq, cost, errCode)
	}()

	action := effectaction.NewReportAction(ts.viper, ts.dataMgrCli, req)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("Report[%s][%s]| %+v", tunnelInfo, req.Seq, err)
		errCode = pbcommon.ErrCode_E_TS_SYSTEM_UNKNOWN
		return err
	}
	return nil
}

// PullRelease returns target or newest release, sidecar requests on time period
// or requests when recviced a publish notification.
func (ts *TunnelServer) PullRelease(msgSeqID uint64, agent *modules.AgentInformation,
	req *pb.GTCMDPullReleaseReq) (*pb.GTCMDPullReleaseResp, error) {

	tunnelInfo := fmt.Sprintf("%d|%d|%s", msgSeqID, agent.CloudID, agent.HostIP)

	rtime := time.Now()
	logger.V(2).Infof("PullRelease[%s][%s]| input[%+v]", tunnelInfo, req.Seq, req)
	response := &pb.GTCMDPullReleaseResp{Seq: req.Seq, Code: pbcommon.ErrCode_E_OK, Message: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("PullRelease", response.Code, rtime, time.Now())

		if response.Release != nil {
			copied := *response.Release
			copied.Strategies = common.EmptyStr()

			logger.V(2).Infof("PullRelease[%s][%s]| output[%dms][code:%+v message:%+v "+
				"need_effect:%+v release:%+v, content_id:%s]",
				tunnelInfo, req.Seq, cost, response.Code, response.Message,
				response.NeedEffect, copied, response.ContentId)
		} else {
			logger.V(2).Infof("PullRelease[%s][%s]| output[%dms][%+v]", tunnelInfo, req.Seq, cost, response)
		}
	}()

	action := releaseaction.NewPullAction(ts.viper, ts.gseControllerCli, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("PullRelease[%s][%s]| %+v", tunnelInfo, req.Seq, err)
	}

	return response, nil
}

// PullConfigList returns all configs.
func (ts *TunnelServer) PullConfigList(msgSeqID uint64, agent *modules.AgentInformation,
	req *pb.GTCMDPullConfigListReq) (*pb.GTCMDPullConfigListResp, error) {

	tunnelInfo := fmt.Sprintf("%d|%d|%s", msgSeqID, agent.CloudID, agent.HostIP)

	rtime := time.Now()
	logger.V(2).Infof("PullConfigList[%s][%s]| input[%+v]", tunnelInfo, req.Seq, req)
	response := &pb.GTCMDPullConfigListResp{Seq: req.Seq, Code: pbcommon.ErrCode_E_OK, Message: "OK"}

	defer func() {
		cost := ts.collector.StatRequest("PullConfigList", response.Code, rtime, time.Now())
		logger.V(2).Infof("PullConfigList[%s][%s]| output[%dms][%+v]", tunnelInfo, req.Seq, cost, response)
	}()

	action := configaction.NewListAction(ts.viper, ts.dataMgrCli, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("PullConfigList[%s][%s]| %+v", tunnelInfo, req.Seq, err)
	}

	return response, nil
}

// PublishRelease is TunnelServer rpc for publishing.
func (ts *TunnelServer) PublishRelease(ctx context.Context, req *pb.PublishReleaseReq) (*pb.PublishReleaseResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PublishReleaseResp)

	defer func() {
		cost := ts.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewPublishAction(ctx, ts.viper, ts.dataMgrCli, ts.publishMgr, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// RollbackRelease is TunnelServer rpc for rollback.
func (ts *TunnelServer) RollbackRelease(ctx context.Context, req *pb.RollbackReleaseReq) (
	*pb.RollbackReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.RollbackReleaseResp)

	defer func() {
		cost := ts.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewRollbackAction(ctx, ts.viper, ts.dataMgrCli, ts.publishMgr, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// Reload is TunnelServer rpc for reload.
func (ts *TunnelServer) Reload(ctx context.Context, req *pb.ReloadReq) (*pb.ReloadResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.ReloadResp)

	defer func() {
		cost := ts.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewReloadAction(ctx, ts.viper, ts.dataMgrCli, ts.publishMgr, req, response)
	if err := ts.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// Healthz returns server health informations.
func (ts *TunnelServer) Healthz(ctx context.Context, req *pb.HealthzReq) (*pb.HealthzResp, error) {
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
