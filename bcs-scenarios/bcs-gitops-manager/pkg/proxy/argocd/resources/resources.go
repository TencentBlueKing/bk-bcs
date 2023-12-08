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

package resources

import (
	"context"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned"
)

const (
	ResourceReplicaSet      = "ReplicaSet"
	ResourceStatefulSet     = "StatefulSet"
	ResourceDaemonSet       = "DaemonSet"
	ResourceJob             = "Job"
	ResourceGameDeployment  = "GameDeployment"
	ResourceGameStatefulSet = "GameStatefulSet"
)

// PodQuery query the pods resources from kubernetes cluster
type PodQuery struct {
	Storage store.Store

	singlePods       map[string][]string
	replicasets      map[string][]string
	statefulsets     map[string][]string
	daemonsets       map[string][]string
	jobs             map[string][]string
	gamedeployments  map[string][]string
	gamestatefulsets map[string][]string

	ctx           context.Context
	clientSet     *kubernetes.Clientset
	tkexClientSet *tkexclientset.Clientset
}

// Query will parse the resource-tree of application, nad create the cluster client.
// Then it will query the detail pod resources from cluster
func (p *PodQuery) Query(ctx context.Context, argoApp *v1alpha1.Application) ([]corev1.Pod, error) {
	p.ctx = ctx
	resourceTree, err := p.Storage.GetApplicationResourceTree(ctx, argoApp.Name)
	if err != nil {
		return nil, errors.Wrap(err, "get application resource tree failed")
	}
	p.parseResourceTree(ctx, argoApp, resourceTree)
	argoCluster, err := p.Storage.GetClusterFromDB(ctx, argoApp.Spec.Destination.Server)
	if err != nil {
		return nil, errors.Wrapf(err, "get cluster '%s' from db failed", argoApp.Spec.Destination.Server)
	}
	var config = &rest.Config{
		Host:        argoApp.Spec.Destination.Server,
		BearerToken: argoCluster.Config.BearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "build kubernetes client for cluster failed")
	}
	p.clientSet = clientSet
	if len(p.gamedeployments)+len(p.gamestatefulsets) != 0 {
		p.tkexClientSet, err = tkexclientset.NewForConfig(config)
		if err != nil {
			return nil, errors.Wrapf(err, "build kubernetes game client failed")
		}
	}

	funcs := []func() ([]corev1.Pod, error){
		p.queryReplicaSetPods, p.queryStatefulSetPods, p.queryJobPods, p.queryDaemonSetPods,
		p.queryGameDeployPods, p.queryGameStatefulSetPods,
	}
	result := make([]corev1.Pod, 0)
	for i := range funcs {
		var pods []corev1.Pod
		pods, err = funcs[i]()
		if err != nil {
			return nil, errors.Wrapf(err, "query pods by workload failed")
		}
		result = append(result, pods...)
	}
	return result, nil
}

// querySinglePods query the single pods
func (p *PodQuery) querySinglePods() ([]corev1.Pod, error) {
	result := make([]corev1.Pod, 0)
	for ns, pods := range p.statefulsets {
		for _, pod := range pods {
			k8sPod, err := p.clientSet.CoreV1().Pods(ns).Get(p.ctx, pod, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, errors.Wrapf(err, "get pod '%s/%s' failed", ns, pod)
			}
			result = append(result, *k8sPod)
		}
	}
	return result, nil
}

// queryReplicaSetPods query the replicaset pods
func (p *PodQuery) queryReplicaSetPods() ([]corev1.Pod, error) {
	result := make([]corev1.Pod, 0)
	if len(p.replicasets) == 0 {
		return result, nil
	}
	for ns, rses := range p.replicasets {
		for _, rs := range rses {
			k8sRs, err := p.clientSet.AppsV1().ReplicaSets(ns).Get(p.ctx, rs, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, errors.Wrapf(err, "get replicaset '%s/%s' failed", ns, rs)
			}
			pods, err := p.listPodsByWorkload(k8sRs.Spec.Selector, ns, rs, ResourceReplicaSet)
			if err != nil {
				return nil, err
			}
			result = append(result, pods...)
		}
	}
	return result, nil
}

// queryStatefulSetPods query the statefulset pods
func (p *PodQuery) queryStatefulSetPods() ([]corev1.Pod, error) {
	result := make([]corev1.Pod, 0)
	if len(p.statefulsets) == 0 {
		return result, nil
	}
	for ns, states := range p.statefulsets {
		for _, stat := range states {
			k8sStat, err := p.clientSet.AppsV1().StatefulSets(ns).Get(p.ctx, stat, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, errors.Wrapf(err, "get statefulset '%s/%s' failed", ns, stat)
			}
			pods, err := p.listPodsByWorkload(k8sStat.Spec.Selector, ns, stat, ResourceStatefulSet)
			if err != nil {
				return nil, err
			}
			result = append(result, pods...)
		}
	}
	return result, nil
}

// queryJobPods query the replicaset pods
func (p *PodQuery) queryJobPods() ([]corev1.Pod, error) {
	result := make([]corev1.Pod, 0)
	if len(p.jobs) == 0 {
		return result, nil
	}
	for ns, jobs := range p.jobs {
		for _, job := range jobs {
			k8sJob, err := p.clientSet.BatchV1().Jobs(ns).Get(p.ctx, job, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, errors.Wrapf(err, "get job '%s/%s' failed", ns, k8sJob)
			}
			pods, err := p.listPodsByWorkload(k8sJob.Spec.Selector, ns, job, ResourceJob)
			if err != nil {
				return nil, err
			}
			result = append(result, pods...)
		}
	}
	return result, nil
}

// queryJobPods query the replicaset pods
func (p *PodQuery) queryDaemonSetPods() ([]corev1.Pod, error) {
	result := make([]corev1.Pod, 0)
	if len(p.daemonsets) == 0 {
		return result, nil
	}
	for ns, daemonSets := range p.daemonsets {
		for _, ds := range daemonSets {
			k8sDS, err := p.clientSet.AppsV1().DaemonSets(ns).Get(p.ctx, ds, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, errors.Wrapf(err, "get daemonset '%s/%s' failed", ns, ds)
			}
			pods, err := p.listPodsByWorkload(k8sDS.Spec.Selector, ns, ds, ResourceReplicaSet)
			if err != nil {
				return nil, err
			}
			result = append(result, pods...)
		}
	}
	return result, nil
}

// queryGameDeployPods query game deployment pods
func (p *PodQuery) queryGameDeployPods() ([]corev1.Pod, error) {
	result := make([]corev1.Pod, 0)
	if len(p.gamedeployments) == 0 {
		return result, nil
	}
	for ns, gamedeployes := range p.gamedeployments {
		for _, gdp := range gamedeployes {
			k8sGameDeploy, err := p.tkexClientSet.TkexV1alpha1().GameDeployments(ns).Get(p.ctx, gdp, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, errors.Wrapf(err, "get game deploy '%s/%s' failed", ns, gdp)
			}
			pods, err := p.listPodsByWorkload(k8sGameDeploy.Spec.Selector, ns, gdp, ResourceGameDeployment)
			if err != nil {
				return nil, err
			}
			result = append(result, pods...)
		}
	}
	return result, nil
}

// queryGameStatefulSetPods query the game statefulset pods
func (p *PodQuery) queryGameStatefulSetPods() ([]corev1.Pod, error) {
	result := make([]corev1.Pod, 0)
	if len(p.gamestatefulsets) == 0 {
		return result, nil
	}
	for ns, gamestates := range p.gamestatefulsets {
		for _, gstat := range gamestates {
			k8sState, err := p.tkexClientSet.TkexV1alpha1().GameStatefulSets(ns).Get(p.ctx, gstat, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return nil, errors.Wrapf(err, "get game statefulset '%s/%s' failed", ns, gstat)
			}
			pods, err := p.listPodsByWorkload(k8sState.Spec.Selector, ns, gstat, ResourceGameStatefulSet)
			if err != nil {
				return nil, err
			}
			result = append(result, pods...)
		}
	}
	return result, nil
}

func (p *PodQuery) parseResourceTree(ctx context.Context, argoApp *v1alpha1.Application,
	resourceTree *v1alpha1.ApplicationTree) {
	p.singlePods = make(map[string][]string)
	p.replicasets = make(map[string][]string)
	p.statefulsets = make(map[string][]string)
	p.jobs = make(map[string][]string)
	p.daemonsets = make(map[string][]string)
	p.gamedeployments = make(map[string][]string)
	p.gamestatefulsets = make(map[string][]string)
	for i := range resourceTree.Nodes {
		node := resourceTree.Nodes[i]
		// continue if the node not pod
		if node.Group != "" || node.Version != "v1" || node.Kind != "Pod" {
			continue
		}
		if len(node.ParentRefs) == 0 {
			v, ok := p.singlePods[node.Namespace]
			if ok {
				p.singlePods[node.Namespace] = append(v, node.Name)
			} else {
				p.singlePods[node.Namespace] = []string{node.Name}
			}
			continue
		}
		knownWorkload := false
		for _, ref := range node.ParentRefs {
			switch ref.Kind {
			case ResourceReplicaSet:
				knownWorkload = true
				v, ok := p.replicasets[ref.Namespace]
				if ok {
					p.replicasets[ref.Namespace] = append(v, ref.Name)
				} else {
					p.replicasets[ref.Namespace] = []string{ref.Name}
				}
			case ResourceStatefulSet:
				knownWorkload = true
				v, ok := p.statefulsets[ref.Namespace]
				if ok {
					p.statefulsets[ref.Namespace] = append(v, ref.Name)
				} else {
					p.statefulsets[ref.Namespace] = []string{ref.Name}
				}
			case ResourceJob:
				knownWorkload = true
				v, ok := p.jobs[ref.Namespace]
				if ok {
					p.jobs[ref.Namespace] = append(v, ref.Name)
				} else {
					p.jobs[ref.Namespace] = []string{ref.Name}
				}
			case ResourceDaemonSet:
				knownWorkload = true
				v, ok := p.daemonsets[ref.Namespace]
				if ok {
					p.daemonsets[ref.Namespace] = append(v, ref.Name)
				} else {
					p.daemonsets[ref.Namespace] = []string{ref.Name}
				}
			case ResourceGameStatefulSet:
				knownWorkload = true
				v, ok := p.gamestatefulsets[ref.Namespace]
				if ok {
					p.gamestatefulsets[ref.Namespace] = append(v, ref.Name)
				} else {
					p.gamestatefulsets[ref.Namespace] = []string{ref.Name}
				}
			case ResourceGameDeployment:
				knownWorkload = true
				v, ok := p.gamedeployments[ref.Namespace]
				if ok {
					p.gamedeployments[ref.Namespace] = append(v, ref.Name)
				} else {
					p.gamedeployments[ref.Namespace] = []string{ref.Name}
				}
			default:
			}
		}
		if !knownWorkload {
			blog.Warnf("RequestID[%s] pod_query application '%s' pod '%s/%s' not belong to known workload",
				mw.RequestID(ctx), argoApp.Name, node.Namespace, node.Name)
			v, ok := p.singlePods[node.Namespace]
			if ok {
				p.singlePods[node.Namespace] = append(v, node.Name)
			} else {
				p.singlePods[node.Namespace] = []string{node.Name}
			}
		}
	}
}

func (p *PodQuery) listPodsByWorkload(selector *metav1.LabelSelector,
	ns, ownerName, ownerKind string) ([]corev1.Pod, error) {
	podList, err := p.clientSet.CoreV1().Pods(ns).List(p.ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(selector),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "list pods for %s '%s/%s' failed", ownerKind, ns, ownerName)
	}
	pods := make([]corev1.Pod, 0, len(podList.Items))
	for i := range podList.Items {
		t := podList.Items[i]
		for j := range t.OwnerReferences {
			r := t.OwnerReferences[j]
			if r.Name == ownerName && r.Kind == ownerKind {
				pods = append(pods, t)
				break
			}
		}
	}
	return pods, nil
}
