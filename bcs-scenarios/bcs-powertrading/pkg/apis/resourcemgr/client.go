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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"

	resourcemanager "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/resourcemgr/resourcemanagerv4"
)

// Client interface
type Client interface {
	ListDevicePool(ctx context.Context, poolType []string) ([]*resourcemanager.DevicePool, error)
	CreateDeviceRecord(ctx context.Context, deviceId []string, deadline string) (*resourcemanager.DeviceRecord, error)
	ListDeviceByAssetIds(ctx context.Context, limit int64, assetList []string) ([]*resourcemanager.Device, error)
	ListDeviceByIps(ctx context.Context, limit int64, assetList []string) ([]*resourcemanager.Device, error)
	UpdateDevice(ctx context.Context, labels map[string]string, annotations map[string]string,
		ip string) (*resourcemanager.Device, error)
	ListDeviceByPool(ctx context.Context, limit int64, poolId []string) ([]*resourcemanager.Device, error)
}

type rmClient struct {
	client      resourcemanager.ResourceManagerService
	concurrency int
}

var (
	// PoolType self
	PoolType = "self"
	// RecordType self return task
	RecordType = "SELF_RETURN_TASK"
)

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
	// Cache storage.Storage
}

// NewClient new client
func NewClient(opt *ClientOptions, concurrency int) Client {
	// init go-micro v2 client instance
	c := grpc.NewClient(
		client.Registry(etcd.NewRegistry(
			registry.Addrs(opt.Etcd...),
			registry.TLSConfig(opt.EtcdConfig)),
		),
		grpc.AuthTLS(opt.ClientConfig),
	)
	// create resource-manager go-micro client api
	return &rmClient{
		client:      resourcemanager.NewResourceManagerService(opt.Name, c),
		concurrency: concurrency,
	}
}

// ListDevicePool list device pool
func (c *rmClient) ListDevicePool(ctx context.Context, poolType []string) ([]*resourcemanager.DevicePool, error) {
	limit := int64(5000)
	req := &resourcemanager.ListDevicePoolReq{
		Limit:    &limit,
		Provider: poolType,
	}
	rsp, err := c.client.ListDevicePool(ctx, req)
	if err != nil {
		blog.Errorf("list device pool error:%s", err.Error())
		return nil, err
	}
	if !*rsp.Result {
		blog.Errorf("list device pool error:%s", *rsp.Message)
		return nil, fmt.Errorf("list device pool error:%s", *rsp.Message)
	}
	return rsp.Data, nil
}

// CreateDeviceRecord create device record
func (c *rmClient) CreateDeviceRecord(ctx context.Context, deviceId []string,
	deadline string) (*resourcemanager.DeviceRecord, error) {
	req := &resourcemanager.CreateDeviceRecordReq{Data: &resourcemanager.DeviceRecord{
		Type:    &RecordType,
		Devices: deviceId,
		ExtraLabels: map[string]string{
			"creator":  "powertrading",
			"deadline": deadline,
		},
		Provider: &PoolType,
	}}
	rsp, err := c.client.CreateDeviceRecord(ctx, req)
	if err != nil {
		blog.Errorf("create device record error:%s", err.Error())
		return nil, err
	}
	if !*rsp.Result {
		blog.Errorf("create device record error:%s", *rsp.Message)
		return nil, fmt.Errorf("create device record error:%s", *rsp.Message)
	}
	return rsp.Data, nil
}

// ListDeviceRecordByPool list record by pool id
func (c *rmClient) ListDeviceRecordByPool(ctx context.Context, pool string) ([]*resourcemanager.DeviceRecord, error) {
	req := &resourcemanager.ListDeviceRecordReq{
		Pool: []string{pool},
	}
	rsp, err := c.client.ListDeviceRecord(ctx, req)
	if err != nil {
		blog.Errorf("list device record error:%s", err.Error())
		return nil, err
	}
	if !*rsp.Result {
		blog.Errorf("list device record error:%s", *rsp.Message)
		return nil, fmt.Errorf("list device record error:%s", *rsp.Message)
	}
	return rsp.Data, nil
}

// ListDeviceByAssetIds list device by asset id
func (c *rmClient) ListDeviceByAssetIds(ctx context.Context, limit int64,
	assetList []string) ([]*resourcemanager.Device, error) {
	req := &resourcemanager.ListDevicesReq{
		Limit:   &limit,
		AssetID: assetList,
	}
	rsp, err := c.client.ListDevices(ctx, req)
	if err != nil {
		blog.Errorf("list device error:%s", err.Error())
		return nil, err
	}
	if !*rsp.Result {
		blog.Errorf("list device error:%s", *rsp.Message)
		return nil, fmt.Errorf("list device error:%s", *rsp.Message)
	}
	return rsp.Data, nil
}

// ListDeviceByIps list devices by ip
func (c *rmClient) ListDeviceByIps(ctx context.Context, limit int64, ips []string) ([]*resourcemanager.Device, error) {
	req := &resourcemanager.ListDevicesReq{
		Limit: &limit,
		Ip:    ips,
	}
	rsp, err := c.client.ListDevices(ctx, req)
	if err != nil {
		blog.Errorf("list device error:%s", err.Error())
		return nil, err
	}
	if !*rsp.Result {
		blog.Errorf("list device error:%s", *rsp.Message)
		return nil, fmt.Errorf("list device error:%s", *rsp.Message)
	}
	return rsp.Data, nil
}

// ListDeviceByPool list device by pool
func (c *rmClient) ListDeviceByPool(ctx context.Context, limit int64,
	poolId []string) ([]*resourcemanager.Device, error) {
	req := &resourcemanager.ListDevicesReq{
		Limit: &limit,
		Pool:  poolId,
	}
	rsp, err := c.client.ListDevices(ctx, req)
	if err != nil {
		blog.Errorf("list device error:%s", err.Error())
		return nil, err
	}
	if !*rsp.Result {
		blog.Errorf("list device error:%s", *rsp.Message)
		return nil, fmt.Errorf("list device error:%s", *rsp.Message)
	}
	return rsp.Data, nil
}

// UpdateDevice update device
func (c *rmClient) UpdateDevice(ctx context.Context, labels map[string]string, annotations map[string]string,
	ip string) (*resourcemanager.Device, error) {
	deviceRsp, err := c.ListDeviceByIps(ctx, 1, []string{ip})
	if err != nil {
		return nil, fmt.Errorf("get device by ip %s failed:%s", ip, err.Error())
	}
	if deviceRsp == nil || len(deviceRsp) != 1 {
		return nil, fmt.Errorf("get device by ip %s is empty", ip)
	}
	device := deviceRsp[0]
	mergeDevice(device, labels, annotations)
	req := &resourcemanager.UpdateDeviceReq{
		DeviceID:    device.Id,
		Labels:      device.Labels,
		Annotations: device.Annotations,
	}
	rsp, updateErr := c.client.UpdateDevice(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update device %s failed:%s", ip, updateErr.Error())
	}
	if !*rsp.Result {
		return nil, fmt.Errorf("update device %s failed, result is false:%s", ip, *rsp.Message)
	}
	return rsp.Data, nil
}

func mergeDevice(device *resourcemanager.Device, newLabels map[string]string, newAnnotations map[string]string) {
	if device.Labels == nil {
		device.Labels = make(map[string]string)
	}
	if device.Annotations == nil {
		device.Annotations = make(map[string]string)
	}
	for labelKey := range newLabels {
		if device.Labels[labelKey] != newLabels[labelKey] {
			device.Labels[labelKey] = newLabels[labelKey]
		}
	}
	for annoKey := range newAnnotations {
		if device.Annotations[annoKey] != newAnnotations[annoKey] {
			device.Annotations[annoKey] = newAnnotations[annoKey]
		}
	}
}
