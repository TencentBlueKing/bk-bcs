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
 *
 */

package cmdb

import (
	"fmt"

	"github.com/kirito41dd/xslice"
	"github.com/patrickmn/go-cache"
)

func splitCountToPage(counts int, pageLimit int) []Page {
	var pages = make([]Page, 0)

	cntSlice := make([]int, 0)
	for i := 0; i < counts; i++ {
		cntSlice = append(cntSlice, i)
	}
	i := xslice.SplitToChunks(cntSlice, pageLimit)
	ss, ok := i.([][]int)
	if !ok {
		return nil
	}

	for _, s := range ss {
		if len(s) > 0 {
			pages = append(pages, Page{
				Start: s[0],
				Limit: pageLimit,
			})
		}
	}

	return pages
}

const (
	cacheBizTopoKeyPrefix             = "cached_biz_topo"
	cacheBizHostTopoRelationKeyPrefix = "cached_biz_host_topo_relation"
	cacheBizHostDataKeyPrefix         = "cached_biz_host_data"
	cacheCloudIDPrefix                = "cached_cloud_id"
)

// CacheType xxx
type CacheType string

var (
	bizTopo     CacheType = "bizTopo"
	bizHostTopo CacheType = "bizHostTopo"
	bizHostData CacheType = "bizHostData"
)

func cacheName(keyPrefix string, bizID int) string {
	return fmt.Sprintf("%s_%v", keyPrefix, bizID)
}

// GetBizHostData get host data from biz
func GetBizHostData(hostCache *cache.Cache, bizID int) ([]HostData, bool) {
	cacheName := cacheName(cacheBizHostDataKeyPrefix, bizID)

	val, ok := hostCache.Get(cacheName)
	if ok && val != nil {
		if hostData, ok1 := val.([]HostData); ok1 {
			return hostData, true
		}
	}

	return nil, false
}

// SetBizHostData set biz hostData
func SetBizHostData(hostCache *cache.Cache, bizID int, hostList []HostData) error {
	cacheName := cacheName(cacheBizHostDataKeyPrefix, bizID)

	err := hostCache.Add(cacheName, hostList, cache.DefaultExpiration)
	if err != nil {
		return err
	}

	return nil
}

// GetBizTopoData get topo data from biz
func GetBizTopoData(topoCache *cache.Cache, bizID int) (*SearchBizInstTopoData, bool) {
	cacheName := cacheName(cacheBizTopoKeyPrefix, bizID)

	val, ok := topoCache.Get(cacheName)
	if ok && val != nil {
		if topoData, ok1 := val.(*SearchBizInstTopoData); ok1 {
			return topoData, true
		}
	}

	return nil, false
}

// SetBizTopoData set biz topoData
func SetBizTopoData(topoCache *cache.Cache, bizID int, bizTopo *SearchBizInstTopoData) error {
	cacheName := cacheName(cacheBizTopoKeyPrefix, bizID)

	err := topoCache.Add(cacheName, bizTopo, cache.DefaultExpiration)
	if err != nil {
		return err
	}

	return nil
}

// GetBizHostTopoData get host topo data from biz
func GetBizHostTopoData(topoCache *cache.Cache, bizID int) ([]HostTopoRelation, bool) {
	cacheName := cacheName(cacheBizHostTopoRelationKeyPrefix, bizID)

	val, ok := topoCache.Get(cacheName)
	if ok && val != nil {
		if topoData, ok1 := val.([]HostTopoRelation); ok1 {
			return topoData, true
		}
	}

	return nil, false
}

// SetBizHostTopoData set biz hostTopoData
func SetBizHostTopoData(topoCache *cache.Cache, bizID int, bizHostTopo []HostTopoRelation) error {
	cacheName := cacheName(cacheBizHostTopoRelationKeyPrefix, bizID)

	err := topoCache.Add(cacheName, bizHostTopo, cache.DefaultExpiration)
	if err != nil {
		return err
	}

	return nil
}

// SetCloudData set cloud data by cloudID
func SetCloudData(cloudIDCache *cache.Cache, cloudID int, cloudData *SearchCloudAreaInfo) error {
	cacheName := cacheName(cacheCloudIDPrefix, cloudID)

	err := cloudIDCache.Add(cacheName, cloudData, cache.DefaultExpiration)
	if err != nil {
		return err
	}

	return nil
}

// GetCloudData get cloud data by cloudID
func GetCloudData(cloudCache *cache.Cache, cloudID int) (*SearchCloudAreaInfo, bool) {
	cacheName := cacheName(cacheCloudIDPrefix, cloudID)

	val, ok := cloudCache.Get(cacheName)
	if ok && val != nil {
		if cloud, ok1 := val.(*SearchCloudAreaInfo); ok1 {
			return cloud, true
		}
	}

	return nil, false
}
