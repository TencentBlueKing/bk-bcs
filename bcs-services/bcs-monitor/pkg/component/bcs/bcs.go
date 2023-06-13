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

// Package bcs xxx
package bcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
)

const (
	// VirtualClusterType vcluster
	VirtualClusterType = "virtual"
)

// Cluster 集群信息
type Cluster struct {
	ProjectId       string `json:"projectID"`
	ClusterId       string `json:"clusterID"`
	ClusterName     string `json:"clusterName"`
	BKBizID         string `json:"businessID"`
	Status          string `json:"status"`
	IsShared        bool   `json:"is_shared"`
	ClusterType     string `json:"clusterType"`
	NetworkSettings struct {
		MaxNodePodNum int `json:"maxNodePodNum"`
		MaxServiceNum int `json:"maxServiceNum"`
	} `json:"networkSettings"`
	ExtraInfo struct {
		NamespaceInfo   string `json:"namespaceInfo"`
		Provider        string `json:"provider"`
		VclusterNetwork string `json:"vclusterNetwork"`
	} `json:"extraInfo"`
	VclusterInfo VclusterInfo `json:"-"`
}

// VclusterInfo vcluster info, parse from extraInfo.namespaceInfo
type VclusterInfo struct {
	Name  string        `json:"name"`
	Quota VclusterQuota `json:"quota"`
}

// VclusterQuota vcluster quota, parse from extraInfo.namespaceInfo
type VclusterQuota struct {
	CPURequests    string `json:"cpuRequests"`
	CPULimits      string `json:"cpuLimits"`
	MemoryRequests string `json:"MemoryRequests"`
	MemoryLimits   string `json:"memoryLimits"`
}

// String :
func (c *Cluster) String() string {
	return fmt.Sprintf("cluster<%s, %s>", c.ClusterName, c.ClusterId)
}

// IsVirtual check cluster is vcluster
func (c *Cluster) IsVirtual() bool {
	return c.ClusterType == VirtualClusterType
}

// CacheListClusters 定时同步 cluster 列表
func CacheListClusters() {
	go func() {
		ListClusters()
		for range time.Tick(time.Minute * 10) {
			klog.Infof("list clusters running")
			ListClusters()
			klog.Infof("list clusters end")
		}
	}()
}

const listClustersCacheKey = "bcs.ListClusters"

// ListClusters 获取集群列表
func ListClusters() {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", config.G.BCS.Host)

	resp, err := component.GetClient().R().
		SetAuthToken(config.G.BCS.Token).
		Get(url)

	if err != nil {
		klog.Errorf("list clusters error, %s", err.Error())
		return
	}

	var result []*Cluster
	if err = component.UnmarshalBKResult(resp, &result); err != nil {
		klog.Errorf("unmarshal clusters error, %s", err.Error())
		return
	}

	clusterMap := map[string]*Cluster{}
	for _, cluster := range result {
		// 集群状态 https://github.com/Tencent/bk-bcs/blob/master/bcs-services/bcs-cluster-manager/
		// api/clustermanager/clustermanager.proto#L1003
		if cluster.Status != "RUNNING" {
			continue
		}
		cls := cluster
		if cls.IsVirtual() {
			cls.VclusterInfo, err = parseVClusterInfo(cls.ExtraInfo.NamespaceInfo)
			if err != nil {
				klog.Errorf("parse clusters %s namespaceInfo %s error, %s", cls.ClusterId, cls.ExtraInfo.NamespaceInfo,
					err.Error())
			}
		}
		clusterMap[cluster.ClusterId] = cls
	}

	storage.LocalCache.Slot.Set(listClustersCacheKey, clusterMap, -1)
	return
}

func parseVClusterInfo(s string) (VclusterInfo, error) {
	info := VclusterInfo{}
	if s == "" {
		return info, nil
	}
	err := json.Unmarshal([]byte(s), &info)
	if err != nil {
		return info, err
	}
	return info, nil
}

// GetClusterMap 获取全部集群数据, map格式
func GetClusterMap() (map[string]*Cluster, error) {
	if cacheResult, ok := storage.LocalCache.Slot.Get(listClustersCacheKey); ok {
		return cacheResult.(map[string]*Cluster), nil
	}
	return nil, fmt.Errorf("not found clusters")
}

var errNotFoundCluster = errors.New("not found cluster")

// GetCluster 获取集群详情
func GetCluster(clusterID string) (*Cluster, error) {
	getCluster := func() (*Cluster, error) {
		var cacheResult interface{}
		var ok bool
		if cacheResult, ok = storage.LocalCache.Slot.Get(listClustersCacheKey); !ok {
			return nil, fmt.Errorf("not found cluster")
		}
		if clusterMap, ok := cacheResult.(map[string]*Cluster); ok {
			if cls, ok := clusterMap[clusterID]; ok {
				return cls, nil
			}
			return nil, errNotFoundCluster
		}
		return nil, fmt.Errorf("cluster cache is invalid")
	}

	cls, err := getCluster()
	if err != nil {
		// 新创建的集群，未在缓存中，刷新一下缓存
		if errors.Is(err, errNotFoundCluster) {
			ListClusters()
			return getCluster()
		}
		return nil, err
	}
	return cls, nil
}
