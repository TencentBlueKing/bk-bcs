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

// Package ingresscache 缓存service/workload到ingress的对应信息，提高ingress的调谐效率
package ingresscache

import (
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// Cache store how much ingress related to resource
type Cache struct {
	serviceCache  cacheUnit
	workloadCache cacheUnit
}

// NewDefaultCache return new cache
func NewDefaultCache() *Cache {
	return &Cache{
		serviceCache:  cacheUnit{ingressMap: make(map[string]mapset.Set)},
		workloadCache: cacheUnit{ingressMap: make(map[string]mapset.Set)},
	}
}

// Add 新增/更新Ingress时调用，确保缓存中为最新值
func (c *Cache) Add(ingress *networkextensionv1.Ingress) {
	ingressKey := buildIngressKey(ingress.GetNamespace(), ingress.GetName())
	for _, rule := range ingress.Spec.Rules {
		if common.InLayer4Protocol(rule.Protocol) {
			for _, route := range rule.Services {
				ns := route.ServiceNamespace
				if ns == "" {
					ns = ingress.GetNamespace()
				}
				svcKey := buildServiceKey(ns, route.ServiceName)
				c.serviceCache.add(svcKey, ingressKey)
			}
		}
		if common.InLayer7Protocol(rule.Protocol) {
			for _, httpRoute := range rule.Routes {
				for _, route := range httpRoute.Services {
					ns := route.ServiceNamespace
					if ns == "" {
						ns = ingress.GetNamespace()
					}
					svcKey := buildServiceKey(ns, route.ServiceName)
					c.serviceCache.add(svcKey, ingressKey)
				}
			}
		}
	}
	for _, mapping := range ingress.Spec.PortMappings {
		workloadKey := buildWorkloadKey(mapping.WorkloadKind, mapping.WorkloadNamespace, mapping.WorkloadName)
		c.workloadCache.add(workloadKey, ingressKey)
	}
}

// Remove 删除/更新Ingress时调用，确保缓存中不存储旧值
func (c *Cache) Remove(ingress *networkextensionv1.Ingress) {
	ingressKey := buildIngressKey(ingress.GetNamespace(), ingress.GetName())
	for _, rule := range ingress.Spec.Rules {
		if common.InLayer4Protocol(rule.Protocol) {
			for _, route := range rule.Services {
				ns := route.ServiceNamespace
				if ns == "" {
					ns = ingress.GetNamespace()
				}
				svcKey := buildServiceKey(ns, route.ServiceName)
				c.serviceCache.remove(svcKey, ingressKey)
			}
		}
		if common.InLayer7Protocol(rule.Protocol) {
			for _, httpRoute := range rule.Routes {
				for _, route := range httpRoute.Services {
					ns := route.ServiceNamespace
					if ns == "" {
						ns = ingress.GetNamespace()
					}
					svcKey := buildServiceKey(ns, route.ServiceName)
					c.serviceCache.remove(svcKey, ingressKey)
				}
			}
		}
	}

	for _, mapping := range ingress.Spec.PortMappings {
		workloadKey := buildWorkloadKey(mapping.WorkloadKind, mapping.WorkloadNamespace, mapping.WorkloadName)
		c.workloadCache.remove(workloadKey, ingressKey)
	}
}

// GetRelatedIngressOfService 获取service相关的ingress信息
func (c *Cache) GetRelatedIngressOfService(serviceNamespace, serviceName string) []IngressMeta {
	serviceKey := buildServiceKey(serviceNamespace, serviceName)
	return c.serviceCache.getRelatedIngress(serviceKey)
}

// GetRelatedIngressOfWorkload 获取workload相关的Ingress信息
func (c *Cache) GetRelatedIngressOfWorkload(workloadKind, workloadNamespace, workloadName string) []IngressMeta {
	workloadKey := buildWorkloadKey(workloadKind, workloadNamespace, workloadName)
	return c.workloadCache.getRelatedIngress(workloadKey)
}

type cacheUnit struct {
	// key => {ingressKey set}
	// key is service/workload's namespaced name build by util.go
	ingressMap map[string]mapset.Set
	sync.RWMutex
}

// add ingressKey to key
func (cu *cacheUnit) add(key string, ingressKey string) {
	cu.Lock()
	defer cu.Unlock()
	ingressSet, ok := cu.ingressMap[key]
	if !ok {
		ingressSet = mapset.NewThreadUnsafeSet()
	}
	ingressSet.Add(ingressKey)
	cu.ingressMap[key] = ingressSet
}

// remove ingressKey tp key
func (cu *cacheUnit) remove(key string, ingressKey string) {
	cu.Lock()
	defer cu.Unlock()
	ingressSet, ok := cu.ingressMap[key]
	if !ok {
		err := errors.Errorf("ingress[%s] not related to key[%s]", ingressKey, key)
		blog.Warnf("ingress cache release failed: %s", err.Error())
		return
	}
	ingressSet.Remove(ingressKey)

	if ingressSet.Cardinality() == 0 {
		delete(cu.ingressMap, key)
	}
}

// getRelatedIngress return related ingress of key
func (cu *cacheUnit) getRelatedIngress(key string) []IngressMeta {
	cu.RLock()
	defer cu.RUnlock()
	ingressSet, ok := cu.ingressMap[key]
	if !ok {
		return nil
	}

	ingressKeyList := make([]IngressMeta, 0, ingressSet.Cardinality())
	for ingressKey := range ingressSet.Iter() {
		splits := strings.Split(ingressKey.(string), "/")
		if len(splits) != 2 {
			blog.Errorf("%+v", errors.Errorf("unknown ingressKey '%s', in key '%s'", ingressKey, key))
			continue
		}
		ingressKeyList = append(ingressKeyList, IngressMeta{
			Namespace: splits[0],
			Name:      splits[1],
		})
	}

	return ingressKeyList
}
