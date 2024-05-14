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

package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// SyncOption option of sync
type SyncOption struct {
	conf.LogConfig
	conf.FileConfig

	// Kubeconfig config of kubernetes
	Kubeconfig string `json:"kubeconfig" value:"" usage:"kubeconfig for kubernetes"` // nolint
	// IPMasqConfigmapName name of ip masq configmap
	IPMasqConfigmapName string `json:"ip_masq_configmap_name" value:"ip-masq-agent-config" usage:"configmap name of ip masq agent config"` // nolint
	// IPMasqConfigmapNamespace namespace of ip masq configmap
	IPMasqConfigmapNamespace string `json:"ip_masq_configmap_namespace" value:"kube-system" usage:"configmap namespace of ip masq agent config"` // nolint
	// SyncIntervalSecond sync interval second
	SyncIntervalSecond int `json:"sync_interval_second" value:"10" usage:"sync interval in second"`
}

// New new SyncOption
func New() *SyncOption {
	return &SyncOption{}
}

// Parse parse options
func Parse(opt *SyncOption) {
	conf.Parse(opt)

	if opt.SyncIntervalSecond < 1 {
		blog.Fatal("invalid sync interval second %d", opt.SyncIntervalSecond)
	}

	blog.Infof("get option %+v", opt)
}
