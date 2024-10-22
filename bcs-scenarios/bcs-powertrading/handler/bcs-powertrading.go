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

// Package handler xxx
package handler

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/golang/protobuf/ptypes"
	"github.com/panjf2000/ants/v2"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/bkcc"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/clustermgr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/resourcemgr"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/scenes/data"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/storage"
	powertrading "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/proto"
)

// PowerTradingHandler handler struct
type PowerTradingHandler struct {
	cmCli       clustermgr.Client
	rmCli       resourcemgr.Client
	dataService data.Service
	bkccCli     bkcc.Client
	storage     storage.Storage
}

// New handler
func New(cmCli clustermgr.Client, rmCli resourcemgr.Client,
	bkccCli bkcc.Client, storageCli storage.Storage) *PowerTradingHandler {
	return &PowerTradingHandler{
		cmCli:       cmCli,
		rmCli:       rmCli,
		dataService: nil,
		bkccCli:     bkccCli,
		storage:     storageCli,
	}
}

// SetClients set clients
func (p *PowerTradingHandler) SetClients(cmCli clustermgr.Client, rmCli resourcemgr.Client) {
	p.cmCli = cmCli
	p.rmCli = rmCli
}

var (
	// InternalErr internal err code
	InternalErr = uint32(500)
	// Success success code
	Success = uint32(0)
)

// ProxyClusterManagerNodeCordon node cordon
func (p *PowerTradingHandler) ProxyClusterManagerNodeCordon(ctx context.Context,
	req *powertrading.ProxyClusterManagerNodeCordonReq, rsp *powertrading.ProxyClusterManagerNodeCordonResp) error {
	blog.Infof("receive ProxyClusterManagerNodeCordon request, ips:%s, businessID:%s", req.InnerIPs, req.BusinessID)
	err := p.cmCli.BatchCordonNodeWithoutCluster(ctx, req.InnerIPs, false)
	if err != nil {
		blog.Errorf("ProxyClusterManagerNodeCordon request failed: %s", err.Error())
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		return nil
	}
	rsp.Code = &Success
	msg := "cordon success"
	rsp.Message = msg
	return nil
}

// ProxyClusterManagerNodeUnCordon node uncordon
func (p *PowerTradingHandler) ProxyClusterManagerNodeUnCordon(ctx context.Context,
	req *powertrading.ProxyClusterManagerNodeCordonReq, rsp *powertrading.ProxyClusterManagerNodeCordonResp) error {
	blog.Infof("receive ProxyClusterManagerNodeUnCordon request, ips:%s, businessID:%s", req.InnerIPs, req.BusinessID)
	err := p.cmCli.BatchCordonNodeWithoutCluster(ctx, req.InnerIPs, true)
	if err != nil {
		blog.Errorf("ProxyClusterManagerNodeCordon request failed: %s", err.Error())
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		return nil
	}
	rsp.Code = &Success
	msg := "uncordon success"
	rsp.Message = msg
	return nil
}

// ProxyClusterManagerNodeDrain node drain
func (p *PowerTradingHandler) ProxyClusterManagerNodeDrain(ctx context.Context,
	req *powertrading.ProxyClusterManagerNodeDrainReq, rsp *powertrading.ProxyClusterManagerNodeDrainResp) error {
	blog.Infof("receive ProxyClusterManagerNodeDrain request, ips:%s, businessID:%d", req.InnerIPs, req.BusinessID)
	err := p.cmCli.BatchDrainNodeWithoutCluster(ctx, req.InnerIPs)
	if err != nil {
		blog.Errorf("ProxyClusterManagerNodeDrain request failed: %s", err.Error())
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		return nil
	}
	rsp.Code = &Success
	return nil
}

// ProxyResourceManagerCreateDeviceRecord create device record
func (p *PowerTradingHandler) ProxyResourceManagerCreateDeviceRecord(ctx context.Context, // nolint
	req *powertrading.ProxyResourceManagerCreateDeviceRecordReq,
	rsp *powertrading.ProxyResourceManagerCreateDeviceRecordResp) error {
	blog.Infof("receive ProxyResourceManagerCreateDeviceRecord request, devices:%s, businessID:%d",
		req.DeviceIDs, req.Deadline)
	if req.Deadline == "" {
		req.Deadline = time.Now().Add(10 * time.Minute).Format(time.RFC3339)
	}
	deviceIds := make([]string, 0)
	if req.DeviceIDs != nil && len(deviceIds) != 0 {
		deviceIds = append(deviceIds, req.DeviceIDs...)
	} else if len(req.Ips) != 0 {
		deviceRsp, err := p.rmCli.ListDeviceByIps(ctx, int64(len(req.Ips)), req.Ips)
		if err != nil {
			blog.Errorf("list device by ip request failed: %s", err.Error())
			rsp.Code = &InternalErr
			msg := err.Error()
			rsp.Message = msg
			return nil
		}
		if len(deviceRsp) == 0 {
			blog.Errorf("list device by ip request failed, rsp is empty")
			rsp.Code = &InternalErr
			msg := "list device by ip request failed, rsp is empty"
			rsp.Message = msg
			return nil
		}
		for _, device := range deviceRsp {
			deviceIds = append(deviceIds, *device.Id)
		}
	}
	record, err := p.rmCli.CreateDeviceRecord(ctx, deviceIds, req.Deadline)
	if err != nil {
		blog.Errorf("ProxyResourceManagerCreateDeviceRecord request failed: %s", err.Error())
		rsp.Code = &InternalErr
		msg := err.Error()
		rsp.Message = msg
		// rsp.Data = record
		return nil
	}
	result, err := ptypes.MarshalAny(record)
	if err != nil {
		return errors.Wrapf(err, "marshal response data failed")
	}
	rsp.Data = result
	return nil
}

//
// func (p *PowerTradingHandler) ProxyResourceManagerListDevices(ctx context.Context,
//	req *powertrading.ListDevicesReq,
//	rsp *powertrading.ListDevicesResp) error {
//	blog.Infof("receive ProxyResourceManagerListDevices request, devices:%s, devices:%v, ips:%v",
//		req.DeviceIDs, req.DeviceIDs, req.Ips)
//	deviceIds := make([]string, 0)
//	if req.DeviceIDs != nil && len(deviceIds) != 0 {
//		rmRsp, err := p.rmCli.ListDeviceByAssetIds()
//	} else if req.Ips != nil && len(req.Ips) != 0 {
//		deviceRsp, err := p.rmCli.ListDeviceByIps(ctx, int64(len(req.Ips)), req.Ips)
//		if err != nil {
//			blog.Errorf("list device by ip request failed: %s", err.Error())
//			rsp.Code = &InternalErr
//			msg := err.Error()
//			rsp.Message = &msg
//			return nil
//		}
//		if deviceRsp == nil || len(deviceRsp) == 0 {
//			blog.Errorf("list device by ip request failed, rsp is empty")
//			rsp.Code = &InternalErr
//			msg := "list device by ip request failed, rsp is empty"
//			rsp.Message = &msg
//			return nil
//		}
//		for _, device := range deviceRsp {
//			deviceIds = append(deviceIds, *device.Id)
//		}
//	}
//	record, err := p.rmCli.CreateDeviceRecord(ctx, req.DeviceIDs, req.)
//	if err != nil {
//		blog.Errorf("ProxyResourceManagerCreateDeviceRecord request failed: %s", err.Error())
//		rsp.Code = &InternalErr
//		msg := err.Error()
//		rsp.Message = &msg
//		// rsp.Data = record
//		return nil
//	}
//	result, err := ptypes.MarshalAny(record)
//	if err != nil {
//		return errors.Wrapf(err, "marshal response data failed")
//	}
//	rsp.Data = result
//	return nil
//}

//
// func (p *PowerTradingHandler) ListDevicePoolOperationData(ctx context.Context,
//	req *powertrading.ListDevicePoolOperationDataReq,
//	rsp *powertrading.ListDevicePoolOperationDataResp) error {
//	blog.Infof("receive ListDevicePoolOperationData request, businessID:%d, pool:%s",
//		req.BusinessID, req.Pool)
//	operationData, err := p.dataService.ListDevicePoolOperationData(ctx)
//	if err != nil {
//		blog.Errorf("ListDevicePoolOperationData request failed: %s", err.Error())
//		rsp.Code = &InternalErr
//		msg := err.Error()
//		rsp.Message = &msg
//		// rsp.Data = record
//		return nil
//	}
//	result := make([]*ptypes.DynamicAny, 0)
//	for _, item := range operationData {
//		itemStruct, structErr := ptypes.MarshalAny(item)
//		if structErr != nil {
//			return errors.Wrapf(structErr, "marshal response data failed")
//		}
//		result = append(result, itemStruct)
//	}
//	rsp.Data = result
//	return nil
//}

// MachineTest machine test
func (p *PowerTradingHandler) MachineTest(ctx context.Context,
	req *powertrading.MachineTestReq,
	rsp *powertrading.MachineTestRsp) error {
	blog.Infof("receive MachineTest request, businessID:%d, source:%s",
		req.BusinessID, req.Source)
	task := storage.InitNewCheckMachineTask()
	task.BusinessID = req.BusinessID
	task.IPList = req.IpList
	task.Source = req.Source
	createErr := p.storage.CreateMachineTestTask(ctx, task, &storage.CreateOptions{OverWriteIfExist: false})
	if createErr != nil {
		blog.Errorf("CreateMachineTestTask error:%s", createErr.Error())
		rsp.Code = &InternalErr
		rsp.Message = fmt.Sprintf("CreateMachineTestTask error:%s", createErr.Error())
		return nil
	}
	rsp.Code = &Success
	rsp.Message = "create task success"
	rsp.Data = &powertrading.MachineTestData{
		TaskId: task.TaskID,
		Status: task.Status,
	}
	return nil
}

// MachineClean machine clean
func (p *PowerTradingHandler) MachineClean(ctx context.Context,
	req *powertrading.MachineTestReq,
	rsp *powertrading.MachineTestRsp) error {
	blog.Infof("receive MachineClean request, businessID:%d, source:%s",
		req.BusinessID, req.Source)
	hosts, err := p.bkccCli.ListHostByCC(ctx, req.IpList, req.BusinessID)
	if err != nil {
		rsp.Code = &InternalErr
		rsp.Message = fmt.Sprintf("check device belong to business failed: %s", err.Error())
		return nil
	}
	hostMap := make(map[string]*bkcc.CCHostInfo)
	for i := range hosts {
		hostMap[hosts[i].BKHostInnerIP] = &hosts[i]
	}
	notExist := make([]string, 0)
	for i := range req.IpList {
		if _, ok := hostMap[req.IpList[i]]; !ok {
			notExist = append(notExist, req.IpList[i])
		}
	}
	if len(notExist) != 0 {
		rsp.Code = &InternalErr
		rsp.Message = fmt.Sprintf("device(%s) not exist in business %s",
			strings.Join(notExist, ","), req.BusinessID)
		return nil
	}
	task := storage.InitNewCleanMachineTask()
	task.BusinessID = req.BusinessID
	task.IPList = req.IpList
	task.Source = req.Source
	createErr := p.storage.CreateMachineTestTask(ctx, task, &storage.CreateOptions{OverWriteIfExist: false})
	if createErr != nil {
		blog.Errorf("Create Task error:%s", createErr.Error())
		rsp.Code = &InternalErr
		rsp.Message = fmt.Sprintf("Create Task error:%s", createErr.Error())
		return nil
	}
	rsp.Code = &Success
	rsp.Message = "create task success"
	rsp.Data = &powertrading.MachineTestData{
		TaskId: task.TaskID,
		Status: task.Status,
	}
	return nil
}

// GetMachineTask get task
func (p *PowerTradingHandler) GetMachineTask(ctx context.Context,
	req *powertrading.GetMachineTestTaskReq,
	rsp *powertrading.MachineTestRsp) error {
	blog.Infof("receive GetMachineTestTask request, taskID:%d, source:%s",
		req.TaskID)
	task, err := p.storage.GetTask(ctx, req.TaskID, &storage.GetOptions{ErrIfNotExist: true})
	if err != nil {
		blog.Errorf("get task %s error:%s", req.TaskID, err.Error())
		rsp.Code = &InternalErr
		rsp.Message = fmt.Sprintf("get task %s error:%s", req.TaskID, err.Error())
		return nil
	}
	rsp.Code = &Success
	rsp.Message = "get task success"
	rsp.Data = &powertrading.MachineTestData{
		TaskId:     task.TaskID,
		Status:     task.Status,
		Ips:        task.IPList,
		TaskType:   task.Type,
		BusinessID: task.BusinessID,
		Detail:     make(map[string]*powertrading.MachineTaskDetail),
	}
	for taskName := range task.Detail {
		rsp.Data.Detail[taskName] = &powertrading.MachineTaskDetail{
			Status: task.Detail[taskName].Status,
		}
		if task.Detail[taskName].BksOpsTaskID != "" {
			rsp.Data.Detail[taskName].BksOpsTaskID = task.Detail[taskName].BksOpsTaskID
		}
		if task.Detail[taskName].JobID != "" {
			rsp.Data.Detail[taskName].JobID = task.Detail[taskName].JobID
		}
		rsp.Data.Detail[taskName].Messages = make([]*powertrading.MachineTestMessage, 0)
		for ip := range task.Detail[taskName].DetailList {
			rsp.Data.Detail[taskName].Messages = append(rsp.Data.Detail[taskName].Messages,
				task.Detail[taskName].DetailList[ip])
		}
	}
	rsp.Data.TaskSummary = &powertrading.TaskSummary{
		Failure:   make([]*powertrading.SummaryMessage, 0),
		Success:   make([]*powertrading.SummaryMessage, 0),
		NeedClean: make([]*powertrading.SummaryMessage, 0),
	}
	for status := range task.Summary {
		for process := range task.Summary[status] {
			switch status {
			case storage.MachineCheckSuccess:
				if len(task.Summary[status][process]) != 0 {
					rsp.Data.TaskSummary.Success = append(rsp.Data.TaskSummary.Success, &powertrading.SummaryMessage{
						CheckProcessName: process,
						Ips:              task.Summary[status][process],
					})
				}
			case storage.MachineCheckFailure:
				if len(task.Summary[status][process]) != 0 {
					rsp.Data.TaskSummary.Failure = append(rsp.Data.TaskSummary.Failure, &powertrading.SummaryMessage{
						CheckProcessName: process,
						Ips:              task.Summary[status][process],
					})
				}
			case storage.MachineNeedClean:
				if len(task.Summary[status][process]) != 0 {
					rsp.Data.TaskSummary.NeedClean = append(rsp.Data.TaskSummary.NeedClean, &powertrading.SummaryMessage{
						CheckProcessName: process,
						Ips:              task.Summary[status][process],
					})
				}
			}
		}
	}
	return nil
}

// ProxyClusterManagerNodeDetail get node detail
func (p *PowerTradingHandler) ProxyClusterManagerNodeDetail(ctx context.Context,
	req *powertrading.ProxyClusterManagerNodeDetailReq,
	rsp *powertrading.ProxyClusterManagerNodeDetailResp) error {
	blog.Infof("receive ProxyClusterManagerNodeDetail request, ips:%s", req.Ips)
	rsp.Data = make([]*powertrading.NodeDetail, 0)
	for _, ip := range req.Ips {
		node, err := p.cmCli.GetNodeDetail(ctx, ip)
		if err != nil {
			blog.Errorf("GetNodeDetail request failed: %s", err.Error())
			rsp.Data = append(rsp.Data, &powertrading.NodeDetail{
				Ip:     ip,
				Status: "REQUEST-FAILURE",
			})
			continue
		}
		if node == nil {
			device, getDeviceErr := p.rmCli.ListDeviceByIps(ctx, 1, []string{ip})
			if getDeviceErr != nil {
				blog.Errorf("ListDeviceByIps failed, ip:%s, error:%s", ip, getDeviceErr.Error())
				rsp.Data = append(rsp.Data, &powertrading.NodeDetail{
					Ip:     ip,
					Status: "REQUEST-FAILURE",
				})
			} else if len(device) == 0 {
				rsp.Data = append(rsp.Data, &powertrading.NodeDetail{
					Ip:     ip,
					Status: "NOT-FOUND",
				})
			} else {
				rsp.Data = append(rsp.Data, &powertrading.NodeDetail{
					Ip:     ip,
					Status: "IMPORTED",
				})
			}
			continue
		}
		result, err := ptypes.MarshalAny(node)
		if err != nil {
			return errors.Wrapf(err, "marshal response data failed")
		}
		rsp.Data = append(rsp.Data, &powertrading.NodeDetail{
			Ip:     ip,
			Status: node.Status,
			Detail: result,
		})
	}
	rsp.Code = &Success
	rsp.Message = "request success"
	return nil
}

// EditDeviceInfo edit device info
func (p *PowerTradingHandler) EditDeviceInfo(ctx context.Context,
	req *powertrading.EditDevicesReq,
	rsp *powertrading.EditDevicesResp) error {
	blog.Infof("receive EditDeviceInfo request, ips:%s, labels:%v, annotations:%v, onlyEditDevice:%t, onlyEditNode:%t",
		req.Ips, req.Labels, req.Annotations, req.OnlyEditInfo, req.OnlyEditNode)
	wg := sync.WaitGroup{}
	pool, err := ants.NewPool(100)
	if err != nil {
		blog.Errorf("init new pool err:%s", err.Error())
		rsp.Code = &InternalErr
		msg := fmt.Sprintf("init new pool err:%v", err.Error())
		rsp.Message = msg
		return nil
	}
	defer pool.Release()
	rsp.Data = &powertrading.EditDevicesDetail{
		Success: make([]string, 0),
		Fail:    make([]string, 0),
	}
	for index := range req.Ips {
		ip := req.Ips[index]
		wg.Add(1)
		submitErr := pool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					blog.Errorf("panic in handle one edit device task, info: %v, stack:%s", r,
						string(debug.Stack()))
				}
				wg.Done()
			}()
			if req.OnlyEditNode {
				updateErr := p.cmCli.UpdateNodeWithoutCluster(ctx, ip, req.Labels, req.Annotations)
				if updateErr != nil {
					blog.Errorf("update node fail, ip:%s, err:%s", ip, updateErr.Error())
					rsp.Data.Fail = append(rsp.Data.Fail, ip)
					return
				}
				rsp.Data.Success = append(rsp.Data.Success, ip)
			} else {
				_, updateErr := p.rmCli.UpdateDevice(ctx, req.Labels, req.Annotations, ip)
				if updateErr != nil {
					blog.Errorf("update device fail, ip:%s, err:%s", ip, updateErr.Error())
					rsp.Data.Fail = append(rsp.Data.Fail, ip)
					return
				} else if !req.OnlyEditInfo {
					updateErr = p.cmCli.UpdateNodeWithoutCluster(ctx, ip, req.Labels, req.Annotations)
					if updateErr != nil {
						blog.Errorf("update node fail, ip:%s, err:%s", ip, updateErr.Error())
						rsp.Data.Fail = append(rsp.Data.Fail, ip)
						return
					}
					rsp.Data.Success = append(rsp.Data.Success, ip)
				} else {
					rsp.Data.Success = append(rsp.Data.Success, ip)
				}
			}

		})
		if submitErr != nil {
			blog.Errorf("submit task to ch pool err:%s", submitErr.Error())
			wg.Done()
		}
	}
	wg.Wait()
	if len(rsp.Data.Fail) != 0 {
		rsp.Code = &InternalErr
		rsp.Message = "execute failed"
		return nil
	}
	rsp.Code = &Success
	rsp.Message = "execute success"
	return nil
}
