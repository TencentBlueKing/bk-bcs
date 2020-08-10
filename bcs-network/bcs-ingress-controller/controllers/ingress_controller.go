/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/option"
)

// IngressReconciler reconciler for bcs ingress in network extension
type IngressReconciler struct {
	Ctx context.Context

	client.Client
	Log    logr.Logger
	Option *option.ControllerOption

	IngressEventer record.EventRecorder

	SvcFilter *ServiceFilter
	PodFilter *PodFilter
	stsFilter *StatefulSetFilter
}

// Reconcile reconcile bcs ingress
func (ir *IngressReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	
	return ctrl.Result{}, nil
}

// SetupWithManager set reconciler
func (ir *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.Ingress{}).
		Watches(&source.Kind{Type: &k8scorev1.Pod{}}, ir.PodFilter).
		Watches(&source.Kind{Type: &k8scorev1.Service{}}, ir.SvcFilter).
		Watches(&source.Kind{Type: &k8sappsv1.StatefulSet{}}, ir.stsFilter).
		Complete(ir)
}
