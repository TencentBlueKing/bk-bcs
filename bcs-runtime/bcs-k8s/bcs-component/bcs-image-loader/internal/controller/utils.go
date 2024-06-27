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

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/controller/history"
	"k8s.io/utils/pointer"
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

func getJobName(loader *tkexv1alpha1.ImageLoader, index int) string {
	return fmt.Sprintf("%s-%d", loader.Name, index)
}

func newJob(loader *tkexv1alpha1.ImageLoader) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: loader.Namespace,
			Labels: map[string]string{
				ImageLoaderNameKey: loader.Name,
			},
			Annotations: loader.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: tkexv1alpha1.GroupVersion.String(),
					Kind:       tkexv1alpha1.KindImageLoader,
					Name:       loader.Name,
					UID:        loader.UID,
					Controller: pointer.Bool(true),
				},
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          &loader.Spec.BackoffLimit,
			ActiveDeadlineSeconds: &loader.Spec.JobTimeout,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      make(map[string]string),
					Annotations: make(map[string]string),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "image-loader",
							ImagePullPolicy: loader.Spec.ImagePullPolicy,
						},
					},
					ImagePullSecrets: loader.Spec.ImagePullSecrets,
					RestartPolicy:    "Never",
					Tolerations:      loader.Spec.Tolerations,
				},
			},
			// the job will be delete by reconciler, no need to set ttl
		},
	}
	if loader.Labels != nil && len(loader.Labels) != 0 {
		job.Labels = loader.Labels
		job.Spec.Template.Labels = loader.Labels
	}
	if loader.Annotations != nil && len(loader.Annotations) != 0 {
		job.Annotations = loader.Annotations
		job.Spec.Template.Annotations = loader.Annotations
	}
	return job
}

func (r *ImageLoaderReconciler) handlePodSelector(ctx context.Context, loader *tkexv1alpha1.ImageLoader,
	job *batchv1.Job) error {
	// ensure every node with specific pod will be scheduled to run this job
	job.Spec.Template.Spec.Affinity = &corev1.Affinity{
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
		nodes[pod.Spec.NodeName] = struct{}{}
	}
	unqieNodes := make([]string, 0, len(nodes))
	for key := range nodes {
		unqieNodes = append(unqieNodes, key)
	}
	// set nodes info in annotaions
	job.Annotations[NodeNameKey] = strings.Join(unqieNodes, ",")
	job.Spec.Parallelism = pointer.Int32(int32(len(unqieNodes)))
	job.Spec.Completions = pointer.Int32(int32(len(unqieNodes)))
	return nil
}

func (r *ImageLoaderReconciler) handleNodeSelector(ctx context.Context, loader *tkexv1alpha1.ImageLoader,
	job *batchv1.Job) error {

	nodes := make([]corev1.Node, 0)
	if len(loader.Spec.NodeSelector.MatchLabels) != 0 {
		// for selector. use nodeSelector
		nodeList := &corev1.NodeList{}
		job.Spec.Template.Spec.NodeSelector = loader.Spec.NodeSelector.MatchLabels
		selector := labels.Set(loader.Spec.NodeSelector.MatchLabels).AsSelector()
		err := r.Client.List(ctx, nodeList, client.MatchingLabelsSelector{Selector: selector})
		if err != nil {
			return fmt.Errorf("failed to list nodes: %v", err)
		}
		nodes = nodeList.Items
	} else if len(loader.Spec.NodeSelector.Names) != 0 {
		// for nodeName, use affinity
		requirement := convertMatchExpressions(loader.Spec.NodeSelector.Names)
		job.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
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

	unqieNodes := make([]string, len(nodes))
	for i := range nodes {
		unqieNodes[i] = nodes[i].Name
	}
	job.Annotations[NodeNameKey] = strings.Join(unqieNodes, ",")
	job.Spec.Parallelism = pointer.Int32(int32(len(unqieNodes)))
	job.Spec.Completions = pointer.Int32(int32(len(unqieNodes)))

	return nil
}

func (r *ImageLoaderReconciler) handleAllNode(ctx context.Context, job *batchv1.Job) error {
	nodeList := &corev1.NodeList{}
	err := r.Client.List(ctx, nodeList)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}
	if len(nodeList.Items) == 0 {
		logger.Info("no node found")
		return nil
	}
	job.Spec.Parallelism = pointer.Int32(int32(len(nodeList.Items)))
	job.Spec.Completions = pointer.Int32(int32(len(nodeList.Items)))
	return nil
}

func modifyJob(job *batchv1.Job, loader *tkexv1alpha1.ImageLoader, index int) {
	job.Name = getJobName(loader, index)
	job.Annotations[LoaderJobNameKey] = job.Name
	job.Spec.Template.Labels[LoaderJobNameKey] = job.Name
	job.Spec.Template.Spec.Containers[0].Image = loader.Spec.Images[index]
	job.Spec.Template.Spec.Containers[0].Command = []string{
		"echo", "pull " + loader.Spec.Images[index],
	}
	// ensure the job runs one pod on each node
	job.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
			{
				Namespaces: []string{loader.Namespace},
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						LoaderJobNameKey: job.Name,
					},
				},
				TopologyKey: corev1.LabelHostname,
			},
		},
	}
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
