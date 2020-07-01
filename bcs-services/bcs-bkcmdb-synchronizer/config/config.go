/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

const (
	// TaskManagerTypePaas paas cc type task manager
	TaskManagerTypePaas = "paas"
	// TaskManagerTypeFile file type task manager
	TaskManagerTypeFile = "file"
)

// SyncOption sync option
type SyncOption struct {

	// configs for synchronizer server
	conf.ServiceConfig
	conf.CertConfig
	conf.MetricConfig
	conf.LogConfig
	conf.FileConfig
	ZkAddr string `json:"zk" value:"127.0.0.1:2181" usage:"zk address"`

	// configs for bcs storage
	StorageZk   string `json:"storage_zk" value:"127.0.0.1:2181" usage:"zk for bcs storage"`
	StorageCa   string `json:"storage_ca" value:"" usage:"ca cert for bcs storage"`
	StorageKey  string `json:"storage_key" value:"" usage:"key file for bcs storage"`
	StorageCert string `json:"storage_cert" value:"" usage:"cert file for bcs storage"`

	// configs for cmdb
	CmdbAddr       string `json:"cmdb_addr" value:"" usage:"cmdb address"`
	CmdbSupplierID string `json:"cmdb_supplier_id" value:"0" usage:"bk supplier id"`
	CmdbUser       string `json:"cmdb_user" value:"" usage:"bk_user for cmdb"`

	// configs for paas
	PaasAddr       string `json:"paas_addr" value:"" usage:"bk paas address"`
	PaasEnv        string `json:"paas_env" value:"" usage:"bk paas environment, optional [test, prod, debug, uat]"`
	PaasClusterEnv string `json:"paas_cluster_env" value:"" usage:"bcs cluster env in bk paas, optional [debug, prod]"`
	PaasAppCode    string `json:"paas_app_code" value:"" usage:"app code for bk paas"`
	PaasAppSecret  string `json:"paas_app_secret" value:"" usage:"app code for bk paas"`
	// interval for discover clusters
	ClusterPullInterval int64 `json:"cluster_pull_interval" value:"600" usage:"interval for discover clusters, seconds"`
	// interval for do full sync
	FullSyncInterval int64 `json:"full_sync_interval" value:"600" usage:"full sync interval, seconds"`

	// TaskManagerType task manager type
	TaskManagerType string `json:"task_manager_type" value:"paas" usage:"task manager type"`

	// TaskFile task file
	TaskFile string `json:"task_file" value:"" usage:"task file, only for file type task manager"`
}

// Load load from config file or command line
func (so *SyncOption) Load() {
	conf.Parse(so)
}

// Validate validate config
func (so *SyncOption) Validate() (bool, string) {
	if so.ClusterPullInterval < 10 || so.FullSyncInterval < 10 {
		return false, fmt.Sprintf("invalid full_sync_interval or full_sync_interval must")
	}
	if len(so.PaasAddr) == 0 {
		return false, fmt.Sprintf("paas_addr cannot be empty")
	}
	return true, ""
}
