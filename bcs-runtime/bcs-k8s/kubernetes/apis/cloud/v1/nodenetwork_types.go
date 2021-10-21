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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// NodeNetworkStatusNotReady node network created, but not effects on node
	NodeNetworkStatusNotReady = "NotReady"
	// NodeNetworkStatusReady node network created and effects on node
	NodeNetworkStatusReady = "Ready"
)

// IPAddress data for ip address
type IPAddress struct {
	IP        string `json:"ip"`
	DNSName   string `json:"dnsName,omitempty"`
	IsPrimary bool   `json:"isPrimary"`
	TestField string `json:"test"`
}

// NetworkInterfaceAttachment attachment for network interface
type NetworkInterfaceAttachment struct {
	Index int `json:"index,omitempty"`
	// for aws
	AttachmentID string `json:"attachmentID,omitempty"`
	// for tencent cloud
	EniID      string `json:"eniID,omitempty"`
	InstanceID string `json:"instanceId"`
}

// ElasticNetworkInterface status for elastic network interface
type ElasticNetworkInterface struct {
	Index              int                                `json:"index"`
	EniID              string                             `json:"eniID"`
	RouteTableID       int                                `json:"routeTableID"`
	EniName            string                             `json:"eniName,omitempty"`
	EniIfaceName       string                             `json:"eniIfaceName"`
	EniSubnetID        string                             `json:"eniSubnetID"`
	EniSubnetCidr      string                             `json:"eniSubnetCidr"`
	MacAddress         string                             `json:"macAddress"`
	Attachment         *NetworkInterfaceAttachment        `json:"attachment"`
	IPNum              int                                `json:"ipNum"`
	Address            *IPAddress                         `json:"address"`
	SecondaryAddresses []*IPAddress                       `json:"secondaryAddresses,omitempty"`
	Status             string                             `json:"status,omitempty"`
}

// FloatingIPNetworkInterface status for elastic network interface used to bind floating ip
type FloatingIPNetworkInterface struct {
	Eni     *ElasticNetworkInterface `json:"eni"`
	IPLimit int                      `json:"ipLimit"`
}

// VMInfo vm info
type VMInfo struct {
	NodeZone     string `json:"zone"`
	NodeRegion   string `json:"region"`
	NodeVpcID    string `json:"vpcID"`
	NodeSubnetID string `json:"subnetID"`
	InstanceID   string `json:"instanceID"`
	InstanceIP   string `json:"instanceIP"`
	CoreNum      int    `json:"coreNum"`
	MemNum       int    `json:"memNum"`
}

// NodeNetworkSpec defines the desired state of NodeNetwork
type NodeNetworkSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Cluster     string  `json:"cluster"`
	Hostname    string  `json:"hostname"`
	NodeAddress string  `json:"nodeAddress"`
	VM          *VMInfo `json:"vmInfo"`
	ENINum      int     `json:"eniNum"`
	IPNumPerENI int     `json:"ipNumPerENI"`
}

// NodeNetworkStatus defines the observed state of NodeNetwork
type NodeNetworkStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Enis          []*ElasticNetworkInterface  `json:"enis,omitempty"`
	FloatingIPEni *FloatingIPNetworkInterface `json:"floatingIPEni,omitempty"`
	Status        string                      `json:"status"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// NodeNetwork is the Schema for the nodenetworks API
type NodeNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeNetworkSpec   `json:"spec,omitempty"`
	Status NodeNetworkStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// NodeNetworkList contains a list of NodeNetwork
type NodeNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeNetwork `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeNetwork{}, &NodeNetworkList{})
}
