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

package types

const (
	BCS_SERV_BASEPATH          = "/bcs/services/endpoints"
	BCS_MODULE_APISERVER       = "apiserver"
	BCS_MODULE_USERMGR         = "usermanager"
	BCS_MODULE_CCAPI           = "ccapi"
	BCS_MODULE_MESOSDATAWATCH  = "mesosdatawatch"
	BCS_MODULE_KUBEDATAWATCH   = "kubedatawatch"
	BCS_MODULE_MESOSDRIVER     = "mesosdriver"
	BCS_MODULE_KUBERNETEDRIVER = "kubernetedriver"
	BCS_MODULE_SCHEDULER       = "scheduler"
	BCS_MODULE_CLUSTERKEEPER   = "clusterkeeper"
	BCS_MODULE_HEALTH          = "health"
	BCS_MODULE_LOADBALANCE     = "loadbalance"
	BCS_MODULE_Check           = "check"
	BCS_MODULE_NETSERVICE      = "netservice"
	BCS_MODULE_DNS             = "dns"
	BCS_MODULE_STORAGE         = "storage"
	BCS_MODULE_DISCOVERY       = "discovery"
	BCS_MODULE_METRICSERVICE   = "metricservice"
	BCS_MODULE_METRICCOLLECTOR = "metriccollector"
	BCS_MODULE_EXPORTER        = "exporter"
	BCS_MODULE_DCSERVER        = "dcserver"
	BCS_MODULE_DCCLINET        = "dcclient"
	BCS_MODULE_QCLOUDCLB       = "qcloudclb"
	BCS_MODULE_MESOSSLAVE      = "mesosslave"
	BCS_MODULE_IPSERVICE       = "ipservice"
	BCS_MODULE_MESOSADAPTER    = "mesosadapter"

	//bcstest 2018.11.07
	BCS_MODULE_K8SAPISERVER     = "kubernetedriver"
	BCS_MODULE_MESOSAPISERVER   = "mesosdriver"
	BCS_MODULE_NETWORKDETECTION = "networkdetection"

	//bcs-api-gateway refactor 2020-04-10
	BCS_MODULE_KUBEAGENT        = "kubeagent"
	BCS_MODULE_USERMANAGER      = "usermanager"
	BCS_MODULE_GATEWAYDISCOVERY = "gatewaydiscovery"
	BCS_MODULE_MESOSWEBCONSOLE  = "mesoswebconsole"

	// for bcs-bkcmdb-synchronizer
	BCS_MODULE_BKCMDB_SYNCHRONIZER = "bkcmdb-synchronizer"
)

var (
	// BCS_PROC_LIST module list information
	BCS_PROC_LIST = []string{
		BCS_MODULE_APISERVER,
		BCS_MODULE_CCAPI,
		BCS_MODULE_MESOSDATAWATCH,
		BCS_MODULE_KUBEDATAWATCH,
		BCS_MODULE_MESOSDRIVER,
		BCS_MODULE_KUBERNETEDRIVER,
		BCS_MODULE_SCHEDULER,
		BCS_MODULE_CLUSTERKEEPER,
		BCS_MODULE_HEALTH,
		BCS_MODULE_LOADBALANCE,
		BCS_MODULE_Check,
		BCS_MODULE_NETSERVICE,
		BCS_MODULE_DNS,
		BCS_MODULE_STORAGE,
		BCS_MODULE_DISCOVERY,
		BCS_MODULE_METRICSERVICE,
		BCS_MODULE_METRICCOLLECTOR,
		BCS_MODULE_EXPORTER,
		BCS_MODULE_DCSERVER,
		BCS_MODULE_DCCLINET,
		BCS_MODULE_QCLOUDCLB,
		BCS_MODULE_MESOSSLAVE,
		BCS_MODULE_IPSERVICE,
		BCS_MODULE_MESOSADAPTER,
		BCS_MODULE_KUBEAGENT,
	}
)

// bcss modules
const (
	BCSS_SERV_BASEPATH        = "/bcss/services/endpoints"
	BCSS_MODULE_CONTAINERWARE = "containerware"
	BCSS_MODULE_PROXY         = "proxy"
	BCSS_MODULE_CONSOLESERVER = "consoleserver"
	BCSS_MODULE_MESHAPI       = "bcss-mesh-api"
)

//ServerInfo base server information
type ServerInfo struct {
	IP           string `json:"ip"`
	IPv6         string `json:"ipv6"`
	Port         uint   `json:"port"`
	GrpcPort     uint   `json:"grpc_port"`
	MetricPort   uint   `json:"metric_port"`
	HostName     string `json:"hostname"`
	Scheme       string `json:"scheme"` //http, https
	Version      string `json:"version"`
	Cluster      string `json:"cluster"`
	Pid          int    `json:"pid"`
	ExternalIp   string `json:"external_ip"`
	ExternalIPv6 string `json:"external_ipv6"`
	ExternalPort uint   `json:"external_port"`
}

//APIServInfo apiserver information
type APIServInfo struct {
	ServerInfo
}

//AuthServInfo auth server information
type AuthServInfo struct {
	ServerInfo
}

//CCAPIServInfo ccapi server information
type CCAPIServInfo struct {
	ServerInfo
}

//RouteServInfo route server information
type RouteServInfo struct {
	ServerInfo
}

//MesosDataWatchServInfo mesos-data-watch server information
type MesosDataWatchServInfo struct {
	ServerInfo
	//Cluster string `json:"cluster"`
}

//MesosDriverServInfo mesosdriver server information
type MesosDriverServInfo struct {
	ServerInfo
	//Cluster string `json:"cluster"`
}

//BcsUserMgrServInfo bcs-user-manager server information
type BcsUserMgrServInfo struct {
	ServerInfo
}

//NetworkDetectionServInfo netwrok-detection server information
type NetworkDetectionServInfo struct {
	ServerInfo
}

type DCServInfo struct {
	ServerInfo
}

//KuberneteDataWatchServInfo kubernete-data-watch server information
type KuberneteDataWatchServInfo struct {
	ServerInfo
	//Cluster string `json:"cluster"`
}

//KuberneteDriverServInfo kubernetedriver server information
type KuberneteDriverServInfo struct {
	ServerInfo
}

//SaDriverServInfo sa driver server information
type SaDriverServInfo struct {
	ServerInfo
	//Cluster string `json:"cluster"`
}

//SchedulerServInfo scheduler server information
type SchedulerServInfo struct {
	ServerInfo
}

// MesosServInfo mesos server
type MesosServInfo struct {
	ServerInfo
}

// KubeNodeInfo node infor
type KubeNodeInfo struct {
	ServerInfo
}

// BcsHealthInfo health server
type BcsHealthInfo struct {
	ServerInfo
}

// BcsCheckInfo check server
type BcsCheckInfo struct {
	ServerInfo
}

// BcsStorageInfo storage server
type BcsStorageInfo struct {
	ServerInfo
}

// BcsK8sApiserverInfo apiserver
type BcsK8sApiserverInfo struct {
	ServerInfo
	CaCertData string //certificates
	UserToken  string //user token
}

// BcsMesosApiserverInfo mesos driver
type BcsMesosApiserverInfo struct {
	ServerInfo
}

// ClusterEndpoints mesos node info
type ClusterEndpoints struct {
	MesosSchedulers []SchedulerServInfo `json:"mesosscheduler,omitempty"`
	MesosMasters    []MesosServInfo     `json:"mesosmaster,omitempty"`
	KubeNodes       []KubeNodeInfo      `json:"kubenodes,omitempty"`
}

//ClusterKeeperServInfo cluster keeper server information
type ClusterKeeperServInfo struct {
	ServerInfo
}

//NetServiceInfo for bcs-netservice
type NetServiceInfo struct {
	ServerInfo
}

//LoadBalanceInfo for bcs-loadBalance
type LoadBalanceInfo struct {
	ServerInfo
}

//DNSInfo for bcs-loadBalance
type DNSInfo struct {
	ServerInfo
}

//DiscoveryInfo for bcs-loadBalance
type DiscoveryInfo struct {
	ServerInfo
}

// MetricServiceInfo for bcs-metricservice
type MetricServiceInfo struct {
	ServerInfo
}

// MetricCollectorInfo for bcs-metriccollector
type MetricCollectorInfo struct {
	ServerInfo
}

// DataExporterInfo for metric data
type DataExporterInfo struct {
	ServerInfo
}

// ContainerWareInfo for container ware
type ContainerWareInfo struct {
	ServerInfo
}

//AWSELBInfo for aws elb
type AWSELBInfo struct {
	ServerInfo
}

//QcloudCLBInfo for qcloud clb
type QcloudCLBInfo struct {
	ServerInfo
}

// IPServiceInfo for module ipservice
type IPServiceInfo struct {
	ServerInfo
}

// ProxyInfo for proxy module
type ProxyInfo struct {
	ServerInfo
}

//ConsoleManagerInfo for console manager
type ConsoleManagerInfo struct {
	ServerInfo
}
