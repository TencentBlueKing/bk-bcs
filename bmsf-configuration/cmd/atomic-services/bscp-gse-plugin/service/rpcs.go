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

	configaction "bk-bscp/cmd/atomic-services/bscp-gse-plugin/actions/config"
	effectaction "bk-bscp/cmd/atomic-services/bscp-gse-plugin/actions/effect"
	metadataaction "bk-bscp/cmd/atomic-services/bscp-gse-plugin/actions/metadata"
	releaseaction "bk-bscp/cmd/atomic-services/bscp-gse-plugin/actions/release"
	signallingaction "bk-bscp/cmd/atomic-services/bscp-gse-plugin/actions/signalling"
	"bk-bscp/internal/healthz"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/version"
)

// QueryAppMetadata returns metadata informations of target app.
func (p *GSEPlugin) QueryAppMetadata(ctx context.Context,
	req *pb.QueryAppMetadataReq) (*pb.QueryAppMetadataResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.QueryAppMetadataResp)

	defer func() {
		cost := p.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := metadataaction.NewQueryAction(ctx, p.viper, p.gseTunnel, req, response)
	if err := p.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// Access returns gse plugin connection server endpoints used for
// bcs sidecar accesses a stream signalling channel at first.
func (p *GSEPlugin) Access(ctx context.Context, req *pb.AccessReq) (*pb.AccessResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.AccessResp)

	defer func() {
		cost := p.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := signallingaction.NewAccessAction(ctx, p.viper, req, response)
	if err := p.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// Report recvices local config informations reported by sidecar instances.
func (p *GSEPlugin) Report(ctx context.Context, req *pb.ReportReq) (*pb.ReportResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.ReportResp)

	defer func() {
		cost := p.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := effectaction.NewReportAction(ctx, p.viper, p.gseTunnel, req, response)
	if err := p.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// PullRelease returns target or newest release, sidecar requests on time period
// or requests when recviced a publish notification.
func (p *GSEPlugin) PullRelease(ctx context.Context, req *pb.PullReleaseReq) (*pb.PullReleaseResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PullReleaseResp)

	defer func() {
		cost := p.collector.StatRequest(method, response.Code, rtime, time.Now())

		if response.Release != nil {
			copied := *response.Release
			copied.Strategies = common.EmptyStr()
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v need_effect:%+v release:%+v, content_id:%s]",
				method, req.Seq, cost, response.Code, response.Message, response.NeedEffect, copied, response.ContentId)
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := releaseaction.NewPullAction(ctx, p.viper, p.gseTunnel, req, response)
	if err := p.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// PullConfigList returns all configs.
func (p *GSEPlugin) PullConfigList(ctx context.Context, req *pb.PullConfigListReq) (*pb.PullConfigListResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PullConfigListResp)

	defer func() {
		cost := p.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := configaction.NewListAction(ctx, p.viper, p.gseTunnel, req, response)
	if err := p.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// SignallingChannel creates a stream signalling channel, processes config publish notification.
func (p *GSEPlugin) SignallingChannel(stream pb.Connection_SignallingChannelServer) error {
	logger.Info("SignallingChannel| new signalling channel creating now, %+v", stream)

	action := signallingaction.NewSignalAction(stream.Context(), p.viper, p.gseTunnel, p.sessionMgr, stream)

	if err := p.executor.Execute(action); err != nil {
		logger.Errorf("%+v| %+v", stream, err)
	}

	return nil
}

// Healthz returns server health informations.
func (p *GSEPlugin) Healthz(ctx context.Context, req *pb.HealthzReq) (*pb.HealthzResp, error) {
	return &pb.HealthzResp{
		Seq:     req.Seq,
		Code:    pbcommon.ErrCode_E_OK,
		Message: "OK",
		Data: &pbcommon.ModuleHealthzInfo{
			Module:    "gse-plugin",
			Version:   version.VERSION,
			BuildTime: version.BUILDTIME,
			GitHash:   version.GITHASH,
			IsHealthy: true,
			Message:   healthz.HealthStateMessage,
		},
	}, nil
}
