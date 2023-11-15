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

package common

const (
	// ModeTunnel tunnel
	ModeTunnel = "tunnel"
	// ModeService service
	ModeService = "service"
	// ProxyName for proxy
	ProxyName = "gitopsproxy.bkbcs.tencent.com"
	// ServiceName for manager
	ServiceName = "gitopsmanager.bkbcs.tencent.com"
	// MetaHTTPKey key for http port
	MetaHTTPKey = "httpport"
	// ConnectURL for peer interconnection
	ConnectURI = "/gitopsmanager/websocket/connect"
	// GatewayURL for gitops manager connect through bcs gateway
	GatewayURL = "/bcsapi/v4" + ConnectURI

	// GitOpsProxyURL proxy path for gitops
	GitOpsProxyURL = "/gitopsmanager/proxy"

	// HeaderServerAddressKey header key for cluster info in websocket tunnel
	HeaderServerAddressKey = "BCS-GITOPS-ServerAddress"
	// HeaderBCSClient header for bcs client
	HeaderBCSClient = "X-BCS-Client"
	// ServiceNameShort used for bcs-client header
	ServiceNameShort = "bcs-gitops-manager"

	// ProjectAliaName the alia name for project
	ProjectAliaName = "bkbcs.tencent.com/projectAliaName"
	// ProjectIDKey ID key indexer
	ProjectIDKey = "bkbcs.tencent.com/projectID"
	// ProjectBusinessIDKey for bcs business indexer
	ProjectBusinessIDKey = "bkbcs.tencent.com/businessID"
	// ProjectBusinessName for bcs business name
	ProjectBusinessName = "bkbcs.tencent.com/businessName"

	// ClusterAliaName defines the alia's name for project
	ClusterAliaName = "bkbcs.tencent.com/clusterAliaName"
	// ClusterEnv defines the cluster env
	ClusterEnv = "bkbcs.tencent.com/clusterEnv"

	// InClusterName defines the cluster which is 'in-cluster', reserve cluster
	InClusterName = "in-cluster"

	// SecretKey defines the secretManager k8s secret namespace:name
	// NOCC:gas/crypto(工具误报)
	SecretKey = "bkbcs.tencent.com/secretManager"
)
