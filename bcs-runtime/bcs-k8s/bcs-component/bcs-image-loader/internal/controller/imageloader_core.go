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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-image-loader/api/v1alpha1"
)

func (r *ImageLoaderReconciler) reconcileImageLoader(ctx context.Context,
	imageLoader *tkexv1alpha1.ImageLoader) (
	*tkexv1alpha1.ImageLoaderStatus, *time.Duration, error,
) {
	var requeue time.Duration
	var err error
	newStatus := imageLoader.Status.DeepCopy()
	// 1. check if the spec is changed
	newRevision := getRevisionHash(&imageLoader.Spec)
	if newRevision == newStatus.Revision &&
		newStatus.Desired != int32(0) &&
		newStatus.Completed == newStatus.Desired {
		logger.Info("ImageLoader complete, skip reconcile")
		return newStatus, nil, nil
	}
	if newStatus.Revision == "" {
		r.resetStatus(imageLoader, newStatus)
		now := metav1.Now()
		newStatus.StartTime = &now
		newStatus.Revision = newRevision
		r.Recorder.Eventf(imageLoader, corev1.EventTypeNormal, "Start", "start to preload images")
	}
	if newRevision != newStatus.Revision {
		logger.Info("ImageLoader spec changed")
		r.resetStatus(imageLoader, newStatus)
		finished, cleanErr := r.cleanPods(ctx, imageLoader, newStatus.Revision)
		if cleanErr != nil {
			return newStatus, nil, cleanErr
		}
		if !finished {
			logger.Info("wait for previous pods to completely cleanup")
			requeue = time.Second
			return newStatus, &requeue, nil
		}
		logger.Info("finish cleaning previous pods")
		now := metav1.Now()
		newStatus.StartTime = &now
		newStatus.Revision = newRevision
	}

	// 2. new pod based on spec
	basePod := newPod(imageLoader, newStatus)
	err = r.handleSelector(ctx, imageLoader, basePod)
	if err != nil {
		return newStatus, nil, err
	}
	if len(basePod.Annotations[NodeNameKey]) == 0 {
		r.resetStatus(imageLoader, newStatus)
		newStatus.Desired = -1
		newStatus.ObservedGeneration = imageLoader.Generation
		newStatus.Completed = newStatus.Desired
		newStatus.Succeeded = newStatus.Desired
		logger.Info("no node need to preload image")
		r.Recorder.Eventf(imageLoader, corev1.EventTypeWarning, "Complete", "no node need to preload image")
		return newStatus, nil, nil
	}

	// 3. load image
	err = r.loadImage(ctx, imageLoader, basePod, newStatus)
	if err != nil {
		return newStatus, nil, err
	}

	// 4. renew status
	r.renewStatus(imageLoader, newStatus)

	return newStatus, nil, nil
}

func (r *ImageLoaderReconciler) cleanPods(ctx context.Context,
	loader *tkexv1alpha1.ImageLoader, revision string,
) (bool, error) {
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList, client.MatchingLabels{
		ImageLoaderNameKey:     loader.Name,
		ImageLoaderRevisionKey: revision,
	}, client.InNamespace(loader.Namespace)); err != nil {
		return false, err
	}
	if len(podList.Items) == 0 {
		return true, nil
	}
	for i := range podList.Items {
		if podList.Items[i].DeletionTimestamp != nil {
			logger.Info("pod is deleting", "pod", podList.Items[i].Name, "node", podList.Items[i].Spec.NodeName)
			continue
		}
		logger.Info("delete pod", "pod", podList.Items[i].Name, "node", podList.Items[i].Spec.NodeName)
		if err := r.Delete(ctx, &podList.Items[i]); err != nil {
			logger.Error(err, "failed to delete pod", "pod", podList.Items[i].Name, "node", podList.Items[i].Spec.NodeName)
			return false, err
		}
	}
	return false, nil
}

func (r *ImageLoaderReconciler) loadImage(ctx context.Context, loader *tkexv1alpha1.ImageLoader,
	basePod *corev1.Pod, newStatus *tkexv1alpha1.ImageLoaderStatus,
) error {
	expectedNodes := strings.Split(basePod.Annotations[NodeNameKey], ",")
	delete(basePod.Annotations, NodeNameKey)
	newStatus.Desired = int32(len(expectedNodes))
	newStatus.Active = 0

	// 检查现有 pods
	toDeletePods, toCreatePods, ignoredNodes, err := r.processPods(ctx, loader, newStatus)
	if err != nil {
		return err
	}

	// 执行删除 pod
	err = deletePods(ctx, r.Client, toDeletePods)
	if err != nil {
		return err
	}

	// 检查节点
	for _, node := range expectedNodes {
		// 忽略节点直接跳过
		if _, ok := ignoredNodes[node]; ok {
			continue
		}
		newPod := basePod.DeepCopy()
		newPod.Name += node
		newPod.Spec.NodeName = node
		if err = ctrl.SetControllerReference(loader, newPod, r.Client.Scheme()); err != nil {
			logger.Error(err, "failed to set owner for pod", "pod", newPod.Name)
			continue
		}
		logger.Info("want to create pod", "pod", newPod.Name, "node", node)
		toCreatePods = append(toCreatePods, newPod)
	}

	// 创建 pod
	err = createPods(ctx, r.Client, toCreatePods)
	if len(toCreatePods) > 0 || len(toDeletePods) > 0 {
		logger.Info("new status in loadImage", "active", newStatus.Active, "succeed", newStatus.Succeeded, "desired",
			newStatus.Desired)
	}
	return err
}

func (r *ImageLoaderReconciler) processPods(ctx context.Context, loader *tkexv1alpha1.ImageLoader,
	newStatus *tkexv1alpha1.ImageLoaderStatus,
) ([]*corev1.Pod, []*corev1.Pod, map[string]struct{}, error) {
	// 已加载成功/失败节点
	ignoredNodes := map[string]struct{}{}
	loadedNodes := map[string]struct{}{}
	failedNodes := map[string]struct{}{}
	for _, n := range newStatus.LoadedNodes {
		ignoredNodes[n] = struct{}{}
		loadedNodes[n] = struct{}{}
	}
	for _, n := range newStatus.FailedNodes {
		ignoredNodes[n] = struct{}{}
		failedNodes[n] = struct{}{}
	}

	existPods := &corev1.PodList{}
	err := r.Client.List(ctx, existPods, client.InNamespace(loader.Namespace), client.MatchingLabels{
		ImageLoaderNameKey: loader.Name, ImageLoaderRevisionKey: newStatus.Revision,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	toDeletePods := make([]*corev1.Pod, 0)
	toCreatePods := make([]*corev1.Pod, 0)

	for i := range existPods.Items {
		pod := &existPods.Items[i]
		// 删除中，忽略
		if !pod.DeletionTimestamp.IsZero() {
			ignoredNodes[pod.Spec.NodeName] = struct{}{}
			// logger.Info("loading pod deleting, ignore it", "pod", pod.Name)
			continue
		}
		// succeed: 删除 Pod 并记录节点
		if pod.Status.Phase == corev1.PodSucceeded {
			toDeletePods = append(toDeletePods, pod)
			loadedNodes[pod.Spec.NodeName] = struct{}{}
			ignoredNodes[pod.Spec.NodeName] = struct{}{}
			logger.Info("loading pod succeed, delete it", "pod", pod.Name)
			continue
		}
		// failed or unknown: 删除 Pod 并跳过此次创建，等待下次创建
		if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
			logger.Info("loading pod failed, rebuild it", "pod", pod.Name, "phase", pod.Status.Phase,
				"reason", pod.Status.Reason)
			toDeletePods = append(toDeletePods, pod)
			ignoredNodes[pod.Spec.NodeName] = struct{}{}
			continue
		}
		// 持续一段时间没完成，强制删除 pod
		if pod.Status.StartTime != nil &&
			time.Since(pod.Status.StartTime.Time) > time.Second*time.Duration(loader.Spec.JobTimeout) {
			// 如果已经失败了多次，保留 pod 不删除
			if time.Since(newStatus.StartTime.Time) >
				time.Second*time.Duration(loader.Spec.BackoffLimit)*time.Duration(loader.Spec.JobTimeout) {
				logger.Info("load image timeout, please check it", "pod", pod.Name)
				failedNodes[pod.Spec.NodeName] = struct{}{}
				r.Recorder.Eventf(loader, corev1.EventTypeWarning,
					"LoadImageFailed", "load image failed on node %s", pod.Spec.NodeName)
				imageLoaderFailed.WithLabelValues(loader.Namespace, loader.Name, pod.Spec.NodeName).Inc()
			} else {
				toDeletePods = append(toDeletePods, pod)
				logger.Info("load pod running timeout, delete it", "pod", pod.Name)
			}
			ignoredNodes[pod.Spec.NodeName] = struct{}{}
			continue
		}

		// running or pending: 跳过创建
		newStatus.Active++
		ignoredNodes[pod.Spec.NodeName] = struct{}{}
		// logger.Info("loading pod running or pending, ignore it", "pod", pod.Name, "phase", pod.Status.Phase)
	}

	newStatus.LoadedNodes = make([]string, 0, len(loadedNodes))
	newStatus.FailedNodes = make([]string, 0, len(failedNodes))
	for n := range loadedNodes {
		newStatus.LoadedNodes = append(newStatus.LoadedNodes, n)
	}
	for n := range failedNodes {
		newStatus.FailedNodes = append(newStatus.FailedNodes, n)
	}
	newStatus.Completed = int32(len(newStatus.LoadedNodes)) + int32(len(newStatus.FailedNodes))
	newStatus.Succeeded = int32(len(newStatus.LoadedNodes))
	return toDeletePods, toCreatePods, ignoredNodes, nil
}

func (r *ImageLoaderReconciler) resetStatus(loader *tkexv1alpha1.ImageLoader,
	newStatus *tkexv1alpha1.ImageLoaderStatus,
) {
	newStatus.ObservedGeneration = loader.Generation
	newStatus.Desired = 0
	newStatus.Active = 0
	newStatus.Completed = 0
	newStatus.Succeeded = 0
	newStatus.FailedStatuses = make([]*tkexv1alpha1.FailedStatus, 0)
	newStatus.CompletionTime = nil
	newStatus.LoadedNodes = make([]string, 0)
	newStatus.FailedNodes = make([]string, 0)
}

func (r *ImageLoaderReconciler) renewStatus(imageLoader *tkexv1alpha1.ImageLoader,
	newStatus *tkexv1alpha1.ImageLoaderStatus,
) {
	if newStatus.Desired == newStatus.Completed {
		imageLoaderRuningSeconds.WithLabelValues(imageLoader.Namespace, imageLoader.Name).Set(0)
		now := metav1.Now()
		newStatus.CompletionTime = &now
		if newStatus.Succeeded == newStatus.Desired {
			logger.Info("imageloader completed successfully")
			r.Recorder.Eventf(imageLoader, corev1.EventTypeNormal, "Succeed", "All imageloader jobs succeeded")
			imageLoaderCompletedSeconds.WithLabelValues(imageLoader.Namespace, imageLoader.Name,
				"Succeeded").Set(time.Since(newStatus.StartTime.Time).Seconds())
		} else {
			logger.Info("imageloader completed with partial jobs succeed", "succeeded", newStatus.Succeeded, "desired",
				newStatus.Desired)
			r.Recorder.Eventf(imageLoader, corev1.EventTypeWarning, "Completed", "Some imageloader jobs failed")
			imageLoaderCompletedSeconds.WithLabelValues(imageLoader.Namespace, imageLoader.Name,
				"Completed").Set(time.Since(newStatus.StartTime.Time).Seconds())
		}
	} else {
		imageLoaderRuningSeconds.WithLabelValues(imageLoader.Namespace, imageLoader.Name).Set(
			time.Since(newStatus.StartTime.Time).Seconds())
	}

	if len(newStatus.FailedNodes) > 0 {
		newStatus.FailedStatuses = make([]*tkexv1alpha1.FailedStatus, 0)
		for _, node := range newStatus.FailedNodes {
			newStatus.FailedStatuses = append(newStatus.FailedStatuses, &tkexv1alpha1.FailedStatus{
				Name: node,
			})
		}
	}
}
