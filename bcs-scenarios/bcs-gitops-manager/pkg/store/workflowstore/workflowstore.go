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

// Package workflowstore xx
package workflowstore

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	workflowv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/apis/gitopsworkflow/v1"
	workflowversioned "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/client/clientset/versioned"
	workflowpkg "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/client/clientset/versioned/typed/gitopsworkflow/v1"
	workflowexternal "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/client/informers/externalversions"
	informers "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/client/informers/externalversions/gitopsworkflow/v1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
)

// WorkflowInterface defines the interface of workflow
type WorkflowInterface interface {
	Init() error

	ListWorkflows(ctx context.Context, projects []string) ([]workflowv1.Workflow, error)
	CreateWorkflow(ctx context.Context, wf *workflowv1.Workflow) error
	GetWorkflowByID(ctx context.Context, id string) (*workflowv1.Workflow, error)
	DeleteWorkflow(ctx context.Context, id string) error
	UpdateWorkflow(ctx context.Context, wf *workflowv1.Workflow) error
	ExecuteWorkflow(ctx context.Context, history *workflowv1.WorkflowHistory) error

	ListWorkflowHistories(ctx context.Context, id string) ([]workflowv1.WorkflowHistory, error)
	GetWorkflowHistoryDetail(ctx context.Context, historyID string) (*workflowv1.WorkflowHistory, error)
}

type workflowHandler struct {
	op                    *options.Options
	workflowClient        *workflowpkg.GitopsworkflowV1Client
	workflowDynamicClient *workflowversioned.Clientset

	gitopsworkflowFactory workflowexternal.SharedInformerFactory
	workflowInformer      informers.WorkflowInformer
	historyInformer       informers.WorkflowHistoryInformer

	stopChan chan struct{}
}

// NewWorkflowHandler create the workflow controller instance
func NewWorkflowHandler() WorkflowInterface {
	return &workflowHandler{
		op:       options.GlobalOptions(),
		stopChan: make(chan struct{}),
	}
}

var (
	resyncPeriod = time.Duration(30) * time.Minute
)

// Init init the workflow store
func (h *workflowHandler) Init() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "get k8s in-cluster config failed")
	}
	h.workflowClient, err = workflowpkg.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create workflow k8s client failed")
	}
	h.workflowDynamicClient, err = workflowversioned.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create workflow dynamic client failed")
	}
	h.gitopsworkflowFactory = workflowexternal.NewSharedInformerFactory(h.workflowDynamicClient, resyncPeriod)
	h.workflowInformer = h.gitopsworkflowFactory.Gitopsworkflow().V1().Workflows()
	h.workflowInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	})
	h.historyInformer = h.gitopsworkflowFactory.Gitopsworkflow().V1().WorkflowHistories()
	h.historyInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	})
	h.gitopsworkflowFactory.Start(h.stopChan)
	go h.waitForSync()
	blog.Infof("workflow store informer started")
	return nil
}

func (h *workflowHandler) waitForSync() {
	if !cache.WaitForCacheSync(h.stopChan, h.workflowInformer.Informer().HasSynced) {
		blog.Errorf("failed to sync workflow informer")
	} else {
		blog.Infof("workflow informer synced success")
	}
	if !cache.WaitForCacheSync(h.stopChan, h.historyInformer.Informer().HasSynced) {
		blog.Errorf("failed to sync workflow history informer")
	} else {
		blog.Infof("workflow history informer synced success")
	}
}

// ListWorkflows list workflows with projects query
func (h *workflowHandler) ListWorkflows(ctx context.Context, projects []string) ([]workflowv1.Workflow, error) {
	projMap := make(map[string]struct{})
	ns := h.op.GitOps.AdminNamespace
	for i := range projects {
		projMap[projects[i]] = struct{}{}
	}
	result := make([]workflowv1.Workflow, 0)
	if h.workflowInformer.Informer().HasSynced() {
		workflows, err := h.workflowInformer.Lister().Workflows(ns).List(labels.NewSelector())
		if err != nil {
			return nil, errors.Wrapf(err, "list workflows from informer failed")
		}
		for _, wf := range workflows {
			if _, ok := projMap[wf.Spec.Project]; ok {
				result = append(result, *wf.DeepCopy())
			}
		}
		return result, nil
	}
	workflows, err := h.workflowClient.Workflows(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "list workflows failed")
	}
	for i := range workflows.Items {
		wf := workflows.Items[i]
		if _, ok := projMap[wf.Spec.Project]; ok {
			result = append(result, *wf.DeepCopy())
		}
	}
	return result, nil
}

// CreateWorkflow create workflow
func (h *workflowHandler) CreateWorkflow(ctx context.Context, wf *workflowv1.Workflow) error {
	if _, err := h.workflowClient.Workflows(h.op.GitOps.AdminNamespace).
		Create(ctx, wf, metav1.CreateOptions{}); err != nil {
		return errors.Wrapf(err, "create workflow failed")
	}
	return nil
}

// GetWorkflowByID get workflow by id
func (h *workflowHandler) GetWorkflowByID(ctx context.Context, id string) (*workflowv1.Workflow, error) {
	if h.workflowInformer.Informer().HasSynced() {
		wf, err := h.workflowInformer.Lister().Workflows(h.op.GitOps.AdminNamespace).Get(id)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, nil
			}
			return nil, errors.Wrapf(err, "get workflow from informer failed")
		}
		return wf.DeepCopy(), nil
	}
	wf, err := h.workflowClient.Workflows(h.op.GitOps.AdminNamespace).Get(ctx, id, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "get workflow failed")
	}
	return wf, nil
}

// DeleteWorkflow delete workflow by id
func (h *workflowHandler) DeleteWorkflow(ctx context.Context, id string) error {
	if err := h.workflowClient.Workflows(h.op.GitOps.AdminNamespace).
		Delete(ctx, id, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "delete workflow failed")
	}
	return nil
}

// UpdateWorkflow update workflow
func (h *workflowHandler) UpdateWorkflow(ctx context.Context, wf *workflowv1.Workflow) error {
	_, err := h.workflowClient.Workflows(h.op.GitOps.AdminNamespace).Update(ctx, wf, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "update workflow failed")
	}
	return nil
}

// ExecuteWorkflow execute workflow
func (h *workflowHandler) ExecuteWorkflow(ctx context.Context, wf *workflowv1.WorkflowHistory) error {
	_, err := h.workflowClient.WorkflowHistories(h.op.GitOps.AdminNamespace).Create(ctx, wf, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrapf(err, "create workflow history failed")
	}
	return nil
}

// ListWorkflowHistories list workflow histories
func (h *workflowHandler) ListWorkflowHistories(ctx context.Context, id string) (
	[]workflowv1.WorkflowHistory, error) {
	if h.historyInformer.Informer().HasSynced() {
		histories, err := h.historyInformer.Lister().WorkflowHistories(h.op.GitOps.AdminNamespace).
			List(labels.SelectorFromSet(map[string]string{
				workflowv1.HistoryLabelWorkflow: id,
			}))
		if err != nil {
			return nil, errors.Wrapf(err, "list hsitories from informer failed")
		}
		result := make([]workflowv1.WorkflowHistory, 0, len(histories))
		for _, item := range histories {
			result = append(result, *item.DeepCopy())
		}
		return result, nil
	}
	historyList, err := h.workflowClient.WorkflowHistories(h.op.GitOps.AdminNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{workflowv1.HistoryLabelWorkflow: id}).String(),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "list histories failed")
	}
	result := make([]workflowv1.WorkflowHistory, 0, len(historyList.Items))
	result = append(result, historyList.Items...)
	return result, nil
}

// GetWorkflowHistoryDetail get workflow history detail
func (h *workflowHandler) GetWorkflowHistoryDetail(ctx context.Context, historyID string) (
	*workflowv1.WorkflowHistory, error) {
	history, err := h.workflowClient.WorkflowHistories(h.op.GitOps.AdminNamespace).
		Get(ctx, historyID, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "get workflow history failed")
	}
	return history, nil
}
