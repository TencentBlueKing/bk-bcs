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

	healthzaction "bk-bscp/cmd/atomic-services/bscp-gse-controller/actions/healthz"
	releaseaction "bk-bscp/cmd/atomic-services/bscp-gse-controller/actions/release"
	pb "bk-bscp/internal/protocol/gse-controller"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// PublishRelease publishs target release.
func (c *GSEController) PublishRelease(ctx context.Context,
	req *pb.PublishReleaseReq) (*pb.PublishReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PublishReleaseResp)

	defer func() {
		cost := c.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewPublishAction(ctx, c.viper, c.dataMgrCli, c.tunnelServerCli, req, response)
	if err := c.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// RollbackRelease rollbacks target release.
func (c *GSEController) RollbackRelease(ctx context.Context,
	req *pb.RollbackReleaseReq) (*pb.RollbackReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.RollbackReleaseResp)

	defer func() {
		cost := c.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewRollbackAction(ctx, c.viper, c.dataMgrCli, c.tunnelServerCli, req, response)
	if err := c.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// PullRelease returns target or newest release.
func (c *GSEController) PullRelease(ctx context.Context, req *pb.PullReleaseReq) (*pb.PullReleaseResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PullReleaseResp)

	defer func() {
		var cost int64
		if len(req.ReleaseId) != 0 {
			cost = c.collector.StatRequest("PullRelease(target)", response.Code, rtime, time.Now())
		} else {
			cost = c.collector.StatRequest("PullRelease(newest)", response.Code, rtime, time.Now())
		}

		if response.Release != nil {
			copied := *response.Release
			copied.Strategies = common.EmptyStr()
			logger.V(2).Infof("%s[%s]| output[%dms][code:%+v message:%+v release:%+v]",
				method, req.Seq, cost, response.Code, response.Message, copied)
		} else {
			logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
		}
	}()

	action := releaseaction.NewPullAction(ctx, c.viper, c.dataMgrCli, c.strategyHandler, req, response)
	if err := c.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// Reload reloads target release or multi release.
func (c *GSEController) Reload(ctx context.Context, req *pb.ReloadReq) (*pb.ReloadResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.ReloadResp)

	defer func() {
		cost := c.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewReloadAction(ctx, c.viper, c.dataMgrCli, c.tunnelServerCli, req, response)
	if err := c.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// Healthz returns server health informations.
func (c *GSEController) Healthz(ctx context.Context, req *pb.HealthzReq) (*pb.HealthzResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.HealthzResp)

	defer func() {
		cost := c.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := healthzaction.NewAction(ctx, c.viper, req, response)
	if err := c.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}
