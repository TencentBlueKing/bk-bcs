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

package ingresscontroller

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/generator"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/option"
)

// IngressReconciler reconciler for bcs ingress in network extension
type IngressReconciler struct {
	Ctx context.Context

	Client client.Client
	Log    logr.Logger
	Option *option.ControllerOption

	IngressEventer record.EventRecorder

	SvcFilter *ServiceFilter
	EpsFilter *EndpointsFilter
	PodFilter *PodFilter
	StsFilter *StatefulSetFilter

	IngressConverter *generator.IngressConverter
}

// getIngressPredicate filter ingress events
func getIngressPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newIngress, okNew := e.ObjectNew.(*networkextensionv1.Ingress)
			oldIngress, okOld := e.ObjectOld.(*networkextensionv1.Ingress)
			if !okNew || !okOld {
				return true
			}
			if reflect.DeepEqual(newIngress.Spec, oldIngress.Spec) &&
				reflect.DeepEqual(newIngress.Annotations, oldIngress.Annotations) {
				blog.V(5).Infof("ingress %+v updated, but spec and annotation not change", newIngress)
				return false
			}
			return true
		},
	}
}

// Reconcile reconcile bcs ingress
func (ir *IngressReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	metrics.IncreaseEventCounter("ingress", metrics.EventTypeUnknown)

	blog.V(3).Infof("ingress %+v triggered", req.NamespacedName)
	ingress := &networkextensionv1.Ingress{}
	if err := ir.Client.Get(ir.Ctx, req.NamespacedName, ingress); err != nil {
		if k8serrors.IsNotFound(err) {
			if inErr := ir.IngressConverter.ProcessDeleteIngress(req.Name, req.Namespace); inErr != nil {
				blog.Errorf("process deleted ingress %s/%s failed, err %s", req.Name, req.Namespace, inErr.Error())
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: time.Duration(5 * time.Second),
				}, nil
			}
			return ctrl.Result{}, nil
		}
		blog.Errorf("get ingress %s/%s failed, err %s", req.Name, req.Namespace, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(5 * time.Second),
		}, err
	}
	if err := ir.IngressConverter.ProcessUpdateIngress(ingress); err != nil {
		// create event for ingress
		ir.IngressEventer.Eventf(ingress, k8scorev1.EventTypeWarning,
			"process ingress failed", "error: %s", err.Error())
		blog.Errorf("process ingress %s/%s event failed, err %s", req.Name, req.Namespace, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Duration(5 * time.Second),
		}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager set reconciler
func (ir *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.Ingress{}).
		Watches(&source.Kind{Type: &k8scorev1.Pod{}}, ir.PodFilter).
		Watches(&source.Kind{Type: &k8scorev1.Service{}}, ir.SvcFilter).
		Watches(&source.Kind{Type: &k8scorev1.Endpoints{}}, ir.EpsFilter).
		Watches(&source.Kind{Type: &k8sappsv1.StatefulSet{}}, ir.StsFilter).
		WithEventFilter(getIngressPredicate()).
		Complete(ir)
}
