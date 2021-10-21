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

package discovery

import "github.com/micro/go-micro/v2/registry"

// all bkbcs module register itself with basic domain .bkbcs.tencent.com
// for example, bcs-mesh-manager register it as meshmanager.bkbcs.tencent.com
// module names refer to common/modules/modules.go, formats likes BCSModule${name}

// EventHandler callback function when server changes
type EventHandler func(module string)

// Discovery grpc discovery definition, interface is designed for
// multiple module discovery.
type Discovery interface {
	// Start to work
	Start() error
	//GetModuleServer get local watch module: modules.BCSModuleScheduler
	//if mesos-apiserver/k8s-apiserver module=clusterId.{module}, for examples: 10001.mesosdriver
	GetModuleServer(module string) (*registry.Service, error)
	// GetRandomServerInstance get random one instance of local cache server information
	//if mesos-apiserver/k8s-apiserver module=clusterId.{module}, for examples: 10001.mesosdriver
	GetRandomServerInstance(module string) (*registry.Node, error)
	//ListAllServer list all registered server information
	ListAllServer() ([]*registry.Service, error)
	// AddModuleWatch add new watch for specified module, Discovery will cache watched module info
	AddModuleWatch(module string) error
	// DeleteModuleWatch clean watch for specified module
	DeleteModuleWatch(module string) error
	// RegisterEventFunc register event handle function
	RegisterEventFunc(handleFunc EventHandler)
	// Stop close module discovery
	Stop()
}
