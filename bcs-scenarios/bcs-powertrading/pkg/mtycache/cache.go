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

// Package mtycache xxx
package mtycache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis"
	resourcemanager "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/resourcemgr/resourcemanagerv4"
)

var (
	defaultTicker = 20 * time.Second
)

// CacheInterface interface
type CacheInterface interface {
	Start(ctx context.Context) error

	GetCache() map[string]map[string]*resourcemanager.Device
	UpdateCacheWithAssetID(assetIDs []string)
	StoreDeviceRecord(deviceIDs []string)
	ExistDeviceRecord(deviceID string) bool

	GetMTYDevicePools(ctx context.Context) (map[string]string, error)
	CheckDeviceRegistered(ctx context.Context, assetIDs []string) (map[string]struct{}, error)
	CreateMTYDefaultDevicePool(ctx context.Context, poolName string) (string, error)
}

type cacheInfo struct {
	sync.RWMutex

	businessId int64
	rmClient   resourcemanager.ResourceManagerService

	devices       map[string]map[string]*resourcemanager.Device
	deviceRecords *sync.Map
}

// NewCache new cache
func NewCache(businessID int64, rmClient resourcemanager.ResourceManagerService) CacheInterface {
	return &cacheInfo{
		businessId:    businessID,
		rmClient:      rmClient,
		devices:       make(map[string]map[string]*resourcemanager.Device),
		deviceRecords: &sync.Map{},
	}
}

// Start 定时同步获取 MTY 的节点信息
func (c *cacheInfo) Start(ctx context.Context) error {
	if err := c.sync(ctx); err != nil {
		return errors.Wrapf(err, "sync mty cache failed")
	}
	ticker := time.NewTicker(defaultTicker)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := c.sync(ctx); err != nil {
				blog.Errorf("[MTY_CACHE] sync failed: %s", err.Error())
			} else {
				blog.Infof("[MTY_CACHE] sync success")
			}
		case <-ctx.Done():
			blog.Infof("[MTY_CACHE] sync closed")
			return nil
		}
	}
}

// UpdateCacheWithAssetID updateCache
func (c *cacheInfo) UpdateCacheWithAssetID(assetIDs []string) {
	resp, err := c.rmClient.ListDevices(context.Background(), &resourcemanager.ListDevicesReq{
		AssetID: assetIDs,
	})
	if err != nil {
		blog.Errorf("[MTY_CACHE] list devices with assets failed: %s", err.Error())
		return
	}
	if *resp.Code != 0 {
		blog.Errorf("[MTY_CACHE] list devices with assets resp code not 0 but %d: %s", *resp.Code, *resp.Message)
		return
	}
	c.storeDevices(resp.Data)
	blog.Infof("[MTY_CACHE] update cache with assets '%v' success", assetIDs)
}

// StoreDeviceRecord store device record
func (c *cacheInfo) StoreDeviceRecord(deviceIDs []string) {
	for _, deviceID := range deviceIDs {
		c.deviceRecords.Store(deviceID, struct{}{})
	}
}

// ExistDeviceRecord edit device record
func (c *cacheInfo) ExistDeviceRecord(deviceID string) bool {
	_, ok := c.deviceRecords.Load(deviceID)
	return ok
}

// GetMTYDevicePools get mty device pools
func (c *cacheInfo) GetMTYDevicePools(ctx context.Context) (map[string]string, error) {
	resp, err := c.rmClient.ListDevicePool(ctx, &resourcemanager.ListDevicePoolReq{
		BusinessID: []int64{c.businessId},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "list mty device pools failed")
	}
	if *resp.Code != 0 {
		return nil, errors.Errorf("list mty device pools resp code not 0 but %d: %s", *resp.Code, *resp.Message)
	}
	result := make(map[string]string)
	for _, pool := range resp.Data {
		if strings.HasPrefix(*pool.Name, "mty_") {
			result[*pool.Name] = *pool.Id
		}
	}
	return result, nil
}

// CheckDeviceRegistered check device
func (c *cacheInfo) CheckDeviceRegistered(ctx context.Context, assetIDs []string) (map[string]struct{}, error) {
	var limit int64 = 10000
	resp, err := c.rmClient.ListDevices(ctx, &resourcemanager.ListDevicesReq{
		Limit:   &limit,
		AssetID: assetIDs,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "list devices failed")
	}
	if *resp.Code != 0 {
		return nil, errors.Errorf("list devices resp code not 0 but %d: %s", *resp.Code, *resp.Message)
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}

	result := make(map[string]struct{})
	for _, device := range resp.Data {
		if device == nil {
			continue
		}
		if device.Info == nil {
			continue
		}
		result[*device.Info.AssetID] = struct{}{}
		blog.Warnf("RequestID[%s] device '%s/%s/%s' already exist in pool '%s' with labels '%v'",
			apis.GetRequestIDFromCtx(ctx), *device.Id, *device.Info.InnerIP, *device.Info.AssetID,
			*device.DevicePoolID, device.Labels)
	}
	return result, nil
}

// CreateMTYDefaultDevicePool create mty device pool
func (c *cacheInfo) CreateMTYDefaultDevicePool(ctx context.Context, poolName string) (string, error) {
	provider := "self"
	operator := "powertrading"
	creator := "bcs-powertrading-mty"
	enableAS := true
	devicePoolType := "host"
	reserved := false
	asType := "HOST"
	region := strings.TrimPrefix(poolName, "mty_")
	createReq := &resourcemanager.CreateDevicePoolReq{
		Name:     &poolName,
		Provider: &provider,
		BaseConfig: &resourcemanager.ConsumeDesire{
			Zone: &resourcemanager.DeviceZone{
				Region: &region,
			},
			BusinessID: &c.businessId,
		},
		Labels: map[string]string{
			"managedby":                        "selfmanager",
			"offerBizList":                     fmt.Sprintf("[\"%d\"]", c.businessId),
			"onscale":                          "true",
			"priceCoefficient":                 "100",
			"recycleExpired":                   "10",
			"recycleStrategy":                  "anytime",
			"reinstall":                        "false",
			"reservedCPU":                      "0",
			"reservedMem":                      "0",
			"reservedType":                     "number",
			"resourceReserve":                  "0",
			"resourceType":                     "offline",
			"supplyWay":                        "static",
			"supplyTime":                       "unlimited",
			"node.info.kubernetes.io/cpu-type": "cvm_low",
		},
		AsOption: &resourcemanager.DevicePoolAutoScalerOption{
			Type:     &asType,
			Settings: map[string]float64{},
		},
		Operator: &operator,
		Creator:  &creator,
		EnableAS: &enableAS,
		Type:     &devicePoolType,
		Reserved: &reserved,
	}
	blog.Infof("RequestID[%s] create devicepool req: %s", apis.GetRequestIDFromCtx(ctx), ToJson(createReq))
	createResp, err := c.rmClient.CreateDevicePool(ctx, createReq)
	if err != nil {
		return "", errors.Wrapf(err, "create devicepool '%s' failed", poolName)
	}
	if *createResp.Code != 0 {
		return "", errors.Errorf("create device pool '%s' resp code not 0 but %d: %s",
			poolName, &createResp.Code, *createResp.Message)
	}
	blog.Infof("RequestID[%s] create device pool resp: %s", apis.GetRequestIDFromCtx(ctx),
		ToJson(createResp))
	return *createResp.Data.Id, nil
}

func (c *cacheInfo) sync(ctx context.Context) error {
	poolResp, err := c.rmClient.ListDevicePool(ctx, &resourcemanager.ListDevicePoolReq{
		BusinessID: []int64{c.businessId},
	})
	if err != nil {
		return errors.Wrapf(err, "list device pools with business '%d' failed", c.businessId)
	}
	if *poolResp.Code != 0 {
		return errors.Errorf("list device pools resp code not 0 but %d: %s", *poolResp.Code, *poolResp.Message)
	}
	mtyPools := make(map[string]string)
	for _, pool := range poolResp.Data {
		if strings.HasPrefix(*pool.Name, "mty_") {
			mtyPools[*pool.Name] = *pool.Id
		}
	}
	records := make(map[string]struct{})
	c.deviceRecords.Range(func(key, value any) bool {
		records[key.(string)] = struct{}{}
		return true
	})
	blog.Infof("[MTY_CACHE] query mty pools: %v", mtyPools)
	for poolName, poolId := range mtyPools {
		var limit int64 = 100000
		var deviceResp *resourcemanager.ListDevicesResp
		blog.Infof("[MTY_CACHE] start list devices for pool '%s/%s'", poolName, poolId)
		deviceResp, err = c.rmClient.ListDevices(ctx, &resourcemanager.ListDevicesReq{
			Limit: &limit,
			Pool:  []string{poolId},
		})
		if err != nil {
			return errors.Wrapf(err, "list devices with pools(%s/%s) failed", poolName, poolId)
		}
		if *deviceResp.Code != 0 {
			return errors.Errorf("list devices with pools(%s/%s) resp code not 0 but %d: %s",
				poolName, poolId, *deviceResp.Code, *deviceResp.Message)
		}
		blog.Infof("[MTY_CACHE] list devices for pool '%s/%s' success: %d",
			poolName, poolId, len(deviceResp.Data))
		c.storePoolDevices(poolId, deviceResp.Data)
		for _, device := range deviceResp.Data {
			_, ok := records[*device.Id]
			if ok {
				delete(records, *device.Id)
			}
		}
	}
	for k := range records {
		c.deviceRecords.Delete(k)
	}
	return nil
}

func (c *cacheInfo) storePoolDevices(poolId string, devices []*resourcemanager.Device) {
	c.Lock()
	defer c.Unlock()
	_, ok := c.devices[poolId]
	if !ok {
		c.devices[poolId] = make(map[string]*resourcemanager.Device)
	}
	devicesMap := make(map[string]*resourcemanager.Device)
	for _, device := range devices {
		devicesMap[*device.Id] = device
	}
	c.devices[poolId] = devicesMap
}

func (c *cacheInfo) storeDevices(devices []*resourcemanager.Device) {
	c.Lock()
	defer c.Unlock()
	for _, device := range devices {
		_, ok := c.devices[*device.DevicePoolID]
		if !ok {
			c.devices[*device.DevicePoolID] = make(map[string]*resourcemanager.Device)
		}
		c.devices[*device.DevicePoolID][*device.Id] = device
	}
}

// GetCache get cache
func (c *cacheInfo) GetCache() map[string]map[string]*resourcemanager.Device {
	c.RLock()
	defer c.RUnlock()
	result := make(map[string]map[string]*resourcemanager.Device)
	for poolID, poolDevices := range c.devices {
		pd := make(map[string]*resourcemanager.Device)
		for k, v := range poolDevices {
			pd[k] = v
		}
		result[poolID] = pd
	}
	return result
}

// ToJson trans to json string
func ToJson(obj interface{}) string {
	bs, err := json.Marshal(obj)
	if err != nil {
		blog.Errorf("marshal object '%v' failed: %s", obj, err.Error())
		return ""
	}
	return string(bs)
}
