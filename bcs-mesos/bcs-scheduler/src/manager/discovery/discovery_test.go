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

package discovery

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
	"testing"
)

func newDiscoveryService() Discovery {
	discovery, err := NewDiscoveryService(util.SchedConfig{
		EtcdConf: registry.CMDOptions{
			Address: "xxx:2379",
			CA:      "./etcd/ca.pem",
			Cert:    "./etcd/client.pem",
			Key:     "./etcd/client-key.pem",
		},
	})

	if err != nil {
		return nil
	}

	return discovery
}

func TestDiscoveryService_GetMicroServiceByName(t *testing.T) {
	disService := newDiscoveryService()
	serviceNames := []ModuleName{AlertManager, "hellomanager"}

	for i := range serviceNames {
		service, err := disService.GetMicroServiceByName(serviceNames[i])
		if err != nil {
			t.Fatalf("GetMicroServiceByName[%s] failed: %v", serviceNames[i], err)
		}

		t.Log(service)
	}
}
