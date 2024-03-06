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
	"reflect"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/internal/logctx"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/tfhandler"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/worker"
)

// TerraformReconciler reconciles a Terraform object
type TerraformReconciler struct {
	// Client k8s api server client
	client.Client
	// Scheme runtime scheme
	Scheme *runtime.Scheme
	// Config opt
	Config *option.ControllerOption

	Queue     worker.TerraformQueue
	TFHandler tfhandler.TerraformHandler
}

// Reconcile reconcile terraform
// +kubebuilder:rbac:groups=terraformextensions.bkbcs.tencent.com,resources=terraforms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=terraformextensions.bkbcs.tencent.com,resources=terraforms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=terraformextensions.bkbcs.tencent.com,resources=terraforms/finalizers,verbs=update
func (r *TerraformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nn := req.NamespacedName.String()
	ctx = context.WithValue(ctx, logctx.ObjectKey, nn)
	logctx.Infof(ctx, "reconcile received terraform '%s'", nn)
	var terraform = new(tfv1.Terraform)
	if err := r.Get(ctx, req.NamespacedName, terraform); err != nil {
		if k8serrors.IsNotFound(err) {
			logctx.Warnf(ctx, "terraform '%s' is deleted, skipped", nn)
			return ctrl.Result{}, nil
		}
	}
	if terraform.DeletionTimestamp != nil {
		return r.deleteTerraform(ctx, terraform)
	}
	if !controllerutil.ContainsFinalizer(terraform, tfv1.TerraformFinalizer) {
		patch := client.MergeFrom(terraform.DeepCopy())
		controllerutil.AddFinalizer(terraform, tfv1.TerraformFinalizer)
		if err := r.Patch(ctx, terraform, patch); err != nil {
			logctx.Errorf(ctx, "add finalizer for '%s' failed: %s", nn, err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
		logctx.Infof(ctx, "add finalizer for '%s' success", nn)
	}
	r.Queue.Push(ctx, terraform)
	return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
}

func (r *TerraformReconciler) deleteTerraform(ctx context.Context, terraform *tfv1.Terraform) (ctrl.Result, error) {
	logctx.Infof(ctx, "terraform deleting with destroy: %v", terraform.Spec.DestroyResourcesOnDeletion)
	if terraform.Spec.DestroyResourcesOnDeletion {
		if err := r.TFHandler.Destroy(ctx, terraform); err != nil {
			logctx.Errorf(ctx, "destroy terraform failed: %s", err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
	}
	rawPatch := client.RawPatch(k8stypes.JSONPatchType, []byte(`[{"op":"remove","path":"/metadata/finalizers"}]`))
	if err := r.Client.Patch(ctx, terraform, rawPatch); err != nil {
		logctx.Errorf(ctx, "terraform patch delete finalizer failed: %s", err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}
	logctx.Infof(ctx, "terraform delete success")
	return ctrl.Result{}, nil
}

// terraformPredicate filter terraform
func (r *TerraformReconciler) terraformPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newTf, okNew := e.ObjectNew.(*tfv1.Terraform)
			oldTf, okOld := e.ObjectOld.(*tfv1.Terraform)
			if !okNew || !okOld {
				return true
			}
			if newTf.ObjectMeta.DeletionTimestamp != nil {
				return true
			}
			if !reflect.DeepEqual(newTf.Spec, oldTf.Spec) {
				blog.Infof("terraform '%s/%s' spec changed, newSpec: %s, oldSpec: %s",
					newTf.GetNamespace(), newTf.GetName(),
					utils.ToJsonString(newTf.Spec), utils.ToJsonString(oldTf.Spec))
				return true
			}
			_, syncNewOK := newTf.Annotations[tfv1.TerraformOperationSync]
			_, syncOldOK := oldTf.Annotations[tfv1.TerraformOperationSync]
			_, cleanNewOK := newTf.Annotations[tfv1.TerraformOperationClean]
			_, cleanOldOK := oldTf.Annotations[tfv1.TerraformOperationClean]
			// 屏蔽掉因为删除 operation annotation 导致的变化
			if (!syncNewOK && syncOldOK) || (!cleanNewOK && cleanOldOK) {
				return false
			}
			if !utils.MapEqualExceptKey(newTf.Annotations, oldTf.Annotations,
				[]string{"kubectl.kubernetes.io/last-applied-configuration"}) {
				blog.Infof("terraform '%s/%s' annotation changed", newTf.GetNamespace(), newTf.GetName())
				return true
			}
			return false
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
