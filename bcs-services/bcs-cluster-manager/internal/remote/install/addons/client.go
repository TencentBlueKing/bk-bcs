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

// Package addons xxx
package addons

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// addonsClient addons client
var addonsClient *AddonsClient

// SetAddonsClient set global addons client
func SetAddonsClient(opts *types.Options) error {
	var err error
	addonsClient, err = NewAddonsClient(opts)
	if err != nil {
		return err
	}

	return nil
}

// GetAddonsClient get addon client
func GetAddonsClient() *AddonsClient {
	return addonsClient
}

// AddonsClient client for addons
type AddonsClient struct { // nolint
	opts      *types.Options
	discovery *discovery.ModuleDiscovery
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewAddonsClient init addon client and start discovery module
func NewAddonsClient(opts *types.Options) (*AddonsClient, error) {
	ok := opts.Validate()
	if !ok {
		return nil, nil
	}

	addonsLocalClient := &AddonsClient{
		opts: opts,
	}
	addonsLocalClient.ctx, addonsLocalClient.cancel = context.WithCancel(context.Background())

	if len(opts.GateWay) == 0 {
		addonsLocalClient.discovery = discovery.NewModuleDiscovery(opts.Module, opts.EtcdRegistry)
		err := addonsLocalClient.discovery.Start()
		if err != nil {
			blog.Errorf("start discovery[%s] client failed: %v", types.ModuleHelmManager, err)
			return nil, err
		}
	}

	return addonsLocalClient, nil
}

// GetAddonsClient get addons client
func (ac *AddonsClient) GetAddonsClient() (helmmanager.ClusterAddonsClient, func(), error) {
	if ac == nil {
		return nil, nil, types.ErrNotInited
	}

	conf := &bcsapi.Config{
		TLSConfig:       ac.opts.ClientTLSConfig,
		InnerClientName: common.ClusterManager,
	}

	if len(ac.opts.GateWay) != 0 {
		conf.Hosts = []string{ac.opts.GateWay}
		conf.AuthToken = ac.opts.Token
	} else {
		nodeServer, err := ac.discovery.GetRandomServiceNode()
		if err != nil {
			return nil, nil, err
		}
		endpoints := utils.GetServerEndpointsFromRegistryNode(nodeServer)
		conf.Hosts = endpoints
	}

	blog.Infof("GetAddonsClient config[%+v]", *conf)

	cli, closeCon := helmmanager.NewHelmAddonsClient(conf)
	return cli, closeCon, nil
}

// Stop stop addonsClient
func (ac *AddonsClient) Stop() {
	if ac == nil {
		return
	}
	if ac.discovery != nil {
		ac.discovery.Stop()
	}
	ac.cancel()
}
