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

package migrator

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// FilterNodes will receive the pod, and return the nodes that pod should locate on.
func (m *descheduleMigratorManager) FilterNodes(ctx context.Context, pod *corev1.Pod) (*corev1.NodeList, error) {
	_, workloadOwnerName, err := m.getPodOwnerName(ctx, pod.Namespace, pod.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "get pod '%s/%s' owner name failed", pod.Namespace, pod.Name)
	}
	v, ok := m.workloadPlansMap.Load(workloadOwnerName)
	if !ok {
		return nil, errors.Errorf("no plans for '%s'", workloadOwnerName)
	}
	plans := v.(WorkloadPlans)
	nodesMap := make(map[string]string)
	for _, plan := range plans {
		nodesMap[plan.To] = plan.To
	}
	nodeList := &corev1.NodeList{
		TypeMeta: metav1.TypeMeta{Kind: "List", APIVersion: "v1"},
		Items:    make([]corev1.Node, 0, len(nodesMap)),
	}
	for nodeName := range nodesMap {
		var node *corev1.Node
		if node, err = m.cacheManager.GetNode(ctx, nodeName); err != nil {
			blog.Warnf("Filter get node '%s' failed: %s", nodeName, err.Error())
			continue
		}
		nodeList.Items = append(nodeList.Items, *node)
	}
	return nodeList, nil
}
