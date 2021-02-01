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

package config

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/apis/bk-bcs/v1"
)

// ManagerConfig is config of k8s.LogManager
type ManagerConfig struct {
	CollectionConfigs []CollectionConfig
	BcsAPIConfig      bcsapi.Config
	CAFile            string
	SystemDataID      string
	KubeConfig        string // manager 所在集群的kubeconfig
	BkUsername        string
	BkAppCode         string
	BkAppSecret       string
	BkBizID           int
	StopCh            chan struct{}
	Ctx               context.Context
}

// CollectionConfig defines some customed information of log collection.
// For example, customed dataid of some Cluster.
type CollectionConfig struct {
	//Config Spec.
	ConfigName      string                 `json:"config_name"`
	ConfigNamespace string                 `json:"config_namespace"`
	ConfigSpec      bcsv1.BcsLogConfigSpec `json:"config_spec"`
	// ClusterIDs comma split clusterid
	ClusterIDs string `json:"cluster_ids"`
}

// CollectionFilterConfig is filter of getting and deleting bcslogconfigs
type CollectionFilterConfig struct {
	ConfigName      string
	ConfigNamespace string
	ClusterIDs      string
}

// ControllerConfig is config for cluster specified bcslogconfigs controller
type ControllerConfig struct {
	Credential      *bcsapi.ClusterCredential
	CAFile          string
	BcsAPIHost      string
	BcsAPIAuthToken string
}

// APIServerConfig is config for APIServer of bcs-log-manager
type APIServerConfig struct {
	conf.ZkConfig
	APICerts      conf.CertConfig
	EtcdCerts     conf.CertConfig
	Host          string
	Port          uint
	BKDataAPIHost string
	EtcdHosts     []string
}
