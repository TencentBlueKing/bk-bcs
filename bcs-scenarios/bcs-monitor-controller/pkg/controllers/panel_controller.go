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

package controllers

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/apiclient"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/fileoperator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/utils"
)

// PanelReconciler reconciles a Panel object
type PanelReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Ctx           context.Context
	FileOp        *fileoperator.FileOperator
	MonitorApiCli apiclient.IMonitorApiClient
}

// +kubebuilder:rbac:groups=monitorextension.bkbcs.tencent.com,
// resources=panels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitorextension.bkbcs.tencent.com,resources=panels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitorextension.bkbcs.tencent.com,resources=panels/finalizers,verbs=update

// Reconcile reconcile panel
func (r *PanelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	blog.Infof("Panel '%s' triggered", req.NamespacedName)

	panel := &monitorextensionv1.Panel{}
	if err := r.Get(context.Background(), req.NamespacedName, panel); err != nil {
		if !k8serrors.IsNotFound(err) {
			blog.Errorf("Get NoticeGroup '%s' failed, err: %s", req.NamespacedName.String(), err.Error())
			return ctrl.Result{}, err
		}

		blog.Infof("NoticeGroup '%s' is deleted, skip...", req.NamespacedName.String())
		return ctrl.Result{}, nil
	}

	if panel.DeletionTimestamp != nil {
		blog.Infof("found deleting panel '%s'", req.NamespacedName)
		if err := r.processDelete(panel); err != nil {
			return ctrl.Result{}, err
		}

		blog.Infof("delete panel '%s' success", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if err := r.checkFinalizer(panel); err != nil {
		return ctrl.Result{}, err
	}

	outputPath, err := r.FileOp.Compress(panel)
	if err != nil {
		blog.Errorf("compress panel '%s/%s' failed, err: %s", panel.Namespace, panel.Name, err.Error())
		if inErr := r.updateSyncStatus(panel, monitorextensionv1.SyncStateFailed, err); inErr != nil {
			blog.Warnf("update panel '%s/%s' sync status failed, err: %s", panel.GetNamespace(),
				panel.GetName(), inErr.Error())
		}
		return ctrl.Result{}, err
	}
	defer os.RemoveAll(outputPath)

	if err = r.MonitorApiCli.UploadConfig(panel.Spec.BizID, panel.Spec.BizToken, outputPath,
		r.getAppName(panel), panel.Spec.Override); err != nil {
		blog.Errorf("upload config to monitor failed, err: %s", err.Error())
		if inErr := r.updateSyncStatus(panel, monitorextensionv1.SyncStateFailed, err); inErr != nil {
			blog.Warnf("update panel '%s/%s' sync status failed, err: %s", panel.GetNamespace(),
				panel.GetName(), inErr.Error())
		}
		return ctrl.Result{}, err
	}

	blog.Infof("sync panel '%s' success", req.NamespacedName)
	if inErr := r.updateSyncStatus(panel, monitorextensionv1.SyncStateCompleted, nil); inErr != nil {
		blog.Warnf("update panel '%s/%s' sync status failed, err: %s", panel.GetNamespace(),
			panel.GetName(), inErr.Error())
	}
	return ctrl.Result{}, nil
}

func (r *PanelReconciler) eventPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			panel, ok := createEvent.Object.(*monitorextensionv1.Panel)
			if !ok {
				return true
			}
			if panel.DeletionTimestamp == nil &&
				panel.Status.SyncStatus.State == monitorextensionv1.SyncStateCompleted && panel.Spec.IgnoreChange {
				blog.V(3).Infof("panel '%s/%s' got create event, but is synced and ignore change",
					panel.GetNamespace(), panel.GetName())
				return false
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newPanel, okNew := e.ObjectNew.(*monitorextensionv1.Panel)
			oldPanel, okOld := e.ObjectOld.(*monitorextensionv1.Panel)
			if !okNew || !okOld {
				return true
			}
			if reflect.DeepEqual(newPanel.Spec, oldPanel.Spec) &&
				reflect.DeepEqual(newPanel.Finalizers, oldPanel.Finalizers) &&
				reflect.DeepEqual(newPanel.DeletionTimestamp, oldPanel.DeletionTimestamp) {
				blog.V(5).Infof("panel %+v updated, but spec and finalizer and deletionTimestamp not change",
					newPanel)
				return false
			}
			if newPanel.DeletionTimestamp == nil && newPanel.Status.
				SyncStatus.State == monitorextensionv1.SyncStateCompleted && newPanel.Spec.IgnoreChange {
				blog.V(3).Infof("panel '%s/%s' updated, but is synced and ignore change",
					newPanel.GetNamespace(), newPanel.GetName())
				return false
			}
			return true
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PanelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitorextensionv1.Panel{}).
		Watches(&source.Kind{Type: &v1.ConfigMap{}}, &configmapFilter{r.Client}).
		WithEventFilter(r.eventPredicate()).
		Complete(r)
}

func (r *PanelReconciler) updateSyncStatus(panel *monitorextensionv1.Panel, state monitorextensionv1.SyncState,
	err error) error {
	blog.Infof("Update sync state of panel (%s/%s) to %s", panel.GetNamespace(), panel.GetName(), state)
	panel.Status.SyncStatus.State = state
	// err message
	if err != nil {
		panel.Status.SyncStatus.Message = err.Error()
	} else {
		panel.Status.SyncStatus.Message = ""
	}
	panel.Status.SyncStatus.LastSyncTime = metav1.NewTime(time.Now())
	panel.Status.SyncStatus.App = r.getAppName(panel)
	if inErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return r.Client.Status().Update(r.Ctx, panel)
	}); inErr != nil {
		blog.Warnf("update panel'%s/%s' failed, err: %s", panel.GetNamespace(), panel.GetName(), inErr.Error())
		return inErr
	}

	return nil
}

func (r *PanelReconciler) getAppName(panel *monitorextensionv1.Panel) string {
	return fmt.Sprintf("bcs-panel-%s-%s", panel.Spec.Scenario, panel.GetName())
}

// checkFinalizer add finalizer if not exist
func (r *PanelReconciler) checkFinalizer(panel *monitorextensionv1.Panel) error {
	if utils.ContainsString(panel.Finalizers, FinalizerMonitorController) {
		return nil
	}

	panel.Finalizers = append(panel.Finalizers, FinalizerMonitorController)
	if err := r.Update(r.Ctx, panel); err != nil {
		blog.Warnf("Update panel '%s/%s' failed, err: %s", panel.Namespace, panel.Name,
			err.Error())
		return err
	}
	return nil
}

func (r *PanelReconciler) removeFinalizer(panel *monitorextensionv1.Panel) error {
	panel.Finalizers = utils.RemoveString(panel.Finalizers, FinalizerMonitorController)
	if err := r.Update(context.Background(), panel, &client.UpdateOptions{}); err != nil {
		blog.Warnf("remove finalizer for panel %s/%s failed, err %s", panel.GetNamespace(), panel.GetName(),
			err.Error())
		return fmt.Errorf("remove finalizer for panel %s/%s failed, err %s", panel.GetNamespace(),
			panel.GetName(), err.Error())
	}
	blog.V(3).Infof("remove finalizer for panel %s/%s successfully", panel.GetNamespace(),
		panel.GetName())
	return nil
}

func (r *PanelReconciler) processDelete(panel *monitorextensionv1.Panel) error {
	if err := r.MonitorApiCli.UploadConfig(panel.Spec.BizID, panel.Spec.BizToken, EmptyTARLocation,
		r.getAppName(panel), panel.Spec.Override); err != nil {
		blog.Errorf("upload config to monitor failed, err: %s", err.Error())
		if inErr := r.updateSyncStatus(panel, monitorextensionv1.SyncStateFailed, err); inErr != nil {
			blog.Warnf("update panel '%s/%s' sync status failed, err: %s", panel.GetNamespace(),
				panel.GetName(), inErr.Error())
		}
		return err
	}

	if err := r.removeFinalizer(panel); err != nil {
		return err
	}

	blog.Infof("delete panel '%s/%s' success", panel.GetNamespace(), panel.GetName())
	return nil
}

type configmapFilter struct {
	cli client.Client
}

// Create implement EventFilter
func (cf *configmapFilter) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
}

// Update implement EventFilter
func (cf *configmapFilter) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	newCm, okNew := e.ObjectNew.(*v1.ConfigMap)
	oldCm, okOld := e.ObjectOld.(*v1.ConfigMap)
	if !okNew || !okOld {
		blog.Warnf("recv create new object is not Pod, event %+v", e)
		return
	}

	if reflect.DeepEqual(newCm.Data, oldCm.Data) {
		return
	}

	scenarioName, ok := newCm.GetLabels()[monitorextensionv1.LabelKeyForScenarioName]
	if !ok {
		blog.V(4).Infof("cm '%s/%s' not found related scenario, break...")
		return
	}
	selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForScenarioName: scenarioName,
	}))
	if err != nil {
		blog.Errorf("generate scenario '%s' label selector failed, err: %s", scenarioName, err.Error())
		return
	}

	panelList := &monitorextensionv1.PanelList{}
	if inErr := cf.cli.List(context.Background(), panelList, &client.ListOptions{
		LabelSelector: selector,
	}); inErr != nil {
		blog.Errorf("list panel failed, err: %s", err.Error())
		return
	}

	for _, panel := range panelList.Items {
		q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      panel.GetName(),
			Namespace: panel.GetNamespace(),
		}})
		blog.Infof("configmap['%s/%s'] scenario['%s'] enqueue related panel '%s/%s'", newCm.GetNamespace(),
			newCm.GetName(), scenarioName, panel.GetNamespace(), panel.GetName())
	}
}

// Delete implement EventFilter
func (cf *configmapFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
}

// Generic implement EventFilter
func (cf *configmapFilter) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
}
