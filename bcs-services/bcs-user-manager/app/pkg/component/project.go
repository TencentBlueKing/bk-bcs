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
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	pkgutils "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
	util "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

// ProjectData project data
type ProjectData struct {
	Total   int       `json:"total"`
	Results []Project `json:"results"`
}

// Project project
type Project struct {
	Creator           string `json:"creator"`
	Updater           string `json:"updater"`
	Managers          string `json:"managers"`
	ProjectID         string `json:"projectID"`
	Name              string `json:"name"`
	ProjectCode       string `json:"projectCode"`
	TenantProjectCode string `json:"tenantProjectCode"`
	TenantID          string `json:"tenantID"`
}

// GetProjectCode get project code
func (p *Project) GetProjectCode() string {
	if p.TenantProjectCode != "" {
		return p.TenantProjectCode
	}
	return p.ProjectCode
}

// GetProjectWithCache 通过 project_id/code 获取项目信息
func GetProjectWithCache(ctx context.Context, projectIDOrCode string) (*Project, error) {
	cacheKey := fmt.Sprintf("bcs.GetProject:%s", projectIDOrCode)
	if cacheResult, ok := cache.LocalCache.Get(cacheKey); ok {
		return cacheResult.(*Project), nil
	}

	proj, err := GetProject(ctx, projectIDOrCode)
	if err != nil {
		return nil, err
	}

	cache.LocalCache.Set(cacheKey, proj, time.Minute*5)

	return proj, nil
}

// GetProject 通过 project_id/code 获取项目信息
func GetProject(ctx context.Context, projectIDOrCode string) (*Project, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s", config.GetGlobalConfig().BcsAPI.InnerHost,
		projectIDOrCode)
	resp, err := GetClient().R().
		SetContext(ctx).
		SetHeaders(pkgutils.GetLaneIDByCtx(ctx)).
		SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
		Get(url)

	if err != nil {
		return nil, err
	}

	project := new(Project)
	if err := UnmarshalBKResult(resp, project); err != nil {
		return nil, err
	}
	return project, nil
}

// QueryProjects query projects
func QueryProjects(ctx context.Context, tenantID string, limit, offset int, params map[string]string) (*ProjectData,
	error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects", config.GetGlobalConfig().BcsAPI.InnerHost)

	if params == nil {
		params = make(map[string]string)
	}
	params["limit"] = strconv.Itoa(limit)
	params["offset"] = strconv.Itoa(offset)

	resp, err := GetClient().R().
		SetContext(ctx).
		SetHeaders(pkgutils.GetLaneIDByCtx(ctx)).
		SetAuthToken(config.GetGlobalConfig().BcsAPI.Token).
		SetQueryParams(params).
		SetHeader(util.HeaderTenantID, tenantID).
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
	Name     string   `json:"name"`
	Managers []string `json:"managers"`
}

// GetClusterNamespaces get cluster namespaces
func GetClusterNamespaces(ctx context.Context, projectCode, clusterID string) ([]Namespace, error) {
	url := fmt.Sprintf("%s/bcsapi/v4/bcsproject/v1/projects/%s/clusters/%s/namespaces",
		config.GetGlobalConfig().BcsAPI.InnerHost, projectCode, clusterID)

	resp, err := GetClient().R().
		SetContext(ctx).
		SetHeaders(pkgutils.GetLaneIDByCtx(ctx)).
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

// GetCachedClusterNamespaces get cached cluster namespaces
func GetCachedClusterNamespaces(ctx context.Context, projectCode, clusterID string) ([]Namespace, error) {
	cacheName := func(projectCode, clusterID string) string {
		return fmt.Sprintf("cluster_namespaces_%s_%s", projectCode, clusterID)
	}
	val, ok := cache.LocalCache.Get(cacheName(projectCode, clusterID))
	if ok && val != nil {
		if namespaces, ok1 := val.([]Namespace); ok1 {
			return namespaces, nil
		}
	}

	namespaces, err := GetClusterNamespaces(ctx, projectCode, clusterID)
	if err != nil {
		return nil, err
	}
	cache.LocalCache.Set(cacheName(projectCode, clusterID), namespaces, 0)
	return namespaces, nil
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
	project, err := GetProjectWithCache(ctx, cluster.ProjectID)
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

// GetProjectFromAttribute get project from attribute
func GetProjectFromAttribute(request *restful.Request) *Project {
	project := request.Attribute(constant.ProjectAttr)
	if p, ok := project.(*Project); ok {
		return p
	}
	return nil
}
