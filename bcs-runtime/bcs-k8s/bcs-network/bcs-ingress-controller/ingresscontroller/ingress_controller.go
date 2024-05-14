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

package ingresscontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	federationv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/federation/v1"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/ingresscache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
	netcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// IngressReconciler reconciler for bcs ingress in network extension
type IngressReconciler struct {
	Ctx context.Context

	Client client.Client
	Log    logr.Logger
	Option *option.ControllerOption

	IngressEventer record.EventRecorder

	EpsFIlter                       *EndpointsFilter
	PodFilter                       *PodFilter
	MultiClusterEndpointSliceFilter *MultiClusterEpsFilter

	IngressConverter *generator.IngressConverter

	Cache ingresscache.IngressCache
}

// getIngressPredicate filter ingress events
func (ir *IngressReconciler) getIngressPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			ingress, ok := createEvent.Object.(*networkextensionv1.Ingress)
			if ok {
				blog.V(5).Infof("add ingress'%s/%s' cache", ingress.GetNamespace(), ingress.GetName())
				ir.Cache.Add(ingress)
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newIngress, okNew := e.ObjectNew.(*networkextensionv1.Ingress)
			oldIngress, okOld := e.ObjectOld.(*networkextensionv1.Ingress)
			if !okNew || !okOld {
				return true
			}
			if reflect.DeepEqual(newIngress.Spec, oldIngress.Spec) &&
				reflect.DeepEqual(newIngress.Annotations, oldIngress.Annotations) &&
				reflect.DeepEqual(newIngress.Finalizers, oldIngress.Finalizers) &&
				reflect.DeepEqual(newIngress.DeletionTimestamp, oldIngress.DeletionTimestamp) {
				blog.V(5).Infof("ingress %+v updated, but spec and annotation and finalizer not change", newIngress)
				return false
			}
			blog.V(5).Infof("update ingress'%s/%s' cache", newIngress.GetNamespace(), newIngress.GetName())
			ir.Cache.Remove(oldIngress)
			ir.Cache.Add(newIngress)
			return true
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			ingress, ok := deleteEvent.Object.(*networkextensionv1.Ingress)
			if ok {
				blog.V(5).Infof("delete ingress'%s/%s' cache", ingress.GetNamespace(), ingress.GetName())
				ir.Cache.Remove(ingress)
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
			blog.Infof("ingress %s/%s deleted successfully", req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		blog.Errorf("get ingress %s/%s failed, err %s", req.Name, req.Namespace, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, err
	}

	// ingress is deleted
	if ingress.DeletionTimestamp != nil {
		// should remove ingress finalizer in ProcessDeleteIngress
		if retry, err := ir.IngressConverter.ProcessDeleteIngress(req.Name, req.Namespace); err != nil {
			metrics.IncreaseFailMetric(metrics.ObjectIngress, metrics.EventTypeDelete)
			blog.Errorf("process deleted ingress %s/%s failed, err %s", req.Name, req.Namespace, err.Error())
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, nil
		} else if retry {
			blog.V(4).Infof("process deleted ingress %s/%s retry", req.Name, req.Namespace)
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, nil
		}
		if err := ir.removeFinalizerForIngress(ingress); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, nil
		}
		return ctrl.Result{}, nil
	}

	// if doesn't has finalizer, add finalizer
	if !netcommon.ContainsString(ingress.Finalizers, constant.FinalizerNameBcsIngressController) {
		if err := ir.addFinalizerForIngress(ingress); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, nil
		}
		return ctrl.Result{}, nil
	}

	warnings, err := ir.IngressConverter.ProcessUpdateIngress(ingress)
	if err != nil {
		warnings = append(warnings, err.Error())
	}
	if derr := ir.patchWarningAnnotationForIngress(ingress, warnings); derr != nil {
		blog.Warnf(derr.Error())
	}

	if err != nil {
		metrics.IncreaseFailMetric(metrics.ObjectIngress, metrics.EventTypeUnknown)
		// create event for ingress
		ir.IngressEventer.Eventf(ingress, k8scorev1.EventTypeWarning,
			"process ingress failed", "error: %s", err.Error())
		blog.Errorf("process ingress %s/%s event failed, err %s", req.Name, req.Namespace, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, nil
	}
	ir.IngressEventer.Eventf(ingress, k8scorev1.EventTypeNormal, "EnsureSuccess", "Ensure success")

	return ctrl.Result{}, nil
}

func (ir *IngressReconciler) removeFinalizerForIngress(ingress *networkextensionv1.Ingress) error {
	ingress.Finalizers = netcommon.RemoveString(ingress.Finalizers, constant.FinalizerNameBcsIngressController)
	if err := ir.Client.Update(context.Background(), ingress, &client.UpdateOptions{}); err != nil {
		blog.Warnf("remove finalizer for ingress %s/%s failed, err %s",
			ingress.GetName(), ingress.GetNamespace(), err.Error())
		return fmt.Errorf("remove finalizer for ingress %s/%s failed, err %s",
			ingress.GetName(), ingress.GetNamespace(), err.Error())
	}
	blog.V(3).Infof("remove finalizer for ingress %s/%s successfully", ingress.GetName(), ingress.GetNamespace())
	return nil
}

// nolint unused
func (ir *IngressReconciler) addFinalizerForIngress(ingress *networkextensionv1.Ingress) error {
	ingress.Finalizers = append(ingress.Finalizers, constant.FinalizerNameBcsIngressController)
	if err := ir.Client.Update(context.Background(), ingress, &client.UpdateOptions{}); err != nil {
		blog.Warnf("add finalizer for ingress %s/%s failed, err %s",
			ingress.GetName(), ingress.GetNamespace(), err.Error())
	}
	blog.V(3).Infof("add finalizer for ingress %s/%s successfully", ingress.GetName(), ingress.GetNamespace())
	return nil
}

func (ir IngressReconciler) patchWarningAnnotationForIngress(ingress *networkextensionv1.Ingress,
	warnings []string) error {
	attachWarning := strings.Join(warnings, ";")

	// if warning annotation not changed, not patch
	if existedWarning, ok := ingress.Annotations[networkextensionv1.
		AnnotationKeyForWarnings]; ok && existedWarning == attachWarning {
		return nil
	}
	patchStruct := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				networkextensionv1.AnnotationKeyForWarnings: attachWarning,
			},
		},
	}
	patchBytes, err := json.Marshal(patchStruct)
	if err != nil {
		return errors.Wrapf(err, "marshal patchStruct for ingress '%s/%s' failed", ingress.GetNamespace(),
			ingress.GetName())
	}
	rawPatch := client.RawPatch(k8stypes.MergePatchType, patchBytes)
	updateIngress := &networkextensionv1.Ingress{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      ingress.GetName(),
			Namespace: ingress.GetNamespace(),
		},
	}
	if err := ir.Client.Patch(context.Background(), updateIngress, rawPatch, &client.PatchOptions{}); err != nil {
		return errors.Wrapf(err, "patch ingress %s/%s annotation failed, patcheStruct: %s", ingress.GetName(),
			ingress.GetNamespace(), string(patchBytes))
	}

	if len(warnings) > 0 {
		ir.IngressEventer.Eventf(updateIngress, k8scorev1.EventTypeWarning, "ingress warning", attachWarning)
	}
	return nil
}

// SetupWithManager set reconciler
func (ir *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.Ingress{}).
		Watches(&source.Kind{Type: &k8scorev1.Pod{}}, ir.PodFilter).
		Watches(&source.Kind{Type: &k8scorev1.Endpoints{}}, ir.EpsFIlter).
		WithEventFilter(ir.getIngressPredicate())
	if ir.Option.IsFederationMode {
		builder = builder.Watches(&source.Kind{Type: &federationv1.MultiClusterEndpointSlice{}},
			ir.MultiClusterEndpointSliceFilter)
	}

	return builder.Complete(ir)
}
