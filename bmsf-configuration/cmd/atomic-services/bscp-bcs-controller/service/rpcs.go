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

	healthzaction "bk-bscp/cmd/atomic-services/bscp-bcs-controller/actions/healthz"
	releaseaction "bk-bscp/cmd/atomic-services/bscp-bcs-controller/actions/release"
	pb "bk-bscp/internal/protocol/bcs-controller"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// PublishReleasePre checks target release for publishing.
func (c *BCSController) PublishReleasePre(ctx context.Context,
	req *pb.PublishReleasePreReq) (*pb.PublishReleasePreResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PublishReleasePreResp)

	defer func() {
		cost := c.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewPublishPreAction(ctx, c.viper, c.dataMgrCli, req, response)
	c.executor.Execute(action)

	return response, nil
}

// PublishRelease publishs target release.
func (c *BCSController) PublishRelease(ctx context.Context,
	req *pb.PublishReleaseReq) (*pb.PublishReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.PublishReleaseResp)

	defer func() {
		cost := c.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewPublishAction(ctx, c.viper, c.dataMgrCli, c.publisher, c.pubTopic, req, response)
	c.executor.Execute(action)

	return response, nil
}

// RollbackRelease rollbacks target release.
func (c *BCSController) RollbackRelease(ctx context.Context,
	req *pb.RollbackReleaseReq) (*pb.RollbackReleaseResp, error) {

	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.RollbackReleaseResp)

	defer func() {
		cost := c.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewRollbackAction(ctx, c.viper, c.dataMgrCli, c.publisher, c.pubTopic, req, response)
	c.executor.Execute(action)

	return response, nil
}

// PullRelease returns target or newest release.
func (c *BCSController) PullRelease(ctx context.Context, req *pb.PullReleaseReq) (*pb.PullReleaseResp, error) {
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
		logger.V(2).Infof("PullRelease[%s]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewPullAction(ctx, c.viper, c.dataMgrCli, c.strategyHandler, req, response)
	c.executor.Execute(action)

	return response, nil
}

// Reload reloads target release or multi release.
func (c *BCSController) Reload(ctx context.Context, req *pb.ReloadReq) (*pb.ReloadResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.ReloadResp)

	defer func() {
		cost := c.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := releaseaction.NewReloadAction(ctx, c.viper, c.dataMgrCli, c.publisher, c.pubTopic, req, response)
	c.executor.Execute(action)

	return response, nil
}

// Healthz returns server health informations.
func (c *BCSController) Healthz(ctx context.Context, req *pb.HealthzReq) (*pb.HealthzResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.HealthzResp)

	defer func() {
		cost := c.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := healthzaction.NewAction(ctx, c.viper, req, response)
	c.executor.Execute(action)

	return response, nil
}
