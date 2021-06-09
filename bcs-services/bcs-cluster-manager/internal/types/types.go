/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"time"
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

const (
	// TkeCidrStatusAvailable available tke cidr status
	TkeCidrStatusAvailable = "available"
	// TkeCidrStatusUsed used tke cidr status
	TkeCidrStatusUsed = "used"
	// TkeCidrStatusReserved reserved tke cidr status
	TkeCidrStatusReserved = "reserved"
)

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
