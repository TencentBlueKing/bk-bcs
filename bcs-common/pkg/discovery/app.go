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

const (
	//BcsDiscoveryVersion definition for version parameter in container ENV
	BcsDiscoveryVersion = "BCS_DISCOVERY_VERSION"
)

//AppSvcRequest base info for request
type AppSvcRequest struct {
	Meta `json:",inline"`
}

//AppSvc service definition for bcs-discovery
type AppSvc struct {
	Meta     `json:",inline"`
	Frontend []string            `json:"frontend,omitempty"` //frontend, come from BcsService.ClusterIP
	Master   string              `json:"master,omitempty"`   //reserved
	SvcPorts []*SvcPort          `json:"ports,omitempty"`    //BcsService.Ports
	Nodes    map[string]*AppNode `json:"nodes,omitempty"`    //TaskGroup/Pod info
}

//SvcPort port definition from BcsService
type SvcPort struct {
	Name        string `json:"name"`
	Protocol    string `json:"protocol"`
	Domain      string `json:"domain,omitempty"`
	Path        string `json:"path,omitempty"`
	ServicePort int    `json:"serviceport"`
	NodePort    int    `json:"nodeport,omitempty"`
}

//SvcPortList list for sorting
type SvcPortList []*SvcPort

//Len is the number of elements in the collection.
func (list SvcPortList) Len() int {
	return len(list)
}

//Less reports whether the element with
// index i should sort before the element with index j.
func (list SvcPortList) Less(i, j int) bool {
	return list[i].ServicePort < list[j].ServicePort
}

//Swap swaps the elements with indexes i and j.
func (list SvcPortList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

//AppNode node info from Taskgroup/Pod
type AppNode struct {
	Meta        `json:",inline"`
	Index       string   `json:"index"`             //node key, pod instance name / taskgroup name
	Version     string   `json:"version,omitempty"` //node version, like v1, v1.1, v12.01.1, come from env[BCS_DISCOVERY_VERSION]
	Network     string   `json:"network"`           //app node network mode
	ContainerIP string   `json:"containerIP"`       //node container ip address
	NodeIP      string   `json:"nodeIP"`            //container deployed host ip address
	Ports       PortList `json:"ports,omitempty"`   //port info for container
}

//Key AppNode key func
func (node *AppNode) Key() string {
	return node.Cluster + "." + node.Namespace + "." + node.Name + "." + node.Index
}

//NodePort port info for container
type NodePort struct {
	Name          string `json:"name"`
	Protocol      string `json:"protocol"`
	ContainerPort int    `json:"containerport"`
	NodePort      int    `json:"nodeport,omitempty"`
}

//PortList list for ports
type PortList []*NodePort

//Len is the number of elements in the collection.
func (list PortList) Len() int {
	return len(list)
}

//Less reports whether the element with
// index i should sort before the element with index j.
func (list PortList) Less(i, j int) bool {
	return list[i].ContainerPort < list[j].ContainerPort
}

//Swap swaps the elements with indexes i and j.
func (list PortList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
