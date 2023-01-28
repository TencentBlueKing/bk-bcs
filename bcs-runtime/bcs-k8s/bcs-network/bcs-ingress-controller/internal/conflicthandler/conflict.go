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

package conflicthandler

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// ConflictHandler handle ingress/portPool conflict
type ConflictHandler struct {
	conflictCheckOpen bool // if false, skip all conflict checking
	defaultRegion     string
	IsTCPUDPPortReuse bool

	k8sClient        client.Client
	ingressConverter *generator.IngressConverter
}

// NewConflictHandler return new conflictHandler
func NewConflictHandler(conflictCheckOpen bool, IsTCPUDPPortReuse bool, defaultRegion string, k8sCli client.Client,
	igc *generator.IngressConverter) *ConflictHandler {
	return &ConflictHandler{
		conflictCheckOpen: conflictCheckOpen,
		defaultRegion:     defaultRegion,
		IsTCPUDPPortReuse: IsTCPUDPPortReuse,
		k8sClient:         k8sCli,
		ingressConverter:  igc,
	}
}

// IsIngressConflict return true if new ingress conflict with existed ingress/portpool
func (h *ConflictHandler) IsIngressConflict(ingress *networkextensionv1.Ingress) (bool, string, error) {
	if !h.conflictCheckOpen {
		return false, "", nil
	}
	res, err := h.getIngressResourceMap(ingress)
	if err != nil {
		return false, "", err
	}
	return h.checkConflict(res, constant.KindIngress, ingress.GetNamespace(), ingress.GetName())
}

// IsPortPoolConflict return true if new port pool conflict with existed ingress/portpool
func (h *ConflictHandler) IsPortPoolConflict(pool *networkextensionv1.PortPool) (bool, string, error) {
	if !h.conflictCheckOpen {
		return false, "", nil
	}
	res, err := h.getPortPoolResourceMap(pool)
	if err != nil {
		return false, "", err
	}
	return h.checkConflict(res, constant.KindPortPool, pool.GetNamespace(), pool.GetName())
}

func (h *ConflictHandler) getIngressResourceMap(ingress *networkextensionv1.Ingress) (map[string]*resource, error) {
	lbObjs, err := h.ingressConverter.GetIngressLoadbalances(ingress)
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

func (h *ConflictHandler) checkConflict(newRes map[string]*resource, newKind, newNamespace, newName string) (bool,
	string, error) {
	conflict, msg, err := h.checkConflictWithIngress(newRes, newKind, newNamespace, newName)
	if err != nil {
		return false, "", err
	}
	if conflict {
		return true, msg, nil
	}

	conflict, msg, err = h.checkConflictWithPortPool(newRes, newKind, newNamespace, newName)
	if err != nil {
		return false, "", err
	}
	if conflict {
		return true, msg, nil
	}

	return false, "", nil
}

// checkConflictWithIngress check if newRes conflict with current ingress
func (h *ConflictHandler) checkConflictWithIngress(requiredResMap map[string]*resource, newKind, newNamespace,
	newName string) (bool, string, error) {
	ingressList := &networkextensionv1.IngressList{}
	if err := h.k8sClient.List(context.TODO(), ingressList); err != nil {
		err = errors.Wrapf(err, "k8s api server list failed")
		return false, "", err
	}

	for _, ingress := range ingressList.Items {
		if newKind == constant.KindIngress && newNamespace == ingress.GetNamespace() && newName == ingress.GetName() {
			continue
		}

		lbObjs, err := h.ingressConverter.GetIngressLoadbalances(&ingress)
		if err != nil {
			return false, "", err
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
					return false, "", err
				}
			}
			ingressRes := ingressResMap[regionID]
			if requiredRes.IsConflict(h.IsTCPUDPPortReuse, ingressRes) {
				return true, fmt.Sprintf(constant.PortConflictMsg, constant.KindIngress,
					ingress.GetNamespace(), ingress.GetName(), regionID), nil
			}

		}
	}
	return false, "", nil
}

func (h *ConflictHandler) checkConflictWithPortPool(requireResourceMap map[string]*resource, newKind, newNamespace,
	newName string) (bool, string, error) {
	portPoolList := &networkextensionv1.PortPoolList{}
	if err := h.k8sClient.List(context.TODO(), portPoolList); err != nil {
		err = errors.Wrapf(err, "k8s api server list failed")
		return false, "", err
	}

	for _, portPool := range portPoolList.Items {
		if newKind == constant.KindPortPool && newNamespace == portPool.GetNamespace() && newName == portPool.
			GetName() {
			continue
		}

		poolResourceMap, err := h.getPortPoolResourceMap(&portPool)
		if err != nil {
			return false, "", err
		}

		for regionID, res := range poolResourceMap {
			if requireRes, ok := requireResourceMap[regionID]; ok {
				if requireRes.IsConflict(h.IsTCPUDPPortReuse, res) {
					return true, fmt.Sprintf(constant.PortConflictMsg, constant.KindPortPool,
						portPool.GetNamespace(), portPool.GetName(), regionID), nil
				}
			}
		}
	}
	return false, "", nil
}
