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

package cnm

import (
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/network"
)

//NewNetManager interface for return DockerNetManager
func NewNetManager() network.NetManager {
	manager := &DockerNetManager{}
	return manager
}

//DockerNetManager docker network manager for using docker network
type DockerNetManager struct {
}

//Init loading all configuration in directory
func (manager *DockerNetManager) Init() error {
	return nil
}

//Stop manager stop if necessary
func (manager *DockerNetManager) Stop() {
	//empty
}

//GetPlugin get plugin by name
func (manager *DockerNetManager) GetPlugin(name string) network.NetworkPlugin {
	return nil
}

//AddPlugin Add plugin to manager dynamic if necessary
func (manager *DockerNetManager) AddPlugin(name string, plguin network.NetworkPlugin) error {
	return nil
}

//SetUpPod for setting Pod network interface
func (manager *DockerNetManager) SetUpPod(podInfo container.Pod) error {
	return nil
}

//TearDownPod for release pod network resource
func (manager *DockerNetManager) TearDownPod(podInfo container.Pod) error {
	return nil
}
