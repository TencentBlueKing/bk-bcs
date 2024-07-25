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

// Package bcs xxx
package bcs

import (
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"

	bcsoptions "github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
)

// GetStorageService returns storage InnerService object for discovery.
// in container deployment mode, get storage endpoints from configuration directly
func GetStorageService(zkHosts string, bcsTLSConfig bcsoptions.TLS, customIPStr string, isExternal bool) (*InnerService,
	*RegisterDiscover.RegDiscover, error) {
	customEndpoints := strings.Split(customIPStr, ",")
	storageService := NewInnerService(types.BCS_MODULE_STORAGE, nil, customEndpoints, isExternal)
	storageService.update(customEndpoints, bcsTLSConfig)
	return storageService, nil, nil
}

// GetNetService returns netservice InnerService object for discovery.
func GetNetService(zkHosts string, bcsTLSConfig bcsoptions.TLS, customIPStr string, isExternal bool) (*InnerService,
	*RegisterDiscover.RegDiscover, error) {
	discovery := RegisterDiscover.NewRegDiscoverEx(zkHosts, 5*time.Second)
	if err := discovery.Start(); err != nil {
		return nil, nil, fmt.Errorf("get netservice from ZK failed, %+v", err)
	}

	// e.g.
	// zk: 127.0.0.11
	// zknode: bcs/services/endpoints/netservice
	path := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_NETSERVICE)
	eventChan, err := discovery.DiscoverService(path)
	if err != nil {
		_ = discovery.Stop()
		return nil, nil, fmt.Errorf("discover netservice failed, %+v", err)
	}
	var customEndpoints []string
	if len(customIPStr) != 0 {
		customEndpoints = strings.Split(customIPStr, ",")
	}
	netService := NewInnerService(types.BCS_MODULE_NETSERVICE, eventChan, customEndpoints, isExternal)
	go netService.Watch(bcsTLSConfig) // nolint

	return netService, discovery, nil
}
