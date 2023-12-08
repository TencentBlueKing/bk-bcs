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

package webhookserver

import (
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func (s *Server) portAllocate(portEntryList []*portEntry) ([]*networkextensionv1.PortPoolItemStatus,
	[][]portpoolcache.AllocatedPortItem, error) {
	portPoolItemStatusList := make([]*networkextensionv1.PortPoolItemStatus, 0, len(portEntryList))
	portItemListArr := make([][]portpoolcache.AllocatedPortItem, 0, len(portEntryList))

	s.poolCache.Lock()
	defer s.poolCache.Unlock()
	for _, portEntry := range portEntryList {
		poolKey := getPoolKey(portEntry.PoolName, portEntry.PoolNamespace)
		var portPoolItemStatus *networkextensionv1.PortPoolItemStatus
		var err error
		// deal with TCP_UDP protocol
		// for TCP_UDP protocol, one container port needs both TCP listener port and UDP listener port
		if portEntry.Protocol == constant.PortPoolPortProtocolTCPUDP {
			var cachePortItemMap map[string]portpoolcache.AllocatedPortItem
			portPoolItemStatus, cachePortItemMap, err = s.poolCache.AllocateAllProtocolPortBinding(poolKey, portEntry.ItemName)
			if err != nil {
				s.cleanAllocatedResource(portItemListArr)
				return nil, nil, errors.Errorf("allocate protocol %s port from pool %s failed, err %s",
					portEntry.Protocol, poolKey, err.Error())
			}
			var tmpPortItemList []portpoolcache.AllocatedPortItem
			for _, cachePortItem := range cachePortItemMap {
				tmpPortItemList = append(tmpPortItemList, cachePortItem)
			}
			portItemListArr = append(portItemListArr, tmpPortItemList)
			portPoolItemStatusList = append(portPoolItemStatusList, portPoolItemStatus)
		} else {
			// deal with TCP protocol and UDP protocol
			var cachePortItem portpoolcache.AllocatedPortItem
			portPoolItemStatus, cachePortItem, err = s.poolCache.AllocatePortBinding(poolKey, portEntry.Protocol,
				portEntry.ItemName)
			if err != nil {
				s.cleanAllocatedResource(portItemListArr)
				return nil, nil, errors.Errorf("allocate protocol %s port from pool %s failed, err %s",
					portEntry.Protocol, poolKey, err.Error())
			}
			portItemListArr = append(portItemListArr, []portpoolcache.AllocatedPortItem{cachePortItem})
			portPoolItemStatusList = append(portPoolItemStatusList, portPoolItemStatus)
		}
	}

	return portPoolItemStatusList, portItemListArr, nil
}
