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

	configsetaction "bk-bscp/cmd/bscp-bcs-controller/actions/configset"
	publishaction "bk-bscp/cmd/bscp-bcs-controller/actions/publish"
	releaseaction "bk-bscp/cmd/bscp-bcs-controller/actions/release"
	reloadaction "bk-bscp/cmd/bscp-bcs-controller/actions/reload"
	reportaction "bk-bscp/cmd/bscp-bcs-controller/actions/report"
	pb "bk-bscp/internal/protocol/bcs-controller"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/logger"
)

// PublishReleasePre checks target release for publishing.
func (c *BCSController) PublishReleasePre(ctx context.Context, req *pb.PublishReleasePreReq) (*pb.PublishReleasePreResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PublishReleasePre[%d]| input[%+v]", req.Seq, req)
	response := &pb.PublishReleasePreResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := c.collector.StatRequest("PublishReleasePre", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PublishReleasePre[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := publishaction.NewPublishPreAction(c.viper, c.dataMgrCli, req, response)
	c.executor.Execute(action)

	return response, nil
}

// PublishRelease publishs target release.
func (c *BCSController) PublishRelease(ctx context.Context, req *pb.PublishReleaseReq) (*pb.PublishReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PublishRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.PublishReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := c.collector.StatRequest("PublishRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PublishRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := publishaction.NewPublishAction(c.viper, c.dataMgrCli, c.publisher, c.pubTopic, req, response)
	c.executor.Execute(action)

	return response, nil
}

// RollbackRelease rollbacks target release.
func (c *BCSController) RollbackRelease(ctx context.Context, req *pb.RollbackReleaseReq) (*pb.RollbackReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("RollbackRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.RollbackReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := c.collector.StatRequest("RollbackRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("RollbackRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := publishaction.NewRollbackAction(c.viper, c.dataMgrCli, c.publisher, c.pubTopic, req, response)
	c.executor.Execute(action)

	return response, nil
}

// Report handles information reported by app instance.
func (c *BCSController) Report(ctx context.Context, req *pb.ReportReq) (*pb.ReportResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("Report[%d]| input[%+v]", req.Seq, req)
	response := &pb.ReportResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := c.collector.StatRequest("Report", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("Report[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := reportaction.NewReportAction(c.viper, c.dataMgrCli, req, response)
	c.executor.Execute(action)

	return response, nil
}

// PullRelease returns target or newest release.
func (c *BCSController) PullRelease(ctx context.Context, req *pb.PullReleaseReq) (*pb.PullReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PullRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.PullReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		var cost int64
		if len(req.Releaseid) != 0 {
			cost = c.collector.StatRequest("PullRelease(target)", response.ErrCode, rtime, time.Now())
		} else {
			cost = c.collector.StatRequest("PullRelease(newest)", response.ErrCode, rtime, time.Now())
		}
		logger.V(2).Infof("PullRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewPullAction(c.viper, c.dataMgrCli, c.strategyHandler, req, response)
	c.executor.Execute(action)

	return response, nil
}

// PullConfigSetList returns configset list.
func (c *BCSController) PullConfigSetList(ctx context.Context, req *pb.PullConfigSetListReq) (*pb.PullConfigSetListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PullConfigSetList[%d]| input[%+v]", req.Seq, req)
	response := &pb.PullConfigSetListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := c.collector.StatRequest("PullConfigSetList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PullConfigSetList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewListAction(c.viper, c.dataMgrCli, req, response)
	c.executor.Execute(action)

	return response, nil
}

// Reload reloads target release or multi release.
func (c *BCSController) Reload(ctx context.Context, req *pb.ReloadReq) (*pb.ReloadResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("Reload[%d]| input[%+v]", req.Seq, req)
	response := &pb.ReloadResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := c.collector.StatRequest("Reload", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("Reload[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := reloadaction.NewReloadAction(c.viper, c.dataMgrCli, c.publisher, c.pubTopic, req, response)
	c.executor.Execute(action)

	return response, nil
}
