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

package modules

const (
	// bcs-services module list

	// BCSModuleUserManager for bcs-user-manager
	BCSModuleUserManager = "usermanager"
	// BCSModuleClusterManager for bcs-cluster-manager
	BCSModuleClusterManager = "clustermanager"
	// BCSModuleGatewayDiscovery for bcs-gateway-discovery
	BCSModuleGatewayDiscovery = "gatewaydiscovery"
	// BCSModuleNetworkdetection for bcs-network-detection
	BCSModuleNetworkdetection = "networkdetection"
	// BCSModuleBKCMDBSynchronizer for bcs-bkcmdb-synchronizer
	BCSModuleBKCMDBSynchronizer = "bkcmdb-synchronizer"
	// BCSModuleStorage for bcs-storage
	BCSModuleStorage = "storage"
	// BCSModuleIPService for bcs-ipservice
	BCSModuleIPService = "ipservice"
	// BCSModuleNetService for bcs-netservice
	BCSModuleNetService = "netservice"
	// BCSModuleDNS for bcs-dns
	BCSModuleDNS = "dns"
	// BCSModuleMetricService for bcs-metricservice
	BCSModuleMetricService = "metricservice"
	// BCSModuleMetricCollector for bcs-metriccollector
	BCSModuleMetricCollector = "metriccollector"
	//end of bcs-services module list

	//bcs mesos module list

	//BCSModuleMesosdriver for bcs-mesos-driver
	BCSModuleMesosdriver = "mesosdriver"
	// BCSModuleMesoswatch for bcs-mesos-watch
	BCSModuleMesoswatch = "mesosdatawatch"
	// BCSModuleMesosWebconsole for bcs-mesoswebconcole
	BCSModuleMesosWebconsole = "mesoswebconsole"
	// BCSModuleScheduler for bcs-scheduler
	BCSModuleScheduler = "scheduler"
	//end of bcs mesos module list

	// BCSModuleKubeagent for bcs-kube-agent
	BCSModuleKubeagent = "kubeagent"
	// BCSModuleKubewatch for bcs-k8s-watch
	BCSModuleKubewatch = "kubedatawatch"

	//mode for mesosdriver & kubeagent

	// BCSConnectModeTunnel mode for tunnel
	BCSConnectModeTunnel = "websockettunnel"
	// BCSConnectModeDirect mode for direct connection
	BCSConnectModeDirect = "direct"
)
