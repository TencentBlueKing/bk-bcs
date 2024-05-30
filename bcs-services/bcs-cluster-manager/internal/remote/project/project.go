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

// Package project xxx
package project

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"
	rutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	// projectManagerServiceName project manager service name
	projectManagerServiceName = "project.bkbcs.tencent.com"
	// cache key
	cacheProjectKeyPrefix = "project_%s"
)

// Options for rm client
type Options struct {
	// Module module name
	Module string
	// other configInfo
	TLSConfig *tls.Config
}

// ProjectClient global project client
var ProjectClient *ProManClient

// SetProjectClient set global project client
func SetProjectClient(opts *Options, disc *discovery.ModuleDiscovery) {
	if opts.Module == "" {
		opts.Module = projectManagerServiceName
	}

	ProjectClient = &ProManClient{
		opts:  opts,
		disc:  disc,
		cache: cache.New(5*time.Minute, 60*time.Minute),
	}
}

// GetProjectManagerClient get project client
func GetProjectManagerClient() *ProManClient {
	return ProjectClient
}

// ProManClient project client
type ProManClient struct {
	opts  *Options
	disc  *discovery.ModuleDiscovery
	cache *cache.Cache
}

// getProjectManagerClient get project client by discovery
func (pm *ProManClient) getProjectManagerClient() (*bcsproject.ProjectClient, func(), error) {
	if pm == nil {
		return nil, nil, rutils.ErrServerNotInit
	}

	if pm.disc == nil {
		return nil, nil, fmt.Errorf("resourceManager module not enable dsicovery")
	}

	// random server
	nodeServer, err := pm.disc.GetRandomServiceNode()
	if err != nil {
		return nil, nil, err
	}
	endpoints := utils.GetServerEndpointsFromRegistryNode(nodeServer)

	blog.Infof("ProManClient get node[%s] from disc", nodeServer.Address)
	conf := &bcsapi.Config{
		Hosts:           endpoints,
		TLSConfig:       pm.opts.TLSConfig,
		InnerClientName: "bcs-cluster-manager",
	}
	cli, closeCon := bcsproject.NewProjectManagerClient(conf)

	return cli, closeCon, nil
}

// GetProjectInfo get project detailed info
func (pm *ProManClient) GetProjectInfo(projectIdOrCode string, isCache bool) (*bcsproject.Project, error) {
	if pm == nil {
		return nil, rutils.ErrServerNotInit
	}

	// load project data from cache
	key := fmt.Sprintf(cacheProjectKeyPrefix, projectIdOrCode)
	if isCache {
		v, ok := pm.cache.Get(key)
		if ok {
			if project, ok := v.(*bcsproject.Project); ok {
				return project, nil
			}
		}
	}

	cli, closeCon, errGet := pm.getProjectManagerClient()
	if errGet != nil {
		blog.Errorf("GetProjectInfo[%s] getProjectManagerClient failed: %v", projectIdOrCode, errGet)
		return nil, errGet
	}
	defer func() {
		if closeCon != nil {
			closeCon()
		}
	}()

	resp, err := cli.Project.GetProject(context.Background(),
		&bcsproject.GetProjectRequest{ProjectIDOrCode: projectIdOrCode})
	if err != nil {
		blog.Errorf("GetProjectInfo[%s] GetProject failed: %v", projectIdOrCode, err)
		return nil, err
	}

	if resp.Code != 0 {
		blog.Errorf("GetProjectInfo[%s] GetProject err: %v", projectIdOrCode, resp.GetMessage())
		return nil, fmt.Errorf(resp.Message)
	}

	if isCache {
		pm.cache.Set(key, resp.GetData(), cache.DefaultExpiration)
	}

	return resp.GetData(), nil
}
