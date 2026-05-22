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

package clustercheck

import (
	"context"
	"fmt"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	admv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
)

// CheckWebhookNamespaceCrossHook 检查是否存在 webhook A 的 NamespaceSelector 命中 webhook B 所在的命名空间，且 webhook B 的 NamespaceSelector 也命中 webhook A 所在的命名空间的“互相 hook”情况。
// 仅限于“会拦截 Pod 创建”的 webhook（Rules 中包含对资源 pods 的 CREATE 操作）。
// 返回出现互相 hook 的命名空间列表（去重）、涉及的 webhook 名称列表（去重）以及 error。若命中情况发生，会使用 klog 打印日志。
func CheckWebhookNamespaceCrossHook(cfg *pluginmanager.ClusterConfig) ([]string, []string, error) {
	if cfg == nil || cfg.ClientSet == nil {
		return nil, nil, fmt.Errorf("invalid cluster config: cfg or cfg.ClientSet is nil")
	}

	ctx := context.TODO()

	// 1. 获取所有 Namespace
	nsList, err := cfg.ClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("list namespaces failed: %w", err)
	}
	// 建立 namespace 映射便于快速匹配
	nsMap := make(map[string]*v1.Namespace, len(nsList.Items))
	for i := range nsList.Items {
		ns := &nsList.Items[i]
		nsMap[ns.Name] = ns
	}

	// 2. 获取所有 ValidatingWebhookConfiguration 与 MutatingWebhookConfiguration
	vwhList, err := cfg.ClientSet.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("list ValidatingWebhookConfigurations failed: %w", err)
	}

	mwhList, err := cfg.ClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("list MutatingWebhookConfigurations failed: %w", err)
	}

	// 3. 归一化每个 webhook：名称、服务所在 namespace、NamespaceSelector（仅纳入拦截 Pod 创建的 webhook）
	type hookInfo struct {
		ID               string                // <kind>/<configName>/<webhookName>
		ServiceNamespace string                // webhook 的 clientConfig.service.namespace
		Selector         *metav1.LabelSelector // NamespaceSelector
	}

	var hooks []hookInfo

	// 收集 VWH
	for i := range vwhList.Items {
		conf := &vwhList.Items[i]
		for j := range conf.Webhooks {
			wh := conf.Webhooks[j]
			if wh.ClientConfig.Service == nil {
				continue
			}
			if !webhookHooksPodCreateV(wh) {
				// 非拦截 Pod 创建的跳过
				continue
			}
			hooks = append(hooks, hookInfo{
				ID:               fmt.Sprintf("VWH/%s/%s", conf.Name, wh.Name),
				ServiceNamespace: wh.ClientConfig.Service.Namespace,
				Selector:         wh.NamespaceSelector,
			})
		}
	}

	// 收集 MWH
	for i := range mwhList.Items {
		conf := &mwhList.Items[i]
		for j := range conf.Webhooks {
			wh := conf.Webhooks[j]
			if wh.ClientConfig.Service == nil {
				continue
			}
			if !webhookHooksPodCreateM(wh) {
				continue
			}
			hooks = append(hooks, hookInfo{
				ID:               fmt.Sprintf("MWH/%s/%s", conf.Name, wh.Name),
				ServiceNamespace: wh.ClientConfig.Service.Namespace,
				Selector:         wh.NamespaceSelector,
			})
		}
	}

	if len(hooks) < 2 {
		return []string{}, []string{}, nil
	}

	// Selector 与 namespace 匹配判断
	matchesNS := func(sel *metav1.LabelSelector, ns *v1.Namespace) bool {
		if ns == nil {
			return false
		}
		// NamespaceSelector 为空表示匹配所有命名空间
		if sel == nil {
			return true
		}
		ls, err := metav1.LabelSelectorAsSelector(sel)
		if err != nil {
			// 解析失败，保守为不匹配
			klog.V(4).Infof("invalid namespace selector for webhook, err=%v", err)
			return false
		}
		return ls.Matches(labels.Set(ns.Labels))
	}

	// 4. 两两检查“互相 hook”
	nsSet := make(map[string]struct{})
	hookSet := make(map[string]struct{})
	for i := 0; i < len(hooks); i++ {
		for j := i + 1; j < len(hooks); j++ {
			a := hooks[i]
			b := hooks[j]
			nsA := nsMap[a.ServiceNamespace]
			nsB := nsMap[b.ServiceNamespace]
			if nsA == nil || nsB == nil {
				// 若 service namespace 不存在（异常情况），跳过
				continue
			}

			aSelectsB := matchesNS(a.Selector, nsB)
			bSelectsA := matchesNS(b.Selector, nsA)

			if aSelectsB && bSelectsA {
				// 命中互相 hook
				klog.Warningf("clusterID=%s: detected mutual namespace hooking: [%s](svc-ns=%s) <-> [%s](svc-ns=%s)",
					cfg.ClusterID, a.ID, a.ServiceNamespace, b.ID, b.ServiceNamespace)
				nsSet[a.ServiceNamespace] = struct{}{}
				nsSet[b.ServiceNamespace] = struct{}{}
				hookSet[a.ID] = struct{}{}
				hookSet[b.ID] = struct{}{}
			}
		}
	}

	// 5. 输出列表（去重、排序）
	nsListOut := make([]string, 0, len(nsSet))
	for ns := range nsSet {
		nsListOut = append(nsListOut, ns)
	}
	sort.Strings(nsListOut)

	hookListOut := make([]string, 0, len(hookSet))
	for h := range hookSet {
		hookListOut = append(hookListOut, h)
	}
	sort.Strings(hookListOut)

	return nsListOut, hookListOut, nil
}

// webhookHooksPodCreateV 判断 ValidatingWebhook 是否拦截 Pod 创建
func webhookHooksPodCreateV(wh admv1.ValidatingWebhook) bool {
	return webhookRulesHookPodCreate(wh.Rules)
}

// webhookHooksPodCreateM 判断 MutatingWebhook 是否拦截 Pod 创建
func webhookHooksPodCreateM(wh admv1.MutatingWebhook) bool {
	return webhookRulesHookPodCreate(wh.Rules)
}

// webhookRulesHookPodCreate 规则是否包含对 pods 的 CREATE 操作
func webhookRulesHookPodCreate(rules []admv1.RuleWithOperations) bool {
	for _, rwo := range rules {
		// 需要包含 CREATE 操作
		hasCreate := false
		for _, op := range rwo.Operations {
			if op == admv1.Create {
				hasCreate = true
				break
			}
		}
		if !hasCreate {
			continue
		}

		// 资源需要包含 pods 或 pods/*
		hooksPods := false
		for _, res := range rwo.Resources {
			if res == "pods" || res == "pods/*" {
				hooksPods = true
				break
			}
		}
		if !hooksPods {
			continue
		}

		// 若上述条件满足，则认为该 webhook 会拦截 Pod 创建
		return true
	}
	return false
}
