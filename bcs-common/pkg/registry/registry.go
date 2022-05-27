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
 *
 */

package registry

import (
	"crypto/tls"
	"time"

	microRegistry "github.com/asim/go-micro/v3/registry"
)

//Options registry options
type Options struct {
	//UUID for registry
	id string
	//Registry address, formation like ip:port
	RegistryAddr []string
	//register name, like $module.bkbcs.tencent.com
	Name string
	//bkbcs version information
	Version string
	//Meta info for go-micro registry
	Meta map[string]string
	//Address information for other module discovery
	// format likes ip:port
	RegAddr  string
	Config   *tls.Config
	TTL      time.Duration
	Interval time.Duration
	//EventHandler & modules that registry watchs
	EvtHandler EventHandler
	Modules    []string
}

//EventHandler handler for module update notification
type EventHandler func(name string)

// Registry interface for go-micro etcd discovery
type Registry interface {
	//Register service information to registry
	// register do not block, if module want to
	// clean registe information, call Deregister
	Register() error
	//Deregister clean service information from registry
	// it means that registry is ready to exit
	Deregister() error
	//Get get specified service by name in local cache
	Get(name string) (*microRegistry.Service, error)
}
