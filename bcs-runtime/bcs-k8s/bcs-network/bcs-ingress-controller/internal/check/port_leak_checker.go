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

package check

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
)

// PortLeakChecker 校验端口泄漏
type PortLeakChecker struct {
	cli       client.Client
	poolCache *portpoolcache.Cache
}

// NewPortLeakChecker return new port leakchecker
func NewPortLeakChecker(cli client.Client, poolCache *portpoolcache.Cache) *PortLeakChecker {
	return &PortLeakChecker{
		cli:       cli,
		poolCache: poolCache,
	}
}

// Run start check
func (p *PortLeakChecker) Run() {
	st := time.Now()
	p.poolCache.Lock()
	defer blog.Infof("port leak check cost: %f", time.Since(st).Seconds())
	defer p.poolCache.Unlock()
	for _, pool := range p.poolCache.GetPortPoolMap() {
		for _, item := range pool.ItemList {
			for _, list := range item.PortListMap {
				for _, port := range list.Ports {
					if port.IsUsed() {
						// 端口被占用但是没有占用者信息， 可能发生了端口泄漏
						// Ref是在PortBinding创建时被设置的， 可能存在Port被webhook分配但PortBinding还没来得及创建的情况。
						if port.RefName == "" && port.RefNamespace == "" && port.RefType == "" {
							if time.Since(port.RefStartTime) > time.Minute*30 {
								// 超过一定时间没有占用者信息，认为已经发生了端口泄漏
								blog.Warnf("pool cache leaked, port[%v] released", port)
								p.poolCache.ReleasePortBinding(pool.PoolKey, item.ItemStatus.ItemName, list.Protocol,
									port.StartPort, port.EndPort)
							}
						}
					}
				}
			}
		}
	}
}
