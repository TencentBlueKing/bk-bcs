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

// Package conflicthandler 判断新增ingress/portpool是否和集群内现有ingress/portpool产生端口冲突
package conflicthandler

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

const (
	// 限制并发检查的ingress数量
	concurrentIngressCheckLimit = 5
)

// ConflictHandler handle ingress/portPool conflict
type ConflictHandler struct {
	conflictCheckOpen bool // if false, skip all conflict checking
	defaultRegion     string
	IsTCPUDPPortReuse bool

	k8sClient        client.Client
	ingressConverter *generator.IngressConverter

	eventer record.EventRecorder
}

// NewConflictHandler return new conflictHandler
func NewConflictHandler(conflictCheckOpen bool, isTCPUDPPortReuse bool, defaultRegion string, k8sCli client.Client,
	igc *generator.IngressConverter, eventer record.EventRecorder) *ConflictHandler {
	return &ConflictHandler{
		conflictCheckOpen: conflictCheckOpen,
		defaultRegion:     defaultRegion,
		IsTCPUDPPortReuse: isTCPUDPPortReuse,
		k8sClient:         k8sCli,
		ingressConverter:  igc,
		eventer:           eventer,
	}
}

// IsIngressConflict return true if new ingress conflict with existed ingress/portpool
func (h *ConflictHandler) IsIngressConflict(ingress *networkextensionv1.Ingress) error {
	if !h.conflictCheckOpen {
		return nil
	}
	res, err := h.getIngressResourceMap(ingress)
	if err != nil {
		return err
	}
	return h.checkConflict(res, constant.KindIngress, ingress.GetNamespace(), ingress.GetName())
}

// IsPortPoolConflict return true if new port pool conflict with existed ingress/portpool
func (h *ConflictHandler) IsPortPoolConflict(pool *networkextensionv1.PortPool) error {
	if !h.conflictCheckOpen {
		return nil
	}
	res, err := h.getPortPoolResourceMap(pool)
	if err != nil {
		return err
	}
	return h.checkConflict(res, constant.KindPortPool, pool.GetNamespace(), pool.GetName())
}

func (h *ConflictHandler) getIngressResourceMap(ingress *networkextensionv1.Ingress) (map[string]*resource, error) {
	lbObjs, err := h.ingressConverter.GetIngressLoadBalancers(ingress)
	if err != nil {
		return nil, err
	}
	// region:lbID => resource
	usedResource := make(map[string]*resource)

	res := newResource()
	for _, rule := range ingress.Spec.Rules {
		res.usedPort[rule.Port] = portStruct{
			val:       rule.Port,
			Protocols: []string{rule.Protocol},
		}
	}
	for _, mapping := range ingress.Spec.PortMappings {
		segmentLen := mapping.SegmentLength
		if segmentLen == 0 {
			segmentLen = 1
		}
		istart := mapping.StartPort + mapping.StartIndex*segmentLen
		iend := mapping.StartPort + mapping.EndIndex*segmentLen
		res.usedPortSegment = append(res.usedPortSegment, portSegment{
			Start:     istart,
			End:       iend,
			Protocols: []string{mapping.Protocol},
		})
	}
	for _, lbObj := range lbObjs {
		usedResource[common.BuildRegionName(lbObj.Region, lbObj.LbID)] = res
	}
	return usedResource, nil
}

func (h *ConflictHandler) getPortPoolResourceMap(pool *networkextensionv1.PortPool) (map[string]*resource, error) {
	// region:lbID => resource
	usedResource := make(map[string]*resource)

	for _, item := range pool.Spec.PoolItems {
		seg := portSegment{
			Start:     int(item.StartPort),
			End:       int(item.EndPort),
			Protocols: common.GetPortPoolItemProtocols(item.Protocol),
		}
		for _, lbID := range item.LoadBalancerIDs {
			regionID, err := getRegionLBID(lbID, h.defaultRegion)
			if err != nil {
				return nil, err
			}

			res, ok := usedResource[regionID]
			if !ok {
				res = newResource()
			}
			res.usedPortSegment = append(res.usedPortSegment, seg)
			usedResource[regionID] = res
		}
	}

	return usedResource, nil
}

func (h *ConflictHandler) checkConflict(newRes map[string]*resource, newKind, newNamespace, newName string) error {
	err := h.checkConflictWithIngress(newRes, newKind, newNamespace, newName)
	if err != nil {
		return err
	}

	err = h.checkConflictWithPortPool(newRes, newKind, newNamespace, newName)
	if err != nil {
		return err
	}

	return nil
}

// checkConflictWithIngress check if newRes conflict with current ingress
func (h *ConflictHandler) checkConflictWithIngress(requiredResMap map[string]*resource, newKind, newNamespace,
	newName string) error {
	ingressList := &networkextensionv1.IngressList{}
	if err := h.k8sClient.List(context.TODO(), ingressList); err != nil {
		err = errors.Wrapf(err, "k8s api server list failed")
		return err
	}

	// 可能存在问题： 当缓存失效时，遍历所有ingress上的lbID，可能导致对腾讯云API的访问高峰，触发qps限制。
	// 但如果不在GetIngressLoadBalancer方法下加多协程，可能导致一个lb上挂多个clb id时查询超时。
	workGroup := errgroup.Group{}
	// 设置limit避免同时检查多个ingress导致缓存失效时describeLoadBalancer请求过多
	workGroup.SetLimit(concurrentIngressCheckLimit)
	for _, ingress := range ingressList.Items {
		ingress := ingress
		workGroup.Go(func() error {
			if newKind == constant.KindIngress && newNamespace == ingress.GetNamespace() && newName == ingress.GetName() {
				return nil
			}

			// 如果集群中某个ingress的clb失效，可能导致整个集群无法继续工作。 优先跳过检查
			lbObjs, err := h.ingressConverter.GetIngressLoadBalancers(&ingress)
			if err != nil {
				errMsg := fmt.Sprintf("get LoadBalancers for ingress '%s/%s failed, "+
					"please check ingress and its load balancer', err : %s", ingress.GetNamespace(),
					ingress.GetName(), err.Error())
				blog.Errorf(errMsg)
				h.eventer.Event(&ingress, k8scorev1.EventTypeWarning, "ingress get lb failed", errMsg)
				return nil
			}

			var ingressResMap map[string]*resource
			for _, lbObj := range lbObjs {
				regionID := common.BuildRegionName(lbObj.Region, lbObj.LbID)
				requiredRes, ok := requiredResMap[regionID]
				if !ok {
					continue
				}

				// load resource only when need
				if ingressResMap == nil {
					ingressResMap, err = h.getIngressResourceMap(&ingress)
					if err != nil {
						return err
					}
				}
				ingressRes := ingressResMap[regionID]
				if requiredRes.IsConflict(h.IsTCPUDPPortReuse, ingressRes) {
					return errors.Errorf(constant.PortConflictMsg, constant.KindIngress,
						ingress.GetNamespace(), ingress.GetName(), regionID)
				}

			}
			return nil
		})
	}

	return workGroup.Wait()
}

func (h *ConflictHandler) checkConflictWithPortPool(requireResourceMap map[string]*resource, newKind, newNamespace,
	newName string) error {
	portPoolList := &networkextensionv1.PortPoolList{}
	if err := h.k8sClient.List(context.TODO(), portPoolList); err != nil {
		err = errors.Wrapf(err, "k8s api server list failed")
		return err
	}

	for _, portPool := range portPoolList.Items {
		if newKind == constant.KindPortPool && newNamespace == portPool.GetNamespace() && newName == portPool.
			GetName() {
			continue
		}

		poolResourceMap, err := h.getPortPoolResourceMap(&portPool)
		if err != nil {
			return err
		}

		for regionID, res := range poolResourceMap {
			if requireRes, ok := requireResourceMap[regionID]; ok {
				if requireRes.IsConflict(h.IsTCPUDPPortReuse, res) {
					return errors.Errorf(constant.PortConflictMsg, constant.KindPortPool,
						portPool.GetNamespace(), portPool.GetName(), regionID)
				}
			}
		}
	}
	return nil
}
