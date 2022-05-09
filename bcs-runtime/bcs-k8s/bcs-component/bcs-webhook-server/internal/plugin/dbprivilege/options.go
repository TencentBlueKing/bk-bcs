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

package dbprivilege

import (
	"fmt"
)

// DbPrivOptions options for db privilege plugin
type DbPrivOptions struct {
	KubeMaster         string `json:"kube_master"`
	Kubeconfig         string `json:"kubeconfig"`
	NetworkType        string `json:"network_type"`
	EsbURL             string `json:"esb_url"`
	CDPGCSURL          string `json:"cdp_gcs_url"`
	AccessToken        string `json:"access_token"`
	InitContainerImage string `json:"init_container_image"`
}

// Validate validate options
func (dpo *DbPrivOptions) Validate() error {
	if len(dpo.NetworkType) == 0 {
		dpo.NetworkType = NetworkTypeOverlay
	}
	if dpo.NetworkType != NetworkTypeOverlay &&
		dpo.NetworkType != NetworkTypeUnderlay {
		return fmt.Errorf("invalid network_type %s", dpo.NetworkType)
	}
	if len(dpo.EsbURL) == 0 {
		return fmt.Errorf("esb_url cannot be empty")
	}
	if len(dpo.InitContainerImage) == 0 {
		return fmt.Errorf("init_container_image cannot be empty")
	}
	return nil
}
