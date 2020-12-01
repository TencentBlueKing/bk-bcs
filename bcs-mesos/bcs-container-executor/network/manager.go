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

package network

import (
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
)

//NetManager manager for NetworkPlugin
type NetManager interface {
	Init() error                                       //manager init
	Stop()                                             //manager stop if necessary
	GetPlugin(name string) NetworkPlugin               //get plugin by name
	AddPlugin(name string, plguin NetworkPlugin) error //Add plugin to manager dynamic if necessary
	SetUpPod(podInfo container.Pod) error              //for setting Pod network interface
	TearDownPod(podInfo container.Pod) error           //for release pod network resource
}
