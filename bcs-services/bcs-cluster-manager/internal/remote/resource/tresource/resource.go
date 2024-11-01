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

// NOCC:tosa/comment_ratio(none)

// Package tresource xxx
package tresource

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ResourceClient global resource client
var ResourceClient *ResManClient

// SetResourceClient set global resource client
func SetResourceClient(opts *Options, disc *discovery.ModuleDiscovery) {
	ResourceClient = &ResManClient{
		opts: opts,
		disc: disc,
	}
}

// GetResourceManagerClient get resource client
func GetResourceManagerClient() *ResManClient {
	return ResourceClient
}

// ResManClient rm client
type ResManClient struct {
	opts *Options
	disc *discovery.ModuleDiscovery
}

// getResourceManagerClient get rm client by discovery
func (rm *ResManClient) getResourceManagerClient() (ResourceManagerClient, func(), error) {
	if rm == nil {
		return nil, nil, ErrNotInited
	}

	if rm.disc == nil {
		return nil, nil, fmt.Errorf("resourceManager module not enable dsicovery")
	}

	// random server
	nodeServer, err := rm.disc.GetRandomServiceNode()
	if err != nil {
		return nil, nil, err
	}
	endpoints := utils.GetServerEndpointsFromRegistryNode(nodeServer)

	blog.Infof("ResManClient get node[%s] from disc", nodeServer.Address)
	conf := &Config{
		Hosts:     endpoints,
		TLSConfig: rm.opts.TLSConfig,
	}
	cli, closeCon := NewResourceManager(conf)

	return cli, closeCon, nil
}

// ApplyInstances apply cvm instances form resource pool to generate orderID
func (rm *ResManClient) ApplyInstances(ctx context.Context, instanceCount int,
	paras *resource.ApplyInstanceReq) (*resource.ApplyInstanceResp, error) {
	if rm == nil {
		return nil, ErrNotInited
	}
	traceID := utils.GetTraceIDFromContext(ctx)

	var (
		desireDevices *ConsumeDeviceReq
		err           error
	)

	// apply cvm or IDC nodes
	switch paras.NodeType {
	case resource.CVM:
		desireDevices, err = buildCVMConsumeDeviceDesireReq(uint32(instanceCount), paras)
	case resource.IDC:
		desireDevices, err = buildIDCConsumeDeviceDesireReq(uint32(instanceCount), paras)
	}
	if err != nil {
		blog.Errorf("ApplyInstances[%s] buildConsumeDeviceDesireReq failed: %v", traceID, err)
		return nil, err
	}

	var (
		orderID string
	)

	// loop check consumer order
	err = retry.Do(func() error {
		cli, closeCon, errGet := rm.getResourceManagerClient()
		if errGet != nil {
			blog.Errorf("ApplyInstances[%s] GetResourceManagerClient failed: %v", traceID, errGet)
			return errGet
		}
		defer func() {
			if closeCon != nil {
				closeCon()
			}
		}()

		start := time.Now()
		// consume devices
		resp, errCon := cli.ConsumeDevice(ctx, desireDevices)
		if errCon != nil {
			metrics.ReportLibRequestMetric("resource", "ConsumeDevice", "grpc", metrics.LibCallStatusErr, start)
			blog.Errorf("ApplyInstances[%s] ConsumeDevice failed: %v", traceID, errCon)
			return errCon
		}
		metrics.ReportLibRequestMetric("resource", "ConsumeDevice", "grpc", metrics.LibCallStatusOK, start)

		if !*resp.Result {
			retErr := fmt.Errorf("ApplyInstances[%s] ConsumeDevice failed: %v", traceID, *resp.Message)
			return retErr
		}

		orderID = *resp.Data.ID
		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(5*time.Second))
	if err != nil {
		blog.Errorf("ApplyInstances[%s] ConsumeDevice failed: %v", traceID, err)
		return nil, err
	}

	blog.Infof("ApplyInstances[%s] ConsumeDevice[%+v] orderID[%s] success", traceID, paras, orderID)

	return &resource.ApplyInstanceResp{
		OrderID: orderID,
	}, nil
}

// DestroyInstances destroy instance and return instance
func (rm *ResManClient) DestroyInstances(ctx context.Context, paras *resource.DestroyInstanceReq) (
	*resource.DestroyInstanceResp, error) {
	if rm == nil {
		return nil, ErrNotInited
	}

	traceID := utils.GetTraceIDFromContext(ctx)
	var (
		resp *ReturnDeviceResp
	)

	// return devices to resourcePool
	err := retry.Do(func() error {
		cli, closeCon, err := rm.getResourceManagerClient()
		if err != nil {
			blog.Errorf("DestroyInstances[%s] GetResourceManagerClient failed: %v", traceID, err)
			return err
		}
		defer func() {
			if closeCon != nil {
				closeCon()
			}
		}()

		blog.Infof("DestroyInstances[%s] returnDevices[%s:%s:%s] devices[%+v]", traceID, paras.PoolID,
			paras.SystemID, paras.Operator, paras.InstanceIDs)

		start := time.Now()
		// return devices
		resp, err = cli.ReturnDevice(context.Background(), &ReturnDeviceReq{
			DeviceConsumerID: &paras.PoolID,
			Devices:          paras.InstanceIDs,
			Operator:         &paras.Operator,
			ExtraSystemID:    &paras.SystemID,
		})
		if err != nil {
			metrics.ReportLibRequestMetric("resource", "ReturnDevice", "grpc", metrics.LibCallStatusErr, start)
			blog.Errorf("DestroyInstances[%s] ReturnDevice failed: %v", traceID, err)
			return err
		}
		metrics.ReportLibRequestMetric("resource", "ReturnDevice", "grpc", metrics.LibCallStatusOK, start)
		if !*resp.Result {
			return fmt.Errorf("DestroyInstances[%s] call resourceManager interface ReturnDevice failed: %v",
				traceID, *resp.Message)
		}

		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(5*time.Second))
	if err != nil {
		blog.Errorf("DestroyInstances[%s] ReturnDevice failed: %v", traceID, err)
		return nil, err
	}

	blog.Infof("DestroyInstances[%s] ReturnDevice[%s] successfully, orders[%+v]", traceID,
		paras.InstanceIDs, *resp.Data.ID)

	return &resource.DestroyInstanceResp{
		OrderID: *resp.Data.ID,
	}, nil
}

// CheckOrderStatus check order status
func (rm *ResManClient) CheckOrderStatus(ctx context.Context, orderID string) (*resource.OrderInstanceList, error) {
	if rm == nil {
		return nil, ErrNotInited
	}
	traceID := utils.GetTraceIDFromContext(ctx)

	timeOutCtx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	var (
		record *DeviceRecord
	)

	// loop check order state
	err := loop.LoopDoFunc(timeOutCtx, func() error {
		cli, closeCon, err := rm.getResourceManagerClient()
		if err != nil {
			blog.Errorf("CheckInstanceOrderStatus[%s] GetResourceManagerClient[%s] failed: %v", traceID, orderID, err)
			return nil
		}
		defer func() {
			if closeCon != nil {
				closeCon()
			}
		}()

		start := time.Now()
		// get device record
		resp, err := cli.GetDeviceRecord(context.Background(), &GetDeviceRecordReq{
			DeviceRecordID: &orderID,
		})
		if err != nil {
			metrics.ReportLibRequestMetric("resource", "GetDeviceRecord", "grpc", metrics.LibCallStatusErr, start)
			blog.Errorf("CheckInstanceOrderStatus[%s] call resource interface GetDeviceRecord[%s] failed: %v",
				traceID, orderID, err)
			return nil
		}
		metrics.ReportLibRequestMetric("resource", "GetDeviceRecord", "grpc", metrics.LibCallStatusOK, start)
		if resp == nil || !*resp.Result {
			blog.Errorf("CheckInstanceOrderStatus[%s] GetDeviceRecord[%s] failed: %v", traceID, orderID, err)
			return nil
		}

		// order status check
		switch *resp.Data.Status {
		case OrderFinished.String():
			blog.Infof("CheckInstanceOrderStatus[%s] orderID[%s] orderState[%s] successful",
				traceID, orderID, *resp.Data.Status)
			record = resp.Data
			return loop.EndLoop
		case OrderFailed.String():
			blog.Errorf("CheckInstanceOrderStatus[%s] orderID[%s] failed: %v", traceID, orderID, err)
			retErr := fmt.Errorf("DeviceRecorderID[%s] failed: %v", orderID, *resp.Data.Message)
			return retErr
		case OrderRequested.String():
			blog.Infof("CheckInstanceOrderStatus[%s] orderID[%s] orderState: %s",
				traceID, orderID, *resp.Data.Status)
		default:
			blog.Errorf("CheckInstanceOrderStatus[%s] orderID[%s] notSupportStatus: %v",
				traceID, orderID, *resp.Data.Status)
		}

		return nil
	}, loop.LoopInterval(time.Second*20))
	if err != nil {
		blog.Errorf("CheckInstanceOrderStatus[%s] orderID[%s] failed: %v", traceID, orderID, err)
		return nil, err
	}

	// get device instanceIDs & instanceIPs
	var (
		instanceIPs, instanceIDs, deviceIDs = make([]string, 0), make([]string, 0), make([]string, 0)
	)
	for _, device := range record.DeviceDetails {
		if device == nil {
			continue
		}
		instanceIPs = append(instanceIPs, *device.Info.InnerIP)
		instanceIDs = append(instanceIDs, *device.Info.Instance)
		deviceIDs = append(deviceIDs, *device.Id)
	}

	// return order instance list
	return &resource.OrderInstanceList{
		InstanceIDs: instanceIDs,
		InstanceIPs: instanceIPs,
		ExtraIDs:    deviceIDs,
	}, nil
}

// CheckInstanceStatus inner system not implement
func (rm *ResManClient) CheckInstanceStatus(ctx context.Context, instanceIDs []string) (*resource.OrderInstanceList,
	error) {
	return nil, ErrNotImplement
}

// GetInstanceTypesV2 get instance types
func (rm *ResManClient) GetInstanceTypesV2(ctx context.Context, region string, spec resource.InstanceSpec) ( // nolint
	[]resource.InstanceType, error) {
	if rm == nil {
		return nil, ErrNotInited
	}

	traceID := utils.GetTraceIDFromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	var (
		pools []*DevicePool
		err   error
	)

	pools, err = rm.listDevicePools(ctx, spec.Provider, region, "")
	if err != nil {
		blog.Errorf("GetInstanceTypesV2[%s] failed: %v", traceID, err)
		return nil, err
	}
	blog.Infof("GetInstanceTypesV2[%s] successful", traceID)

	// trans device pool to instances type
	targetTypes := make([]resource.InstanceType, 0)
	for _, pool := range pools {
		blog.Infof("GetInstanceTypesV2 pool[%v] detail: %v", pool.GetId(), pool)

		// filter cpu & mem
		if spec.Cpu != 0 && spec.Cpu != pool.GetBaseConfig().GetCpu() {
			continue
		}
		if spec.Mem != 0 && spec.Mem != pool.GetBaseConfig().GetMem() {
			continue
		}
		// labels
		labels := pool.GetLabels()
		// bizID
		blog.Infof("GetInstanceTypesV2 pool[%v] business[%v:%v]", pool.GetId(),
			pool.GetBaseConfig().GetBusinessID(), spec.BizID)
		if spec.BizID != "" && pool.GetBaseConfig().GetBusinessID() > 0 {
			dstBiz, _ := strconv.Atoi(spec.BizID)
			if int64(dstBiz) != pool.GetBaseConfig().GetBusinessID() {
				continue
			}
		}

		// resourceType: online && offline
		if spec.ResourceType != "" && spec.ResourceType != labels[ResourceType.String()] {
			continue
		}

		// available quota
		poolUsage := getDevicePoolUsage(pool)

		// instanceType sell status
		status := common.InstanceSell
		if poolUsage.OversoldAvailable <= 0 {
			status = common.InstanceSoldOut
		}

		// target instance types
		targetTypes = append(targetTypes, resource.InstanceType{
			NodeType:       *pool.GetBaseConfig().InstanceType,
			TypeName:       labels[InstanceSpecs.String()],
			NodeFamily:     "",
			Cpu:            *pool.GetBaseConfig().Cpu,
			Memory:         *pool.GetBaseConfig().Mem,
			Gpu:            *pool.GetBaseConfig().Gpu,
			Status:         status,
			UnitPrice:      0,
			Zones:          []string{pool.GetBaseConfig().GetZone().GetZone()},
			Provider:       *pool.Provider,
			ResourcePoolID: pool.GetId(),
			SystemDisk: func() *resource.DataDisk {
				if pool.GetBaseConfig().GetSystemDisk() != nil {
					return &resource.DataDisk{
						DiskType: pool.GetBaseConfig().GetSystemDisk().GetType(),
						DiskSize: pool.GetBaseConfig().GetSystemDisk().GetSize(),
					}
				}

				return nil
			}(),
			DataDisks: func() []*resource.DataDisk {
				if len(pool.GetBaseConfig().GetDataDisks()) > 0 {
					disks := make([]*resource.DataDisk, 0)
					for i := range pool.GetBaseConfig().GetDataDisks() {
						disks = append(disks, &resource.DataDisk{
							DiskType: pool.GetBaseConfig().GetDataDisks()[i].GetType(),
							DiskSize: pool.GetBaseConfig().GetDataDisks()[i].GetSize(),
						})
					}
					return disks
				}

				return nil
			}(),
			OversoldAvailable: poolUsage.OversoldAvailable,
			Region:            pool.GetBaseConfig().GetZone().GetRegion(),
		})
	}

	return targetTypes, nil
}

// GetInstanceTypes get region instance types
func (rm *ResManClient) GetInstanceTypes(ctx context.Context, region string, spec resource.InstanceSpec) ( // nolint
	[]resource.InstanceType, error) {
	if rm == nil {
		return nil, ErrNotInited
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60) // nolint
	defer cancel()

	return rm.GetInstanceTypesV2(ctx, region, spec)
}

// CreateResourcePool create resource pool for resource manager
func (rm *ResManClient) CreateResourcePool(ctx context.Context, info resource.ResourcePoolInfo) (string, error) {
	if rm == nil {
		return "", ErrNotInited
	}
	traceID := utils.GetTraceIDFromContext(ctx)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var (
		poolID string
		err    error
	)

	// create node pool
	err = retry.Do(func() error {
		cli, closeCon, errGet := rm.getResourceManagerClient()
		if errGet != nil {
			blog.Errorf("CreateResourcePool[%s] GetResourceManagerClient failed: %v", traceID, errGet)
			return errGet
		}
		defer func() {
			if closeCon != nil {
				closeCon()
			}
		}()

		start := time.Now()
		// CreateDeviceConsumer device consumer
		resp, errCreate := cli.CreateDeviceConsumer(ctx, &CreateDeviceConsumerReq{
			Name:                 &info.Name,
			Provider:             &info.Provider,
			ClusterID:            &info.ClusterID,
			AssociatedDevicePool: info.RelativeDevicePool,
			Labels: func() map[string]string {
				if len(info.PoolID) == 0 {
					return nil
				}
				return resource.BuildResourcePoolLabels(info.PoolID[0])
			}(),
			Operator: &info.Operator,
		})
		if errCreate != nil {
			metrics.ReportLibRequestMetric("resource", "CreateDeviceConsumer", "grpc", metrics.LibCallStatusErr, start)
			blog.Errorf("CreateResourcePool[%s] CreateDeviceConsumer failed: %v", traceID, errCreate)
			return errCreate
		}
		metrics.ReportLibRequestMetric("resource", "CreateDeviceConsumer", "grpc", metrics.LibCallStatusOK, start)
		if *resp.Code != 0 || !*resp.Result {
			blog.Errorf("CreateResourcePool[%s] CreateDeviceConsumer failed: %v", traceID, *resp.Message)
			return errors.New(*resp.Message)
		}
		poolID = *resp.Data.Id

		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second*5))
	if err != nil {
		blog.Errorf("CreateResourcePool[%s] failed: %v", traceID, err)
		return "", err
	}

	blog.Infof("CreateResourcePool[%s] successful[%s]", traceID, poolID)
	return poolID, nil
}

// DeleteResourcePool delete resource pool for resource manager
func (rm *ResManClient) DeleteResourcePool(ctx context.Context, poolID string) error {
	if rm == nil {
		return ErrNotInited
	}
	traceID := utils.GetTraceIDFromContext(ctx)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var err = retry.Do(func() error {
		cli, closeCon, errGet := rm.getResourceManagerClient()
		if errGet != nil {
			blog.Errorf("DeleteResourcePool[%s] GetResourceManagerClient failed: %v", traceID, errGet)
			return errGet
		}
		defer func() {
			if closeCon != nil {
				closeCon()
			}
		}()

		start := time.Now()
		// DeleteDeviceConsumer delete consumer
		resp, errDelete := cli.DeleteDeviceConsumer(ctx, &DeleteDeviceConsumerReq{
			DeviceConsumerID: &poolID,
		})
		if errDelete != nil {
			metrics.ReportLibRequestMetric("resource", "DeleteDeviceConsumer", "grpc", metrics.LibCallStatusErr, start)
			blog.Errorf("DeleteResourcePool[%s] DeleteDeviceConsumer failed: %v", traceID, errDelete)
			return errDelete
		}
		metrics.ReportLibRequestMetric("resource", "DeleteDeviceConsumer", "grpc", metrics.LibCallStatusOK, start)
		if *resp.Code != 0 || !*resp.Result {
			blog.Errorf("DeleteResourcePool[%s] DeleteDeviceConsumer failed: %v", traceID, *resp.Message)
			return errors.New(*resp.Message)
		}

		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second*3))
	if err != nil {
		blog.Errorf("DeleteResourcePool[%s] failed: %v", traceID, err)
		return err
	}

	blog.Infof("DeleteResourcePool[%s] successful[%s]", traceID, poolID)
	return nil
}

func (rm *ResManClient) listDevices(ctx context.Context, provider string) ([]*Device, error) {
	if rm == nil {
		return nil, ErrNotInited
	}

	traceID := utils.GetTraceIDFromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	var (
		devices []*Device
		err     error
	)

	err = retry.Do(func() error {
		cli, closeCon, errGet := rm.getResourceManagerClient()
		if errGet != nil {
			blog.Errorf("listDevices[%s] GetResourceManagerClient failed: %v", traceID, errGet)
			return errGet
		}
		defer func() {
			if closeCon != nil {
				closeCon()
			}
		}()

		var (
			// limit for all device
			limit int64 = 20000
		)

		req := &ListDevicesReq{
			Limit:  &limit,
			Status: []int64{int64(3)},
		}
		if len(provider) > 0 {
			req.Provider = append(req.Provider, provider)
		}

		start := time.Now()
		// list devices
		resp, errList := cli.ListDevices(ctx, req)
		if errList != nil {
			metrics.ReportLibRequestMetric("resource", "ListDevices", "grpc", metrics.LibCallStatusErr, start)
			blog.Errorf("listDevices[%s] ListDevices failed: %v", traceID, errList)
			return errList
		}
		metrics.ReportLibRequestMetric("resource", "ListDevices", "grpc", metrics.LibCallStatusOK, start)
		if *resp.Code != 0 || !*resp.Result {
			blog.Errorf("listDevices[%s] ListDevices failed: %v", traceID, resp.Message)
			return errors.New(*resp.Message)
		}
		devices = resp.Data

		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second*3))
	if err != nil {
		blog.Errorf("listDevices[%s] failed: %v", traceID, err)
		return nil, err
	}

	blog.Infof("listDevices[%s] successful", traceID)

	return devices, nil
}

// GetDevicesInfoMap get map devices
func (rm *ResManClient) GetDevicesInfoMap(ctx context.Context, provider string, isId bool) (
	map[string]resource.DeviceInfo, error) {
	if rm == nil {
		return nil, ErrNotInited
	}
	traceID := utils.GetTraceIDFromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	devices, err := rm.listDevices(ctx, provider)
	if err != nil {
		blog.Errorf("GetDeviceInfoIdMap[%s] listDevices failed: %v", traceID, err)
		return nil, err
	}

	devicesInfo := make(map[string]resource.DeviceInfo, 0)
	for _, device := range devices {
		if isId {
			if *device.Id != "" && *device.Info.InnerIP != "" {
				v, ok := devicesInfo[*device.Id]
				if !ok {
					devicesInfo[*device.Id] = resource.DeviceInfo{
						Provider:           *device.Provider,
						DeviceID:           *device.Id,
						InnerIP:            *device.Info.InnerIP,
						Instance:           *device.Info.Instance,
						InstanceType:       *device.Info.InstanceType,
						Cpu:                *device.Info.Cpu,
						Mem:                *device.Info.Mem,
						Gpu:                *device.Info.Gpu,
						Vpc:                *device.Info.Vpc,
						Region:             *device.Info.Region,
						Zone:               device.Info.GetZone().GetZone(),
						LastConsumerId:     *device.LastConsumerID,
						LastRecordId:       *device.LastRecordID,
						LastReturnRecordId: *device.LastReturnRecordID,
						DevicePoolID:       *device.DevicePoolID,
					}

					continue
				}

				blog.Errorf("%v 和 %v deviceId 重复", v.DeviceID, *device.Id)
			}
		} else {
			if *device.Id != "" && *device.Info.InnerIP != "" {
				v, ok := devicesInfo[*device.Info.InnerIP]
				if !ok {
					devicesInfo[*device.Info.InnerIP] = resource.DeviceInfo{
						Provider:           *device.Provider,
						DeviceID:           *device.Id,
						InnerIP:            *device.Info.InnerIP,
						Instance:           *device.Info.Instance,
						InstanceType:       *device.Info.InstanceType,
						Cpu:                *device.Info.Cpu,
						Mem:                *device.Info.Mem,
						Gpu:                *device.Info.Gpu,
						Vpc:                *device.Info.Vpc,
						Region:             *device.Info.Region,
						Zone:               device.Info.GetZone().GetZone(),
						LastConsumerId:     *device.LastConsumerID,
						LastRecordId:       *device.LastRecordID,
						LastReturnRecordId: *device.LastReturnRecordID,
						DevicePoolID:       *device.DevicePoolID,
					}

					continue
				}

				blog.Errorf("%v 和 %v deviceIp 重复", v.DeviceID, *device.Id)
			}
		}
	}

	return devicesInfo, nil
}

func (rm *ResManClient) listDevicePools(ctx context.Context, provider, region, instanceType string) ([]*DevicePool, error) {
	if rm == nil {
		return nil, ErrNotInited
	}

	traceID := utils.GetTraceIDFromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	var (
		pools []*DevicePool
		err   error
	)

	err = retry.Do(func() error {
		cli, closeCon, errGet := rm.getResourceManagerClient()
		if errGet != nil {
			blog.Errorf("listDevicePools[%s] GetResourceManagerClient failed: %v", traceID, errGet)
			return errGet
		}
		defer func() {
			if closeCon != nil {
				closeCon()
			}
		}()

		var (
			// limit for all devicePool
			limit  int64 = 10000
			onSale       = true
		)

		req := &ListDevicePoolReq{
			Limit:  &limit,
			Onsale: &onSale,
		}
		if len(provider) > 0 {
			req.Provider = append(req.Provider, provider)
		}
		if len(region) > 0 {
			req.Region = &region
		}

		start := time.Now()
		// list device pool
		resp, errList := cli.ListDevicePool(ctx, req)
		if errList != nil {
			metrics.ReportLibRequestMetric("resource", "ListDevicePool", "grpc", metrics.LibCallStatusErr, start)
			blog.Errorf("listDevicePools[%s] ListDevicePool failed: %v", traceID, errList)
			return errList
		}
		metrics.ReportLibRequestMetric("resource", "ListDevicePool", "grpc", metrics.LibCallStatusOK, start)
		if *resp.Code != 0 || !*resp.Result {
			blog.Errorf("listDevicePools[%s] ListDevicePool failed: %v", traceID, resp.Message)
			return errors.New(*resp.Message)
		}
		pools = resp.Data

		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second*3))
	if err != nil {
		blog.Errorf("listDevicePools[%s] failed: %v", traceID, err)
		return nil, err
	}
	blog.Infof("listDevicePools[%s] successful", traceID)

	filterPools := make([]*DevicePool, 0)
	// filter snapshot pools
	for i := range pools {
		if strings.Contains(pools[i].GetName(), "snapshot") {
			continue
		}

		filterPools = append(filterPools, pools[i])
	}

	if instanceType == "" {
		return filterPools, nil
	}

	insTypesPools := make([]*DevicePool, 0)
	for _, pool := range filterPools {
		if pool.GetBaseConfig().GetInstanceType() == instanceType {
			insTypesPools = append(insTypesPools, pool)
		}
	}

	return insTypesPools, nil
}

// GetRegionInstanceTypesFromPools get region instanceTypes机型 & zones可用区
func (rm *ResManClient) GetRegionInstanceTypesFromPools(ctx context.Context, provider string) (
	map[string][]string, error) {
	if rm == nil {
		return nil, ErrNotInited
	}

	pools, err := rm.listDevicePools(ctx, provider, "", "")
	if err != nil {
		return nil, err
	}

	var (
		regionInsTypes = make(map[string][]string, 0)
	)

	for i := range pools {
		region := pools[i].GetBaseConfig().GetZone().GetRegion()
		insType := pools[i].GetBaseConfig().GetInstanceType()

		v, ok := regionInsTypes[region]
		if !ok {
			if regionInsTypes[region] == nil {
				regionInsTypes[region] = make([]string, 0)
			}
			regionInsTypes[region] = append(regionInsTypes[region], insType)
			continue
		}

		if !utils.StringInSlice(insType, v) {
			regionInsTypes[region] = append(regionInsTypes[region], insType)
		}
	}

	return regionInsTypes, nil
}

func getDevicePoolUsage(pool *DevicePool) resource.PoolUsage {
	var (
		err error

		poolUsage resource.PoolUsage
		oversold  float64 = 1
	)
	v, ok := pool.GetLabels()[userQuota]
	if ok {
		total, _ := utils.StringToInt(v)
		poolUsage.Total = int32(total)
	}
	v, ok = pool.GetLabels()[usedQuota]
	if ok {
		used, _ := utils.StringToInt(v)
		poolUsage.Used = int32(used)
	}
	v, ok = pool.GetLabels()[availableQuota]
	if ok {
		available, _ := utils.StringToInt(v)
		poolUsage.Available = int32(available)
	}

	// oversold ratio
	v, ok = pool.GetLabels()[OverSold]
	if ok {
		oversold, err = strconv.ParseFloat(v, 64)
		if err != nil || oversold <= 1 {
			oversold = 1
		}
	}

	poolUsage.OversoldTotal = int32(math.Floor(float64(poolUsage.Total) * oversold))
	poolUsage.OversoldAvailable = int32(math.Floor(float64(poolUsage.Total)*oversold)) - int32(poolUsage.Used)

	return poolUsage
}

// ListRegionZonePools 获取可用区维度的资源池信息 & 可用区列表
func (rm *ResManClient) ListRegionZonePools(ctx context.Context, provider string, region, insType string) (
	map[string]*resource.DevicePoolInfo, []string, error) {
	if rm == nil {
		return nil, nil, ErrNotInited
	}

	pools, err := rm.listDevicePools(ctx, provider, region, insType)
	if err != nil {
		return nil, nil, err
	}

	var (
		zonePool = make(map[string]*resource.DevicePoolInfo, 0)
		zones    = make([]string, 0)
	)

	for _, pool := range pools {
		poolUsage := getDevicePoolUsage(pool)

		if len(pool.GetBaseConfig().Zone.GetZone()) == 0 {
			continue
		}

		zonePool[pool.GetBaseConfig().Zone.GetZone()] = &resource.DevicePoolInfo{
			PoolId:       *pool.Id,
			PoolName:     *pool.Name,
			Region:       pool.GetBaseConfig().GetZone().GetRegion(),
			Zone:         pool.GetBaseConfig().GetZone().GetZone(),
			InstanceType: pool.GetBaseConfig().GetInstanceType(),
			Total:        poolUsage.Total,
			Used:         poolUsage.Used,
			Available:    poolUsage.Available,
			Status:       pool.GetStatus(),

			OversoldTotal:     poolUsage.OversoldTotal,
			OversoldAvailable: poolUsage.OversoldAvailable,
		}

		zones = append(zones, pool.GetBaseConfig().Zone.GetZone())
	}

	sort.Sort(sort.StringSlice(zones))

	return zonePool, zones, nil
}

// ListAvailableInsufficientPools 获取可用资源不足的节点池
func (rm *ResManClient) ListAvailableInsufficientPools(ctx context.Context, provider string,
	region, insType string, ratio resource.UsageRatio) ([]*resource.DevicePoolInfo, error) {
	if rm == nil {
		return nil, ErrNotInited
	}

	pools, err := rm.listDevicePools(ctx, provider, region, insType)
	if err != nil {
		return nil, err
	}

	var (
		filterPools = make([]*resource.DevicePoolInfo, 0)
	)

	for _, pool := range pools {
		poolUsage := getDevicePoolUsage(pool)

		if len(pool.GetBaseConfig().Zone.GetZone()) == 0 {
			continue
		}

		if ratio.QuotaRatio != nil && ((float64(poolUsage.Available)/float64(poolUsage.Total))*100 >
			float64(*ratio.QuotaRatio)) {
			continue
		}

		if ratio.QuotaCount != nil && (poolUsage.Available > int32(*ratio.QuotaCount)) {
			continue
		}

		filterPools = append(filterPools, &resource.DevicePoolInfo{
			PoolId:            *pool.Id,
			PoolName:          *pool.Name,
			Region:            pool.GetBaseConfig().GetZone().GetRegion(),
			Zone:              pool.GetBaseConfig().GetZone().GetZone(),
			InstanceType:      pool.GetBaseConfig().GetInstanceType(),
			Total:             poolUsage.Total,
			Available:         poolUsage.Available,
			Status:            pool.GetStatus(),
			OversoldTotal:     poolUsage.OversoldTotal,
			OversoldAvailable: poolUsage.OversoldAvailable,
		})
	}

	return filterPools, nil
}

// GetDeviceConsumer get device consumer
func (rm *ResManClient) GetDeviceConsumer(ctx context.Context, consumerId string) (*DeviceConsumer, error) {
	if rm == nil {
		return nil, ErrNotInited
	}
	traceID := utils.GetTraceIDFromContext(ctx)

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cli, closeCon, err := rm.getResourceManagerClient()
	if err != nil {
		blog.Errorf("GetDeviceConsumer[%s][%s] GetResourceManagerClient failed: %v", traceID, consumerId, err)
		return nil, err
	}
	defer func() {
		if closeCon != nil {
			closeCon()
		}
	}()

	start := time.Now()
	// GetDeviceConsumer get consumer info
	resp, err := cli.GetDeviceConsumer(ctx, &GetDeviceConsumerReq{DeviceConsumerID: &consumerId})
	if err != nil {
		metrics.ReportLibRequestMetric("resource", "GetDeviceConsumer", "grpc", metrics.LibCallStatusErr, start)
		blog.Errorf("GetDeviceConsumer[%s][%s] GetDeviceConsumer failed: %v", traceID, consumerId, err)
		return nil, err
	}
	metrics.ReportLibRequestMetric("resource", "GetDeviceConsumer", "grpc", metrics.LibCallStatusOK, start)
	if *resp.Code != 0 || !*resp.Result {
		blog.Errorf("GetDeviceConsumer[%s][%s] GetDeviceConsumer failed: %v", traceID, consumerId, *resp.Message)
		return nil, errors.New(*resp.Message)
	}

	blog.Infof("GetDeviceConsumer[%s][%s] GetDeviceConsumer[%s] success", traceID, consumerId)
	return resp.GetData(), nil
}

// GetDeviceInfoByDeviceID get device detailed info by deviceID
func (rm *ResManClient) GetDeviceInfoByDeviceID(ctx context.Context, deviceID string) (*resource.DeviceInfo, error) {
	if rm == nil {
		return nil, ErrNotInited
	}
	traceID := utils.GetTraceIDFromContext(ctx)

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cli, closeCon, err := rm.getResourceManagerClient()
	if err != nil {
		blog.Errorf("GetDeviceInfoByDeviceID[%s][%s] GetResourceManagerClient failed: %v", traceID, deviceID, err)
		return nil, err
	}
	defer func() {
		if closeCon != nil {
			closeCon()
		}
	}()

	start := time.Now()
	// GetDevice get device detailed info
	resp, err := cli.GetDevice(ctx, &GetDeviceReq{
		DeviceID: &deviceID,
	})
	if err != nil {
		metrics.ReportLibRequestMetric("resource", "GetDevice", "grpc", metrics.LibCallStatusErr, start)
		blog.Errorf("GetDeviceInfoByDeviceID[%s][%s] GetDevice failed: %v", traceID, deviceID, err)
		return nil, err
	}
	metrics.ReportLibRequestMetric("resource", "GetDevice", "grpc", metrics.LibCallStatusOK, start)
	if *resp.Code != 0 || !*resp.Result {
		blog.Errorf("GetDeviceInfoByDeviceID[%s][%s] GetDevice failed: %v", traceID, deviceID, *resp.Message)
		return nil, errors.New(*resp.Message)
	}

	blog.Infof("GetDeviceInfoByDeviceID[%s][%s] GetDevice[%s] success", traceID, deviceID)
	return &resource.DeviceInfo{
		DeviceID:     deviceID,
		Provider:     resp.Data.GetProvider(),
		Labels:       resp.Data.GetLabels(),
		Annotations:  resp.Data.GetAnnotations(),
		Status:       resp.Data.GetStatus(),
		DevicePoolID: resp.Data.GetDevicePoolID(),
		Instance:     resp.Data.Info.GetInstance(),
		InnerIP:      resp.Data.Info.GetInnerIP(),
		InstanceType: resp.Data.Info.GetInstanceType(),
		Cpu:          resp.Data.Info.GetCpu(),
		Mem:          resp.Data.Info.GetMem(),
		Gpu:          resp.Data.Info.GetGpu(),
		Vpc:          resp.Data.Info.GetVpc(),
		Region:       resp.Data.Info.GetRegion(),
	}, nil
}

// getResourceAvailableZones resource available zones
func getResourceAvailableZones(req *resource.ApplyInstanceReq) ([]resource.SubnetZone, error) {
	if req == nil {
		return nil, fmt.Errorf("getResourceAvailableSubnetZones request nil")
	}
	zones := req.ZoneList

	// get available zones
	availableZones := make([]resource.SubnetZone, 0)
	for i := 0; i < len(zones); i++ {
		availableZones = append(availableZones, resource.SubnetZone{
			Zone: zones[i],
		})
	}

	// availableZones empty
	if len(availableZones) == 0 {
		availableZones = append(availableZones, resource.SubnetZone{
			Subnet: "",
			Zone:   "",
		})
	}

	return availableZones, nil
}

// buildIDCConsumeDeviceDesireReq build resource request
func buildIDCConsumeDeviceDesireReq(desiredNodes uint32, req *resource.ApplyInstanceReq) (*ConsumeDeviceReq, error) {
	// get available zones
	availableZones, err := getResourceAvailableZones(req)
	if err != nil {
		return nil, err
	}

	desires := make([]*ConsumeDesire, 0)
	for _, sz := range availableZones {
		desires = append(desires, &ConsumeDesire{
			InstanceType: &req.InstanceType,
			Cpu:          &req.CPU,
			Mem:          &req.Memory,
			Gpu:          &req.Gpu,
			Vpc:          &req.VpcID,
			Zone: &DeviceZone{
				Region: &sz.Zone,
				// IDC instance zone cmdb region
				Zone:   &sz.Zone,
				Subnet: &sz.Subnet,
			},
			Labels: req.Selector,
		})
	}

	// build consume device request
	return &ConsumeDeviceReq{
		DeviceConsumerID: &req.PoolID,
		Num:              &desiredNodes,
		Desire:           desires,
		Operator:         &req.Operator,
	}, nil
}

// buildCVMConsumeDeviceDesireReq build resource request for qcloud instance
func buildCVMConsumeDeviceDesireReq(desiredNodes uint32, req *resource.ApplyInstanceReq) (*ConsumeDeviceReq, error) {
	// get available zones
	availableZones, err := getResourceAvailableZones(req)
	if err != nil {
		return nil, err
	}

	var zones = make([]string, 0)
	for i := range availableZones {
		zones = append(zones, availableZones[i].Zone)
	}
	zoneStr := strings.Join(zones, ",")
	subnetStr := ""

	desires := make([]*ConsumeDesire, 0)

	desires = append(desires, &ConsumeDesire{
		InstanceType: &req.InstanceType,
		SystemDisk: &DeviceDisk{
			Type: &req.SystemDisk.DiskType,
			Size: &req.SystemDisk.DiskSize,
		},
		DataDisks: func() []*DeviceDisk {
			dataDisks := make([]*DeviceDisk, 0)
			for _, disk := range req.DataDisks {
				dataDisks = append(dataDisks, &DeviceDisk{
					Type: &disk.DiskType,
					Size: &disk.DiskSize,
				})
			}

			return dataDisks
		}(),
		Security: &DeviceSecurity{
			Group: func() *string {
				var (
					securityIDs = req.SecurityGroupIds
				)
				security := strings.Join(securityIDs, ",")

				return &security
			}(),
		},
		Image: &DeviceImage{
			Image: &req.Image.ImageID,
			Name:  &req.Image.ImageName,
		},
		Vpc: &req.VpcID,
		Zone: &DeviceZone{
			Region: &req.Region,
			Zone:   &zoneStr,
			Subnet: &subnetStr,
		},
	})

	// build consume device request
	return &ConsumeDeviceReq{
		DeviceConsumerID: &req.PoolID,
		Num:              &desiredNodes,
		Desire:           desires,
		Operator:         &req.Operator,
	}, nil
}
