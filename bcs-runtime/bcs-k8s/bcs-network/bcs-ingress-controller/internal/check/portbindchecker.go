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
	"fmt"
	"strings"
	"time"

	k8scorev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

const (
	// poolNamespace/poolName/poolItemName
	portBindingItemKey = "%s/%s/%s"
	// poolNamespace/poolName/poolItemName:startPort
	conflictKetFormat = "%s/%s/%s:%d"

	// portBindingTimeoutMin 端口绑定超时告警分钟数
	portBindingTimeoutMin = 10
)

// PortBindChecker 校验端口重复分配
type PortBindChecker struct {
	cli     client.Client
	eventer record.EventRecorder
}

// NewPortBindChecker return new port bind checker
func NewPortBindChecker(cli client.Client, eventer record.EventRecorder) *PortBindChecker {
	return &PortBindChecker{
		cli:     cli,
		eventer: eventer,
	}
}

// Run start check
func (p *PortBindChecker) Run() {
	portBindingList := &networkextensionv1.PortBindingList{}
	if err := p.cli.List(context.TODO(), portBindingList); err != nil {
		blog.Warnf("list portbinding failed, err: %s", err.Error())
		return
	}

	p.recordPortBindingStatusMetric(portBindingList)

	// poolNamespace/poolName/poolItemName:port -> [podName...] 记录itemKey+port下有哪些pod发生冲突
	conflictMap := p.getPortConflictMap(portBindingList)

	metrics.PortBindingConflictGauge.Reset()
	for conflictKey, conflictNamespaceNameList := range conflictMap {
		conflictNamespaceNames := strings.Join(conflictNamespaceNameList, ",")
		msg := fmt.Sprintf("[%s] conflict on pod %s", conflictKey, conflictNamespaceNames)
		blog.Errorf("port allocate conflict: %s", msg)
		p.sendEventToPortPool(conflictKey, msg)

		metrics.PortBindingConflictGauge.WithLabelValues(conflictKey).Set(float64(len(conflictNamespaceNameList)))
	}
}

func (p *PortBindChecker) getPortConflictMap(portBindingList *networkextensionv1.PortBindingList) map[string][]string {
	if portBindingList == nil {
		return nil
	}
	// itemKey -> {protocol/startPort -> podName}
	itemMap := make(map[string]map[string]string)
	// itemKey:port -> [podName...] 记录itemKey+port下有哪些pod发生冲突
	conflictMap := make(map[string][]string)
	for _, portBinding := range portBindingList.Items {
		portBindingItemList := portBinding.Spec.PortBindingList
		for _, item := range portBindingItemList {
			itemKey := buildPortBindingItemKey(item)
			portMap, ok := itemMap[itemKey]
			if !ok {
				portMap = make(map[string]string)
				itemMap[itemKey] = portMap
			}

			portKey := fmt.Sprintf("%s/%d", item.Protocol, item.StartPort)
			namespaceName, ok := portMap[portKey]
			if !ok {
				portMap[portKey] = fmt.Sprintf("%s/%s", portBinding.Namespace, portBinding.Name)
			} else {
				// startPort冲突说明发生端口重复分配,
				conflictKey := buildConflictKey(item)

				// 记录发生冲突的pod name
				conflictList, cok := conflictMap[conflictKey]
				if !cok {
					conflictList = make([]string, 0)
					conflictList = append(conflictList, namespaceName)
				}
				conflictList = append(conflictList, fmt.Sprintf("%s/%s", portBinding.Namespace,
					portBinding.Name))
				conflictMap[conflictKey] = conflictList
			}
		}
	}
	return conflictMap
}

func (p *PortBindChecker) recordPortBindingStatusMetric(portBindingList *networkextensionv1.PortBindingList) {
	if portBindingList == nil {
		return
	}

	cntMap := make(map[string]int)
	for _, portBinding := range portBindingList.Items {
		cntMap[portBinding.Status.Status] = cntMap[portBinding.Status.Status] + 1

		// 检查端口时间是否过久
		if notReadyTimeStr, ok := portBinding.Annotations[constant.
			AnnotationForPortBindingNotReadyTimestamp]; ok && notReadyTimeStr != "" {
			notReadyTime, err := time.Parse(time.RFC3339Nano, notReadyTimeStr)
			if err != nil {
				blog.Warnf("parse not ready time for portbinding '%s/%s' failed, err: %s",
					portBinding.GetNamespace(), portBinding.Name, err.Error())
				continue
			}
			if time.Now().Sub(notReadyTime).Minutes() > portBindingTimeoutMin {
				msg := fmt.Sprintf("port binding '%s/%s' bind exceed 10min", portBinding.GetNamespace(),
					portBinding.GetName())
				blog.Warn(msg)
				p.eventer.Event(&portBinding, k8scorev1.EventTypeWarning, "port bind slow", msg)
			}
		}
	}

	metrics.PortBindingTotal.Reset()
	for status, cnt := range cntMap {
		metrics.PortBindingTotal.WithLabelValues(status).Set(float64(cnt))
	}
}

func buildPortBindingItemKey(item *networkextensionv1.PortBindingItem) string {
	if item == nil {
		return ""
	}

	return fmt.Sprintf(portBindingItemKey, item.PoolNamespace, item.PoolName, item.PoolItemName)
}

func buildConflictKey(item *networkextensionv1.PortBindingItem) string {
	if item == nil {
		return ""
	}

	return fmt.Sprintf(conflictKetFormat, item.PoolNamespace, item.PoolName, item.PoolItemName, item.StartPort)
}

func (p *PortBindChecker) sendEventToPortPool(conflictKey, msg string) {
	portPool := &networkextensionv1.PortPool{}

	splits := strings.Split(conflictKey, "/")
	if len(splits) < 3 {
		blog.Errorf("conflictKey format error, get %s", conflictKey)
		return
	}
	if err := p.cli.Get(context.TODO(), k8stypes.NamespacedName{
		Namespace: splits[0],
		Name:      splits[1],
	}, portPool); err != nil {
		blog.Errorf("k8s get portpool failed, err: %s", err.Error())
		return
	}
	p.eventer.Event(portPool, k8scorev1.EventTypeWarning, "port allocate conflict", msg)
}
