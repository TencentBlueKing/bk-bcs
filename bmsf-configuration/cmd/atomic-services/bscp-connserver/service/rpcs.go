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

	configaction "bk-bscp/cmd/atomic-services/bscp-connserver/actions/config"
	effectaction "bk-bscp/cmd/atomic-services/bscp-connserver/actions/effect"
	healthzaction "bk-bscp/cmd/atomic-services/bscp-connserver/actions/healthz"
	metadataaction "bk-bscp/cmd/atomic-services/bscp-connserver/actions/metadata"
	releaseaction "bk-bscp/cmd/atomic-services/bscp-connserver/actions/release"
	signallingaction "bk-bscp/cmd/atomic-services/bscp-connserver/actions/signalling"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// QueryAppMetadata returns metadata informations of target app.
func (cs *ConnServer) QueryAppMetadata(ctx context.Context,
	req *pb.QueryAppMetadataReq) (*pb.QueryAppMetadataResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryAppMetadataResp)

	defer func() {
		cost := cs.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := metadataaction.NewQueryAction(ctx, cs.viper, cs.dataMgrCli, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// Access returns connection server endpoints used for
// bcs sidecar accesses a stream signalling channel at first.
func (cs *ConnServer) Access(ctx context.Context, req *pb.AccessReq) (*pb.AccessResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.AccessResp)

	defer func() {
		cost := cs.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := signallingaction.NewAccessAction(ctx, cs.viper, cs.accessResource, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// Report recvices local config informations reported by sidecar instances.
func (cs *ConnServer) Report(ctx context.Context, req *pb.ReportReq) (*pb.ReportResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.ReportResp)

	defer func() {
		cost := cs.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := effectaction.NewReportAction(ctx, cs.viper, cs.dataMgrCli, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// PullRelease returns target or newest release, sidecar requests on time period
// or requests when recviced a publish notification.
func (cs *ConnServer) PullRelease(ctx context.Context, req *pb.PullReleaseReq) (*pb.PullReleaseResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PullReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewPullAction(ctx, cs.viper, cs.bcsControllerCli, cs.dataMgrCli, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// PullConfigList returns all configs.
func (cs *ConnServer) PullConfigList(ctx context.Context, req *pb.PullConfigListReq) (*pb.PullConfigListResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PullConfigListResp)

	defer func() {
		cost := cs.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := configaction.NewListAction(ctx, cs.viper, cs.dataMgrCli, req, response)
	cs.executor.Execute(action)

	return response, nil
}

// SignallingChannel creates a stream signalling channel, processes config publish notification.
func (cs *ConnServer) SignallingChannel(stream pb.Connection_SignallingChannelServer) error {
	logger.Info("SignallingChannel| new signalling channel creating now, %+v", stream)

	action := signallingaction.NewSignalAction(stream.Context(), cs.viper, cs.dataMgrCli,
		cs.sessionMgr, cs.collector, stream)

	cs.executor.Execute(action)

	return nil
}

// Healthz returns server health informations.
func (cs *ConnServer) Healthz(ctx context.Context, req *pb.HealthzReq) (*pb.HealthzResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.HealthzResp)

	defer func() {
		cost := cs.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := healthzaction.NewAction(ctx, cs.viper, req, response)
	cs.executor.Execute(action)

	return response, nil
}
