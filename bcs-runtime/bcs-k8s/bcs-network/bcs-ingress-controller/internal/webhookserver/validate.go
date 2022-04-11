/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webhookserver

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/aws/aws-sdk-go-v2/aws/arn"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8sselection "k8s.io/apimachinery/pkg/selection"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Server) validatePortPool(newPool *networkextensionv1.PortPool) error {
	if err := s.checkPortPool(newPool); err != nil {
		return err
	}
	oldPool := &networkextensionv1.PortPool{}
	found := true
	if err := s.k8sClient.Get(context.Background(), k8stypes.NamespacedName{
		Name:      newPool.GetName(),
		Namespace: newPool.GetNamespace(),
	}, oldPool); err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("get port pool failed by %s/%s, err %s",
				newPool.GetName(), newPool.GetNamespace(), err.Error())
		}
		found = false
	}
	if found {
		if err := s.checkPortPoolChanges(newPool, oldPool); err != nil {
			return err
		}
	}
	if err := s.checkPortPoolConflicts(newPool); err != nil {
		return err
	}
	if err := s.checkPortPoolConflictWithIngress(newPool); err != nil {
		return err
	}
	return nil
}

func (s *Server) checkPortPool(newPool *networkextensionv1.PortPool) error {
	lbIDMap := make(map[string]string)
	itemNameMap := make(map[string]struct{})
	for _, item := range newPool.Spec.PoolItems {
		if err := item.Validate(); err != nil {
			return err
		}
		if item.SegmentLength == 0 || item.SegmentLength == 1 {
			if item.EndPort-item.StartPort > constant.MaxPortQuantityForEachLoadbalancer {
				return fmt.Errorf("too many port for item %s", item.ItemName)
			}
		} else {
			if (item.EndPort-item.StartPort)/item.SegmentLength > constant.MaxPortQuantityForEachLoadbalancer {
				return fmt.Errorf("too many port segment for item %s", item.ItemName)
			}
		}
		if _, ok := itemNameMap[item.ItemName]; ok {
			return fmt.Errorf("duplicated item name %s", item.ItemName)
		}
		itemNameMap[item.ItemName] = struct{}{}

		isARN := false
		for _, lbIDStr := range item.LoadBalancerIDs {
			lbID, err := getLbIDFromRegionID(lbIDStr)
			if err != nil {
				return fmt.Errorf("lbIDStr %s of item %s is invalid", lbIDStr, item.ItemName)
			}
			if itemName, ok := lbIDMap[lbID]; ok {
				return fmt.Errorf("lbID %s of item %s conflicts with id of item %s", lbID, item.ItemName, itemName)
			}
			if arn.IsARN(lbID) {
				isARN = true
			}
			lbIDMap[lbID] = item.ItemName
		}

		// check protocol
		if isARN && item.Protocol != constant.PortPoolPortProtocolTCP &&
			item.Protocol != constant.PortPoolPortProtocolUDP {
			return fmt.Errorf("protocol %s of item %s invalid", item.Protocol, item.ItemName)
		}
	}
	return nil
}

func (s *Server) checkPortPoolChanges(newPool, oldPool *networkextensionv1.PortPool) error {
	for _, newItem := range newPool.Spec.PoolItems {
		lbIDMap := make(map[string]string)
		for _, oldItem := range oldPool.Spec.PoolItems {
			if newItem.ItemName == oldItem.ItemName {
				if !reflect.DeepEqual(newItem.LoadBalancerIDs, oldItem.LoadBalancerIDs) ||
					newItem.SegmentLength != oldItem.SegmentLength ||
					newItem.StartPort != oldItem.StartPort {
					return fmt.Errorf(
						"loadBalancerIDs, startPort, endPort or segmentLength of item %s cannot be changeed",
						newItem.ItemName)
				}
				if newItem.EndPort < oldItem.EndPort {
					return fmt.Errorf("endPort of item %s can only be increased", newItem.ItemName)
				}
				continue
			}
			for _, lbIDStr := range oldItem.LoadBalancerIDs {
				lbID, err := getLbIDFromRegionID(lbIDStr)
				if err != nil {
					return fmt.Errorf("lbIDStr %s of item %s is invalid", oldItem.ItemName, lbIDStr)
				}
				lbIDMap[lbID] = oldItem.ItemName
			}
		}
		for _, lbIDStr := range newItem.LoadBalancerIDs {
			lbID, err := getLbIDFromRegionID(lbIDStr)
			if err != nil {
				return fmt.Errorf("lbIDStr %s of item %s is invalid", newItem.ItemName, lbIDStr)
			}
			if oldItemName, ok := lbIDMap[lbID]; ok {
				return fmt.Errorf("lbID %s of item %s conflicts with item %s", lbID, newItem.ItemName, oldItemName)
			}
		}
	}
	return nil
}

func (s *Server) checkPortPoolConflicts(newPool *networkextensionv1.PortPool) error {
	portPoolList := &networkextensionv1.PortPoolList{}
	if err := s.k8sClient.List(context.Background(), portPoolList, &client.ListOptions{}); err != nil {
		return fmt.Errorf("list port pool list failed, err %s", err.Error())
	}

	lbIDMap := make(map[string]string)
	for _, existedPool := range portPoolList.Items {
		if newPool.GetName() == existedPool.GetName() && newPool.GetNamespace() == existedPool.GetNamespace() {
			continue
		}
		for _, item := range existedPool.Spec.PoolItems {
			for _, lbIDStr := range item.LoadBalancerIDs {
				lbID, err := getLbIDFromRegionID(lbIDStr)
				if err != nil {
					return fmt.Errorf("lbIDStr %s of existed item %s is invalid", item.ItemName, lbIDStr)
				}
				lbIDMap[lbID] = existedPool.GetName() + "/" + existedPool.GetNamespace()
			}
		}
	}

	for _, newItem := range newPool.Spec.PoolItems {
		for _, lbIDStr := range newItem.LoadBalancerIDs {
			lbID, err := getLbIDFromRegionID(lbIDStr)
			if err != nil {
				return fmt.Errorf("lbIDStr %s of new item %s is invalid", newItem.ItemName, lbIDStr)
			}
			if existedPoolKey, ok := lbIDMap[lbID]; ok {
				return fmt.Errorf("lbID %s of new item %s is conflict with pool %s",
					lbID, newItem.ItemName, existedPoolKey)
			}
		}
	}
	return nil
}

// There is a time difference between the update of the remote api and the update of the local cache
// when conflicting ingress and portpool are created at the same time, there will be some unexpected behavior
func (s *Server) checkPortPoolConflictWithIngress(newPool *networkextensionv1.PortPool) error {
	existedListenerList := &networkextensionv1.ListenerList{}
	selector := k8slabels.NewSelector()
	requirement, err := k8slabels.NewRequirement(
		networkextensionv1.LabelKeyForPortPoolListener, k8sselection.NotIn, []string{networkextensionv1.LabelValueTrue})
	if err != nil {
		return fmt.Errorf("create new requirement failed, err %s", err.Error())
	}
	selector = selector.Add(*requirement)
	err = s.k8sClient.List(context.Background(), existedListenerList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return fmt.Errorf("list existed listener by selector %s failed, err %s", selector.String(), err.Error())
	}
	lbIDMap := make(map[string]struct{})
	for _, listener := range existedListenerList.Items {
		lbIDMap[listener.Spec.LoadbalancerID] = struct{}{}
	}
	for _, newItem := range newPool.Spec.PoolItems {
		for _, lbIDStr := range newItem.LoadBalancerIDs {
			lbID, err := getLbIDFromRegionID(lbIDStr)
			if err != nil {
				return fmt.Errorf("lbIDStr %s of new item %s is invalid", newItem.ItemName, lbIDStr)
			}
			if _, ok := lbIDMap[lbID]; ok {
				return fmt.Errorf("lbIDStr %s of new item %s conflicts with existed ingress listener",
					lbIDStr, newItem.ItemName)
			}
		}
	}
	return nil
}
