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

	accessaction "bk-bscp/cmd/bscp-connserver/actions/access"
	configsaction "bk-bscp/cmd/bscp-connserver/actions/configs"
	configsetaction "bk-bscp/cmd/bscp-connserver/actions/configset"
	metadataaction "bk-bscp/cmd/bscp-connserver/actions/metadata"
	releaseaction "bk-bscp/cmd/bscp-connserver/actions/release"
	reportaction "bk-bscp/cmd/bscp-connserver/actions/report"
	signallingaction "bk-bscp/cmd/bscp-connserver/actions/signalling"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/pkg/logger"
)

// QueryAppMetadata returns metadata informations of target app.
func (cs *ConnServer) QueryAppMetadata(ctx context.Context, req *pb.QueryAppMetadataReq) (*pb.QueryAppMetadataResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppMetadata[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppMetadataResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := cs.collector.StatRequest("QueryAppMetadata", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppMetadata[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := metadataaction.NewQueryAction(cs.viper, cs.dataMgrCli, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// Access returns connection server endpoints used for
// bcs sidecar accesses a stream signalling channel at first.
func (cs *ConnServer) Access(ctx context.Context, req *pb.AccessReq) (*pb.AccessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("Access[%d]| input[%+v]", req.Seq, req)
	response := &pb.AccessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := cs.collector.StatRequest("Access", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("Access[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := accessaction.NewAccessAction(cs.viper, cs.accessResource, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// Report recvices local config informations reported by sidecar instances.
func (cs *ConnServer) Report(ctx context.Context, req *pb.ReportReq) (*pb.ReportResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("Report[%d]| input[%+v]", req.Seq, req)
	response := &pb.ReportResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := cs.collector.StatRequest("Report", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("Report[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := reportaction.NewReportAction(cs.viper, cs.bcsControllerCli, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// PullRelease returns target or newest release, sidecar requests on time period
// or requests when recviced a publish notification.
func (cs *ConnServer) PullRelease(ctx context.Context, req *pb.PullReleaseReq) (*pb.PullReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PullRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.PullReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := cs.collector.StatRequest("PullRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PullRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewPullAction(cs.viper, cs.bcsControllerCli, cs.dataMgrCli, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// PullReleaseConfigs returns target or newest release configs.
func (cs *ConnServer) PullReleaseConfigs(ctx context.Context, req *pb.PullReleaseConfigsReq) (*pb.PullReleaseConfigsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PullReleaseConfigs[%d]| input[%+v]", req.Seq, req)
	response := &pb.PullReleaseConfigsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := cs.collector.StatRequest("PullReleaseConfigs", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PullReleaseConfigs[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewPullAction(cs.viper, cs.dataMgrCli, cs.configsCache, cs.collector, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// PullConfigSetList returns all configsets.
func (cs *ConnServer) PullConfigSetList(ctx context.Context, req *pb.PullConfigSetListReq) (*pb.PullConfigSetListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PullConfigSetList[%d]| input[%+v]", req.Seq, req)
	response := &pb.PullConfigSetListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := cs.collector.StatRequest("PullConfigSetList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PullConfigSetList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewListAction(cs.viper, cs.dataMgrCli, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// SignallingChannel creates a stream signalling channel, processes config publish notification.
func (cs *ConnServer) SignallingChannel(stream pb.Connection_SignallingChannelServer) error {
	logger.Info("SignallingChannel| new signalling channel creating now, %+v", stream)

	action := signallingaction.NewSignallingAction(cs.viper, cs.dataMgrCli, cs.sessionMgr, cs.collector, stream)
	cs.executor.Execute(action)

	return nil
}
