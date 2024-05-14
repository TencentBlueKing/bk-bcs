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

// Package plugin_manager xxx
package plugin_manager // nolint

import (
	"k8s.io/client-go/rest"
)

// Config Options bcs log options
type Config struct {
	ClusterConfigs  []ClusterConfig
	InClusterConfig ClusterConfig
}

// ClusterConfig xxx
type ClusterConfig struct {
	Config     *rest.Config
	ClusterID  string
	BusinessID string
}

// Validate validate options
func (o *Config) Validate() error {
	// if len(o.KubeMaster) == 0 {
	//	return fmt.Errorf("kube_master cannot be empty")
	// }
	// if len(o.Kubeconfig) == 0 {
	//	return fmt.Errorf("kubeconfig cannot be empty")
	// }
	return nil
}
