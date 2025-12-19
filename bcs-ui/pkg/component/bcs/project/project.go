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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"github.com/patrickmn/go-cache"
	microRgt "go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
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
	Cache *cache.Cache
}

var client *Client

// NewClient create project service client
func NewClient(tlsConfig *tls.Config, microRgt microRgt.Registry) error {
	if !discovery.UseServiceDiscovery() {
		dis := discovery.NewModuleDiscovery(ProjectManagerServiceName, microRgt)
		err := dis.Start()
		if err != nil {
			return err
		}
		bcsproject.SetClientConfig(tlsConfig, dis)
	} else {
		bcsproject.SetClientConfig(tlsConfig, nil)
	}
	client = &Client{
		Cache: cache.New(defaultExpiration, cache.NoExpiration),
	}
	return nil
}

// GetProjectByCode get project from project code
func GetProjectByCode(ctx context.Context, projectCode string) (*bcsproject.Project, error) {
	// load project data from cache
	key := fmt.Sprintf(cacheProjectKeyPrefix, projectCode)
	v, ok := client.Cache.Get(key)
	if ok {
		if project, ok := v.(*bcsproject.Project); ok {
			return project, nil
		}
	}
	cli, close, err := bcsproject.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.Project.GetProject(ctx,
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
