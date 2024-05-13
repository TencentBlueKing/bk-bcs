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

// Package helm xxx
package helm

import (
	"context"
	"crypto/tls"
	"errors"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	// errNotInited err server not init
	errNotInited = errors.New("server not init")
)

const (
	defaultTimeOut = time.Second * 10
	retryCount     = 10
)

// helmClient helm-manager client
var helmClient *HelmClient

// SetHelmManagerClient set global helm-manager client
func SetHelmManagerClient(opts *Options) error {
	var err error
	helmClient, err = NewHelmClient(opts)
	if err != nil {
		return err
	}

	return nil
}

// GetHelmManagerClient get user-manager client
func GetHelmManagerClient() *HelmClient {
	return helmClient
}

// HelmClient client for helmmanager
type HelmClient struct { // nolint
	opts      *Options
	discovery *discovery.ModuleDiscovery
	ctx       context.Context
	cancel    context.CancelFunc
}

// Options for init clusterManager
type Options struct {
	Enable bool
	// GateWay address
	GateWay         string
	Token           string
	Module          string
	EtcdRegistry    registry.Registry
	ClientTLSConfig *tls.Config
}

func (o *Options) validate() bool {
	if o == nil {
		return false
	}
	if !o.Enable {
		return false
	}

	if o.Module == "" {
		o.Module = ModuleHelmManager
	}

	return true
}

// NewHelmClient init helm manager and start discovery module(helmmanager)
func NewHelmClient(opts *Options) (*HelmClient, error) {
	ok := opts.validate()
	if !ok {
		return nil, nil
	}

	helmClient := &HelmClient{
		opts: opts,
	}
	helmClient.ctx, helmClient.cancel = context.WithCancel(context.Background())

	if len(opts.GateWay) == 0 {
		helmClient.discovery = discovery.NewModuleDiscovery(opts.Module, opts.EtcdRegistry)
		err := helmClient.discovery.Start()
		if err != nil {
			blog.Errorf("start discovery[%s] client failed: %v", ModuleHelmManager, err)
			return nil, err
		}
	}

	return helmClient, nil
}

// GetHelmManagerClient get helm client
func (hm *HelmClient) GetHelmManagerClient() (helmmanager.HelmManagerClient, func(), error) {
	if hm == nil {
		return nil, nil, errNotInited
	}

	conf := &bcsapi.Config{
		TLSConfig:       hm.opts.ClientTLSConfig,
		InnerClientName: common.ClusterManager,
	}

	if len(hm.opts.GateWay) != 0 {
		conf.Hosts = []string{hm.opts.GateWay}
		conf.AuthToken = hm.opts.Token
	} else {
		nodeServer, err := hm.discovery.GetRandomServiceNode()
		if err != nil {
			return nil, nil, err
		}
		endpoints := utils.GetServerEndpointsFromRegistryNode(nodeServer)
		conf.Hosts = endpoints
	}

	blog.Infof("GetHelmManagerClient config[%+v]", *conf)

	cli, closeCon := helmmanager.NewHelmClient(conf)
	return cli, closeCon, nil
}

// Stop stop HelmManagerClient
func (hm *HelmClient) Stop() {
	if hm == nil {
		return
	}
	if hm.discovery != nil {
		hm.discovery.Stop()
	}
	hm.cancel()
}
