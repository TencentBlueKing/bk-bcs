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
	microRgt "github.com/micro/go-micro/v2/registry"
	"github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/discovery"
)

const (
	// ProjectManagerServiceName project manager service name
	ProjectManagerServiceName = "project.bkbcs.tencent.com"

	// cache key
	cacheProjectKeyPrefix = "project_%s"

	// defaultExpiration
	defaultExpiration = 10 * time.Minute
)

// Client xxx
type Client struct {
	Discovery       *discovery.ModuleDiscovery
	ClientTLSConfig *tls.Config
	Cache           *cache.Cache
}

var client *Client

// NewClient create project service client
func NewClient(tlsConfig *tls.Config, microRgt microRgt.Registry) error {
	dis := discovery.NewModuleDiscovery(ProjectManagerServiceName, microRgt)
	err := dis.Start()
	if err != nil {
		return err
	}
	client = &Client{
		Discovery:       dis,
		ClientTLSConfig: tlsConfig,
		Cache:           cache.New(defaultExpiration, cache.NoExpiration),
	}
	return nil
}

func (p *Client) getProjectClient() (*bcsproject.ProjectClient, func(), error) {
	node, err := p.Discovery.GetRandServiceInst()
	if err != nil {
		return nil, nil, err
	}
	blog.V(4).Infof("get random project-manager instance [%s] from etcd registry successful", node.Address)

	cfg := bcsapi.Config{}
	// discovery hosts
	cfg.Hosts = discovery.GetServerEndpointsFromRegistryNode(node)
	cfg.TLSConfig = p.ClientTLSConfig
	cfg.InnerClientName = "bcs-helm-manager"
	cli, close := bcsproject.NewProjectManagerClient(&cfg)
	return cli, close, nil
}

// GetProjectByCode get project from project code
func GetProjectByCode(projectCode string) (*bcsproject.Project, error) {
	// load project data from cache
	key := fmt.Sprintf(cacheProjectKeyPrefix, projectCode)
	v, ok := client.Cache.Get(key)
	if ok {
		if project, ok := v.(*bcsproject.Project); ok {
			return project, nil
		}
	}
	cli, close, err := client.getProjectClient()
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.Project.GetProject(context.Background(),
		&bcsproject.GetProjectRequest{ProjectIDOrCode: projectCode})
	if err != nil {
		return nil, fmt.Errorf("GetProject error: %s", err)
	}
	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("GetProject error, code: %d, message: %s, requestID: %s",
			p.Code, p.GetMessage(), p.GetRequestID())
	}
	// save project data to cache
	client.Cache.Set(key, p.Data, defaultExpiration)
	return p.Data, nil
}

// GetVariable get project from project code
func GetVariable(projectCode, clusterID, namespace string) ([]*bcsproject.VariableValue, error) {
	client, close, err := client.getProjectClient()
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	resp, err := client.Variable.RenderVariables(context.Background(),
		&bcsproject.RenderVariablesRequest{ProjectCode: projectCode, ClusterID: clusterID, Namespace: namespace})
	if err != nil {
		return nil, fmt.Errorf("ListNamespaceVariables error: %s", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("ListNamespaceVariables error, code: %d, message: %s, requestID: %s",
			resp.Code, resp.GetMessage(), resp.GetRequestID())
	}
	return resp.GetData(), nil
}
