/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetagent"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	ipAction "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/action"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/actionexecutor"
)

// AllocIP allocate ip
func (s *Server) AllocIP(ctx context.Context, req *pb.AllocIPReq) (*pb.AllocIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("AllocIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.AllocIPResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := s.metricCollector.StatRequest("AllocIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("AllocIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	allocAction := ipAction.NewAllocateAction(
		ctx, req, response,
		s.k8sClient, s.k8sIPClient, s.cloudNetClient, s.inspector, s.fixedIPWorkloads)
	actionexecutor.NewExecutor().Execute(allocAction)

	return response, nil
}

// ReleaseIP release ip
func (s *Server) ReleaseIP(ctx context.Context, req *pb.ReleaseIPReq) (*pb.ReleaseIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("ReleaseIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.ReleaseIPResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := s.metricCollector.StatRequest("ReleaseIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ReleaseIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	allocAction := ipAction.NewReleaseAction(
		ctx, req, response, s.cloudNetClient, s.k8sIPClient, s.inspector)
	actionexecutor.NewExecutor().Execute(allocAction)

	return response, nil
}
