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
	"reflect"
	"time"

	"github.com/google/uuid"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apitypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/core"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
)

// TerraformReconciler reconciles a Terraform object
type TerraformReconciler struct {
	// Client k8s api server client
	client.Client
	// Scheme runtime scheme
	Scheme *runtime.Scheme
	// Config opt
	Config *option.ControllerOption
}

// Reconcile reconcile terraform
// +kubebuilder:rbac:groups=terraformextesions.bkbcs.tencent.com,resources=terraforms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=terraformextesions.bkbcs.tencent.com,resources=terraforms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=terraformextesions.bkbcs.tencent.com,resources=terraforms/finalizers,verbs=update
func (r *TerraformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	terraform := tfv1.Terraform{}
	traceId := uuid.New().String()
	nn := req.NamespacedName.String()
	if err := r.Get(ctx, req.NamespacedName, &terraform); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Errorf("terraform '%s' is deleted, skipped", nn)
			return ctrl.Result{}, nil
		}
	}
	blog.Infof("reconcile receive terraform, tf: %s, trace-id: %s", nn, traceId)

	if terraform.DeletionTimestamp != nil { // note: 禁用删除功能
		//blog.Infof("delete terraform: %s, trace-id: %s", utils.ToJsonString(terraform), traceId)
		//if err := handler.Init(); err != nil {
		//	blog.Errorf("core handler init failed, err: %s", err.Error())
		//	return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
		//}
		//if err := handler.Delete(); err != nil {
		//	blog.Errorf("core handler delete terraform resource failed, err: %s", err.Error())
		//	return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
		//}
		////defer r.clean(handler)
		terraform.Finalizers = utils.RemoveString(terraform.Finalizers, tfv1.TerraformFinalizer)
		if err := r.Client.Update(ctx, &terraform, &client.UpdateOptions{}); err != nil {
			blog.Errorf("remove finalizer for terraform failed, terraform: %s, trace-id: %s, err: %s",
				nn, traceId, err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
		}
		blog.Infof("remove finalizer for terraform success, terraform: %s, trace-id: %s", nn, traceId)
		return ctrl.Result{}, nil
	}
	if !controllerutil.ContainsFinalizer(&terraform, tfv1.TerraformFinalizer) {
		patch := client.MergeFrom(terraform.DeepCopy())
		controllerutil.AddFinalizer(&terraform, tfv1.TerraformFinalizer)

		if err := r.Patch(ctx, &terraform, patch); err != nil {
			blog.Errorf("add finalizer for terraform '%s' failed, err: %s", nn, err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
		}
	}

	return r.handler(ctx, traceId, &terraform)
}

// handler 核心逻辑
func (r *TerraformReconciler) handler(ctx context.Context, traceId string, tf *tfv1.Terraform) (ctrl.Result, error) {
	currentCommitId, currentCommitIdOk := tf.Annotations[tfv1.TerraformManualAnnotation]
	if utils.RemoveManualAnnotation(tf) {
		if err2 := r.updateAnnotations(ctx, traceId, currentCommitId, tf); err2 != nil {
			return ctrl.Result{}, err2
		}
	}
	defer r.finish(ctx, traceId, tf)

	getTfPlanFlag := false
	h := core.NewTask(ctx, traceId, tf, r.Client)
	nn := apitypes.NamespacedName{Namespace: tf.Namespace, Name: tf.Name}

	if err := h.Init(); err != nil {
		blog.Errorf("init handler failed, err: %s", err.Error())
		r.setPhaseError(tf, fmt.Sprintf("init handler failed, err: %s", err.Error()))
		return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, err
	}

	lastCommitId, err := h.GetLastCommitId()
	if err != nil {
		blog.Errorf("get changes failed, err: %s", err)
		r.setPhaseError(tf, fmt.Sprintf("get changes failed, err: %s", err.Error()))
		return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, err
	}
	tf.Status.SyncStatus = tfv1.OutOfSyncStatus
	if tf.Spec.SyncPolicy == tfv1.AutoSyncPolicy {
		// 自动策略
	} else if tf.Spec.SyncPolicy == tfv1.ManualSyncPolicy && currentCommitIdOk && len(currentCommitId) != 0 {
		// 手动策略
		if currentCommitId != "refresh" && !h.CheckCommitIdConsistency(currentCommitId) {
			// 手动指定不一致, 既不等于 'refresh', commit-id 也不相等
			r.setPhaseError(tf, fmt.Sprintf("the commit-id of manual sync is inconsistent with the commit-id of "+
				"the remote branch, current commit-id: %s, remote branch commit-id: %s", currentCommitId, lastCommitId))
			blog.Infof("the commit-id of manual sync is inconsistent with the commit-id of the remote branch, "+
				"current commit-id: %s, remote branch commit-id: %s, tf: %s, trace-id: %s", currentCommitId,
				lastCommitId, nn, traceId)
			return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
		}
	} else {
		// 其他，则不执行并返回
		r.setPhaseSucceeded(tf, fmt.Sprintf("tf doesn’t have to do anything"))
		blog.Infof("tf doesn’t have to do anything, tf: %s, trace-id: %s", nn, traceId)
		return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
	}

	if !h.CheckForChanges() {
		// LastAppliedRevision == git lastCommitId (无变化)
		getTfPlanFlag = true // 假定执行apply失败, 需要进行apply重试, 获取tfplan, 再去执行apply

		if currentCommitId == "refresh" {
			// todo: 手动刷新 or 强制刷新?
			// getTfPlanFlag = true
		} else if len(tf.Status.LastApplyError) == 0 {
			// apply成功, 不需要再执行
			blog.Infof("tf remote branch last commit-id no change, tf: %s, trace-id: %s", nn, traceId)
			r.setPhaseSucceeded(tf, fmt.Sprintf("tf remote branch last commit-id no change"))
			return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
		}
	}

	if utils.FormatRevision(tf.Spec.Repository.TargetRevision, lastCommitId) == tf.Status.LastPlannedRevision ||
		getTfPlanFlag {
		blog.Infof("try to get the last tfplan result, tf: %s, trace-id: %s", nn, traceId)

		if len(tf.Status.LastPlanError) != 0 {
			blog.Error("tf plan failed, trace-id: %s, err: %s", traceId, tf.Status.LastPlanError)
			r.setPhaseError(tf, fmt.Sprintf("tf plan failed, trace-id: %s, err: %s", traceId, tf.Status.LastPlanError))
			return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
		}

		tf.Status.LastPlanError, err = h.GetTfPlan() // 同步结果至status
		if err != nil {
			blog.Errorf("get tf plan failed, err: %s", err)
			r.setPhaseError(tf, fmt.Sprintf("get tf plan failed, err: %s", err))
			return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, err
		}
		tf.Status.LastPlanError = "" // 如果执行成功则清理
	} else {
		blog.Infof("try creating a tfplan result, tf: %s, trace-id: %s", nn, traceId)

		tf.Status.LastPlanError, err = h.Plan() // 同步结果至status
		if err != nil {
			blog.Errorf("tf plan failed, err: %s", err)
			r.setPhaseError(tf, fmt.Sprintf("tf plan failed, err: %s", err))
			return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, err
		}
		tf.Status.LastPlanError = "" // 如果执行成功则清理

		if err = h.SaveTfPlan(); err != nil { // plan结果保存到secret、configmap
			blog.Errorf("tf plan result save failed, err: %s", err)
			r.setPhaseError(tf, fmt.Sprintf("tf plan result save failed, err: %s", err))
			return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, err
		}
	}

	apply, err := h.Apply()
	tf.Status.LastApplyError = apply
	if err != nil { // 执行失败，设置状态
		blog.Errorf("apply tf failed, err: %s", err)
		r.setPhaseError(tf, fmt.Sprintf("apply tf failed, err: %s", err))
		return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, err
	}
	tf.Status.LastApplyError = ""
	tf.Status.SyncStatus = tfv1.SyncedStatus
	blog.Infof("sync tf success, tf: %s, trace-id: %s", nn, traceId)

	if err = h.SaveApplyOutputToConfigMap(apply); err != nil {
		// 执行失败，设置状态
		blog.Errorf("save apply tf result to configmap failed, err: %s", err)
		r.setPhaseError(tf, fmt.Sprintf("save apply tf result to configmap failed, err: %s", err))
		return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, err
	}
	r.setPhaseSucceeded(tf, fmt.Sprintf("sync tf success"))

	return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
}

// finish 更新tf status至api server
func (r *TerraformReconciler) finish(ctx context.Context, traceId string, tf *tfv1.Terraform) {
	tf.Status.ObservedGeneration += 1
	nn := apitypes.NamespacedName{Namespace: tf.Namespace, Name: tf.Name}
	//blog.Infof("update tf status before, trace-id: %s, tf-json: %s,", traceId, utils.ToJsonString(tf))

	if err := r.Client.Status().Update(ctx, tf, &client.UpdateOptions{}); err != nil {
		blog.Errorf("update tf status failed(finish), tf: %s, trace-id: %s, err: %s", nn, traceId, err)
		return
	} else {
		blog.Infof("update tf status success, tf: %s, trace-id: %s", nn, traceId)
	}
}

// updateAnnotations 更新注解
func (r *TerraformReconciler) updateAnnotations(ctx context.Context, traceId, id string, tf *tfv1.Terraform) error {
	nn := apitypes.NamespacedName{Namespace: tf.Namespace, Name: tf.Name}
	//blog.Infof("updateAnnotations to tf, trace-id: %s, tf-json: %s,", traceId, utils.ToJsonString(tf))

	if err := r.Client.Update(ctx, tf); err != nil {
		// 移除annotations失败
		blog.Errorf("updateAnnotations remove '%s' annotations failed, tf: %s, trace-id: %s, err: %s",
			tfv1.TerraformManualAnnotation, nn, traceId, err.Error())
		return err
	}

	// 移除annotations成功
	blog.Infof("updateAnnotations remove '%s: %s' annotations success, tf: %s, trace-id: %s",
		tfv1.TerraformManualAnnotation, id, nn, traceId)

	return nil
}

// setPhaseError 设置phase
func (r *TerraformReconciler) setPhaseError(tf *tfv1.Terraform, message string) {
	tf.Status.OperationStatus.Message = message
	tf.Status.OperationStatus.Phase = tfv1.PhaseError
	tf.Status.OperationStatus.FinishAt = &metav1.Time{Time: time.Now()}
}

// setPhaseSucceeded 设置phase
func (r *TerraformReconciler) setPhaseSucceeded(tf *tfv1.Terraform, message string) {
	tf.Status.OperationStatus.Message = message
	tf.Status.OperationStatus.Phase = tfv1.PhaseSucceeded
	tf.Status.OperationStatus.FinishAt = &metav1.Time{Time: time.Now()}
}

// terraformPredicate filter terraform
func (r *TerraformReconciler) terraformPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			tf, ok := createEvent.Object.(*tfv1.Terraform)
			if ok {
				blog.V(5).Infof("add terraform(predicate): %s/%s", tf.GetNamespace(), tf.GetName())
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newTf, okNew := e.ObjectNew.(*tfv1.Terraform)
			oldTf, okOld := e.ObjectOld.(*tfv1.Terraform)
			if !okNew || !okOld {
				return true
			}
			blog.V(5).Infof("update terraform(predicate): %s/%s", newTf.GetNamespace(), newTf.GetName())

			if reflect.DeepEqual(newTf.Spec, oldTf.Spec) &&
				reflect.DeepEqual(newTf.Annotations, oldTf.Annotations) &&
				reflect.DeepEqual(newTf.Finalizers, oldTf.Finalizers) &&
				reflect.DeepEqual(newTf.DeletionTimestamp, oldTf.DeletionTimestamp) {
				blog.V(5).Infof("terraform %s updated(predicate), but spec and annotation and finalizer not change",
					utils.ToJsonString(newTf))
				return false
			}
			blog.V(5).Infof("terraform(predicate) changes, new tf: %s, old tf: %s", utils.ToJsonString(newTf),
				utils.ToJsonString(oldTf))

			return true
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			tf, ok := deleteEvent.Object.(*tfv1.Terraform)
			if ok {
				blog.V(5).Infof("delete terraform(predicate): %s/%s", tf.GetNamespace(), tf.GetName())
			}
			return true
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TerraformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tfv1.Terraform{}).
		WithEventFilter(r.terraformPredicate()).
		Complete(r)
}

//// noChanges 没有变化, 执行完成, 进入收尾
//func (r *TerraformReconciler) noChanges(ctx context.Context, traceId string, tf *tfv1.Terraform) (ctrl.Result, error) {
//	if !utils.RemoveManualAnnotation(tf) { // 没有annotations
//		return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
//	}
//
//	// 有annotations, 则移除
//	if err := r.Client.Update(ctx, tf); err != nil { // 移除annotations失败
//		blog.Errorf("remove '%s' annotations failed, tf: %s/%s, trace-id: %s, err: %s",
//			tfv1.TerraformManualAnnotation, tf.Namespace, tf.Name, traceId, err.Error())
//		return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, err
//	}
//
//	// 移除annotations成功
//	blog.Infof("remove '%s' annotations success, tf: %s/%s, trace-id: %s", tfv1.TerraformManualAnnotation,
//		tf.Namespace, tf.Name, traceId)
//
//	return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
//}
