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
	"os"

	"github.com/google/uuid"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/runner"
)

// TerraformReconciler reconciles a Terraform object
type TerraformReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// doto: use interface
	TerraformRunner runner.TerraformLocalRunner
}

// Reconcile reconcile terraform
// +kubebuilder:rbac:groups=terraformextesions.bkbcs.tencent.com,resources=terraforms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=terraformextesions.bkbcs.tencent.com,resources=terraforms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=terraformextesions.bkbcs.tencent.com,resources=terraforms/finalizers,verbs=update
func (r *TerraformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	reconciliationLoopID := uuid.New().String()

	terraform := tfv1.Terraform{}
	if err := r.Get(ctx, req.NamespacedName, &terraform); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("terraform '%s' is deleted, skipped", req.NamespacedName.String())
			return ctrl.Result{}, nil
		}
	}

	if terraform.DeletionTimestamp != nil {
		// doto delete logic
	}

	if !controllerutil.ContainsFinalizer(&terraform, tfv1.TerraformFinalizer) {
		patch := client.MergeFrom(terraform.DeepCopy())
		controllerutil.AddFinalizer(&terraform, tfv1.TerraformFinalizer)
		if err := r.Patch(ctx, &terraform, patch); err != nil {
			blog.Errorf("add finalizer for terraform '%s' failed, err: %s", req.NamespacedName.String(), err.Error())
			return ctrl.Result{}, err
		}
	}

	// 1. get terraform source, i.e. terraform code in git

	// 2. compare source.artifact and terraform's applied revision
	// if changed, do re-plan

	// 3. get tf runner
	tfRunner := &runner.TerraformLocalRunner{
		Cli:      r.Client,
		ExecPath: "/usr/local/bin/terraform", // doto use env (install in docker file)
		Done:     make(chan os.Signal),
	}

	newTerraformReply, err := tfRunner.NewTerraform(ctx, &runner.NewTerraformRequest{
		WorkingDir: "/data/bcs/terraform", // doto use argo cd to get repo info
		Terraform:  terraform,
		InstanceID: reconciliationLoopID,
	})
	if err != nil {
		blog.Errorf("new terraform failed, err: %s", err.Error())
		return ctrl.Result{}, err
	}
	tfInstance := newTerraformReply.Id

	initReply, err := tfRunner.Init(ctx, &runner.InitRequest{
		TfInstance: tfInstance,
		Upgrade:    true,
	})
	if err != nil {
		blog.Errorf("init terraform failed, err: %s", err.Error())
		return ctrl.Result{}, err
	}

	blog.Info("terraform init message: %s", initReply.Message)

	planReply, err := tfRunner.Plan(ctx, &runner.PlanRequest{
		TfInstance: tfInstance,
		Out:        "",
		Refresh:    false,
		Destroy:    false,
		Targets:    nil,
	})
	if err != nil {
		blog.Errorf("plan terraform failed, err: %s", err.Error())
		return ctrl.Result{}, err
	}

	blog.Info("terraform plan message: %s", planReply.Message)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TerraformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tfv1.Terraform{}).
		Complete(r)
}
