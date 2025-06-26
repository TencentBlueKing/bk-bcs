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

package project

import (
	"crypto/tls"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"k8s.io/klog/v2"
)

var (
	clientConfig *bcsapi.ClientConfig
)

// SetClientConfig set bcs project client config
// disc nil 表示使用k8s 内置的service 进行服务访问
func SetClientConfig(tlsConfig *tls.Config, disc *discovery.ModuleDiscovery) {
	clientConfig = &bcsapi.ClientConfig{
		TLSConfig: tlsConfig,
		Discovery: disc,
	}
}

// GetClient get cm client by discovery
func GetClient(innerClientName string) (*ProjectClient, func(), error) {
	if clientConfig == nil {
		return nil, nil, bcsapi.ErrNotInited
	}
	var addr string
	if discovery.UseServiceDiscovery() {
		addr = fmt.Sprintf("%s:%d", discovery.ProjectManagerServiceName, discovery.ServiceGrpcPort)
	} else {
		if clientConfig.Discovery == nil {
			return nil, nil, fmt.Errorf("project manager module not enable discovery")
		}

		nodeServer, err := clientConfig.Discovery.GetRandomServiceNode()
		if err != nil {
			return nil, nil, err
		}
		addr = nodeServer.Address
	}
	klog.Infof("get project manager client with address: %s", addr)
	conf := &bcsapi.Config{
		Hosts:           []string{addr},
		TLSConfig:       clientConfig.TLSConfig,
		InnerClientName: innerClientName,
	}
	cli, closeCon := NewProjectManagerClient(conf)

	return cli, closeCon, nil
}
