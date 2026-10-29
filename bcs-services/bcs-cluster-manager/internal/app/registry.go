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

package app

import (
	"sync"

	"go-micro.dev/v4/registry"
	registryutil "go-micro.dev/v4/util/registry"
)

// shutdownRegistry prevents this server from registering again during shutdown.
type shutdownRegistry struct {
	registry.Registry
	mu       sync.Mutex
	stopping bool
	service  *registry.Service
}

// Register skips registration after shutdown begins.
func (r *shutdownRegistry) Register(service *registry.Service, opts ...registry.RegisterOption) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopping {
		return nil
	}
	if err := r.Registry.Register(service, opts...); err != nil {
		return err
	}
	// Preserve the registered identity without retaining the caller's mutable nodes.
	r.service = registryutil.CopyService(service)
	return nil
}

// StopAndDeregister waits for in-flight registration, blocks renewals and removes this instance.
// The framework still performs its normal deregistration and cleanup in Server.Stop.
func (r *shutdownRegistry) StopAndDeregister() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stopping = true
	if r.service == nil {
		return nil
	}
	if err := r.Registry.Deregister(r.service); err != nil {
		return err
	}
	r.service = nil
	return nil
}
