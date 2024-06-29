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

// Package sdk bcs-sdk
package sdk

import (
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/roundtrip"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/clusterManger"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/helmManger"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/projectManger"
)

// Client bcsprovider-sdk-go 客户端
type Client interface {
	// ClusterManger bcs集群管理服务
	ClusterManger() clusterManger.Service

	// HelmManger bcs helm管理服务
	HelmManger() helmManger.Service

	// ProjectManger bcs 项目管理服务
	ProjectManger() projectManger.Service

	// Config 获取配置
	Config() *options.Config

	// RoundTrip http client
	RoundTrip() roundtrip.Client
}

// NewClient return Client
func NewClient(config *options.Config) (Client, error) {
	if config == nil {
		return nil, errors.Errorf("config cannot be empty.")
	}
	if err := config.Check(); err != nil {
		return nil, errors.Wrap(err, "config check failed.")
	}

	cli := &client{
		config:    config,
		roundtrip: roundtrip.NewClient(config),
	}
	if err := cli.init(); err != nil {
		return nil, errors.Wrapf(err, "sdk client init failed")
	}

	return cli, nil
}

// client impl Client
type client struct {
	// config 配置
	config *options.Config

	// roundtrip http client
	roundtrip roundtrip.Client

	/*
		bcs services
	*/
	cm clusterManger.Service

	hm helmManger.Service

	pm projectManger.Service
}

func (c *client) init() error {
	pm, err := projectManger.NewService(c.config, c.roundtrip)
	if err != nil {
		return errors.Wrapf(err, "new project manager servcie failed")
	}
	c.pm = pm

	hm, err := helmManger.NewService(c.config, c.roundtrip)
	if err != nil {
		return errors.Wrapf(err, "new helm manager servcie failed")
	}
	c.hm = hm

	cm, err := clusterManger.NewService(c.config, c.roundtrip)
	if err != nil {
		return errors.Wrapf(err, "new cluster manager servcie failed")
	}
	c.cm = cm

	return nil
}

// ClusterManger bcs集群管理服务
func (c *client) ClusterManger() clusterManger.Service {
	return c.cm
}

// HelmManger bcs helm管理服务
func (c *client) HelmManger() helmManger.Service {
	return c.hm
}

// ProjectManger bcs 项目管理服务
func (c *client) ProjectManger() projectManger.Service {
	return c.pm
}

// Config 获取配置
func (c *client) Config() *options.Config {
	return c.config
}

// RoundTrip http client
func (c *client) RoundTrip() roundtrip.Client {
	return c.roundtrip
}
