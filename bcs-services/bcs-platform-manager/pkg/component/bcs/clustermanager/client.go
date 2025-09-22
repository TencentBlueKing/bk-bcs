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

// Package clustermanager xxx
package clustermanager

import (
	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	microRgt "go-micro.dev/v4/registry"
)

const (
	// ClusterManagerServiceName cluster manager service name
	ClusterManagerServiceName = "clustermanager.bkbcs.tencent.com"
)

// SetClientConifg set client config
func SetClientConifg(tlsConfig *tls.Config, microRgt microRgt.Registry) error {
	if !discovery.UseServiceDiscovery() {
		dis := discovery.NewModuleDiscovery(ClusterManagerServiceName, microRgt)
		err := dis.Start()
		if err != nil {
			return err
		}
		clustermanager.SetClientConfig(tlsConfig, dis)
	} else {
		clustermanager.SetClientConfig(tlsConfig, nil)
	}

	return nil
}

// Close close client connection
func Close(close func()) {
	if close != nil {
		close()
	}
}
