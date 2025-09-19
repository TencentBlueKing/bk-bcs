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

// Package cmanager xxx
package cmanager

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	discovery "github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"go-micro.dev/v4/registry"
)

const (
	// ModuleClusterManager default discovery clustermanager module
	ModuleClusterManager = "clustermanager.bkbcs.tencent.com"
)

// Options for init clusterManager
type Options struct {
	Module          string
	Address         string
	EtcdRegistry    registry.Registry
	ClientTLSConfig *tls.Config
	AuthToken       string
}

func (o *Options) validate() bool {
	if o == nil {
		return false
	}

	if o.Module == "" {
		o.Module = ModuleClusterManager
	}

	return true
}

// InitClusterManagerDiscovery init cluster manager and start discovery module(clustermanager)
func InitClusterManagerDiscovery(opts *Options) error {
	ok := opts.validate()
	if !ok {
		return errors.New("InitClusterManagerDiscovery failed")
	}

	if !discovery.UseServiceDiscovery() {
		dis := discovery.NewModuleDiscovery(opts.Module, opts.EtcdRegistry)
		err := dis.Start()
		if err != nil {
			return fmt.Errorf("start discovery client failed: %v", err)
		}
		clustermanager.SetClientConfig(opts.ClientTLSConfig, dis)
	} else {
		clustermanager.SetClientConfig(opts.ClientTLSConfig, nil)
	}
	return nil
}
