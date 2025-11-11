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

package component

import (
	"context"
	"fmt"
	neturl "net/url"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
	apputils "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

const (
	// CacheClusterKeyPrefix key prefix for cluster
	CacheClusterKeyPrefix = "cached_cluster"
	// ListClusterKeyPrefix key prefix for list cluster
	ListClusterKeyPrefix = "cached_list_cluster"
)

// CacheClusterList 定时同步 cluster 列表
func CacheClusterList() {
	go func() {
		_, _ = GetAllCluster()
		for range time.Tick(time.Minute) {
			klog.Infof("list clusters running")
			_, _ = GetAllCluster()
			klog.Infof("list clusters end")
		}
	}()
}

// Cluster cluster data
type Cluster struct {
	ClusterID   string `json:"clusterID"`
	ProjectID   string `json:"projectID"`
	BusinessID  string `json:"businessID"`
	ClusterName string `json:"clusterName"`
	Creator     string `json:"creator"`
	Updater     string `json:"updater"`
	Status      string `json:"status"`
	IsShared    bool   `json:"is_shared"`
}

// GetClusterByClusterID get cluster by clusterID
func GetClusterByClusterID(ctx context.Context, clusterID string) (*Cluster, error) {
	cacheName := func(id string) string {
		return fmt.Sprintf("%s_%v", CacheClusterKeyPrefix, id)
	}
	val, ok := cache.LocalCache.Get(cacheName(clusterID))
	if ok && val != nil {
		if cluster, ok1 := val.(*Cluster); ok1 {
			return cluster, nil
		}
	}
	blog.V(3).Infof("GetClusterByClusterID miss clusterID cache")
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster/%s", config.GetGlobalConfig().BcsAPI.InnerHost, clusterID)

	resp, err := GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)).
		SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
		Get(url)
	if err != nil {
		return nil, err
	}

	var cluster *Cluster
	if err = UnmarshalBKResult(resp, &cluster); err != nil {
		return nil, err
	}
	err = cache.LocalCache.Add(cacheName(clusterID), cluster, cache.DefaultExpiration)
	if err != nil {
		blog.Errorf("GetClusterByClusterID set cache by cacheName[%s] failed: %v", cacheName(clusterID), err)
	}
	return cluster, nil
}

// GetClustersByProjectID get clusters by projectID
func GetClustersByProjectID(ctx context.Context, projectID string) ([]*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/projects/%s/clusters", config.GetGlobalConfig().BcsAPI.InnerHost,
		projectID)

	resp, err := GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)).
		SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
		Get(url)
	if err != nil {
		return nil, err
	}

	var clusters []*Cluster
	if err = UnmarshalBKResult(resp, &clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}

// GetAllTenantCluster get all clusters
func GetAllTenantCluster() ([]*Cluster, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", config.GetGlobalConfig().BcsAPI.InnerHost)

	tenants, err := ListTenant(context.Background())
	if err != nil {
		return nil, err
	}

	clusterMap := map[string]*Cluster{}
	clustersData := []*Cluster{}
	for _, tenant := range tenants {
		resp, err := GetClient().R().
			SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
			SetHeader(apputils.HeaderTenantID, tenant.ID).
			Get(url)
		if err != nil {
			return nil, err
		}

		var clusters []*Cluster
		if err = UnmarshalBKResult(resp, &clusters); err != nil {
			return nil, err
		}
		for _, v := range clusters {
			cls := v
			clusterMap[v.ClusterID] = cls
			clustersData = append(clustersData, cls)
		}
	}

	cache.LocalCache.Set(ListClusterKeyPrefix, clusterMap, -1)
	return clustersData, nil
}

// GetAllCluster get all clusters
func GetAllCluster() ([]*Cluster, error) {
	if config.GetGlobalConfig() == nil || config.GetGlobalConfig().BcsAPI == nil {
		return nil, fmt.Errorf("bcs-api config not found")
	}
	if config.GetGlobalConfig().Tenant.Enable {
		return GetAllTenantCluster()
	}
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/cluster", config.GetGlobalConfig().BcsAPI.InnerHost)

	resp, err := GetClient().R().
		SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
		Get(url)
	if err != nil {
		return nil, err
	}

	var clusters []*Cluster
	if err = UnmarshalBKResult(resp, &clusters); err != nil {
		return nil, err
	}
	clusterMap := map[string]*Cluster{}
	for _, v := range clusters {
		cls := v
		clusterMap[v.ClusterID] = cls
	}
	cache.LocalCache.Set(ListClusterKeyPrefix, clusterMap, -1)
	return clusters, nil
}

// GetClusterMap 获取全部集群数据, map格式
func GetClusterMap() (map[string]*Cluster, error) {
	if cacheResult, ok := cache.LocalCache.Get(ListClusterKeyPrefix); ok {
		return cacheResult.(map[string]*Cluster), nil
	}
	return nil, fmt.Errorf("not found clusters")
}

// CloudAccount cloud account data
type CloudAccount struct {
	AccountID   string `json:"accountID"`
	AccountName string `json:"accountName"`
	Creator     string `json:"creator"`
	Updater     string `json:"updater"`
}

// ListCloudAccount get cloudaccount by projectID
func ListCloudAccount(ctx context.Context, projectID string, accountIDs []string) ([]*CloudAccount, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/clustermanager/v1/accounts", config.GetGlobalConfig().BcsAPI.InnerHost)

	params := neturl.Values{}
	if projectID != "" {
		params.Add("projectID", projectID)
	}
	for _, v := range accountIDs {
		params.Add("accountID", v)
	}
	resp, err := GetClient().R().
		SetContext(ctx).
		SetHeaders(utils.GetLaneIDByCtx(ctx)).
		SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
		SetQueryParamsFromValues(params).
		Get(url)
	if err != nil {
		return nil, err
	}

	var accounts []*CloudAccount
	if err = UnmarshalBKResult(resp, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}
