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
 */

// Package types xxx
package types

import (
	"time"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	// ServiceDomain domain name for service
	ServiceDomain = "clustermanager.bkbcs.tencent.com"
)

// Cluster cluster struct
type Cluster struct {
	ClusterID           string            `bson:"clusterID"`
	ClusterName         string            `bson:"clusterName"`
	FederationClusterID string            `bson:"federationClusterID"`
	Provider            string            `bson:"provider"`
	Region              string            `bson:"region"`
	VpcID               string            `bson:"vpcID"`
	ProjectID           string            `bson:"projectID"`
	BusinessID          string            `bson:"businessID"`
	Environment         string            `bson:"environment"`
	EngineType          string            `bson:"engineType"`
	IsExclusive         bool              `bson:"isExclusive"`
	ClusterType         string            `bson:"clusterType"`
	APIServerEndpoints  string            `bson:"apiserverEndpoints"`
	APIServerClientCa   string            `bson:"apiServerClientCa"`
	Token               string            `bson:"token"`
	Kubeconfig          string            `bson:"kubeconfig"`
	WssServerCert       string            `bson:"wssServerCert"`
	WssServerKey        string            `bson:"wssServerKey"`
	WssCa               string            `bson:"wssCa"`
	Labels              map[string]string `bson:"labels,omitempty"`
	Operators           []string          `bson:"operators,omitempty"`
	CreateTime          time.Time         `bson:"createTime"`
	UpdateTime          time.Time         `bson:"updateTime"`
}

// ClusterCredential online cluster struct
type ClusterCredential struct {
	ServerKey     string    `bson:"serverKey"`
	ClusterID     string    `bson:"clusterID"`
	ClientModule  string    `bson:"clientModule"`
	ServerAddress string    `bson:"serverAddress"`
	CaCertData    string    `bson:"caCertData"`
	UserToken     string    `bson:"userToken"`
	ClusterDomain string    `bson:"clusterDomain"`
	ConnectMode   string    `bson:"connectMode"`
	CreateTime    time.Time `bson:"createTime"`
	UpdateTime    time.Time `bson:"updateTime"`
}

// Namespace struct of namespace
type Namespace struct {
	Name                string            `bson:"name"`
	FederationClusterID string            `bson:"federationClusterID"`
	ProjectID           string            `bson:"projectID"`
	BusinessID          string            `bson:"businessID"`
	Labels              map[string]string `bson:"labels,omitempty"`
	MaxQuota            string            `bson:"maxQuota"`
	CreateTime          time.Time         `bson:"createTime"`
	UpdateTime          time.Time         `bson:"updateTime"`
}

// NamespaceQuota resource quota of namespace
type NamespaceQuota struct {
	Namespace           string    `bson:"namespace"`
	FederationClusterID string    `bson:"federationClusterID"`
	ClusterID           string    `bson:"clusterID"`
	Region              string    `bson:"region"`
	ResourceQuota       string    `bson:"resourceQuota"`
	CreateTime          time.Time `bson:"createTime"`
	UpdateTime          time.Time `bson:"updateTime"`
	Status              string    `bson:"status"`
	Message             string    `bson:"message"`
}

// TkeCidr tke cidr
type TkeCidr struct {
	Vpc      string    `bson:"vpc"`
	Cidr     string    `bson:"cidr"`
	IPNumber uint64    `bson:"ipNumber"`
	Status   string    `bson:"status"`
	Cluster  string    `bson:"cluster"`
	CreateAt time.Time `bson:"createAt"`
	UpdateAt time.Time `bson:"updateAt"`
}

// TkeCidrCount tke cidr count
type TkeCidrCount struct {
	Count    uint64 `bson:"count"`
	Vpc      string `bson:"vpc"`
	IPNumber uint64 `bson:"ipNumber"`
	Status   string `bson:"status"`
}

// NodeAddress node address
type NodeAddress struct {
	NodeName    string
	IPv4Address string
	IPv6Address string
}

// ResourceSchema resource schema
type ResourceSchema struct {
	Name        string              `json:"name"`
	DisplayName string              `json:"displayName"`
	Description string              `json:"description"`
	Schema      *v1.JSONSchemaProps `json:"schema"`
	CloudID     string              `json:"cloudID"`
}

// GCPServiceAccount for gcp service account secret
type GCPServiceAccount struct {
	AccountType         string `json:"type"`
	ProjectID           string `json:"project_id"`
	PrivateKeyID        string `json:"private_key_id"`
	PrivateKey          string `json:"private_key"`
	ClientEmail         string `json:"client_email"`
	ClientID            string `json:"client_id"`
	AuthURI             string `json:"auth_uri"`
	TokenURI            string `json:"token_uri"`
	AuthProviderCertURL string `json:"auth_provider_x509_cert_url"`
	ClientCertURL       string `json:"client_x509_cert_url"`
}

// ClusterWholeTaskMetrics cluster whole task metrics
type ClusterWholeTaskMetrics struct {
	ClusterId        string  `json:"clusterId"`
	SuccessRate      float64 `json:"successRate"`
	AvgExecutionTime float64 `json:"avgExecutionTime"`
}

// ClusterSubSuccessTaskMetrics cluster sub success task metrics
type ClusterSubSuccessTaskMetrics struct {
	ClusterId        string  `json:"clusterId"`
	SuccessRate      float64 `json:"successRate"`
	AvgExecutionTime float64 `json:"avgExecutionTime"`
	FailTasks        int     `json:"failTasks"`
	TaskType         string  `json:"taskType"`
}

// ClusterSubFailTaskMetrics cluster sub fail task metrics
type ClusterSubFailTaskMetrics struct {
	ClusterId string `json:"clusterId"`
	TaskType  string `json:"taskType"`
	Message   string `json:"message"`
	FailTasks int    `json:"failTasks"`
}

// BusinessWholeTaskMetrics business whole task metrics
type BusinessWholeTaskMetrics struct {
	BusinessId       string  `json:"businessId"`
	SuccessRate      float64 `json:"successRate"`
	AvgExecutionTime float64 `json:"avgExecutionTime"`
}

// BusinessSubSuccessTaskMetrics business sub success task metrics
type BusinessSubSuccessTaskMetrics struct {
	BusinessId       string  `json:"businessId"`
	SuccessRate      float64 `json:"successRate"`
	AvgExecutionTime float64 `json:"avgExecutionTime"`
	FailTasks        int     `json:"failTasks"`
	TaskType         string  `json:"taskType"`
}

// BusinessSubFailTaskMetrics business sub fail task metrics
type BusinessSubFailTaskMetrics struct {
	BusinessId string `json:"businessId"`
	TaskType   string `json:"taskType"`
	Message    string `json:"message"`
	FailTasks  int    `json:"failTasks"`
}
