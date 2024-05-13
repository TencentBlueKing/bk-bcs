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
	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/action"
	eniAction "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/action/eni"
	ipAction "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/action/ip"
	quotaAction "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/action/quota"
	subnetAction "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/action/subnet"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/metric"
)

// AddSubnet add subnet to cloud netservice
func (cn *CloudNetservice) AddSubnet(ctx context.Context, req *pb.AddSubnetReq) (*pb.AddSubnetResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("AddSubnet seq[%d] input[%+v]", req.Seq, req)
	response := &pb.AddSubnetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("AddSubnet", response.ErrCode, rtime, time.Now())
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
		cost := metric.DefaultCollector.StatRequest("DeleteSubnet", response.ErrCode, rtime, time.Now())
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
		cost := metric.DefaultCollector.StatRequest("ChangeSubnet", response.ErrCode, rtime, time.Now())
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
		cost := metric.DefaultCollector.StatRequest("ListSubnet", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ListSubnet seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	listAction := subnetAction.NewListAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(listAction)

	return response, nil
}

// GetAvailableSubnet get available subnet for certain region
func (cn *CloudNetservice) GetAvailableSubnet(ctx context.Context, req *pb.GetAvailableSubnetReq) (
	*pb.GetAvailableSubnetResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("GetAvailableSubnet seq[%d] input[%+v]", req.Seq, req)
	response := &pb.GetAvailableSubnetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("GetAvailableSubnet", response.ErrCode, rtime, time.Now())
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
		cost := metric.DefaultCollector.StatRequest("AllocateIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("AllocateIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	var allocateAction action.Action
	if req.IsFixed {
		allocateAction = ipAction.NewFixedAllocateAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	} else {
		allocateAction = ipAction.NewAllocateAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	}
	action.NewExecutor().Execute(allocateAction)

	return response, nil
}

// ReleaseIP release ip for certain pod
func (cn *CloudNetservice) ReleaseIP(ctx context.Context, req *pb.ReleaseIPReq) (*pb.ReleaseIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("ReleaseIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.ReleaseIPResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("ReleaseIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ReleaseIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	releaseAction := ipAction.NewReleaseAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	action.NewExecutor().Execute(releaseAction)

	return response, nil
}

// CleanFixedIP clean fixed ip
func (cn *CloudNetservice) CleanFixedIP(ctx context.Context, req *pb.CleanFixedIPReq) (*pb.CleanFixedIPResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("CleanFixedIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.CleanFixedIPResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("CleanFixedIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("CleanFixedIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do clean fixed ip action
	fixedCleanAction := ipAction.NewFixedCleanAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	action.NewExecutor().Execute(fixedCleanAction)
	return response, nil
}

// CleanEni clean eni
func (cn *CloudNetservice) CleanEni(ctx context.Context, req *pb.CleanEniReq) (*pb.CleanEniResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("CleanEni seq[%d] input[%+v]", req.Seq, req)
	response := &pb.CleanEniResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("CleanEni", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("CleanEni seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do clean eni action
	cleanEniAction := ipAction.NewCleanEniAction(ctx, req, response, cn.storeIf, cn.cloudIf)
	action.NewExecutor().Execute(cleanEniAction)
	return response, nil
}

// ListIP list ip objects from cloud netservice
func (cn *CloudNetservice) ListIP(ctx context.Context, req *pb.ListIPsReq) (*pb.ListIPsResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("ListIP seq[%d] input[%+v]", req.Seq, req)
	response := &pb.ListIPsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("ListIP", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ListIP seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do list action
	listAction := ipAction.NewListAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(listAction)
	return response, nil
}

// AllocateEni create eni record, netservice judges whether it can apply for a new eni based on this record
func (cn *CloudNetservice) AllocateEni(ctx context.Context, req *pb.AllocateEniReq) (*pb.AllocateEniResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("AllocateEni seq[%d] input[%+v]", req.Seq, req)
	response := &pb.AllocateEniResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("AllocateEni", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("AllocateEni seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do allocate action
	allocateEniAction := eniAction.NewAllocateAction(ctx, cn.cfg, req, response, cn.storeIf, cn.locker)
	action.NewExecutor().Execute(allocateEniAction)
	return response, nil
}

// ReleaseEni release eni record
func (cn *CloudNetservice) ReleaseEni(ctx context.Context, req *pb.ReleaseEniReq) (*pb.ReleaseEniResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("ReleaseEni seq[%d] input[%+v]", req.Seq, req)
	response := &pb.ReleaseEniResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("ReleaseEni", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ReleaseEni seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do release action
	releaseEniAction := eniAction.NewReleaseAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(releaseEniAction)

	return response, nil
}

// TransIPStatus trans ip status
func (cn *CloudNetservice) TransIPStatus(ctx context.Context, req *pb.TransIPStatusReq) (*pb.TransIPStatusResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("TransIPStatus seq[%d] input[%+v]", req.Seq, req)
	response := &pb.TransIPStatusResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("TransIPStatus", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("TransIPStatus seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do change status action
	transIPAction := ipAction.NewTransStatusAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(transIPAction)

	return response, nil
}

// GetQuota get ip quota
func (cn *CloudNetservice) GetQuota(ctx context.Context, req *pb.GetIPQuotaReq) (*pb.GetIPQuotaResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("GetQuota seq[%d] input[%+v]", req.Seq, req)
	response := &pb.GetIPQuotaResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("GetQuota", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("GetQuota seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do get action
	getAction := quotaAction.NewGetAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(getAction)

	return response, nil
}

// CreateQuota create ip quota
func (cn *CloudNetservice) CreateQuota(ctx context.Context, req *pb.CreateIPQuotaReq) (*pb.CreateIPQuotaResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("CreateQuota seq[%d] input[%+v]", req.Seq, req)
	response := &pb.CreateIPQuotaResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("CreateQuota", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("CreateQuota seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do add action
	createAction := quotaAction.NewAddAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(createAction)

	return response, nil
}

// UpdateQuota update ip quota
func (cn *CloudNetservice) UpdateQuota(ctx context.Context, req *pb.UpdateIPQuotaReq) (*pb.UpdateIPQuotaResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("UpdateQuota seq[%d] input[%+v]", req.Seq, req)
	response := &pb.UpdateIPQuotaResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("UpdateQuota", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("UpdateQuota seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do update action
	createAction := quotaAction.NewUpdateAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(createAction)

	return response, nil
}

// DeleteQuota delete ip quota
func (cn *CloudNetservice) DeleteQuota(ctx context.Context, req *pb.DeleteIPQuotaReq) (*pb.DeleteIPQuotaResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("DeleteQuota seq[%d] input[%+v]", req.Seq, req)
	response := &pb.DeleteIPQuotaResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("DeleteQuota", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("DeleteQuota seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do delete action
	createAction := quotaAction.NewDeleteAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(createAction)

	return response, nil
}

// ListQuota list ip quota
func (cn *CloudNetservice) ListQuota(ctx context.Context, req *pb.ListIPQuotaReq) (*pb.ListIPQuotaResp, error) {
	rtime := time.Now()
	blog.V(3).Infof("ListQuota seq[%d] input[%+v]", req.Seq, req)
	response := &pb.ListIPQuotaResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_ERROR_OK, ErrMsg: "OK"}

	defer func() {
		cost := metric.DefaultCollector.StatRequest("ListQuota", response.ErrCode, rtime, time.Now())
		blog.V(3).Infof("ListQuota seq[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()
	// do create action
	createAction := quotaAction.NewListAction(ctx, req, response, cn.storeIf)
	action.NewExecutor().Execute(createAction)

	return response, nil
}
