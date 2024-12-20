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

// Package portbindingcontroller controller for portbinding
package portbindingcontroller

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/apiclient"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// PortBindingByPassReconciler reconciler for bcs port pool
type PortBindingByPassReconciler struct {
	ctx           context.Context
	k8sClient     client.Client
	monitorHelper *apiclient.PortBindingItemMonitorHelper
	opts          *option.ControllerOption

	eventer record.EventRecorder
}

// NewPortBindingByPassReconciler create PortBindingByPassReconciler
func NewPortBindingByPassReconciler(ctx context.Context, k8sClient client.Client,
	eventer record.EventRecorder, opts *option.ControllerOption) *PortBindingByPassReconciler {
	return &PortBindingByPassReconciler{
		ctx:           ctx,
		k8sClient:     k8sClient,
		monitorHelper: apiclient.NewPortBindingItemMonitorHelper(),
		opts:          opts,
		eventer:       eventer,
	}
}

// Reconcile uptime check task for portBinding
func (pbr *PortBindingByPassReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("PortBinding %+v triggered", req.NamespacedName)
	pb := &networkextensionv1.PortBinding{}
	if err := pbr.k8sClient.Get(context.Background(), req.NamespacedName, pb); err != nil {
		// nolint
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		blog.Warnf("get portbinding %v failed, err %s, requeue it", req.NamespacedName, err.Error())
		return ctrl.Result{}, err
	}
	portBinding := pb.DeepCopy()

	if portBinding.DeletionTimestamp != nil || (portBinding.Status.Status == constant.
		PortBindingStatusCleaned || portBinding.Status.Status == constant.PortBindingStatusCleaning) {
		if !common.ContainsString(portBinding.Finalizers, constant.FinalizerNameUptimeCheck) {
			return ctrl.Result{}, nil
		}

		if err := pbr.monitorHelper.DeleteUptimeCheckTask(context.Background(), portBinding); err != nil {
			blog.Errorf("delete uptime check task for port binding %s/%s failed, err %s", portBinding.GetName(),
				portBinding.GetNamespace(), err.Error())
			_ = pbr.updatePortBindingStatus(portBinding)
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 3 * time.Second,
			}, nil
		}
		_ = pbr.updatePortBindingStatus(portBinding)
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			rawPb := &networkextensionv1.PortBinding{}
			if err := pbr.k8sClient.Get(context.Background(), req.NamespacedName, rawPb); err != nil {
				return err
			}
			cpPb := rawPb.DeepCopy()
			cpPb.Finalizers = common.RemoveString(cpPb.Finalizers, constant.FinalizerNameUptimeCheck)

			return pbr.k8sClient.Update(context.Background(), cpPb)
		}); err != nil {
			blog.Errorf("update port binding %s/%s finalizers failed, err %s", portBinding.GetName(),
				portBinding.GetNamespace(), err.Error())
			return ctrl.Result{}, err
		}

		blog.Infof("remove portbinding %s/%s uptime check finalizers successfully", portBinding.GetNamespace(),
			portBinding.GetName())
		return ctrl.Result{}, nil
	}

	if !portBinding.Spec.HasEnableUptimeCheck() {
		if err := pbr.monitorHelper.DeleteUptimeCheckTask(context.Background(), portBinding); err != nil {
			blog.Errorf("delete uptime check task for port binding %s/%s failed, err %s", portBinding.GetName(),
				portBinding.GetNamespace(), err.Error())
			_ = pbr.updatePortBindingStatus(portBinding)
			return ctrl.Result{}, err
		}
		_ = pbr.updatePortBindingStatus(portBinding)
		return ctrl.Result{}, nil
	}

	if err := pbr.ensureFinalizer(portBinding); err != nil {
		blog.Errorf("ensure port binding %s/%s finalizers failed, err %s", portBinding.GetName(),
			portBinding.GetNamespace(), err.Error())
		return ctrl.Result{}, err
	}

	if needRetry := pbr.monitorHelper.EnsureUptimeCheck(context.Background(), portBinding); needRetry {
		_ = pbr.updatePortBindingStatus(portBinding)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}

	if err := pbr.updatePortBindingStatus(portBinding); err != nil {
		return ctrl.Result{}, err
	}
	blog.Infof("ensure port binding %s/%s successfully", portBinding.GetName(), portBinding.GetNamespace())
	return ctrl.Result{}, nil
}

// SetupWithManager set reconciler
func (pbr *PortBindingByPassReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.PortBinding{}).
		WithEventFilter(pbr.getPortBindingPredicate()).
		WithOptions(controller.Options{MaxConcurrentReconciles: pbr.opts.ListenerBypassMaxConcurrent}).
		Complete(pbr)
}

func (pbr *PortBindingByPassReconciler) getPortBindingPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) (processed bool) {
			defer func() {
				if r := recover(); r != nil {
					blog.Errorf("[panic] PortBinding predicate panic, info: %v, stack:%s", r,
						string(debug.Stack()))
					processed = true
				}
			}()
			objectNew := e.ObjectNew.DeepCopyObject()
			objectOld := e.ObjectOld.DeepCopyObject()
			newPoolBinding, okNew := objectNew.(*networkextensionv1.PortBinding)
			oldPoolBinding, okOld := objectOld.(*networkextensionv1.PortBinding)
			if !okNew || !okOld {
				return true
			}
			if newPoolBinding.DeletionTimestamp != nil {
				return true
			}
			if !newPoolBinding.Spec.HasEnableUptimeCheck() && !oldPoolBinding.Spec.HasEnableUptimeCheck() {
				return false
			}
			if reflect.DeepEqual(newPoolBinding.Spec, oldPoolBinding.Spec) &&
				newPoolBinding.Status.Status == oldPoolBinding.Status.Status {
				return false
			}

			return true
		},
	}
}

func (pbr *PortBindingByPassReconciler) updatePortBindingStatus(portBinding *networkextensionv1.PortBinding) error {
	if e := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		rawPb := &networkextensionv1.PortBinding{}
		if err := pbr.k8sClient.Get(context.Background(), k8stypes.NamespacedName{
			Namespace: portBinding.Namespace,
			Name:      portBinding.Name,
		}, rawPb); err != nil {
			return fmt.Errorf("get portBinding %s/%s failed, err %s", portBinding.Namespace, portBinding.Name, err.Error())
		}

		cpPb := rawPb.DeepCopy()
		for _, itemStatus := range cpPb.Status.PortBindingStatusList {
			for _, editStatus := range portBinding.Status.PortBindingStatusList {
				if itemStatus.GetFullKey() == editStatus.GetFullKey() {
					itemStatus.UptimeCheckStatus = editStatus.UptimeCheckStatus
				}
			}
		}
		if err := pbr.k8sClient.Status().Update(context.Background(), cpPb, &client.UpdateOptions{}); err != nil {
			return fmt.Errorf("update portBinding[%s/%s] status failed, err %s", portBinding.GetNamespace(),
				portBinding.GetName(), err.Error())
		}
		return nil
	}); e != nil {
		blog.Errorf("update portBinding[%s/%s] status failed, err %s", portBinding.GetNamespace(),
			portBinding.GetName(), e.Error())
		return e
	}

	return nil
}

func (pbr *PortBindingByPassReconciler) ensureFinalizer(portBinding *networkextensionv1.PortBinding) error {
	if common.ContainsString(portBinding.Finalizers, constant.FinalizerNameUptimeCheck) {
		return nil
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pb := &networkextensionv1.PortBinding{}
		if err := pbr.k8sClient.Get(context.Background(), k8stypes.NamespacedName{
			Namespace: portBinding.Namespace,
			Name:      portBinding.Name,
		}, pb); err != nil {
			return err
		}

		cpPb := pb.DeepCopy()
		cpPb.Finalizers = append(cpPb.Finalizers, constant.FinalizerNameUptimeCheck)

		return pbr.k8sClient.Update(context.Background(), cpPb)
	}); err != nil {
		return err
	}
	return nil
}
