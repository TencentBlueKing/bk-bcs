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

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	gameAppsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/apps"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"
	sliceutil "k8s.io/kubectl/pkg/util/slice"
	"sigs.k8s.io/yaml"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
)

const (
	// ChangeCauseAnnotation is the annotation key for the change-cause of a ReplicaSet
	ChangeCauseAnnotation = "kubernetes.io/change-cause"
)

// HistoryViewer provides an interface for resources have historical information.
type HistoryViewer interface {
	ViewHistory(namespace, name string, revision int64) (string, error)
	GetHistory(namespace, name string) (map[int64]runtime.Object, error)
}

// HistoryVisitor is a visitor that can visit different kinds of resources
type HistoryVisitor struct {
	clientset kubernetes.Interface
	result    HistoryViewer
}

// VisitDeployment visits a deployment
func (v *HistoryVisitor) VisitDeployment(elem apps.GroupKindElement) {
	v.result = &DeploymentHistoryViewer{v.clientset}
}

// VisitStatefulSet visits a statefulset
func (v *HistoryVisitor) VisitStatefulSet(kind apps.GroupKindElement) {
	v.result = &StatefulSetHistoryViewer{v.clientset}
}

// VisitDaemonSet visits a daemonset
func (v *HistoryVisitor) VisitDaemonSet(kind apps.GroupKindElement) {
	v.result = &DaemonSetHistoryViewer{v.clientset}
}

// VisitJob visits a job
func (v *HistoryVisitor) VisitJob(kind apps.GroupKindElement) {}

// VisitPod visits a pod
func (v *HistoryVisitor) VisitPod(kind apps.GroupKindElement) {}

// VisitReplicaSet visits a replicaset
func (v *HistoryVisitor) VisitReplicaSet(kind apps.GroupKindElement) {}

// VisitReplicationController visits a replicationcontroller
func (v *HistoryVisitor) VisitReplicationController(kind apps.GroupKindElement) {}

// VisitCronJob visits a cronjob
func (v *HistoryVisitor) VisitCronJob(kind apps.GroupKindElement) {}

// CustomHistoryViewerFor returns an implementation of HistoryViewer interface for the given schema kind
func CustomHistoryViewerFor(kind schema.GroupKind, c kubernetes.Interface, g versioned.Interface) (
	HistoryViewer, error) {
	elem := apps.GroupKindElement(kind)
	var historyViwer HistoryViewer
	switch {
	case elem.GroupMatch("apiextensions.k8s.io") && elem.Kind == "GameDeployment":
		historyViwer = &GameDeploymentHistoryViewer{c: c, g: g.TkexV1alpha1()}
	case elem.GroupMatch("apiextensions.k8s.io") && elem.Kind == "GameStatefulSet":
		historyViwer = &GameStatefulSetHistoryViewer{c: c, g: g.TkexV1alpha1()}
	default:
		historyViwer = nil
	}

	if historyViwer == nil {
		return nil, fmt.Errorf("%q is not custom resource, has no history view", kind.String())
	}
	return historyViwer, nil
}

// HistoryViewerFor returns an implementation of HistoryViewer interface for the given schema kind
func HistoryViewerFor(kind schema.GroupKind, c kubernetes.Interface) (HistoryViewer, error) {
	elem := apps.GroupKindElement(kind)
	visitor := &HistoryVisitor{
		clientset: c,
	}

	// Determine which HistoryViewer we need here
	err := elem.Accept(visitor)

	if err != nil {
		return nil, fmt.Errorf("error retrieving history for %q, %v", kind.String(), err)
	}

	if visitor.result == nil {
		return nil, fmt.Errorf("no history viewer has been implemented for %q", kind.String())
	}

	return visitor.result, nil
}

// GameStatefulSetHistoryViewer is an implementation of HistoryViewer for GameStatefulSet
type GameStatefulSetHistoryViewer struct {
	c kubernetes.Interface
	g v1alpha1.TkexV1alpha1Interface
}

// ViewHistory returns a list of the revision history of a statefulset
// DOTO: this should be a describer
func (h *GameStatefulSetHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	sts, history, err := gameStatefulSetHistory(h.c.AppsV1(), h.g, namespace, name)
	if err != nil {
		return "", err
	}
	return printHistory(history, revision, func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error) {
		stsOfHistory, err := applyGameStatefulSetHistory(sts, history)
		if err != nil {
			return nil, err
		}
		return &stsOfHistory.Spec.Template, err
	})
}

// GetHistory returns the revisions associated with a StatefulSet
func (h *GameStatefulSetHistoryViewer) GetHistory(namespace, name string) (map[int64]runtime.Object, error) {
	_, history, err := gameStatefulSetHistory(h.c.AppsV1(), h.g, namespace, name)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]runtime.Object)
	for _, h := range history {
		result[h.Revision] = h.DeepCopy()
	}

	return result, nil
}

// GameDeploymentHistoryViewer is an implementation of HistoryViewer for GameDeployment
type GameDeploymentHistoryViewer struct {
	c kubernetes.Interface
	g v1alpha1.TkexV1alpha1Interface
}

// ViewHistory returns a list of the revision history of a GameDeployment
func (h *GameDeploymentHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	ds, history, err := gameDeploymentHistory(h.c.AppsV1(), h.g, namespace, name)
	if err != nil {
		return "", err
	}
	return printHistory(history, revision, func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error) {
		dsOfHistory, err := applyGameDeploymentHistory(ds, history)
		if err != nil {
			return nil, err
		}
		return &dsOfHistory.Spec.Template, err
	})
}

// GetHistory returns the revisions associated with a GameDeployment
func (h *GameDeploymentHistoryViewer) GetHistory(namespace, name string) (map[int64]runtime.Object, error) {
	_, history, err := gameDeploymentHistory(h.c.AppsV1(), h.g, namespace, name)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]runtime.Object)
	for _, h := range history {
		result[h.Revision] = h.DeepCopy()
	}

	return result, nil
}

// DeploymentHistoryViewer is an implementation of HistoryViewer for deployments
type DeploymentHistoryViewer struct {
	c kubernetes.Interface
}

// ViewHistory returns a revision-to-replicaset map as the revision history of a deployment
// DOTO: this should be a describer
func (h *DeploymentHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	allRSs, err := getDeploymentReplicaSets(h.c.AppsV1(), namespace, name)
	if err != nil {
		return "", err
	}

	historyInfo := make(map[int64]*corev1.PodTemplateSpec)
	for _, rs := range allRSs {
		v, err := deploymentutil.Revision(rs)
		if err != nil {
			klog.Warningf("unable to get revision from replicaset %s for deployment %s in namespace %s: %v",
				rs.Name, name, namespace, err)
			continue
		}
		historyInfo[v] = &rs.Spec.Template
		changeCause := getChangeCause(rs)
		if historyInfo[v].Annotations == nil {
			historyInfo[v].Annotations = make(map[string]string)
		}
		if len(changeCause) > 0 {
			historyInfo[v].Annotations[ChangeCauseAnnotation] = changeCause
		}
	}

	if len(historyInfo) == 0 {
		return "No rollout history found.", nil
	}

	if revision > 0 {
		// Print details of a specific revision
		template, ok := historyInfo[revision]
		if !ok {
			return "", fmt.Errorf("unable to find the specified revision")
		}
		return templateToString(template)
	}

	// Sort the revisionToChangeCause map by revision
	revisions := make([]int64, 0, len(historyInfo))
	for r := range historyInfo {
		revisions = append(revisions, r)
	}
	sliceutil.SortInts64(revisions)
	if len(revisions) == 0 {
		return "", nil
	}
	template, ok := historyInfo[revisions[len(revisions)-1]]
	if !ok {
		return "", fmt.Errorf("unable to find the specified revision")
	}
	return templateToString(template)
}

// GetHistory returns the ReplicaSet revisions associated with a Deployment
func (h *DeploymentHistoryViewer) GetHistory(namespace, name string) (map[int64]runtime.Object, error) {
	allRSs, err := getDeploymentReplicaSets(h.c.AppsV1(), namespace, name)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]runtime.Object)
	for _, rs := range allRSs {
		v, err := deploymentutil.Revision(rs)
		if err != nil {
			klog.Warningf("unable to get revision from replicaset %s for deployment %s in namespace %s: %v",
				rs.Name, name, namespace, err)
			continue
		}
		result[v] = rs
	}

	return result, nil
}

func templateToString(template *corev1.PodTemplateSpec) (string, error) {
	// remove pod-template-hash and other auto-generated labels
	if template.Annotations != nil {
		delete(template.Annotations, formatter.WorkloadRestartAnnotationKey)
		delete(template.Annotations, formatter.WorkloadRestartVersionAnnotationKey)
	}
	if template.Labels != nil {
		delete(template.Labels, "pod-template-hash")
	}
	data, err := json.Marshal(template)
	if err != nil {
		return "", err
	}
	data, err = yaml.JSONToYAML(data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DaemonSetHistoryViewer is an implementation of HistoryViewer for daemonsets
type DaemonSetHistoryViewer struct {
	c kubernetes.Interface
}

// ViewHistory returns a revision-to-history map as the revision history of a deployment
// DOTO: this should be a describer
func (h *DaemonSetHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	ds, history, err := daemonSetHistory(h.c.AppsV1(), namespace, name)
	if err != nil {
		return "", err
	}
	return printHistory(history, revision, func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error) {
		dsOfHistory, err := applyDaemonSetHistory(ds, history)
		if err != nil {
			return nil, err
		}
		return &dsOfHistory.Spec.Template, err
	})
}

// GetHistory returns the revisions associated with a DaemonSet
func (h *DaemonSetHistoryViewer) GetHistory(namespace, name string) (map[int64]runtime.Object, error) {
	_, history, err := daemonSetHistory(h.c.AppsV1(), namespace, name)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]runtime.Object)
	for _, h := range history {
		result[h.Revision] = h.DeepCopy()
	}

	return result, nil
}

// printHistory returns the podTemplate of the given revision if it is non-zero
// else returns the overall revisions
func printHistory(history []*appsv1.ControllerRevision, revision int64, getPodTemplate func(
	history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error)) (string, error) {
	historyInfo := make(map[int64]*appsv1.ControllerRevision)
	for _, history := range history {
		// DOTO: for now we assume revisions don't overlap, we may need to handle it
		historyInfo[history.Revision] = history
	}
	if len(historyInfo) == 0 {
		return "No rollout history found.", nil
	}

	// Print details of a specific revision
	if revision > 0 {
		history, ok := historyInfo[revision]
		if !ok {
			return "", fmt.Errorf("unable to find the specified revision")
		}
		podTemplate, err := getPodTemplate(history)
		if err != nil {
			return "", fmt.Errorf("unable to parse history %s", history.Name)
		}
		return templateToString(podTemplate)
	}

	// Print an overview of all Revisions
	// Sort the revisionToChangeCause map by revision
	revisions := make([]int64, 0, len(historyInfo))
	for r := range historyInfo {
		revisions = append(revisions, r)
	}
	sliceutil.SortInts64(revisions)

	if len(revisions) == 0 {
		return "", nil
	}
	template, ok := historyInfo[revisions[len(revisions)-1]]
	if !ok {
		return "", fmt.Errorf("unable to find the specified revision")
	}
	podTemplate, err := getPodTemplate(template)
	if err != nil {
		return "", fmt.Errorf("unable to parse history %s", template.Name)
	}
	return templateToString(podTemplate)
}

// StatefulSetHistoryViewer is an implementation of HistoryViewer for statefulsets
type StatefulSetHistoryViewer struct {
	c kubernetes.Interface
}

// ViewHistory returns a list of the revision history of a statefulset
// DOTO: this should be a describer
func (h *StatefulSetHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	sts, history, err := statefulSetHistory(h.c.AppsV1(), namespace, name)
	if err != nil {
		return "", err
	}
	return printHistory(history, revision, func(history *appsv1.ControllerRevision) (*corev1.PodTemplateSpec, error) {
		stsOfHistory, err := applyStatefulSetHistory(sts, history)
		if err != nil {
			return nil, err
		}
		return &stsOfHistory.Spec.Template, err
	})
}

// GetHistory returns the revisions associated with a StatefulSet
func (h *StatefulSetHistoryViewer) GetHistory(namespace, name string) (map[int64]runtime.Object, error) {
	_, history, err := statefulSetHistory(h.c.AppsV1(), namespace, name)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]runtime.Object)
	for _, h := range history {
		result[h.Revision] = h.DeepCopy()
	}

	return result, nil
}

func getDeploymentReplicaSets(apps clientappsv1.AppsV1Interface, namespace, name string) ([]*appsv1.ReplicaSet, error) {
	deployment, err := apps.Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve deployment %s: %v", name, err)
	}

	_, oldRSs, newRS, err := deploymentutil.GetAllReplicaSets(deployment, apps)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve replica sets from deployment %s: %v", name, err)
	}

	if newRS == nil {
		return oldRSs, nil
	}
	return append(oldRSs, newRS), nil
}

// controlledHistories returns all ControllerRevisions in namespace that selected by selector and owned by accessor
// DOTO: Rename this to controllerHistory when other controllers have been upgraded
func controlledHistoryV1(
	apps clientappsv1.AppsV1Interface,
	namespace string,
	selector labels.Selector,
	accessor metav1.Object) ([]*appsv1.ControllerRevision, error) {
	var result []*appsv1.ControllerRevision
	historyList, err := apps.ControllerRevisions(namespace).List(context.TODO(),
		metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}
	for i := range historyList.Items {
		history := historyList.Items[i]
		// Only add history that belongs to the API object
		if metav1.IsControlledBy(&history, accessor) {
			result = append(result, &history)
		}
	}
	return result, nil
}

// controlledHistories returns all ControllerRevisions in namespace that selected by selector and owned by accessor
func controlledHistory(
	apps clientappsv1.AppsV1Interface,
	namespace string,
	selector labels.Selector,
	accessor metav1.Object) ([]*appsv1.ControllerRevision, error) {
	var result []*appsv1.ControllerRevision
	historyList, err := apps.ControllerRevisions(namespace).List(context.TODO(),
		metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}
	for i := range historyList.Items {
		history := historyList.Items[i]
		// Only add history that belongs to the API object
		if metav1.IsControlledBy(&history, accessor) {
			result = append(result, &history)
		}
	}
	return result, nil
}

// daemonSetHistory returns the DaemonSet named name in namespace and all ControllerRevisions in its history.
func daemonSetHistory(
	apps clientappsv1.AppsV1Interface,
	namespace, name string) (*appsv1.DaemonSet, []*appsv1.ControllerRevision, error) {
	ds, err := apps.DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve DaemonSet %s: %v", name, err)
	}
	selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create selector for DaemonSet %s: %v", ds.Name, err)
	}
	accessor, err := meta.Accessor(ds)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create accessor for DaemonSet %s: %v", ds.Name, err)
	}
	history, err := controlledHistory(apps, ds.Namespace, selector, accessor)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find history controlled by DaemonSet %s: %v", ds.Name, err)
	}
	return ds, history, nil
}

// statefulSetHistory returns the StatefulSet named name in namespace and all ControllerRevisions in its history.
func statefulSetHistory(
	apps clientappsv1.AppsV1Interface,
	namespace, name string) (*appsv1.StatefulSet, []*appsv1.ControllerRevision, error) {
	sts, err := apps.StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve Statefulset %s: %s", name, err.Error())
	}
	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create selector for StatefulSet %s: %s", name, err.Error())
	}
	accessor, err := meta.Accessor(sts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to obtain accessor for StatefulSet %s: %s", name, err.Error())
	}
	history, err := controlledHistoryV1(apps, namespace, selector, accessor)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find history controlled by StatefulSet %s: %v", name, err)
	}
	return sts, history, nil
}

// statefulSetHistory returns the StatefulSet named name in namespace and all ControllerRevisions in its history.
func gameStatefulSetHistory(
	apps clientappsv1.AppsV1Interface,
	gapps v1alpha1.TkexV1alpha1Interface,
	namespace, name string) (*gameAppsv1.GameStatefulSet, []*appsv1.ControllerRevision, error) {
	sts, err := gapps.GameStatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve gameStatefulSet %s: %s", name, err.Error())
	}
	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create selector for gameStatefulSet %s: %s", name, err.Error())
	}
	accessor, err := meta.Accessor(sts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to obtain accessor for gameStatefulSet %s: %s", name, err.Error())
	}
	history, err := controlledHistoryV1(apps, namespace, selector, accessor)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find history controlled by gameStatefulSet %s: %v", name, err)
	}
	return sts, history, nil
}

func gameDeploymentHistory(
	apps clientappsv1.AppsV1Interface,
	gapps v1alpha1.TkexV1alpha1Interface,
	namespace, name string) (*gameAppsv1.GameDeployment, []*appsv1.ControllerRevision, error) {
	ds, err := gapps.GameDeployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve gameDeployment %s: %s", name, err.Error())
	}
	selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create selector for gameDeployment %s: %s", name, err.Error())
	}
	accessor, err := meta.Accessor(ds)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to obtain accessor for gameDeployment %s: %s", name, err.Error())
	}
	history, err := controlledHistoryV1(apps, namespace, selector, accessor)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find history controlled by gameDeployment %s: %v", name, err)
	}
	return ds, history, nil
}

// applyDaemonSetHistory returns a specific revision of DaemonSet by applying the given history to a copy of
// the given DaemonSet
func applyDaemonSetHistory(ds *appsv1.DaemonSet, history *appsv1.ControllerRevision) (*appsv1.DaemonSet, error) {
	dsBytes, err := json.Marshal(ds)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(dsBytes, history.Data.Raw, ds)
	if err != nil {
		return nil, err
	}
	result := &appsv1.DaemonSet{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// applyStatefulSetHistory returns a specific revision of StatefulSet by applying the given history to a copy of
// the given StatefulSet
func applyStatefulSetHistory(sts *appsv1.StatefulSet, history *appsv1.ControllerRevision) (*appsv1.StatefulSet, error) {
	stsBytes, err := json.Marshal(sts)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(stsBytes, history.Data.Raw, sts)
	if err != nil {
		return nil, err
	}
	result := &appsv1.StatefulSet{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// applyGameStatefulSetHistory returns a specific revision by applying the given history to a copy of the given workload
func applyGameStatefulSetHistory(sts interface{}, history *appsv1.ControllerRevision) (*gameAppsv1.GameStatefulSet,
	error) {
	stsBytes, err := json.Marshal(sts)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(stsBytes, history.Data.Raw, sts)
	if err != nil {
		return nil, err
	}
	result := &gameAppsv1.GameStatefulSet{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// applyGameStatefulSetHistory returns a specific revision by applying the given history to a copy of the given workload
func applyGameDeploymentHistory(sts interface{}, history *appsv1.ControllerRevision) (*gameAppsv1.GameDeployment,
	error) {
	stsBytes, err := json.Marshal(sts)
	if err != nil {
		return nil, err
	}
	patched, err := strategicpatch.StrategicMergePatch(stsBytes, history.Data.Raw, sts)
	if err != nil {
		return nil, err
	}
	result := &gameAppsv1.GameDeployment{}
	err = json.Unmarshal(patched, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type historiesByRevision []*appsv1.ControllerRevision

func (h historiesByRevision) Len() int      { return len(h) }
func (h historiesByRevision) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h historiesByRevision) Less(i, j int) bool {
	return h[i].Revision < h[j].Revision
}

// findHistory returns a controllerrevision of a specific revision from the given controllerrevisions.
// It returns nil if no such controllerrevision exists.
// If toRevision is 0, the last previously used history is returned.
func findHistory(toRevision int64, allHistory []*appsv1.ControllerRevision) *appsv1.ControllerRevision {
	if toRevision == 0 && len(allHistory) <= 1 {
		return nil
	}

	// Find the history to rollback to
	var toHistory *appsv1.ControllerRevision
	if toRevision == 0 {
		// If toRevision == 0, find the latest revision (2nd max)
		sort.Sort(historiesByRevision(allHistory))
		toHistory = allHistory[len(allHistory)-2]
	} else {
		for _, h := range allHistory {
			if h.Revision == toRevision {
				// If toRevision != 0, find the history with matching revision
				return h
			}
		}
	}

	return toHistory
}

// DOTO: copied here until this becomes a describer
// nolint
func tabbedString(f func(io.Writer) error) (string, error) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 2, ' ', 0)

	err := f(out)
	if err != nil {
		return "", err
	}

	out.Flush()
	str := buf.String()
	return str, nil
}

// getChangeCause returns the change-cause annotation of the input object
func getChangeCause(obj runtime.Object) string {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return ""
	}
	return accessor.GetAnnotations()[ChangeCauseAnnotation]
}
