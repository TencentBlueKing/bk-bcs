/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pb "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/action"
	ipAction "github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/action/ip"
	subnetAction "github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/action/subnet"
)

// AddSubnet add subnet to cloud netservice
func (cn *CloudNetservice) AddSubnet(ctx context.Context, req *pb.AddSubnetReq) (*pb.AddSubnetResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("AddSubnet seq[%d] input[%+v]", req.Seq, req)
	response := &pb.AddSubnetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("AddSubnet", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("AddSubnet seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	addAction := subnetAction.NewAddAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	action.NewExecutor().Execute(addAction)

	return response, nil
}

// DeleteSubnet delete subnet from cloud netservice
func (cn *CloudNetservice) DeleteSubnet(ctx context.Context, req *pb.DeleteSubnetReq) (*pb.DeleteSubnetResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("DeleteSubnet seq[%d] input[%+v]", req.Seq, req)
	response := &pb.DeleteSubnetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("DeleteSubnet", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("DeleteSubnet seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	deleteAction := subnetAction.NewDeleteAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(deleteAction)

	return response, nil
}

// ChangeSubnet change subnet state
func (cn *CloudNetservice) ChangeSubnet(ctx context.Context, req *pb.ChangeSubnetReq) (*pb.ChangeSubnetResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("ChangeSubnet seq[%d] input[%+v]", req.Seq, req)
	response := &pb.ChangeSubnetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("ChangeSubnet", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ChangeSubnet seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	changeAction := subnetAction.NewChangeAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(changeAction)

	return response, nil
}

// ListSubnet list subnet
func (cn *CloudNetservice) ListSubnet(ctx context.Context, req *pb.ListSubnetReq) (*pb.ListSubnetResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("ListSubnet seq[%d] input[%+v]", req.Seq, req)
	response := &pb.ListSubnetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("ListSubnet", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ListSubnet seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	listAction := subnetAction.NewListAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(listAction)

	return response, nil
}

// GetAvailableSubnet get available subnet for certain region
func (cn *CloudNetservice) GetAvailableSubnet(ctx context.Context, req *pb.GetAvailableSubnetReq) (*pb.GetAvailableSubnetResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("GetAvailableSubnet seq[%d] input[%+v]", req.Seq, req)
	response := &pb.GetAvailableSubnetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("GetAvailableSubnet", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("GetAvailableSubnet seq[%d] output[%dms][%+v]", req.Seq, cost, response)
	}()

	getAvailableAction := subnetAction.NewFindAvailableAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	action.NewExecutor().Execute(getAvailableAction)

	return response, nil
}

// AllocateIP allocate ip for certain pod
func (cn *CloudNetservice) AllocateIP(ctx context.Context, req *pb.AllocateIPReq) (*pb.AllocateIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("AllocateIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.AllocateIPResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("AllocateIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("AllocateIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	allocateAction := ipAction.NewAllocateAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	action.NewExecutor().Execute(allocateAction)

	return response, nil
}

// ReleaseIP release ip for certain pod
func (cn *CloudNetservice) ReleaseIP(ctx context.Context, req *pb.ReleaseIPReq) (*pb.ReleaseIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("ReleaseIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.ReleaseIPResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("ReleaseIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ReleaseIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	releaseAction := ipAction.NewReleaseAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(releaseAction)

	return response, nil
}

// AllocateFixedIP allocate fixed ip for certain pod
func (cn *CloudNetservice) AllocateFixedIP(ctx context.Context, req *pb.AllocateFixedIPReq) (*pb.AllocateFixedIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("AllocateFixedIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.AllocateFixedIPResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("AllocateFixedIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("AllocateFixedIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	fixedAllocateAction := ipAction.NewFixedAllocateAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	action.NewExecutor().Execute(fixedAllocateAction)

	return response, nil
}

// ReleaseFixedIP release fixed ip for certain pod
func (cn *CloudNetservice) ReleaseFixedIP(ctx context.Context, req *pb.ReleaseFixedIPReq) (*pb.ReleaseFixedIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("ReleaseFixedIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.ReleaseFixedIPResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("ReleaseFixedIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ReleaseFixedIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	fixedReleaseAction := ipAction.NewFixedReleaseAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(fixedReleaseAction)

	return response, nil
}

// CleanFixedIP clean fixed ip
func (cn *CloudNetservice) CleanFixedIP(ctx context.Context, req *pb.CleanFixedIPReq) (*pb.CleanFixedIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("CleanFixedIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.CleanFixedIPResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("CleanFixedIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("CleanFixedIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	fixedCleanAction := ipAction.NewFixedCleanAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	action.NewExecutor().Execute(fixedCleanAction)

	return response, nil
}

// CleanNode clean node ip
func (cn *CloudNetservice) CleanNode(ctx context.Context, req *pb.CleanNodeReq) (*pb.CleanNodeResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("CleanNode seq[%d] input[%+v]", req.Seq, req)
	response := &pb.CleanNodeResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := cn.metricCollector.StatRequest("CleanNode", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("CleanNode seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	fixedCleanNodeAction := ipAction.NewCleanNodeAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	action.NewExecutor().Execute(fixedCleanNodeAction)

	return response, nil
}

// ListIP list ip objects from cloud netservice
func (cn *CloudNetservice) ListIP(ctx context.Context, req *pb.ListIPsReq) (*pb.ListIPsResp, error) {
	return nil, nil
}
