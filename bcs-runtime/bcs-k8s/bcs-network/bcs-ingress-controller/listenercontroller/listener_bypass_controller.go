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

package listenercontroller

import (
	"context"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	gocache "github.com/patrickmn/go-cache"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/apiclient"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// ListenerBypassReconciler reconclier for networkextensionv1 listener
type ListenerBypassReconciler struct {
	Ctx context.Context

	Client client.Client
	Option *option.ControllerOption

	// ListenerEventer record.EventRecorder
	monitorHelper *apiclient.MonitorHelper
}

// NewListenerBypassReconciler create ListenerBypassReconciler
func NewListenerBypassReconciler(ctx context.Context, client client.Client, lbIDCache *gocache.Cache,
	options *option.ControllerOption) *ListenerBypassReconciler {
	return &ListenerBypassReconciler{
		Ctx:           ctx,
		Client:        client,
		monitorHelper: apiclient.NewMonitorHelper(lbIDCache),
		Option:        options,
	}
}

// Reconcile reconclie listener
func (lc *ListenerBypassReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	metrics.IncreaseEventCounter("listener-bypass", metrics.EventTypeUnknown)

	li := &networkextensionv1.Listener{}
	if err := lc.Client.Get(lc.Ctx, req.NamespacedName, li); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		blog.Errorf("get listener %s/%s failed, err %s", req.Namespace, req.Name, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, nil
	}
	listener := li.DeepCopy()

	if listener.DeletionTimestamp != nil {
		if !common.ContainsString(listener.Finalizers, constant.FinalizerNameUptimeCheck) {
			return ctrl.Result{}, nil
		}
		if err := lc.monitorHelper.DeleteUptimeCheckTask(lc.Ctx, listener); err != nil {
			blog.Errorf("delete uptime check task for listener '%s/%s' failed, err: %s", listener.GetNamespace(),
				listener.GetName(), err.Error())
			_ = lc.updateListenerStatus(req.NamespacedName, listener.GetUptimeCheckTaskStatus().ID,
				networkextensionv1.ListenerStatusNotSynced, err.Error())
			return ctrl.Result{}, err
		}

		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			li := &networkextensionv1.Listener{}
			if err := lc.Client.Get(lc.Ctx, req.NamespacedName, li); err != nil {
				return err
			}
			cpListener := li.DeepCopy()
			cpListener.Finalizers = common.RemoveString(cpListener.Finalizers, constant.FinalizerNameUptimeCheck)

			return lc.Client.Update(lc.Ctx, cpListener)
		}); err != nil {
			blog.Errorf("remove finalizer for listeners '%s/%s' failed, err: %s", listener.GetNamespace(),
				listener.GetNamespace(), err.Error())
			return ctrl.Result{}, err
		}
		blog.Infof("remove listener '%s/%s' uptime check task success", listener.GetNamespace(), listener.GetName())
		return ctrl.Result{}, nil
	}

	if !listener.IsUptimeCheckEnable() {
		if err := lc.monitorHelper.DeleteUptimeCheckTask(lc.Ctx, listener); err != nil {
			blog.Errorf("listener '%s/%s' delete uptime check task(close uptime check) failed, err: %s",
				listener.GetNamespace(), listener.GetName(), err.Error())
			_ = lc.updateListenerStatus(req.NamespacedName, listener.GetUptimeCheckTaskStatus().ID,
				networkextensionv1.ListenerStatusNotSynced, err.Error())
			return ctrl.Result{}, err
		}
		_ = lc.updateListenerStatus(req.NamespacedName, 0, networkextensionv1.ListenerStatusSynced, "")
		return ctrl.Result{}, nil
	}

	if err := lc.ensureFinalizer(listener); err != nil {
		blog.Errorf("ensure finalizer for listeners '%s/%s' failed, err: %s", listener.GetNamespace(),
			listener.GetNamespace(), err.Error())
		return ctrl.Result{}, err
	}

	taskID, err := lc.monitorHelper.EnsureUptimeCheck(lc.Ctx, listener)
	if err != nil {
		blog.Errorf("ensure uptime check for listener '%s/%s' failed, err: %s", listener.GetNamespace(),
			listener.GetName(), err.Error())
		// 这里仍保留拨测任务ID， 避免删除错误等原因， 清理了拨测任务ID
		_ = lc.updateListenerStatus(req.NamespacedName, listener.GetUptimeCheckTaskStatus().ID, networkextensionv1.ListenerStatusNotSynced, err.Error())
		return ctrl.Result{}, err
	}

	if err = lc.updateListenerStatus(req.NamespacedName, taskID, networkextensionv1.ListenerStatusSynced,
		""); err != nil {
		blog.Errorf("update uptime check status for listeners '%s/%s' failed, err: %s", listener.GetNamespace(),
			listener.GetName(), err.Error())
		return ctrl.Result{}, err
	}

	blog.V(3).Infof("ensure listeners '%s/%s' uptime check status success", listener.GetNamespace(), listener.GetName())

	return ctrl.Result{}, nil
}

func (lc *ListenerBypassReconciler) ensureFinalizer(listener *networkextensionv1.Listener) error {
	if common.ContainsString(listener.Finalizers, constant.FinalizerNameUptimeCheck) {
		return nil
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		li := &networkextensionv1.Listener{}
		if err := lc.Client.Get(lc.Ctx, k8stypes.NamespacedName{
			Namespace: listener.Namespace,
			Name:      listener.Name,
		}, li); err != nil {
			return err
		}

		cpListener := li.DeepCopy()
		cpListener.Finalizers = append(cpListener.Finalizers, constant.FinalizerNameUptimeCheck)

		return lc.Client.Update(lc.Ctx, cpListener)
	}); err != nil {
		return err
	}
	return nil
}

// SetupWithManager set reconciler
func (lc *ListenerBypassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.Listener{}).
		WithEventFilter(getListenerByPassPredicate()).
		WithOptions(controller.Options{MaxConcurrentReconciles: lc.Option.ListenerBypassMaxConcurrent}).
		Complete(lc)
}

func (lc *ListenerBypassReconciler) updateListenerStatus(namespacedName k8stypes.NamespacedName, taskID int64,
	status string, msg string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		li := &networkextensionv1.Listener{}
		if err := lc.Client.Get(lc.Ctx, namespacedName, li); err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		cpListener := li.DeepCopy()
		cpListener.Status.UptimeCheckStatus = &networkextensionv1.UptimeCheckTaskStatus{
			ID:     taskID,
			Status: status,
			Msg:    msg,
		}
		if err := lc.Client.Update(lc.Ctx, cpListener, &client.UpdateOptions{}); err != nil {
			blog.Errorf("update uptime_check status failed, err: %s", err.Error())
			return err
		}
		return nil
	})
}

// getListenerByPassPredicate filter listener events
func getListenerByPassPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) (processed bool) {
			defer func() {
				if r := recover(); r != nil {
					blog.Errorf("[panic] Listener predicate panic, info: %v, stack:%s", r,
						string(debug.Stack()))
					processed = true
				}
			}()

			objectNew := e.ObjectNew.DeepCopyObject()
			objectOld := e.ObjectOld.DeepCopyObject()
			newListener, okNew := objectNew.(*networkextensionv1.Listener)
			oldListener, okOld := objectOld.(*networkextensionv1.Listener)
			if !okNew || !okOld {
				return false
			}
			if newListener.DeletionTimestamp != nil {
				return true
			}
			if !newListener.IsUptimeCheckEnable() && !oldListener.IsUptimeCheckEnable() {
				return false
			}
			if reflect.DeepEqual(newListener.Spec, oldListener.Spec) {
				blog.V(5).Infof("listener %+v updated, but spec not change", oldListener)
				return false
			}
			return true
		},
	}
}
