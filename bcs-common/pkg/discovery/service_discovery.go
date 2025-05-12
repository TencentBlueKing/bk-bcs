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

// Package discovery service discovery
package discovery

import "os"

const (
	// EnvUseServiceDiscoveryKey 是否使用service discovery
	EnvUseServiceDiscoveryKey = "ENV_USE_SERVICE_DISCOVERY"

	// ServiceName BCS 服务名称

	// AlertManagerServiceName alertmanager 服务名称
	AlertManagerServiceName = "bcs-alert-manager"
	// ApiGatewayServiceName apigateway 服务名称
	ApiGatewayServiceName = "bcs-api-gateway"
	// BkcmdbSynchronizerServiceName bkcmdb-synchronizer 服务名称
	BkcmdbSynchronizerServiceName = "bcs-bkcmdb-synchronizer"
	// BkcmdbSynchronizerServerServiceName bkcmdb-synchronizer-server 服务名称
	BkcmdbSynchronizerServerServiceName = "bcs-bkcmdb-synchronizer-server"
	// CidrManagerServiceName cidrmanager 服务名称
	CidrManagerServiceName = "bcs-cidr-manager"
	// ClusterManagerServiceName clustermanager 服务名称
	ClusterManagerServiceName = "bcs-cluster-manager"
	// ClusterManagerStandaloneServiceName clustermanager standalone 服务名称
	ClusterManagerStandaloneServiceName = "bcs-cluster-manager-standalone"
	// ClusterResourcesServiceName clusterresources 服务名称
	ClusterResourcesServiceName = "bcs-cluster-resources"
	// CostManagerCadenceFrontendServiceName costmanager cadence frontend 服务名称
	CostManagerCadenceFrontendServiceName = "bcs-cost-manager-cadence-frontend"
	// CostManagerCadenceFrontendHeadlessServiceName costmanager cadence frontend headless 服务名称
	CostManagerCadenceFrontendHeadlessServiceName = "bcs-cost-manager-cadence-frontend-headless"
	// CostManagerCadenceHistoryHeadlessServiceName costmanager cadence history headless 服务名称
	CostManagerCadenceHistoryHeadlessServiceName = "bcs-cost-manager-cadence-history-headless"
	// CostManagerCadenceMatchingHeadlessServiceName costmanager cadence matching headless 服务名称
	CostManagerCadenceMatchingHeadlessServiceName = "bcs-cost-manager-cadence-matching-headless"
	// CostManagerCadenceWorkerHeadlessServiceName costmanager cadence worker headless 服务名称
	CostManagerCadenceWorkerHeadlessServiceName = "bcs-cost-manager-cadence-worker-headless"
	// CostManagerRedisServiceName costmanager redis 服务名称
	CostManagerRedisServiceName = "bcs-cost-manager-redis"
	// CostManagerServerServiceName costmanager server 服务名称
	CostManagerServerServiceName = "bcs-cost-manager-server"
	// DataManagerServiceName datamanager 服务名称
	DataManagerServiceName = "bcs-data-manager"
	// DevspaceManagerServiceName devspace manager 服务名称
	DevspaceManagerServiceName = "bcs-devspace-manager"
	// FederationManagerServiceName federation manager 服务名称
	FederationManagerServiceName = "bcs-federation-manager"
	// HelmManagerServiceName helm manager 服务名称
	HelmManagerServiceName = "bcs-helm-manager"
	// MonitorApiServiceName monitor api 服务名称
	MonitorApiServiceName = "bcs-monitor-api"
	// MonitorQueryServiceName monitor query 服务名称
	MonitorQueryServiceName = "bcs-monitor-query"
	// MonitorStoregwServiceName monitor storegw 服务名称
	MonitorStoregwServiceName = "bcs-monitor-storegw"
	// MonitorStoregwSdServiceName monitor storegw sd 服务名称
	MonitorStoregwSdServiceName = "bcs-monitor-storegw-sd"
	// MonitorTencentStoregwServiceName monitor tencent storegw 服务名称
	MonitorTencentStoregwServiceName = "bcs-monitor-tencent-storegw"
	// MonitorTencentStoregwSdServiceName monitor tencent storegw sd 服务名称
	MonitorTencentStoregwSdServiceName = "bcs-monitor-tencent-storegw-sd"
	// NodeCommandServiceName node command 服务名称
	NodeCommandServiceName = "bcs-node-command"
	// NodegroupManagerServiceName nodegroup manager 服务名称
	NodegroupManagerServiceName = "bcs-nodegroup-manager"
	// OperationDataExporterServiceName operation data exporter 服务名称
	OperationDataExporterServiceName = "bcs-operation-data-exporter"
	// OperationDataExporterYundingServiceName operation data exporter yunding 服务名称
	OperationDataExporterYundingServiceName = "bcs-operation-data-exporter-yunding"
	// ProjectManagerServiceName project manager 服务名称
	ProjectManagerServiceName = "bcs-project-manager"
	// RabbitmqServiceName 				rabbitmq 服务名称
	RabbitmqServiceName = "bcs-rabbitmq"
	// RabbitmqHeadlessServiceName 		rabbitmq headless 服务名称
	RabbitmqHeadlessServiceName = "bcs-rabbitmq-headless"
	// RemoteCommandServiceName remote command 服务名称
	RemoteCommandServiceName = "bcs-remote-command"
	// ResourceManagerServiceName resource manager 服务名称
	ResourceManagerServiceName = "bcs-resource-manager"
	// StorageServiceName storage 服务名称
	StorageServiceName = "bcs-storage"
	// StorageDataControllerServiceName storage data controller 服务名称
	StorageDataControllerServiceName = "bcs-storage-data-controller"
	// UiServiceName ui 服务名称
	UiServiceName = "bcs-ui"
	// UserManagerServiceName user manager 服务名称
	UserManagerServiceName = "bcs-user-manager"
	// WebconsoleServiceName webconsole 服务名称
	WebconsoleServiceName = "bcs-webconsole"
	// ZookeeperServiceName zookeeper 服务名称
	ZookeeperServiceName = "bcs-zookeeper"
	// ZookeeperHeadlessServiceName zookeeper headless 服务名称
	ZookeeperHeadlessServiceName = "bcs-zookeeper-headless"

	// 服务的端口

	// ServiceHTTPPort 服务http端口
	ServiceHTTPPort = 8080
	// ServiceGrpcPort 服务grpc端口
	ServiceGrpcPort = 8081
	// ServiceMetricsPort 服务metrics端口
	ServiceMetricsPort = 8082
)

// UseServiceDiscovery 检查是否使用服务发现
func UseServiceDiscovery() bool {
	return os.Getenv(EnvUseServiceDiscoveryKey) == "true"
}
