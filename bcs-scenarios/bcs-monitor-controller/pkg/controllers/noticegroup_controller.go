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
	"fmt"
	"os"
	"reflect"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/apiclient"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/fileoperator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/utils"
)

// NoticeGroupReconciler reconciles a NoticeGroup object
type NoticeGroupReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Ctx           context.Context
	FileOp        *fileoperator.FileOperator
	MonitorApiCli apiclient.IMonitorApiClient
}

// +kubebuilder:rbac:groups=monitorextension.bkbcs.tencent.com,resources=noticegroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitorextension.bkbcs.tencent.com,resources=noticegroups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitorextension.bkbcs.tencent.com,resources=noticegroups/finalizers,verbs=update

// Reconcile notice group
func (r *NoticeGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	blog.Infof("NoticeGroup '%s' triggered", req.NamespacedName)

	noticeGroup := &monitorextensionv1.NoticeGroup{}
	if err := r.Get(context.Background(), req.NamespacedName, noticeGroup); err != nil {
		if !k8serrors.IsNotFound(err) {
			blog.Errorf("Get NoticeGroup '%s' failed, err: %s", req.NamespacedName.String(), err.Error())
			return ctrl.Result{}, err
		}

		blog.Infof("NoticeGroup '%s' is deleted, skip...", req.NamespacedName.String())
		return ctrl.Result{}, nil
	}

	if noticeGroup.DeletionTimestamp != nil {
		blog.Infof("found deleting notice group '%s'", req.NamespacedName)
		if err := r.processDelete(noticeGroup); err != nil {
			return ctrl.Result{}, err
		}

		blog.Infof("delete notice group '%s' success", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if err := r.checkFinalizer(noticeGroup); err != nil {
		return ctrl.Result{}, err
	}

	outputPath, err := r.FileOp.Compress(noticeGroup)
	if err != nil {
		blog.Errorf("compress notice group '%s/%s' failed, err: %s", noticeGroup.Namespace, noticeGroup.Name, err.Error())
		if inErr := r.updateSyncStatus(noticeGroup, monitorextensionv1.SyncStateFailed, err); inErr != nil {
			blog.Warnf("update noticeGroup '%s/%s' sync status failed, err: %s", noticeGroup.GetNamespace(),
				noticeGroup.GetName(), inErr.Error())
		}
		return ctrl.Result{}, err
	}
	defer os.Remove(outputPath)

	if err = r.MonitorApiCli.UploadConfig(noticeGroup.Spec.BizID, noticeGroup.Spec.BizToken, outputPath,
		r.getAppName(noticeGroup), noticeGroup.Spec.Override); err != nil {
		blog.Errorf("upload config to monitor failed, err: %s", err.Error())
		if inErr := r.updateSyncStatus(noticeGroup, monitorextensionv1.SyncStateFailed, err); inErr != nil {
			blog.Warnf("update noticeGroup '%s/%s' sync status failed, err: %s", noticeGroup.GetNamespace(),
				noticeGroup.GetName(), inErr.Error())
		}
		return ctrl.Result{}, err
	}

	blog.Infof("sync notice group '%s' success", req.NamespacedName)
	if inErr := r.updateSyncStatus(noticeGroup, monitorextensionv1.SyncStateCompleted, nil); inErr != nil {
		blog.Warnf("update notice group '%s/%s' sync status failed, err: %s", noticeGroup.GetNamespace(),
			noticeGroup.GetName(), inErr.Error())
	}

	return ctrl.Result{}, nil
}

func (r *NoticeGroupReconciler) eventPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			ng := createEvent.Object.(*monitorextensionv1.NoticeGroup)
			if ng.DeletionTimestamp == nil && ng.Status.SyncStatus.State == monitorextensionv1.SyncStateCompleted &&
				ng.Spec.IgnoreChange == true {
				blog.V(3).Infof("notice group '%s/%s' got create event, but is synced and ignore change",
					ng.GetNamespace(), ng.GetName())
				return false
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newNg, okNew := e.ObjectNew.(*monitorextensionv1.NoticeGroup)
			oldNg, okOld := e.ObjectOld.(*monitorextensionv1.NoticeGroup)
			if !okNew || !okOld {
				return true
			}
			if reflect.DeepEqual(newNg.Spec, oldNg.Spec) &&
				reflect.DeepEqual(newNg.Finalizers, oldNg.Finalizers) &&
				reflect.DeepEqual(newNg.DeletionTimestamp, oldNg.DeletionTimestamp) {
				blog.V(5).Infof("noticeGroup %+v updated, but spec and finalizer and deletionTimestamp not change",
					newNg)
				return false
			}
			if newNg.DeletionTimestamp == nil && newNg.Status.SyncStatus.
				State == monitorextensionv1.SyncStateCompleted && newNg.Spec.IgnoreChange == true {
				blog.V(3).Infof("noticeGroup '%s/%s' updated, but is synced and ignore change",
					newNg.GetNamespace(), newNg.GetName())
				return false
			}
			return true
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NoticeGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitorextensionv1.NoticeGroup{}).
		WithEventFilter(r.eventPredicate()).
		Complete(r)
}

func (r *NoticeGroupReconciler) updateSyncStatus(noticeGroup *monitorextensionv1.NoticeGroup,
	state monitorextensionv1.SyncState, err error) error {
	blog.Infof("Update sync state of noticeGroup (%s/%s) to %s", noticeGroup.GetNamespace(), noticeGroup.GetName(),
		state)
	noticeGroup.Status.SyncStatus.State = state
	// err message
	if err != nil {
		noticeGroup.Status.SyncStatus.Message = err.Error()
	} else {
		noticeGroup.Status.SyncStatus.Message = ""
	}
	noticeGroup.Status.SyncStatus.LastSyncTime = metav1.NewTime(time.Now())
	noticeGroup.Status.SyncStatus.App = r.getAppName(noticeGroup)
	if inErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return r.Client.Status().Update(r.Ctx, noticeGroup)
	}); inErr != nil {
		blog.Warnf("update noticeGroup'%s/%s' failed, err: %s", noticeGroup.GetNamespace(), noticeGroup.GetName(), inErr.Error())
		return inErr
	}

	return nil
}

func (r *NoticeGroupReconciler) getAppName(noticeGroup *monitorextensionv1.NoticeGroup) string {
	return fmt.Sprintf("bcs-ng-%s-%s", noticeGroup.Spec.Scenario, noticeGroup.GetName())
}

// checkFinalizer add finalizer if not exist
func (r *NoticeGroupReconciler) checkFinalizer(noticeGroup *monitorextensionv1.NoticeGroup) error {
	if utils.ContainsString(noticeGroup.Finalizers, FinalizerMonitorController) {
		return nil
	}

	noticeGroup.Finalizers = append(noticeGroup.Finalizers, FinalizerMonitorController)
	if err := r.Update(r.Ctx, noticeGroup); err != nil {
		blog.Warnf("Update notice group '%s/%s' failed, err: %s", noticeGroup.Namespace, noticeGroup.Name,
			err.Error())
		return err
	}
	return nil
}

func (r *NoticeGroupReconciler) removeFinalizer(notice *monitorextensionv1.NoticeGroup) error {
	notice.Finalizers = utils.RemoveString(notice.Finalizers, FinalizerMonitorController)
	if err := r.Update(context.Background(), notice, &client.UpdateOptions{}); err != nil {
		blog.Warnf("remove finalizer for notice %s/%s failed, err %s", notice.GetNamespace(), notice.GetName(),
			err.Error())
		return fmt.Errorf("remove finalizer for notice %s/%s failed, err %s", notice.GetNamespace(),
			notice.GetName(), err.Error())
	}
	blog.V(3).Infof("remove finalizer for notice %s/%s successfully", notice.GetNamespace(),
		notice.GetName())
	return nil
}

func (r *NoticeGroupReconciler) processDelete(noticeGroup *monitorextensionv1.NoticeGroup) error {
	if err := r.MonitorApiCli.UploadConfig(noticeGroup.Spec.BizID, noticeGroup.Spec.BizToken, EmptyTARLocation,
		r.getAppName(noticeGroup), noticeGroup.Spec.Override); err != nil {
		blog.Errorf("upload config to monitor failed, err: %s", err.Error())
		if inErr := r.updateSyncStatus(noticeGroup, monitorextensionv1.SyncStateFailed, err); inErr != nil {
			blog.Warnf("update noticeGroup '%s/%s' sync status failed, err: %s", noticeGroup.GetNamespace(),
				noticeGroup.GetName(), inErr.Error())
		}
		return err
	}

	if err := r.removeFinalizer(noticeGroup); err != nil {
		return err
	}

	blog.Infof("delete noticeGroup '%s/%s' success", noticeGroup.GetNamespace(), noticeGroup.GetName())
	return nil
}
