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
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/utils"
	gitopsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/apis/gitopsworkflow/v1"
	thirdengine "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/third_engine"
)

// HistoryController defines the controller of history
type HistoryController struct {
	// Client k8s api server client
	client.Client
	// Scheme runtime scheme
	Scheme *runtime.Scheme
}

// NewHistoryController create the instance of hsitory controller
func NewHistoryController(client client.Client, scheme *runtime.Scheme) *HistoryController {
	return &HistoryController{
		Client: client,
		Scheme: scheme,
	}
}

const (
	defaultReconcileTime = 180
)

// Reconcile the workflow history object
func (r *HistoryController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = context.WithValue(ctx, logctx.TraceKey, uuid.New().String())
	nn := req.NamespacedName.String()
	ctx = context.WithValue(ctx, logctx.ObjectKey, nn)
	logctx.Infof(ctx, "reconcile received workflow history")

	var history = new(gitopsv1.WorkflowHistory)
	if err := r.Get(ctx, req.NamespacedName, history); err != nil {
		if k8serrors.IsNotFound(err) {
			logctx.Warnf(ctx, "workflow is deleted, skip")
			return ctrl.Result{}, nil
		}
	}
	if history.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}
	stat := history.Status.Phase
	if stat == gitopsv1.HistorySuccess || stat == gitopsv1.HistoryFailed {
		return ctrl.Result{}, nil
	}
	if stat == gitopsv1.HistoryError {
		return ctrl.Result{}, nil
	}

	if stat == gitopsv1.HistoryRunning {
		if err := r.syncHistoryWithTicker(ctx, history); err != nil {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true, RequeueAfter: defaultReconcileTime * time.Second}, nil
	}
	if stat == "" {
		if err := r.executeWorkflow(ctx, history); err != nil {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true, RequeueAfter: defaultReconcileTime * time.Second}, nil
	}
	logctx.Warnf(ctx, "unknown workflow history status '%s'", stat)
	return ctrl.Result{}, nil
}

func (r *HistoryController) executeWorkflow(ctx context.Context, history *gitopsv1.WorkflowHistory) error {
	workflow, err := r.getWorkflow(ctx, history)
	if err != nil {
		logctx.Errorf(ctx, "get workflow failed when execute: %s", err.Error())
		_, _ = r.updateHistoryStatus(ctx, history.Name, history.Namespace, gitopsv1.WorkflowHistoryStatus{
			Phase:   gitopsv1.ErrorStatus,
			Message: err.Error(),
		})
		return err
	}

	user := workflow.Annotations[gitopsv1.WorkflowAnnotationCreateUser]
	h := returnEngineHandler(workflow)
	historyNum, historyID, err := h.ExecutePipeline(ctx, user, workflow, history)
	if err != nil {
		logctx.Errorf(ctx, "execute pipeline failed: %s", err.Error())
		_, _ = r.updateHistoryStatus(ctx, history.Name, history.Namespace, gitopsv1.WorkflowHistoryStatus{
			Phase:   gitopsv1.ErrorStatus,
			Message: fmt.Sprintf("execute pipeline failed: %s", err.Error()),
		})
		return err
	}
	newHistory, err := r.updateHistoryStatus(ctx, history.Name, history.Namespace, gitopsv1.WorkflowHistoryStatus{
		Phase:      gitopsv1.HistoryRunning,
		Message:    "",
		HistoryID:  historyID,
		HistoryNum: historyNum,
	})
	if err != nil {
		return err
	}
	if err = r.syncHistoryWithTicker(ctx, newHistory); err != nil {
		return err
	}
	return nil
}

func (r *HistoryController) syncHistoryWithTicker(ctx context.Context, history *gitopsv1.WorkflowHistory) error {
	historyID := history.Status.HistoryID
	if historyID == "" {
		err := errors.Errorf("history id is empty")
		logctx.Errorf(ctx, err.Error())
		_, _ = r.updateHistoryStatus(ctx, history.Name, history.Namespace, gitopsv1.WorkflowHistoryStatus{
			Phase:   gitopsv1.ErrorStatus,
			Message: err.Error(),
		})
		return err
	}
	workflow, err := r.getWorkflow(ctx, history)
	if err != nil {
		logctx.Errorf(ctx, "get workflow failed when sync history: %s", err.Error())
		_, _ = r.updateHistoryStatus(ctx, history.Name, history.Namespace, gitopsv1.WorkflowHistoryStatus{
			Phase:   gitopsv1.ErrorStatus,
			Message: err.Error(),
		})
		return err
	}
	go func(wk *gitopsv1.Workflow) {
		user := wk.Annotations[gitopsv1.WorkflowAnnotationCreateUser]
		if updateUser := history.Annotations[gitopsv1.WorkflowAnnotationCreateUser]; updateUser != "" {
			user = updateUser
		}
		h := returnEngineHandler(wk)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		timeout := time.After(defaultReconcileTime * time.Second)
		for {
			select {
			case <-ticker.C:
				var historyStatus *thirdengine.PipelineHistoryStatus
				historyStatus, err = h.GetPipelineHistoryStatus(ctx, wk.Spec.Project, user, historyID)
				if err != nil {
					logctx.Errorf(ctx, "get pipeline history status failed: %s", err.Error())
					continue
				}
				if historyStatus.Status == gitopsv1.HistorySuccess || historyStatus.Status == gitopsv1.HistoryFailed {
					status := gitopsv1.WorkflowHistoryStatus{Phase: historyStatus.Status}
					if historyStatus.StartTime != nil {
						status.StartedAt = &metav1.Time{Time: *historyStatus.StartTime}
					}
					if historyStatus.EndTime != nil {
						status.FinishedAt = &metav1.Time{Time: *historyStatus.EndTime}
					}
					_, _ = r.updateHistoryStatus(ctx, history.Name, history.Namespace, status)
					return
				}
			case <-timeout:
				return
			}
		}
	}(workflow)
	return nil
}

func (r *HistoryController) getWorkflow(ctx context.Context, history *gitopsv1.WorkflowHistory) (
	*gitopsv1.Workflow, error) {
	wk, ok := history.Annotations[gitopsv1.HistoryAnnotationWorkflow]
	if !ok {
		return nil, errors.Errorf("history not have annotation '%s' specify workflow",
			gitopsv1.HistoryAnnotationWorkflow)
	}
	queryWorkflow := new(gitopsv1.Workflow)
	if err := r.Client.Get(ctx,
		k8stypes.NamespacedName{Namespace: history.Namespace, Name: wk}, queryWorkflow); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, errors.Errorf("workflow '%s' not found", wk)
		}
		return nil, errors.Wrapf(err, "get workflow '%s' failed", wk)
	}
	return queryWorkflow, nil
}

func (r *HistoryController) updateHistoryStatus(ctx context.Context, name, namespace string,
	status gitopsv1.WorkflowHistoryStatus) (*gitopsv1.WorkflowHistory, error) {
	var queryHistory = new(gitopsv1.WorkflowHistory)
	var resultErr error
	defer func() {
		if resultErr != nil {
			logctx.Errorf(ctx, "update workflow history '%s/%s' failed: %s; target status: %v",
				namespace, name, resultErr, utils.ToJsonString(resultErr))
		}
	}()
	if err := r.Client.Get(ctx, k8stypes.NamespacedName{Namespace: namespace, Name: name}, queryHistory); err != nil {
		resultErr = errors.Wrapf(err, "get workflow history failed when udpate status")
		return nil, resultErr
	}
	if status.Message != "" {
		logctx.Infof(ctx, "update status '%s' -> '%s', msg: %s", queryHistory.Status.Phase,
			status.Phase, status.Message)
	} else {
		logctx.Infof(ctx, "update status '%s' -> '%s'", queryHistory.Status.Phase, status.Phase)
	}
	queryHistory.Status.Phase = status.Phase
	queryHistory.Status.Message = status.Message
	if queryHistory.Status.HistoryNum == 0 {
		queryHistory.Status.HistoryNum = status.HistoryNum
	}
	if queryHistory.Status.HistoryID == "" {
		queryHistory.Status.HistoryID = status.HistoryID
	}
	if queryHistory.Status.StartedAt == nil {
		queryHistory.Status.StartedAt = status.StartedAt
	}
	if queryHistory.Status.FinishedAt == nil {
		queryHistory.Status.FinishedAt = status.FinishedAt
	}
	start := queryHistory.Status.StartedAt
	end := queryHistory.Status.FinishedAt
	if start != nil {
		if end != nil {
			queryHistory.Status.CostTime = end.Sub(start.Time).String()
		} else {
			queryHistory.Status.CostTime = time.Since(start.Time).String()
		}
	}
	if err := r.Client.Status().Update(ctx, queryHistory, &client.UpdateOptions{}); err != nil {
		resultErr = errors.Wrapf(err, "update workflow history status failed")
		return nil, resultErr
	}
	return queryHistory, resultErr
}

func (r *HistoryController) historyPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *HistoryController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gitopsv1.WorkflowHistory{}).
		WithEventFilter(r.historyPredicate()).
		Complete(r)
}
