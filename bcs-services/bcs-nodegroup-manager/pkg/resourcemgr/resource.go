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

// Package resourcemgr xxx
package resourcemgr

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	grpc "github.com/asim/go-micro/plugins/client/grpc/v4"
	etcd "github.com/asim/go-micro/plugins/registry/etcd/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/metric"
	impl "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/resourcemgr/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

const (
	// success definition for service calls
	success = 0
	// LabelKeyDrainDelay 节点抽离时间
	LabelKeyDrainDelay = "nodeDrainDelay"
	// LabelKeyDeadline 节点抽离ddl
	LabelKeyDeadline = "nodeDeadline"
)

// ListOptions options for list resource pools
type ListOptions struct {
	PageSize int
}

// GetOptions options for list resource pools
type GetOptions struct {
	GetCacheIfEmpty bool
}

// ClientOptions for local resource-manager client
type ClientOptions struct {
	// Name for resource-manager registry
	Name string
	// Etcd endpoints information
	Etcd []string
	// EtcdConfig tls config for etcd
	EtcdConfig *tls.Config
	// ClientConfig tls config
	ClientConfig *tls.Config
	// Cache for store ResourcePool information
	Cache storage.Storage
}

// Client for resource-manager
type Client interface {
	ListTasksByConsumer(consumerID string, option *ListOptions) ([]*storage.ScaleDownTask, error)
	// GetTaskByID get task by record id
	GetTaskByID(recordID string, opt *GetOptions) (*storage.ScaleDownTask, error)
	// GetDeviceListByConsumer get device list by consumer
	GetDeviceListByConsumer(consumerID string, option *GetOptions) (*storage.DeviceGroup, error)
	// GetDeviceListByPoolID get device list by pool id
	GetDeviceListByPoolID(consumerID string, devicePoolID []string, option *GetOptions) (*storage.DeviceGroup, error)
	// FillDeviceRecordIp fill scale down ip
	FillDeviceRecordIp(recordID string, ipList []string) error
	// ListTasksByCond list device records by type and status
	ListTasksByCond(recordType, status []int64) ([]*storage.ScaleDownTask, error)
}

// New create resource-manager client instance
func New(opt *ClientOptions) Client {
	// init go-micro v2 client instance
	c := grpc.NewClient(
		client.Registry(etcd.NewRegistry(
			registry.Addrs(opt.Etcd...),
			registry.TLSConfig(opt.EtcdConfig)),
		),
		grpc.AuthTLS(opt.ClientConfig),
	)
	// create resource-manager go-micro client api
	return &innerClient{
		client: impl.NewResourceManagerService(opt.Name, c),
	}
}

// innerClient
type innerClient struct {
	client impl.ResourceManagerService
}

// GetDeviceListByConsumer get device list by nodegroup consumer
func (c *innerClient) GetDeviceListByConsumer(consumerID string, option *GetOptions) (*storage.DeviceGroup, error) {
	devicePoolList, err := c.GetConsumerAssociatedDevicePool(consumerID, option)
	blog.Infof("consumer %s associated device pool:%v", consumerID, devicePoolList)
	if err != nil {
		return nil, err
	}
	confirmDeviceList := make([]string, 0)
	for _, devicePool := range devicePoolList {
		info, err := c.GetDevicePoolInfo(devicePool, nil)
		if err != nil {
			return nil, err
		}
		allowConsumer := info.AllowedDeviceConsumer
		blog.Infof("device pool %s allow consumer: %v", devicePool, allowConsumer)
		if len(allowConsumer) == 0 || containItem(consumerID, allowConsumer) {
			confirmDeviceList = append(confirmDeviceList, devicePool)
			continue
		}
	}
	return c.GetDeviceListByPoolID(consumerID, confirmDeviceList, option)
}

// GetDeviceListByPoolID get device list by device pool id
func (c *innerClient) GetDeviceListByPoolID(consumerID string, devicePoolID []string,
	option *GetOptions) (*storage.DeviceGroup, error) {
	defaultLimit := int64(10000)
	req := &impl.ListDevicesReq{Pool: devicePoolID, Limit: &defaultLimit}
	startTime := time.Now()
	rsp, err := c.client.ListDevices(context.Background(), req)
	defer func() {
		metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "ListDevices", err, startTime)
	}()
	if err != nil {
		blog.Errorf("get device list from resource-manager failed, %s", err.Error())
		return nil, err
	}
	if rsp.GetCode() != success {
		blog.Errorf("get device list failed, resource-manager logic err: %s", rsp.GetMessage())
		err = fmt.Errorf("resource-manager logic failure, %s", rsp.GetMessage())
		return nil, err
	}
	if len(rsp.Data) == 0 {
		blog.Errorf("resource-manager response empty Resource from device pool list %s", devicePoolID)
		return nil, fmt.Errorf("empty resources response")
	}
	deviceGroup := convertToDevicePool(consumerID, rsp.Data)
	return deviceGroup, nil
}

// GetDevicePoolInfo get device pool info
func (c *innerClient) GetDevicePoolInfo(devicePool string, option *GetOptions) (*impl.DevicePool, error) {
	devicePoolReq := &impl.GetDevicePoolReq{DevicePoolID: &devicePool}
	startTime := time.Now()
	devicePoolRsp, err := c.client.GetDevicePool(context.Background(), devicePoolReq)
	defer func() {
		metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "GetDevicePool", err, startTime)
	}()
	if err != nil {
		blog.Errorf("get device pool %s from resource-manager failed, %s", devicePool, err.Error())
		return nil, err
	}
	if devicePoolRsp.GetCode() != success {
		blog.Errorf("get device pool %s failed, resource-manager logic err: %s", devicePool, devicePoolRsp.GetMessage())
		err = fmt.Errorf("resource-manager logic failure, %s", devicePoolRsp.GetMessage())
		return nil, err
	}
	if devicePoolRsp.Data == nil {
		err = fmt.Errorf("devicePool info of consumer(%s) is empty", devicePool)
		blog.Errorf(err.Error())
		return nil, err
	}
	return devicePoolRsp.Data, nil
}

// GetConsumerAssociatedDevicePool get associate device pool
func (c *innerClient) GetConsumerAssociatedDevicePool(consumerID string, option *GetOptions) ([]string, error) {
	consumerReq := &impl.GetDeviceConsumerReq{DeviceConsumerID: &consumerID}
	startTime := time.Now()
	consumerRsp, err := c.client.GetDeviceConsumer(context.Background(), consumerReq)
	defer func() {
		metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "GetDeviceConsumer", err, startTime)
	}()
	if err != nil {
		blog.Errorf("get device consumer from resource-manager failed, %s", err.Error())
		return nil, err
	}
	if consumerRsp.GetCode() != success {
		blog.Errorf("get device consumer failed, resource-manager logic err: %s", consumerRsp.GetMessage())
		err = fmt.Errorf("resource-manager logic failure, %s", consumerRsp.GetMessage())
		return nil, err
	}
	if len(consumerRsp.Data.AssociatedDevicePool) == 0 {
		err = fmt.Errorf("associatedDevicePool of consumer(%s) is empty", consumerID)
		blog.Errorf(err.Error())
		return nil, err
	}
	blog.Infof("consumer:%s, devicePool:%v", consumerID, consumerRsp.Data.AssociatedDevicePool)
	return consumerRsp.Data.AssociatedDevicePool, nil
}

// FillDeviceRecordIp fill ip to device record
func (c *innerClient) FillDeviceRecordIp(recordID string, ipList []string) error {
	_, err := c.GetDeviceRecord(recordID)
	if err != nil {
		return fmt.Errorf("get device record by id %s err:%s", recordID, err.Error())
	}
	updateReq := &impl.UpdateDeviceRecordReq{
		DeviceRecordID: &recordID,
		Devices:        &impl.ListString{Data: ipList},
	}
	startTime := time.Now()
	defer func() {
		metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "UpdateDeviceRecord", err, startTime)
	}()
	resp, err := c.client.UpdateDeviceRecord(context.Background(), updateReq)
	if err != nil {
		blog.Errorf("update device record by id %s failed, %s", recordID, err.Error())
		return err
	}
	if resp.GetCode() != success {
		blog.Errorf("update device records by id %s failed, resource-manager logic err: %s", recordID, resp.GetMessage())
		err = fmt.Errorf("resource-manager logic failure, %s", resp.GetMessage())
		return err
	}
	return nil
}

// GetDeviceRecord get device record
func (c *innerClient) GetDeviceRecord(recordID string) (*impl.DeviceRecord, error) {
	req := &impl.GetDeviceRecordReq{DeviceRecordID: &recordID}
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "GetDeviceRecord", err, startTime)
	}()
	resp, err := c.client.GetDeviceRecord(context.Background(), req)
	if err != nil {
		blog.Errorf("get device record by id %s from resource-manager failed, %s", recordID, err.Error())
		return nil, err
	}
	if resp.GetCode() != success {
		blog.Errorf("get device records failed, resource-manager logic err: %s", resp.GetMessage())
		err = fmt.Errorf("resource-manager logic failure, %s", resp.GetMessage())
		return nil, err
	}
	return resp.GetData(), nil
}

// ListTasksByConsumer list tasks by consumer id
func (c *innerClient) ListTasksByConsumer(consumerID string, opt *ListOptions) ([]*storage.ScaleDownTask, error) {
	devicePoolList, err := c.GetConsumerAssociatedDevicePool(consumerID, nil)
	if err != nil {
		return nil, err
	}
	localTasks := make([]*storage.ScaleDownTask, 0)
	for _, devicePoolID := range devicePoolList {
		req := &impl.ListDeviceRecordByDevicePoolReq{
			PoolID: &devicePoolID,
		}
		startTime := time.Now()
		resp, err := c.client.ListDeviceRecordByDevicePool(context.Background(), req)
		if err != nil {
			blog.Errorf("get device records from resource-manager failed, %s", err.Error())
			metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "ListDeviceRecordByPool", err, startTime)
			return nil, err
		}
		if resp.GetCode() != success {
			blog.Errorf("get device records failed, resource-manager logic err: %s", resp.GetMessage())
			err = fmt.Errorf("resource-manager logic failure, %s", resp.GetMessage())
			metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "ListDeviceRecordByPool", err, startTime)
			return nil, err
		}
		metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "ListDeviceRecordByPool", err, startTime)
		// convert details to local ResourcePool definition
		if len(resp.Data) == 0 {
			blog.Infof("resource-manager response empty device records from ResourcePool %s", devicePoolID)
		}
		for index := range resp.Data {
			task := convertTaskToLocal(resp.Data[index])
			if task == nil {
				continue
			}
			task.DevicePoolID = devicePoolID
			blog.Infof("task:%v", task)
			localTasks = append(localTasks, task)
		}
	}
	return localTasks, nil
}

// GetTaskByID get task by task id
func (c *innerClient) GetTaskByID(recordID string, opt *GetOptions) (*storage.ScaleDownTask, error) {
	req := &impl.GetDeviceRecordReq{DeviceRecordID: &recordID}
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "GetDeviceRecord", err, startTime)
	}()
	resp, err := c.client.GetDeviceRecord(context.Background(), req)
	if err != nil {
		blog.Errorf("get device record by id %s from resource-manager failed, %s", recordID, err.Error())
		return nil, err
	}
	if resp.GetCode() != success {
		blog.Errorf("get device records failed, resource-manager logic err: %s", resp.GetMessage())
		err = fmt.Errorf("resource-manager logic failure, %s", resp.GetMessage())
		return nil, err
	}
	return convertTaskToLocal(resp.Data), nil
}

// ListTasksByCond list device records by type and status
func (c *innerClient) ListTasksByCond(recordType, status []int64) ([]*storage.ScaleDownTask, error) {
	limit := int64(10000)
	req := &impl.ListDeviceRecordReq{
		Type:   recordType,
		Status: status,
		Limit:  &limit,
	}
	startTime := time.Now()
	resp, err := c.client.ListDeviceRecord(context.Background(), req)
	if err != nil {
		blog.Errorf("get device records from resource-manager failed, %s", err.Error())
		metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "ListDeviceRecord", err, startTime)
		return nil, err
	}
	if resp.GetCode() != success {
		blog.Errorf("get device records failed, resource-manager logic err: %s", resp.GetMessage())
		err = fmt.Errorf("resource-manager logic failure, %s", resp.GetMessage())
		metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "ListDeviceRecord", err, startTime)
		return nil, err
	}
	metric.ReportLibRequestMetric(metric.BkBcsResourceManager, "grpc", "ListDeviceRecord", err, startTime)
	// convert details to local ResourcePool definition
	if len(resp.Data) == 0 {
		blog.Infof("resource-manager response empty device records")
	}
	localTasks := make([]*storage.ScaleDownTask, 0)
	for index := range resp.Data {
		task := convertTaskToLocal(resp.Data[index])
		if task == nil {
			continue
		}
		blog.Infof("task:%v", task)
		localTasks = append(localTasks, task)
	}
	return localTasks, nil
}

func convertDeviceToLocal(r *impl.Device) (*storage.Resource, int64) {
	resource := &storage.Resource{
		ID:               r.GetId(),
		InnerIP:          r.GetInfo().GetInnerIP(),
		ResourceType:     r.GetInfo().GetInstanceType(),
		ResourceProvider: r.GetProvider(),
		UpdatedTime:      time.Unix(*r.UpdateTime, 0),
	}
	if r.Status != nil {
		resource.Phase = r.GetStatus()
		resource.DevicePool = r.GetDevicePoolID()
	} else {
		// feature protection(DeveloperJim): consider that unknown status is Consumed
		resource.Phase = storage.NodeConsumedState
	}
	return resource, *r.UpdateTime
}

func convertTaskToLocal(deviceRecord *impl.DeviceRecord) *storage.ScaleDownTask {
	deadline, _ := time.Parse(time.RFC3339, deviceRecord.GetConsumerLabels()[LabelKeyDeadline])
	task := &storage.ScaleDownTask{
		TaskID:     deviceRecord.GetId(),
		TotalNum:   int(deviceRecord.GetNum()),
		Status:     deviceRecord.GetStatus(),
		DrainDelay: deviceRecord.GetConsumerLabels()[LabelKeyDrainDelay],
		Deadline:   deadline,
	}
	if deviceRecord.GetProvider() == "self" && deviceRecord.GetExtraLabels()["creator"] == "powertrading" {
		devices := make([]string, 0)
		if len(deviceRecord.Devices) != 0 && (len(deviceRecord.Devices) != len(deviceRecord.GetDeviceDetails())) {
			blog.Infof("wait for device record %s update detail", deviceRecord.GetId())
			blog.Infof("deviceRecord: %v", deviceRecord)
			blog.Infof("deviceRecord: %v", deviceRecord.GetDeviceDetails())
			return nil
		}
		for _, deviceDetail := range deviceRecord.GetDeviceDetails() {
			devices = append(devices, *deviceDetail.Info.InnerIP)
		}
		task.DeviceList = devices
		if deviceRecord.GetExtraLabels()["deadline"] != "" {
			deadline, _ = time.Parse(time.RFC3339, deviceRecord.GetExtraLabels()["deadline"])
			task.Deadline = deadline
		}
	}
	return task
}

func convertToDevicePool(consumerID string, devices []*impl.Device) *storage.DeviceGroup {
	deviceGroup := &storage.DeviceGroup{ConsumerID: consumerID}
	var latest int64
	for _, r := range devices {
		resource, updated := convertDeviceToLocal(r)
		switch resource.Phase {
		case storage.NodeIdleState:
			deviceGroup.IdleNum++
		case storage.NodeInitState:
			deviceGroup.InitNum++
		case storage.NodeReturnState:
			deviceGroup.ReturnedNum++
		case storage.NodeConsumedState:
			deviceGroup.ConsumedNum++
		default:
			// unknown status set as Consumed
			deviceGroup.ConsumedNum++
		}
		if updated > latest {
			latest = updated
		}
		deviceGroup.Resources = append(deviceGroup.Resources, resource)
	}
	deviceGroup.UpdatedTime = time.Unix(latest, 0)
	return deviceGroup
}

func containItem(item string, itemGroup []string) bool {
	var contain bool
	for index := range itemGroup {
		if itemGroup[index] == item {
			contain = true
			break
		}
	}
	return contain
}
