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

package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/lock"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/lock/locallock"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/utils"
	gitopsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/apis/gitopsworkflow/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/option"
)

// WorkflowController defines the controller of gitops workflow
type WorkflowController struct {
	locker lock.Interface
	// Client k8s api server client
	client.Client
	// Scheme runtime scheme
	Scheme *runtime.Scheme
}

// NewWorkflowController create the instance of workflow
func NewWorkflowController(client client.Client, scheme *runtime.Scheme) *WorkflowController {
	return &WorkflowController{
		locker: locallock.NewLocalLock(),
		Client: client,
		Scheme: scheme,
	}
}

// Reconcile the workflow object
func (r *WorkflowController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = context.WithValue(ctx, logctx.TraceKey, uuid.New().String())
	nn := req.NamespacedName.String()
	ctx = context.WithValue(ctx, logctx.ObjectKey, nn)
	logctx.Infof(ctx, "reconcile received workflow")
	// locking object key to avoid competition with concurrent workers
	_ = r.locker.Lock(ctx, nn)
	defer r.locker.UnLock(ctx, nn) // nolint

	var workflow = new(gitopsv1.Workflow)
	if err := r.Get(ctx, req.NamespacedName, workflow); err != nil {
		if k8serrors.IsNotFound(err) {
			logctx.Warnf(ctx, "workflow is deleted, skip")
			return ctrl.Result{}, nil
		}
	}
	if workflow.DeletionTimestamp != nil {
		return r.deleteWorkflow(ctx, workflow)
	}
	if workflow.Spec.Disable {
		logctx.Infof(ctx, "workflow id disabled, no need handle it")
		return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
	}
	if !controllerutil.ContainsFinalizer(workflow, gitopsv1.WorkflowFinalizer) {
		patch := client.MergeFrom(workflow.DeepCopy())
		controllerutil.AddFinalizer(workflow, gitopsv1.WorkflowFinalizer)
		if err := r.Patch(ctx, workflow, patch); err != nil {
			logctx.Errorf(ctx, "add finalizer failed: %s", err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
		logctx.Infof(ctx, "add finalizer success")
	}
	return r.handleWorkflow(ctx, workflow)
}

func (r *WorkflowController) handleWorkflow(ctx context.Context, workflow *gitopsv1.Workflow) (ctrl.Result, error) {
	currentStatus := workflow.Status.Phase
	if currentStatus == "" || currentStatus == gitopsv1.ErrorStatus {
		if err := r.updateWorkflowStatus(ctx, workflow.Name, workflow.Namespace,
			gitopsv1.WorkflowStatus{Phase: gitopsv1.InitializingStatus, Message: workflow.Status.Message},
		); err != nil {
			logctx.Errorf(ctx, "workflow update status to 'initializing' failed: %s", err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
		}
		logctx.Infof(ctx, "workflow update status to 'initializing' success")
	}
	if err := r.createOrUpdatePipeline(ctx, workflow); err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
	}
	return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
}

func (r *WorkflowController) createOrUpdatePipeline(ctx context.Context, workflow *gitopsv1.Workflow) error {
	eng := returnEngineHandler(workflow)
	user := workflow.Annotations[gitopsv1.WorkflowAnnotationCreateUser]
	var operation = "create"
	ppID := workflow.Status.PipelineID
	if ppID != "" {
		updateUser := workflow.Annotations[gitopsv1.WorkflowAnnotationUpdateUser]
		if updateUser != "" {
			user = updateUser
		}
		operation = "update"
	}
	ctx = context.WithValue(ctx, logctx.UserKey, user)

	logctx.Infof(ctx, "workflow start %s pipeline", operation)
	pipelineID, err := eng.CreateOrUpdatePipeline(ctx, user, workflow)
	if err != nil {
		logctx.Errorf(ctx, "workflow %s pipeline failed: %s", operation, err.Error())
		if updateStatusErr := r.updateWorkflowStatus(ctx, workflow.Name, workflow.Namespace,
			gitopsv1.WorkflowStatus{
				Phase:   gitopsv1.ErrorStatus,
				Message: fmt.Sprintf("%s pipeline failed: %s", operation, err.Error()),
			},
		); updateStatusErr != nil {
			logctx.Errorf(ctx, "workflow update status to 'error' failed: %s", updateStatusErr.Error())
		}
		return err
	}

	if operation == "create" {
		logctx.Infof(ctx, "workflow create pipeline success: %s", pipelineID)
	} else {
		pipelineID = ppID
		logctx.Infof(ctx, "workflow update pipeline '%s' success", ppID)
	}
	if updateStatusErr := r.updateWorkflowStatus(ctx, workflow.Name, workflow.Namespace,
		gitopsv1.WorkflowStatus{Phase: gitopsv1.ReadyStatus, PipelineID: pipelineID},
	); updateStatusErr != nil {
		logctx.Errorf(ctx, "workflow update status to 'ready' failed: %s", updateStatusErr.Error())
	}
	return nil
}

func (r *WorkflowController) updateWorkflowStatus(ctx context.Context, name, namespace string,
	status gitopsv1.WorkflowStatus) error {
	status.LastUpdateTime = &metav1.Time{Time: time.Now()}
	var queryWorkflow = new(gitopsv1.Workflow)
	if err := r.Client.Get(ctx, k8stypes.NamespacedName{Namespace: namespace, Name: name}, queryWorkflow); err != nil {
		return errors.Wrapf(err, "get workflow failed when udpate status")
	}
	queryWorkflow.Status = status
	if err := r.Client.Status().Update(ctx, queryWorkflow, &client.UpdateOptions{}); err != nil {
		return errors.Wrapf(err, "update workflow status failed")
	}
	return nil
}

func (r *WorkflowController) deleteWorkflow(ctx context.Context, workflow *gitopsv1.Workflow) (ctrl.Result, error) {
	user := workflow.Annotations[gitopsv1.WorkflowAnnotationUpdateUser]
	ctx = context.WithValue(ctx, logctx.UserKey, user)
	logctx.Infof(ctx, "workflow start delete pipeline")
	if workflow.Status.PipelineID != "" && workflow.Spec.DestroyOnDeletion {
		eng := returnEngineHandler(workflow)
		if err := eng.DeletePipeline(ctx, user, workflow); err != nil {
			logctx.Errorf(ctx, "workflow delete pipeline failed: %s", err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
	}
	rawPatch := client.RawPatch(k8stypes.JSONPatchType, []byte(`[{"op":"remove","path":"/metadata/finalizers"}]`))
	if err := r.Client.Patch(ctx, workflow, rawPatch); err != nil {
		logctx.Errorf(ctx, "workflow patch delete finalizer failed: %s", err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}
	logctx.Infof(ctx, "workflow delete success")
	return ctrl.Result{}, nil
}

// workflowPredicate filter workflow object
func (r *WorkflowController) workflowPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newWF, okNew := e.ObjectNew.(*gitopsv1.Workflow)
			oldWF, okOld := e.ObjectOld.(*gitopsv1.Workflow)
			if !okNew || !okOld {
				return true
			}
			if newWF.ObjectMeta.DeletionTimestamp != nil {
				return true
			}
			if !reflect.DeepEqual(newWF.Spec, oldWF.Spec) {
				blog.Infof("terraform '%s/%s' spec changed, newSpec: %s, oldSpec: %s",
					newWF.GetNamespace(), newWF.GetName(),
					utils.ToJsonString(newWF.Spec), utils.ToJsonString(oldWF.Spec))
				return true
			}
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WorkflowController) SetupWithManager(mgr ctrl.Manager) error {
	op := option.GlobalOption()
	return ctrl.NewControllerManagedBy(mgr).
		For(&gitopsv1.Workflow{}).
		WithEventFilter(r.workflowPredicate()).
		WithOptions(controller.Options{MaxConcurrentReconciles: op.MaxWorkers}).
		Complete(r)
}
