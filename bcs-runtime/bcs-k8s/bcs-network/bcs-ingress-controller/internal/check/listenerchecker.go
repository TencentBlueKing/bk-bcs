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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/listenercontroller"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// ListenerChecker do listener related check
type ListenerChecker struct {
	cli            client.Client
	listenerHelper *listenercontroller.ListenerHelper
}

// NewListenerChecker return listener checker
func NewListenerChecker(cli client.Client, listenerHelp *listenercontroller.ListenerHelper) *ListenerChecker {
	return &ListenerChecker{
		cli:            cli,
		listenerHelper: listenerHelp,
	}
}

// Run start
func (l *ListenerChecker) Run() {
	listenerList := &networkextensionv1.ListenerList{}
	if err := l.cli.List(context.TODO(), listenerList); err != nil {
		blog.Errorf("list listener failed, err: %s", err.Error())
		return
	}

	go l.setMetric(listenerList)
	go l.deletePortPoolUnusedListener(listenerList)
}

func (l *ListenerChecker) setMetric(listenerList *networkextensionv1.ListenerList) {
	cntMap := make(map[string]int)
	for _, listener := range listenerList.Items {
		status := listener.Status.Status

		targetGroupType := networkextensionv1.LabelValueForTargetGroupNormal
		protocol := listener.Spec.Protocol
		if common.InLayer4Protocol(protocol) {
			if listener.Spec.TargetGroup == nil || len(listener.Spec.TargetGroup.Backends) == 0 {
				targetGroupType = networkextensionv1.LabelValueForTargetGroupEmpty
			}
		} else if common.InLayer7Protocol(protocol) {
			for _, rule := range listener.Spec.Rules {
				if rule.TargetGroup == nil || len(rule.TargetGroup.Backends) == 0 {
					targetGroupType = networkextensionv1.LabelValueForTargetGroupEmpty
					break
				}
			}
		}

		cntMap[buildKey(status, targetGroupType)] = cntMap[buildKey(status, targetGroupType)] + 1

		label := listener.GetLabels()
		value, ok := label[networkextensionv1.LabelKetForTargetGroupType]
		if !ok || value != targetGroupType {
			patchStruct := map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]string{
						networkextensionv1.LabelKetForTargetGroupType: targetGroupType,
					},
				},
			}
			patchData, err := json.Marshal(patchStruct)
			if err != nil {
				blog.Errorf("marshal listener failed, err: %s", err.Error())
				continue
			}
			updatePod := &networkextensionv1.Listener{
				ObjectMeta: metav1.ObjectMeta{
					Name:      listener.Name,
					Namespace: listener.Namespace,
				},
			}
			err = l.cli.Patch(context.TODO(), updatePod, client.RawPatch(types.MergePatchType, patchData))
			if err != nil {
				blog.Errorf("patch listener failed, err: %s", err.Error())
				continue
			}
		}
	}

	metrics.ListenerTotal.Reset()
	for key, cnt := range cntMap {
		status, targetGroupType := transKey(key)
		metrics.ListenerTotal.WithLabelValues(status, targetGroupType).Set(float64(cnt))
	}
}

// deleteUnusedListener 端口池中修改item-lbID后，需要回收不需要的监听器
func (l *ListenerChecker) deletePortPoolUnusedListener(listenerList *networkextensionv1.ListenerList) {
	portPoolList := &networkextensionv1.PortPoolList{}
	if err := l.cli.List(context.TODO(), portPoolList); err != nil {
		blog.Errorf("list portpool failed, err: %s", err.Error())
		return
	}

	// 缓存所有portpool下item相关的lbID
	// portPoolNamespace/portPoolName/itemName : lbIDSet
	poolItemLBIDMap := make(map[string]mapset.Set)
	for _, portpool := range portPoolList.Items {
		for _, item := range portpool.Spec.PoolItems {
			lbIDSet := mapset.NewThreadUnsafeSet()
			for _, lbID := range item.LoadBalancerIDs {
				_, tmpID, err := common.GetLbRegionAndName(lbID)
				if err != nil {
					blog.Errorf("unknown type lbID: %s, portpool: %s/%s", lbID, portpool.GetNamespace(), portpool.GetName())
					continue
				}
				lbIDSet.Add(tmpID)
			}
			key := portpool.GetNamespace() + "/" + common.GetPortPoolListenerLabelKey(portpool.GetName(), item.ItemName)
			poolItemLBIDMap[key] = lbIDSet
		}
	}

	for _, listener := range listenerList.Items {
		// 1. 确认listener是否属于portpool
		ownerKind, kok := listener.Labels[networkextensionv1.LabelKeyForOwnerKind]
		ownerName, nok := listener.Labels[networkextensionv1.LabelKeyForOwnerName]
		if !kok || !nok {
			blog.Warnf("listener '%s/%s' has no owner labels", listener.GetNamespace(), listener.GetName())
			continue
		}

		if ownerKind != constant.KindPortPool {
			continue
		}

		// 2. 找到listener对应的端口池
		portpool := &networkextensionv1.PortPool{}
		err := l.cli.Get(context.TODO(), types.NamespacedName{
			Namespace: listener.Namespace,
			Name:      ownerName,
		}, portpool)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				blog.Warnf("listener '%s/%s' not found related portpool '%s'", listener.GetNamespace(),
					listener.GetName(), ownerName)
				l.listenerHelper.SetDeleteListeners([]networkextensionv1.Listener{listener})
				continue
			}

			blog.Errorf("get portpool '%s/%s' failed, err: %s", ownerName, listener.GetNamespace(), err.Error())
			continue
		}

		// 3. 找到和listener匹配的PortPool item
		for _, item := range portpool.Spec.PoolItems {
			poolNameLabel := common.GetPortPoolListenerLabelKey(portpool.GetName(), item.ItemName)
			if _, ok := listener.Labels[poolNameLabel]; !ok {
				continue
			}

			poolItemKey := portpool.GetNamespace() + "/" + poolNameLabel
			lbIDSet, ok := poolItemLBIDMap[poolItemKey]
			if !ok {
				blog.Errorf("unknown pool item '%s' for listener '%s/%s'", poolNameLabel, listener.GetNamespace(),
					listener.GetName())
				break
			}

			// 4. 判断lbID是否和item定义匹配，不匹配则删除listener
			if !lbIDSet.Contains(listener.Spec.LoadbalancerID) {
				blog.Infof("listener '%s/%s' related loadbalancer is remove from item '%s/%s', delete it",
					listener.GetNamespace(), listener.GetName(), portpool.GetName(), item.ItemName)
				l.listenerHelper.SetDeleteListeners([]networkextensionv1.Listener{listener})
			}
			break
		}
	}
}

func buildKey(status, targetGroupType string) string {
	return fmt.Sprintf("%s/%s", status, targetGroupType)
}

// return status, targetGroup
func transKey(key string) (string, string) {
	splits := strings.Split(key, "/")
	if len(splits) != 2 {
		return "", ""
	}
	return splits[0], splits[1]
}
