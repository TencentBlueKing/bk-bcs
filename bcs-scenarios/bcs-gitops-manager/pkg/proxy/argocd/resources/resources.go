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

// Package resources xxx
package resources

import (
	"context"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsapi "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// PodQuery query the pods resources from kubernetes cluster
type PodQuery struct {
	Storage    store.Store
	BCSStorage bcsapi.Storage
}

// Query will parse the resource-tree of application, nad create the cluster client.
// Then it will query the detail pod resources from cluster
func (p *PodQuery) Query(ctx context.Context, argoApp *v1alpha1.Application) ([]corev1.Pod, error) {
	resourceTree, err := p.Storage.GetApplicationResourceTree(ctx, argoApp.Name)
	if err != nil {
		return nil, errors.Wrap(err, "get application resource tree failed")
	}
	namespacedPods := make(map[string][]*v1alpha1.ResourceNode)
	managedPods := 0
	for i := range resourceTree.Nodes {
		node := resourceTree.Nodes[i]
		// continue if the node not pod
		if node.Group != "" || node.Version != "v1" || node.Kind != "Pod" {
			continue
		}
		_, ok := namespacedPods[node.Namespace]
		if ok {
			namespacedPods[node.Namespace] = append(namespacedPods[node.Namespace], &node)
		} else {
			namespacedPods[node.Namespace] = []*v1alpha1.ResourceNode{&node}
		}
		managedPods++
	}
	t := strings.Split(argoApp.Spec.Destination.Server, "/")
	if len(t) == 0 {
		return nil, errors.Errorf("cluster '%s' format error", argoApp.Spec.Destination.Server)
	}
	clusterID := t[len(t)-1]
	if !strings.HasPrefix(clusterID, "BCS-K8S-") {
		return nil, errors.Errorf("cluster '%s' parse failed", argoApp.Spec.Destination.Server)
	}

	result := make([]corev1.Pod, 0, managedPods)
	for ns, nspods := range namespacedPods {
		pods, err := p.BCSStorage.QueryK8SPod(clusterID, ns)
		if err != nil {
			return nil, errors.Wrapf(err, "query k8s pods for cluster '%s/%s' failed", clusterID, ns)
		}
		podsMap := make(map[string]*corev1.Pod)
		for _, pod := range pods {
			if pod.Data == nil {
				blog.Warnf("RequestID[%s] pod '%s' from bcs-storage data is nil",
					mw.RequestID(ctx), pod.ResourceName)
				continue
			}
			podsMap[pod.Data.Name] = pod.Data
		}
		for _, pod := range nspods {
			v, ok := podsMap[pod.Name]
			if ok {
				result = append(result, *v)
			} else {
				blog.Warnf("RequestID[%s] pod '%s' not queried", mw.RequestID(ctx), pod.Name)
			}
		}
	}
	return result, nil
}
