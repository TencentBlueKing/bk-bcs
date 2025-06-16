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

package clusterops

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// stringInSlice returns true if given string in slice
func stringInSlice(s string, l []string) bool {
	for _, objStr := range l {
		if s == objStr {
			return true
		}
	}
	return false
}

// GetNodePodsOption get node pods option
type GetNodePodsOption struct {
	ClusterID        string
	NodeName         string
	FilterNamespaces []string
	FilterKinds      []string
}

// GetNodePods get node pods
func (ko *K8SOperator) GetNodePods(ctx context.Context, option GetNodePodsOption) ([]*v1.Pod, error) {
	if ko == nil {
		return nil, ErrServerNotInit
	}

	if option.ClusterID == "" || option.NodeName == "" {
		return nil, fmt.Errorf("clusterId or nodeName is empty")
	}

	clientInterface, err := ko.GetClusterClient(option.ClusterID)
	if err != nil {
		blog.Errorf("GetNodePods GetClusterClient failed: %v", err)
		return nil, err
	}

	var (
		pods []*v1.Pod
	)
	podList, err := clientInterface.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", option.NodeName),
	})
	if err != nil {
		blog.Errorf("GetNodePods ListPods[%s] failed: %v", option.ClusterID, err)
		return nil, err
	}

	blog.Infof("cluster[%s] ListClusterNodePods successful: %v", option.ClusterID, len(podList.Items))

	for i := range podList.Items {
		if len(option.FilterNamespaces) > 0 &&
			stringInSlice(podList.Items[i].Namespace, option.FilterNamespaces) {
			continue
		}

		filterKindResult := false
		if len(option.FilterKinds) > 0 {
			for _, item := range podList.Items[i].GetOwnerReferences() {
				if stringInSlice(item.Kind, option.FilterKinds) {
					filterKindResult = true
					break
				}
			}
		}

		if !filterKindResult {
			pods = append(pods, &podList.Items[i])
		}
	}

	return pods, nil
}

// CheckPodPDBOption check pod PDB option
type CheckPodPDBOption struct {
	ClusterID string
	Pod       v1.Pod
}

// CheckPodPDB check pod has PDB
func (ko *K8SOperator) CheckPodPDB(ctx context.Context, option CheckPodPDBOption) (bool, error) {
	if ko == nil {
		return false, ErrServerNotInit
	}

	clientInterface, err := ko.GetClusterClient(option.ClusterID)
	if err != nil {
		blog.Errorf("CheckPodPDB GetClusterClient failed: %v", err)
		return false, err
	}

	// List all PDB in the pod namespace
	pdbs, err := clientInterface.PolicyV1().PodDisruptionBudgets(option.Pod.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("CheckPodPDB Get PodDisruptionBudgets List failed: %v", err)
	}

	// check if any PDB selector matches the pod labels
	for _, pdb := range pdbs.Items {
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			blog.Warnf("CheckPodPDB[%s] parse selector for PDB failed: %v", pdb.Name, err)
			continue
		}

		if selector.Matches(labels.Set(option.Pod.Labels)) {
			return true, nil
		}
	}

	return false, nil
}
