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

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
)

const (
	//ActionAdd add event
	ActionAdd = "Add"
	//ActionDelete delete event
	ActionDelete = "Delete"
	//ActionUpdate update event
	ActionUpdate = "Update"
)

//BcsSyncData holder for sync data
type BcsSyncData struct {
	DataType string      //data type: reflect.TypeOf(Item).Name()
	Action   string      //operation, like Add, Delete, Update
	Item     interface{} //SyncData, data is Endpoint, Service, Pod
}

//CmdConfig hold all command line config item
type CmdConfig struct {
	ClusterID   string
	ClusterInfo string
	IsExternal  bool
	CAFile      string
	CertFile    string
	KeyFile     string
	PassWord    string

	RegDiscvSvr            string
	Address                string
	ApplicationThreadNum   int
	TaskgroupThreadNum     int
	ExportserviceThreadNum int
	DeploymentThreadNum    int

	MetricPort uint

	ServerCAFile   string
	ServerCertFile string
	ServerKeyFile  string
	ServerPassWord string
	ServerSchem    string

	KubeConfig  string
	StoreDriver string

	// NetServiceZK is zookeeper address config for netservice discovery,
	// reuse RegDiscvSvr by default.
	NetServiceZK string

	// Etcd etcd options for service registry and discovery
	Etcd registry.CMDOptions

	// StorageAddresses address for bcs-storage
	StorageAddresses []string
}

const (
	//ApplicationChannelPrefix prefix for event post channel
	ApplicationChannelPrefix = "Application_"
	//TaskgroupChannelPrefix prefix for event post channel
	TaskgroupChannelPrefix = "TaskGroup_"
	//ExportserviceChannelPrefix prefix for event post channel
	ExportserviceChannelPrefix = "Exportservice_"
	// DeploymentChannelPrefix deployment prefix for post channel
	DeploymentChannelPrefix = "Deployment_"
)
