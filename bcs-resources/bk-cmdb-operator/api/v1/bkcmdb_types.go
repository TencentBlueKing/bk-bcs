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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BkcmdbSpec defines the desired state of Bkcmdb
type BkcmdbSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	IngressDomain   string               `json:"ingress-domain,omitempty"`
	Image           string               `json:"image,omitempty"`
	MongoDb         *MongoDbConf         `json:"mongodb,omitempty"`
	Redis           *RedisConf           `json:"redis,omitempty"`
	Zookeeper       *ZookeeperConf       `json:"zookeeper,omitempty"`
	AdminServer     *AdminServerConf     `json:"adminserver,omitempty"`
	ApiServer       *ApiServerConf       `json:"apiserver,omitempty"`
	CoreService     *CoreServiceConf     `json:"coreservice,omitempty"`
	DataCollection  *DataCollectionConf  `json:"datacollection,omitempty"`
	EventServer     *EventServerConf     `json:"eventserver,omitempty"`
	HostServer      *HostServerConf      `json:"hostserver,omitempty"`
	OperationServer *OperationServerConf `json:"operationserver,omitempty"`
	ProcServer      *ProcServerConf      `json:"procserver,omitempty"`
	TaskServer      *TaskServerConf      `json:"taskserver,omitempty"`
	TmServer        *TmServerConf        `json:"tmserver,omitempty"`
	TopoServer      *TopoServerConf      `json:"toposerver,omitempty"`
	WebServer       *WebServerConf       `json:"webserver,omitempty"`
}

type WebServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type TopoServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type TmServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type TaskServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type ProcServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type OperationServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type HostServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type EventServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type DataCollectionConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type CoreServiceConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type ApiServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type AdminServerConf struct {
	Replicas  uint                        `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type ZookeeperConf struct {
	Host string `json:"host,omitempty"`
	Port int32  `json:"port,omitempty"`
}

type RedisConf struct {
	Host     string `json:"host,omitempty"`
	Port     int32  `json:"port,omitempty"`
	Password string `json:"password,omitempty"`
}

type MongoDbConf struct {
	Host     string `json:"host,omitempty"`
	Port     int32  `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// BkcmdbStatus defines the observed state of Bkcmdb
type BkcmdbStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Bkcmdb is the Schema for the bkcmdbs API
type Bkcmdb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BkcmdbSpec   `json:"spec,omitempty"`
	Status BkcmdbStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BkcmdbList contains a list of Bkcmdb
type BkcmdbList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bkcmdb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bkcmdb{}, &BkcmdbList{})
}
