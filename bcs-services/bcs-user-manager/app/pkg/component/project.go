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

package component

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

// ProjectData project data
type ProjectData struct {
	Total   int       `json:"total"`
	Results []Project `json:"results"`
}

// Project project
type Project struct {
	Creator     string `json:"creator"`
	Updater     string `json:"updater"`
	Managers    string `json:"managers"`
	ProjectID   string `json:"projectID"`
	Name        string `json:"name"`
	ProjectCode string `json:"projectCode"`
}

// GetProject 通过 project_id/code 获取项目信息
func GetProject(ctx context.Context, projectIDOrCode string) (*Project, error) {
	cacheKey := fmt.Sprintf("bcs.GetProject:%s", projectIDOrCode)
	if cacheResult, ok := cache.LocalCache.Get(cacheKey); ok {
		return cacheResult.(*Project), nil
	}

	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s", config.GetGlobalConfig().BcsAPI.Host, projectIDOrCode)
	resp, err := GetClient().R().
		SetContext(ctx).
		SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	project := new(Project)
	if err := UnmarshalBKResult(resp, project); err != nil {
		return nil, err
	}

	cache.LocalCache.Set(cacheKey, project, time.Hour*1)

	return project, nil
}

// QueryProjects query projects
func QueryProjects(ctx context.Context, limit, offset int, params map[string]string) (*ProjectData, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects", config.GetGlobalConfig().BcsAPI.Host)

	if params == nil {
		params = make(map[string]string)
	}
	params["limit"] = strconv.Itoa(limit)
	params["offset"] = strconv.Itoa(offset)

	resp, err := GetClient().R().
		SetContext(ctx).
		SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
		SetQueryParams(params).
		Get(url)

	if err != nil {
		return nil, err
	}

	var data ProjectData
	if err := UnmarshalBKResult(resp, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// Namespace ns
type Namespace struct {
	Name string `json:"name"`
}

// GetClusterNamespaces get cluster namespaces
func GetClusterNamespaces(ctx context.Context, projectID, clusterID string) ([]Namespace, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s/clusters/%s/namespaces",
		config.GetGlobalConfig().BcsAPI.Host, projectID, clusterID)

	resp, err := GetClient().R().
		SetContext(ctx).
		SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	var data []Namespace
	if err := UnmarshalBKResult(resp, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// GetCachedNamespace get cached namespace
func GetCachedNamespace(ctx context.Context, clusterID, nsID string) (*Namespace, error) {
	// get namespace from cache
	cacheName := func(nsID string) string {
		return fmt.Sprintf("namespace_%s", nsID)
	}
	val, ok := cache.LocalCache.Get(cacheName(nsID))
	if ok && val != nil {
		if n, ok1 := val.(*Namespace); ok1 {
			return n, nil
		}
	}

	// get cluster namespaces
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	cluster, err := GetClusterByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	project, err := GetProject(ctx, cluster.ProjectID)
	if err != nil {
		return nil, err
	}
	// 解决共享集群问题
	nss, err := GetClusterNamespaces(ctx, project.ProjectCode, clusterID)
	if err != nil {
		return nil, err
	}

	var ns *Namespace
	for _, v := range nss {
		namespace := v
		curNSID := utils.CalcIAMNsID(clusterID, namespace.Name)
		if curNSID == nsID {
			ns = &namespace
		}
		cache.LocalCache.Set(cacheName(curNSID), &namespace, cache.NoExpiration)
	}

	if ns == nil {
		return nil, fmt.Errorf("namespace %s not found", nsID)
	}
	return ns, nil
}
