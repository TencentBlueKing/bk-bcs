/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/controller/strategy"
	mgr "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/resourcemgr"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

// Option function for better injection
type Option func(o *Options)

// Options for NodeGroup controller to control
// inner logic loops
type Options struct {
	// interval for one logic loop, unit is second
	Interval uint
	// Concurrency for handle cluster client request
	Concurrency int
	// resource manager interface for data retrieve
	ResourceManager mgr.Client
	// storage for access database
	Storage storage.Storage
	// client for cluster request
	ClusterClient cluster.Client
	// factory for strategy executor
	StrategyExecutorFactory strategy.Factory
}

// ResourceManager implementation for injection
func ResourceManager(c mgr.Client) Option {
	return func(o *Options) {
		o.ResourceManager = c
	}
}

// Storage implementation for injection
func Storage(s storage.Storage) Option {
	return func(o *Options) {
		o.Storage = s
	}
}

// ClusterClient implementation for injection
func ClusterClient(c cluster.Client) Option {
	return func(o *Options) {
		o.ClusterClient = c
	}
}

// StrategyExecutorFactory implementation for injection
func StrategyExecutorFactory(f strategy.Factory) Option {
	return func(o *Options) {
		o.StrategyExecutorFactory = f
	}
}
