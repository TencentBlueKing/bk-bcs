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

package serviceclient

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppService internal service structure for container service discovery
type AppService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Version           string        `json:"version,omitempty"`
	Type              string        `json:"type,omitempty"`     //service type, ClusterIP, Intergration or Empty
	Frontend          []string      `json:"frontend,omitempty"` //frontend represents service ip address, use for proxy or intergate
	Alias             string        `json:"alias,omitempty"`    //domain alias
	WANIP             []string      `json:"wanip,omitempty"`    //use for wan export
	Master            string        `json:"master,omitempty"`   //reserved
	ServicePorts      []ServicePort `json:"ports"`              //BcsService.Ports
	Nodes             []AppNode     `json:"nodes"`              //TaskGroup/Pod info
	Spec              interface{}   `json:"spec,omitempty"`     //user custom definition attributes
	RawBytes          string        `json:"rawBytes,omitempty"` //raw string for user custom definition
}

// ServicePort port definition for application
type ServicePort struct {
	Name        string `json:"name"`                 //name for service port
	Protocol    string `json:"protocol"`             //protocol for service port
	Domain      string `json:"domain,omitempty"`     //domain value for http proxy
	Path        string `json:"path,omitempty"`       //http url path
	ServicePort int    `json:"serviceport"`          //service port for all AppNode, ServicePort.Name == AppNode.Ports[i].Name
	ProxyPort   int    `json:"proxyport,omitempty"`  //proxy port for this Service Port if exist
	TargetPort  int    `json:"targetport,omitempty"` //target port for this Service Port
}

// ServicePortList list for sorting
type ServicePortList []ServicePort

// Len is the number of elements in the collection.
func (list ServicePortList) Len() int {
	return len(list)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (list ServicePortList) Less(i, j int) bool {
	return list[i].ServicePort < list[j].ServicePort
}

// Swap swaps the elements with indexes i and j.
func (list ServicePortList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// NodeList list for AppNode
type NodeList []AppNode

// Len is the number of elements in the collection.
func (list NodeList) Len() int {
	return len(list)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (list NodeList) Less(i, j int) bool {
	return list[i].Index < list[j].Index
}

// Swap swaps the elements with indexes i and j.
func (list NodeList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// AppNode node info from Taskgroup/Pod
type AppNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Index             string      `json:"index"`              //node key, pod instance name / taskgroup name
	Version           string      `json:"version,omitempty"`  //node version, like v1, v1.1, v12.01.1, come from env[BCS_DISCOVERY_VERSION]
	Weight            uint        `json:"weight,omitempty"`   //node weight, it's a Relative value
	Network           string      `json:"network,omitempty"`  //app node network mode
	NodeIP            string      `json:"nodeIP"`             //node ip address
	ProxyIP           string      `json:"proxyIP,omitempty"`  //proxy ip address for this node
	Ports             []NodePort  `json:"ports,omitempty"`    //port info for container
	Spec              interface{} `json:"spec,omitempty"`     //user custom definition attributes
	RawBytes          string      `json:"rawBytes,omitempty"` //raw string for user custom definition
}

// GetPort get port by port name
func (node *AppNode) GetPort(name string) *NodePort {
	for _, p := range node.Ports {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// NodePort port info for one node of service
type NodePort struct {
	Name      string `json:"name"`                //name for port, must equal to one service port
	Protocol  string `json:"protocol"`            //protocol for this port
	NodePort  int    `json:"nodeport"`            //node port
	ProxyPort int    `json:"proxyport,omitempty"` //proxy port if exists
}

// PortList list for ports
type PortList []NodePort

// Len is the number of elements in the collection.
func (list PortList) Len() int {
	return len(list)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (list PortList) Less(i, j int) bool {
	return list[i].NodePort < list[j].NodePort
}

//Swap swaps the elements with indexes i and j.
func (list PortList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
