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
	"net"
)

//NetworkPlugin defination for all network
type NetworkPlugin interface {
	Name() string //Get plugin name
	//Type() string                            //Get plugin type, executable binary name
	Init(host string) error //init Plugin
	//Status() *NetStatus                      //get network status
	//Version() string                         //Plugin version
	SetUpPod(podInfo container.Pod) error    //Setup Network info for pod
	TearDownPod(podInfo container.Pod) error //Teardown pod network info
}

//NetStatus hold pod network info
type NetStatus struct {
	IfName  string    `json:"ifname"`  //device name
	IP      net.IP    `json:"ip"`      //ip address for pod
	Net     net.IPNet `json:"net"`     //net for ip address, including network mask
	Gateway net.IP    `json:"gateway"` //network gateway
}
