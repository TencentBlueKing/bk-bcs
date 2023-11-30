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

package tresource

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"
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
func GetResourceManagerClient() resource.ManagerResource {
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

		// consume devices
		resp, errCon := cli.ConsumeDevice(ctx, desireDevices)
		if errCon != nil {
			blog.Errorf("ApplyInstances[%s] ConsumeDevice failed: %v", traceID, errCon)
			return errCon
		}

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

		resp, err = cli.ReturnDevice(context.Background(), &ReturnDeviceReq{
			DeviceConsumerID: &paras.PoolID,
			Devices:          paras.InstanceIDs,
			Operator:         &paras.Operator,
			ExtraSystemID:    &paras.SystemID,
		})
		if err != nil {
			blog.Errorf("DestroyInstances[%s] ReturnDevice failed: %v", traceID, err)
			return err
		}
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

		// get device record
		resp, err := cli.GetDeviceRecord(context.Background(), &GetDeviceRecordReq{
			DeviceRecordID: &orderID,
		})
		if err != nil {
			blog.Errorf("CheckInstanceOrderStatus[%s] call resource interface GetDeviceRecord[%s] failed: %v",
				traceID, orderID, err)
			return nil
		}
		if resp == nil || !*resp.Result {
			blog.Errorf("CheckInstanceOrderStatus[%s] GetDeviceRecord[%s] failed: %v", traceID, orderID, err)
			return nil
		}

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
	err = retry.Do(func() error {
		cli, closeCon, errGet := rm.getResourceManagerClient()
		if errGet != nil {
			blog.Errorf("GetInstanceTypesV2[%s] GetResourceManagerClient failed: %v", traceID, errGet)
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
			Region: &region,
		}
		if len(spec.Provider) > 0 {
			req.Provider = append(req.Provider, spec.Provider)
		}

		// list device pool
		resp, errList := cli.ListDevicePool(ctx, req)
		if errList != nil {
			blog.Errorf("GetInstanceTypesV2[%s] ListDevicePool failed: %v", traceID, errList)
			return errList
		}
		if *resp.Code != 0 || !*resp.Result {
			blog.Errorf("GetInstanceTypesV2[%s] ListDevicePool failed: %v", traceID, resp.Message)
			return errors.New(*resp.Message)
		}
		pools = resp.Data

		return nil
	}, retry.Attempts(3), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second*3))
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
		var quota uint64
		availableQuota, ok := labels[AvailableQuota.String()]
		if ok {
			quota, _ = strconv.ParseUint(availableQuota, 10, 64)
		}
		// instanceType sell status
		status := common.InstanceSell
		if quota == 0 {
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
			blog.Errorf("CreateResourcePool[%s] CreateDeviceConsumer failed: %v", traceID, errCreate)
			return errCreate
		}
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

		// DeleteDeviceConsumer delete consumer
		resp, errDelete := cli.DeleteDeviceConsumer(ctx, &DeleteDeviceConsumerReq{
			DeviceConsumerID: &poolID,
		})
		if errDelete != nil {
			blog.Errorf("DeleteResourcePool[%s] DeleteDeviceConsumer failed: %v", traceID, errDelete)
			return errDelete
		}
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

	// GetDevice get device detailed info
	resp, err := cli.GetDevice(ctx, &GetDeviceReq{
		DeviceID: &deviceID,
	})
	if err != nil {
		blog.Errorf("GetDeviceInfoByDeviceID[%s][%s] GetDevice failed: %v", traceID, deviceID, err)
		return nil, err
	}
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

	return &ConsumeDeviceReq{
		DeviceConsumerID: &req.PoolID,
		Num:              &desiredNodes,
		Desire:           desires,
		Operator:         &req.Operator,
	}, nil
}

// buildCVMConsumeDeviceDesireReq build resource request for qcloud instance
func buildCVMConsumeDeviceDesireReq(desiredNodes uint32, req *resource.ApplyInstanceReq) (*ConsumeDeviceReq, error) {
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
					security    = ""
					securityIDs = req.SecurityGroupIds
				)
				if len(securityIDs) > 0 {
					security = securityIDs[0]
				}

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

	return &ConsumeDeviceReq{
		DeviceConsumerID: &req.PoolID,
		Num:              &desiredNodes,
		Desire:           desires,
		Operator:         &req.Operator,
	}, nil
}
