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

package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	workflowv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/apis/gitopsworkflow/v1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/workflowstore"
)

// WorkflowPlugin defines the plugin for workflow
type WorkflowPlugin struct {
	*mux.Router
	op            *options.Options
	middleware    mw.MiddlewareInterface
	workflowStore workflowstore.WorkflowInterface
	permitChecker permitcheck.PermissionInterface
}

// Init the workflow router
func (plugin *WorkflowPlugin) Init() error {
	// 获取 workflow 列表
	plugin.Path("").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.listHandler))
	// 创建 workflow
	plugin.Path("").Methods(http.MethodPost).Handler(plugin.middleware.HttpWrapper(plugin.createHandler))
	// 获取 workflow 详情
	plugin.Path("/{id}").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.getDetailHandler))
	// 删除 workflow
	plugin.Path("/{id}").Methods(http.MethodDelete).Handler(plugin.middleware.HttpWrapper(plugin.deleteHandler))
	// 更新 workflow
	plugin.Path("/{id}").Methods(http.MethodPut).Handler(plugin.middleware.HttpWrapper(plugin.updateHandler))
	// 执行 workflow
	plugin.Path("/{id}/execute").Methods(http.MethodPost).
		Handler(plugin.middleware.HttpWrapper(plugin.executeHandler))

	// 获取 workflow 的执行历史列表
	plugin.Path("/{id}/histories").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.listHistoriesHandler))
	// 获取某个执行历史的详情
	plugin.Path("/histories/{historyID}").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.getHistoryDetailHandler))
	plugin.workflowStore = workflowstore.NewWorkflowHandler()
	if err := plugin.workflowStore.Init(); err != nil {
		return errors.Wrapf(err, "workflow store init failed")
	}
	return nil
}

// listHandler lis the workflows by query projects
func (plugin *WorkflowPlugin) listHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectList, statusCode, err := plugin.middleware.ListProjects(r.Context())
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "list projects failed"))
	}
	allProjects := make(map[string]struct{})
	for i := range projectList.Items {
		allProjects[projectList.Items[i].Name] = struct{}{}
	}
	projects := r.URL.Query()["projects"]
	queryProjects := make([]string, 0, len(allProjects))
	if len(projects) == 0 {
		for proj := range allProjects {
			queryProjects = append(queryProjects, proj)
		}
	} else {
		for i := range projects {
			if _, ok := allProjects[projects[i]]; ok {
				queryProjects = append(queryProjects, projects[i])
			}
		}
	}
	if len(queryProjects) == 0 {
		return r, mw.ReturnJSONResponse(nil)
	}
	workflows, err := plugin.workflowStore.ListWorkflows(r.Context(), queryProjects)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse(workflows)
}

// checkWorkflow check workflow params is legal
func (plugin *WorkflowPlugin) checkWorkflow(wf *workflowv1.Workflow) error {
	proj := wf.Spec.Project
	if proj == "" {
		return errors.Errorf("spec.project cannot be empty")
	}
	if wf.Spec.Name == "" {
		return errors.Errorf("spec.name cannot be empty")
	}
	if wf.Spec.Engine == "" {
		return errors.Errorf("spec.engine cannot be empty")
	}
	return nil
}

// createHandler will create the Workflow object
func (plugin *WorkflowPlugin) createHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	wf := new(workflowv1.Workflow)
	if err = json.Unmarshal(body, wf); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal request body failed"))
	}
	if wf.Status.PipelineID != "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("status.pipelineID should be empty"))
	}
	if err = plugin.checkWorkflow(wf); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "check workflow failed"))
	}
	_, status, err := plugin.permitChecker.CheckProjectPermission(r.Context(), wf.Spec.Project,
		permitcheck.ProjectViewRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}

	user := ctxutils.User(r.Context())
	wf.TypeMeta = metav1.TypeMeta{
		APIVersion: "gitopsworkflow.bkbcs.tencent.com/v1",
		Kind:       "Workflow",
	}
	wf.ObjectMeta = metav1.ObjectMeta{
		Name: workflowv1.SecretName("workflow", fmt.Sprintf("%s/%d",
			wf.Spec.Name, time.Now().UnixMilli())),
		Namespace: plugin.op.GitOps.AdminNamespace,
		Annotations: map[string]string{
			workflowv1.WorkflowAnnotationCreateUser: user.GetUser(),
			workflowv1.WorkflowAnnotationUpdateUser: user.GetUser(),
		},
		Labels: map[string]string{
			workflowv1.WorkflowLabelProject: wf.Spec.Project,
		},
	}
	if err = plugin.workflowStore.CreateWorkflow(r.Context(), wf); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "create workflow failed"))
	}
	return r, mw.ReturnJSONResponse("ok")
}

// updateHandler will check user's permission and update the specified workflow
func (plugin *WorkflowPlugin) updateHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	wf := new(workflowv1.Workflow)
	if err = json.Unmarshal(body, wf); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal request body failed"))
	}
	if wf.Name == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("workflow 'name‘ cannot be empty"))
	}
	if err = plugin.checkWorkflow(wf); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "check workflow failed"))
	}

	k8sWorkflow, httpResp := plugin.checkWorkflowPermission(r.Context(), wf.Name)
	if httpResp != nil {
		return r, httpResp
	}
	user := ctxutils.User(r.Context())
	k8sWorkflow.Annotations[workflowv1.WorkflowAnnotationUpdateUser] = user.GetUser()
	if k8sWorkflow.Spec.Engine != wf.Spec.Engine {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Errorf("spec.engine cannot change: %s -> %s", k8sWorkflow.Spec.Engine, wf.Spec.Engine))
	}
	if k8sWorkflow.Spec.Project != wf.Spec.Project {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Errorf("spec.project cannot change: %s -> %s", k8sWorkflow.Spec.Project, wf.Spec.Project))
	}
	k8sWorkflow.Spec = wf.Spec
	if err = plugin.workflowStore.UpdateWorkflow(r.Context(), k8sWorkflow); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse("ok")
}

// getDetailHandler return the workflow detail
func (plugin *WorkflowPlugin) getDetailHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	name := mux.Vars(r)["id"]
	wf, httpResp := plugin.checkWorkflowPermission(r.Context(), name)
	if httpResp != nil {
		return r, httpResp
	}
	return r, mw.ReturnJSONResponse(wf)
}

// deleteHandler delete workflow by workflow id
func (plugin *WorkflowPlugin) deleteHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	name := mux.Vars(r)["id"]
	_, httpResp := plugin.checkWorkflowPermission(r.Context(), name)
	if httpResp != nil {
		return r, httpResp
	}
	if err := plugin.workflowStore.DeleteWorkflow(r.Context(), name); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse("ok")
}

// executeHandler will create the WorkflowHistory to execute workflow
func (plugin *WorkflowPlugin) executeHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	name := mux.Vars(r)["id"]
	wf, httpResp := plugin.checkWorkflowPermission(r.Context(), name)
	if httpResp != nil {
		return r, httpResp
	}
	user := ctxutils.User(r.Context())
	controller := true
	block := true
	history := &workflowv1.WorkflowHistory{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gitopsworkflow.bkbcs.tencent.com/v1",
			Kind:       "WorkflowHistory",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: workflowv1.SecretName(wf.Spec.Project+"-history",
				fmt.Sprintf("%d-%d", time.Now().UnixMilli(), workflowv1.RandomNum())),
			Namespace: plugin.op.GitOps.AdminNamespace,
			Annotations: map[string]string{
				workflowv1.WorkflowAnnotationCreateUser: user.GetUser(),
				workflowv1.WorkflowAnnotationUpdateUser: user.GetUser(),
				workflowv1.HistoryAnnotationWorkflow:    wf.Name,
			},
			Labels: map[string]string{
				workflowv1.HistoryLabelWorkflow: wf.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "gitopsworkflow.bkbcs.tencent.com/v1",
					Kind:               "Workflow",
					Name:               wf.Name,
					UID:                wf.UID,
					Controller:         &controller,
					BlockOwnerDeletion: &block,
				},
			},
		},
		Spec: workflowv1.WorkflowHistorySpec{
			TriggerByWorkflow: true,
			TriggerType:       "manual",
		},
	}
	if err := plugin.workflowStore.ExecuteWorkflow(r.Context(), history); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, err)
	}
	return r, mw.ReturnJSONResponse(history.Name)
}

// listHistoriesHandler lis histories by workflow-id handler
func (plugin *WorkflowPlugin) listHistoriesHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	workflowName := mux.Vars(r)["id"]
	_, httpResp := plugin.checkWorkflowPermission(r.Context(), workflowName)
	if httpResp != nil {
		return r, httpResp
	}
	histories, err := plugin.workflowStore.ListWorkflowHistories(r.Context(), workflowName)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse(histories)
}

// getHistoryDetailHandler return the history detail handler
func (plugin *WorkflowPlugin) getHistoryDetailHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	historyID := mux.Vars(r)["historyID"]
	history, err := plugin.workflowStore.GetWorkflowHistoryDetail(r.Context(), historyID)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	if history == nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("workflow history not found"))
	}
	if len(history.OwnerReferences) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Errorf("workflow history not have parent workflow"))
	}
	_, httpResp := plugin.checkWorkflowPermission(r.Context(), history.OwnerReferences[0].Name)
	if httpResp != nil {
		return r, httpResp
	}
	return r, mw.ReturnJSONResponse(history)
}

// checkWorkflowPermission get the workflow object, and check the workflow's project permission
// with user login
func (plugin *WorkflowPlugin) checkWorkflowPermission(ctx context.Context, name string) (
	*workflowv1.Workflow, *mw.HttpResponse) {
	wf, err := plugin.workflowStore.GetWorkflowByID(ctx, name)
	if err != nil {
		return nil, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	if wf == nil {
		return nil, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("workflow '%s' not found", name))
	}
	_, status, err := plugin.permitChecker.CheckProjectPermission(ctx, wf.Spec.Project, permitcheck.ProjectViewRSAction)
	if err != nil {
		return nil, mw.ReturnErrorResponse(status, errors.Wrapf(err, "check project permission failed"))
	}
	return wf, nil
}
