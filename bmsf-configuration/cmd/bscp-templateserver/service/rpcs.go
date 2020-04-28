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
