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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller/history"
	"k8s.io/utils/integer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-image-loader/api/v1alpha1"
)

const (
	// ImageLoaderNameKey the annotation key of imageloader name
	ImageLoaderNameKey = "imageloader.tkex.tencent.com/name"

	// LoaderJobNameKey the annotation key of loader job name
	LoaderJobNameKey = "imageloader.tkex.tencent.com/job-name"

	// NodeNameKey the annotation key of node name
	NodeNameKey = "imageloader.tkex.tencent.com/node-name"

	// ImageLoaderRevisionKey the annotation key of imageloader revision
	ImageLoaderRevisionKey = "imageloader.tkex.tencent.com/revision-hash"
)

func getRevisionHash(spec *tkexv1alpha1.ImageLoaderSpec) string {
	marshalled, _ := json.Marshal(spec)
	revision := appsv1.ControllerRevision{
		Data: runtime.RawExtension{
			Raw: marshalled,
		},
	}
	hash := history.HashControllerRevision(&revision, nil)
	return hash
}

func newPod(loader *tkexv1alpha1.ImageLoader, newStatus *tkexv1alpha1.ImageLoaderStatus) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   loader.Namespace,
			Name:        loader.Name + "-",
			Labels:      loader.Labels,
			Annotations: loader.Annotations,
		},
		Spec: corev1.PodSpec{
			Containers:       make([]corev1.Container, 0, len(loader.Spec.Images)),
			RestartPolicy:    corev1.RestartPolicyOnFailure,
			ImagePullSecrets: loader.Spec.ImagePullSecrets,
			Tolerations:      loader.Spec.Tolerations,
		},
	}
	for i := range loader.Spec.Images {
		pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
			Name:            fmt.Sprintf("image-loader-%d", i),
			Image:           loader.Spec.Images[i],
			ImagePullPolicy: loader.Spec.ImagePullPolicy,
			Command:         []string{"echo", "pull", loader.Spec.Images[i]},
		})
	}
	if len(pod.Labels) == 0 {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[ImageLoaderNameKey] = loader.Name
	pod.Labels[ImageLoaderRevisionKey] = newStatus.Revision
	return pod
}

func (r *ImageLoaderReconciler) handleSelector(ctx context.Context, imageLoader *tkexv1alpha1.ImageLoader,
	pod *corev1.Pod,
) error {
	var err error
	switch {
	case imageLoader.Spec.PodSelector != nil:
		err = r.handlePodSelector(ctx, imageLoader, pod)
		if err == nil && len(pod.Annotations[NodeNameKey]) == 0 && imageLoader.Spec.NodeSelector != nil {
			logger.Info("failed to find node with specific pod, try to find node with specific node selector",
				"podSelector", imageLoader.Spec.PodSelector)
			pod.Spec.Affinity = nil
			err = r.handleNodeSelector(ctx, imageLoader, pod)
		}
	case imageLoader.Spec.NodeSelector != nil:
		err = r.handleNodeSelector(ctx, imageLoader, pod)
	default:
		err = r.handleAllNode(ctx, pod)
	}
	return err
}

func (r *ImageLoaderReconciler) handlePodSelector(ctx context.Context, loader *tkexv1alpha1.ImageLoader,
	pod *corev1.Pod,
) error {
	// ensure every node with specific pod will be scheduled to run this job
	pod.Spec.Affinity = &corev1.Affinity{
		PodAffinity: &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					Namespaces:    []string{loader.Namespace},
					LabelSelector: loader.Spec.PodSelector,
					TopologyKey:   corev1.LabelHostname,
				},
			},
		},
	}
	podList := &corev1.PodList{}
	selector, err := metav1.LabelSelectorAsSelector(loader.Spec.PodSelector)
	if err != nil {
		return fmt.Errorf("failed to convert labelSelector: %v", err)
	}
	err = r.Client.List(ctx, podList, client.InNamespace(loader.Namespace),
		client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}
	if len(podList.Items) == 0 {
		logger.Info("pod not found for podSelector", "podSelector", loader.Spec.PodSelector)
		return nil
	}
	nodes := make(map[string]struct{})
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == "" {
			logger.Info("pod's nodeName is empty", "pod", pod.Name, "status", pod.Status.Phase)
			continue
		}
		nodes[pod.Spec.NodeName] = struct{}{}
	}
	uniqueNodes := make([]string, 0, len(nodes))
	for key := range nodes {
		uniqueNodes = append(uniqueNodes, key)
	}
	// set nodes info in annotaions
	pod.Annotations[NodeNameKey] = strings.Join(uniqueNodes, ",")
	return nil
}

func (r *ImageLoaderReconciler) handleNodeSelector(ctx context.Context, loader *tkexv1alpha1.ImageLoader,
	pod *corev1.Pod,
) error {
	nodes := make([]corev1.Node, 0)
	if len(loader.Spec.NodeSelector.MatchLabels) != 0 {
		// for selector. use nodeSelector
		nodeList := &corev1.NodeList{}
		pod.Spec.NodeSelector = loader.Spec.NodeSelector.MatchLabels
		selector := labels.Set(loader.Spec.NodeSelector.MatchLabels).AsSelector()
		err := r.Client.List(ctx, nodeList, client.MatchingLabelsSelector{Selector: selector})
		if err != nil {
			return fmt.Errorf("failed to list nodes: %v", err)
		}
		nodes = nodeList.Items
	} else if len(loader.Spec.NodeSelector.Names) != 0 {
		// for nodeName, use affinity
		requirement := convertMatchExpressions(loader.Spec.NodeSelector.Names)
		pod.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							*requirement,
						},
					},
				},
			},
		}
		for _, name := range loader.Spec.NodeSelector.Names {
			node := &corev1.Node{}
			err := r.Client.Get(ctx, client.ObjectKey{Name: name}, node)
			if err != nil {
				// cannot find node, continue
				logger.Error(err, "failed to get node", "node", name)
				continue
			}
			nodes = append(nodes, *node)
		}
	}
	if len(nodes) == 0 {
		logger.Info("node not found for nodeSelector", "nodeSelector", loader.Spec.NodeSelector)
		return nil
	}

	uniqueNodes := make([]string, len(nodes))
	for i := range nodes {
		uniqueNodes[i] = nodes[i].Name
	}
	pod.Annotations[NodeNameKey] = strings.Join(uniqueNodes, ",")
	return nil
}

func (r *ImageLoaderReconciler) handleAllNode(ctx context.Context, pod *corev1.Pod) error {
	nodeList := &corev1.NodeList{}
	err := r.Client.List(ctx, nodeList)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}
	if len(nodeList.Items) == 0 {
		logger.Info("no node found")
		return nil
	}
	uniqueNodes := make([]string, len(nodeList.Items))
	for i := range nodeList.Items {
		uniqueNodes[i] = nodeList.Items[i].Name
	}
	pod.Annotations[NodeNameKey] = strings.Join(uniqueNodes, ",")
	return nil
}

func deletePods(ctx context.Context, cli client.Client, pods []*corev1.Pod) error {
	initialBatchSize := 1
	podsDeleteChan := make(chan *corev1.Pod, len(pods))
	for _, p := range pods {
		podsDeleteChan <- p
	}
	_, err := DoItSlowly(len(pods), initialBatchSize, func() error {
		pod := <-podsDeleteChan
		if deleteErr := cli.Delete(ctx, pod, client.GracePeriodSeconds(0)); deleteErr != nil {
			logger.Error(deleteErr, "failed to delete pod", "pod", klog.KRef(pod.Namespace, pod.Name))
			return deleteErr
		}
		logger.Info("deleted pod", "pod", klog.KRef(pod.Namespace, pod.Name))
		return nil
	})
	return err
}

func createPods(ctx context.Context, cli client.Client, pods []*corev1.Pod) error {
	initialBatchSize := 1
	podsCreateChan := make(chan *corev1.Pod, len(pods))
	for _, p := range pods {
		podsCreateChan <- p
	}
	_, err := DoItSlowly(len(pods), initialBatchSize, func() error {
		pod := <-podsCreateChan
		if createErr := cli.Create(ctx, pod); createErr != nil && !errors.IsAlreadyExists(createErr) {
			logger.Error(createErr, "failed to create pod", "pod", klog.KRef(pod.Namespace, pod.Name))
			return createErr
		}
		return nil
	})
	return err
}

// DoItSlowly tries to call the provided function a total of 'count' times,
// starting slow to check for errors, then speeding up if calls succeed.
//
// It groups the calls into batches, starting with a group of initialBatchSize.
// Within each batch, it may call the function multiple times concurrently.
//
// If a whole batch succeeds, the next batch may get exponentially larger.
// If there are any failures in a batch, all remaining batches are skipped
// after waiting for the current batch to complete.
//
// It returns the number of successful calls to the function.
func DoItSlowly(count int, initialBatchSize int, fn func() error) (int, error) {
	remaining := count
	successes := 0
	for batchSize := integer.IntMin(remaining, initialBatchSize); batchSize > 0; batchSize = integer.IntMin(
		2*batchSize, remaining) {
		errCh := make(chan error, batchSize)
		var wg sync.WaitGroup
		wg.Add(batchSize)
		for i := 0; i < batchSize; i++ {
			go func() {
				defer wg.Done()
				if err := fn(); err != nil {
					errCh <- err
				}
			}()
		}
		wg.Wait()
		curSuccesses := batchSize - len(errCh)
		successes += curSuccesses
		if len(errCh) > 0 {
			return successes, <-errCh
		}
		remaining -= batchSize
	}
	return successes, nil
}

// convertMatchExpressions covert nodeName to matchExpressions
func convertMatchExpressions(nodeNames []string) *corev1.NodeSelectorRequirement {
	res := &corev1.NodeSelectorRequirement{
		Key:      corev1.LabelHostname,
		Operator: corev1.NodeSelectorOpIn,
		Values:   make([]string, len(nodeNames)),
	}
	copy(res.Values, nodeNames)
	return res
}
