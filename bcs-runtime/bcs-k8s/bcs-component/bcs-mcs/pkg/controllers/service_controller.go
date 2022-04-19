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

package controllers

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	mcsv1alpha1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

const (
	// ServiceControllerName is the name of the service controller
	ServiceControllerName = "service-controller"
)

// ServiceController is the controller for Service
type ServiceController struct {
	client.Client
}

// Reconcile 当service变更之后，同步将serviceImport的IP进行更新
func (c *ServiceController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("Reconcile service %s/%s", req.Namespace, req.Name)
	var service corev1.Service
	if err := c.Get(ctx, req.NamespacedName, &service); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if service.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}
	serviceImportName := serviceImportOwner(service.OwnerReferences)
	if serviceImportName == "" {
		return ctrl.Result{}, nil
	}
	var svcImport mcsv1alpha1.ServiceImport
	if err := c.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: serviceImportName}, &svcImport); err != nil {
		return ctrl.Result{}, err
	}

	if len(svcImport.Spec.IPs) > 0 {
		return ctrl.Result{}, nil
	}

	svcImport.Spec.IPs = []string{service.Spec.ClusterIP}
	if err := c.Update(ctx, &svcImport); err != nil {
		return ctrl.Result{}, err
	}
	klog.Infof("updated serviceImport ip", "ip", service.Spec.ClusterIP, "serviceImport", serviceImportName, "namespace", req.Namespace)
	return ctrl.Result{}, nil
}

// SetupWithManager wires up the controller.
func (c *ServiceController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&corev1.Service{}).Complete(c)
}

// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch
func serviceImportOwner(refs []metav1.OwnerReference) string {
	for _, ref := range refs {
		if ref.APIVersion == mcsv1alpha1.GroupVersion.String() && ref.Kind == utils.ServiceImportKind {
			return ref.Name
		}
	}
	return ""
}
