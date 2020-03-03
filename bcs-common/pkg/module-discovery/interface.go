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

package module_discovery

type ModuleDiscovery interface {
	// module: types.BCS_MODULE_SCHEDULER...
	// list all servers
	//if mesos-apiserver/k8s-apiserver module={module}/clusterid, for examples: mesosdriver/BCS-TESTBCSTEST01-10001
	GetModuleServers(module string) ([]interface{}, error)

	// get random one server
	//if mesos-apiserver/k8s-apiserver module={module}/clusterid, for examples: mesosdriver/BCS-TESTBCSTEST01-10001
	GetRandModuleServer(moduleName string) (interface{}, error)

	// register event handle function
	RegisterEventFunc(handleFunc EventHandleFunc)
}

// module: types.BCS_MODULE_SCHEDULER...
// if mesos-apiserver/k8s-apiserver module={module}/clusterid, for examples: mesosdriver/BCS-TESTBCSTEST01-10001
type EventHandleFunc func(module string)
