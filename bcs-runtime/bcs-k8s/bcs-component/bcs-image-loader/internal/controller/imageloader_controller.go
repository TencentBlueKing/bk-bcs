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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-image-loader/api/v1alpha1"
)

// ImageLoaderReconciler reconciles a ImageLoader object
type ImageLoaderReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
	APIReader client.Reader
}

var logger logr.Logger

// +kubebuilder:rbac:groups=tkex.tencent.com,resources=imageloaders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tkex.tencent.com,resources=imageloaders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tkex.tencent.com,resources=imageloaders/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ImageLoaderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger = log.FromContext(ctx)

	originImageLoader := &tkexv1alpha1.ImageLoader{}
	if err := r.APIReader.Get(ctx, req.NamespacedName, originImageLoader); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if originImageLoader.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	logger.Info("start reconcile")
	imageLoader := originImageLoader.DeepCopy()
	newStatus, requeueTime, err := r.reconcileImageLoader(ctx, imageLoader)
	if err != nil {
		logger.Error(err, "reconcile imageLoader error")
		return ctrl.Result{}, err
	}
	imageLoader.Status = *newStatus
	err = r.Client.Status().Patch(ctx, imageLoader, client.MergeFrom(originImageLoader))
	if err != nil {
		logger.Error(err, "failed to update imageLoader status")
		return ctrl.Result{}, err
	}
	if requeueTime != nil {
		logger.Info("imageLoader will be requeue", "requeue time", *requeueTime)
		return ctrl.Result{Requeue: true, RequeueAfter: *requeueTime}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImageLoaderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tkexv1alpha1.ImageLoader{}).
		Owns(&corev1.Pod{}).
		Watches(&corev1.Pod{}, &EmptyEventHandler{}).
		Watches(&corev1.Node{}, &EmptyEventHandler{}).
		Complete(r)
}
