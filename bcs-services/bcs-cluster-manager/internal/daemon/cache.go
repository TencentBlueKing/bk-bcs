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

// Package daemon xxx
package daemon

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	gocache "github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
)

const (
	cacheResourceDevicePool = "cached_resource_device_pool"
)

func buildCacheName(keyPrefix string, id string) string {
	return fmt.Sprintf("%s_%v", keyPrefix, id)
}

// SetResourceDevicePoolData set devicePool data
func SetResourceDevicePoolData(devicePoolId string, poolInfo *resource.DevicePoolInfo) error {
	cacheName := buildCacheName(cacheResourceDevicePool, devicePoolId)

	var err error

	p, exist := cache.GetCache().Get(cacheName)
	if exist {
		blog.Infof("SetResourceDevicePoolData cacheName:%s, cache exist %+v", cacheName, p)
		err = cache.GetCache().Replace(cacheName, poolInfo, gocache.DefaultExpiration)
	} else {
		err = cache.GetCache().Add(cacheName, poolInfo, gocache.DefaultExpiration)
	}
	if err != nil {
		return err
	}

	return nil
}

// GetResourceDevicePoolData get devicePool  data
func GetResourceDevicePoolData(devicePoolId string) (*resource.DevicePoolInfo, bool) {
	cacheName := buildCacheName(cacheResourceDevicePool, devicePoolId)

	val, ok := cache.GetCache().Get(cacheName)
	if ok && val != nil {
		blog.Infof("SetResourceDevicePoolData cacheName:%s, cache exist %+v", cacheName, val)
		if devicePool, ok1 := val.(*resource.DevicePoolInfo); ok1 {
			return devicePool, true
		}
	}

	return nil, false
}
