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

// Package workflow xx
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	workflowv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/apis/gitopsworkflow/v1"
	"github.com/fatih/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/httputils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

// Handler defines the handler of workflow
type Handler struct {
	op *options.GitOpsOptions
}

// NewHandler create the workflow handler instance
func NewHandler() *Handler {
	return &Handler{
		op: options.GlobalOption(),
	}
}

var (
	listPath          = "/api/v1/workflows"
	createPath        = "/api/v1/workflows"
	getDetailPath     = "/api/v1/workflows/%s"
	deletePath        = "/api/v1/workflows/%s"
	updatePath        = "/api/v1/workflows/%s"
	executePath       = "/api/v1/workflows/%s/execute"
	listHistoryPath   = "/api/v1/workflows/%s/histories"
	historyDetailPath = "/api/v1/workflows/histories/%s"
)

// List workflows by projects specified
func (h *Handler) List(ctx context.Context, projects *[]string) {
	queryParams := make(map[string][]string)
	if len(*projects) != 0 {
		queryParams["projects"] = *projects
	}
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:        listPath,
		Method:      http.MethodGet,
		QueryParams: queryParams,
	})
	workflows := make([]workflowv1.Workflow, 0)
	if err := json.Unmarshal(respBody, &workflows); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response body failed: %s", err.Error()))
	}

	tw := utils.DefaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "PROJECT", "ENGINE", "DESTROY-ON-DELETION", "STATUS", "NAME", "PIPELINE",
		}
	}())
	for i := range workflows {
		wf := workflows[i]
		tw.Append(func() []string {
			return []string{
				wf.Name, wf.Spec.Project, wf.Spec.Engine, fmt.Sprintf("%v", wf.Spec.DestroyOnDeletion),
				wf.Status.Phase, wf.Spec.Name, wf.Status.PipelineID,
			}
		}())
	}
	tw.Render()
}

func (h *Handler) readSpec(fp string) ([]byte, *workflowv1.WorkflowSpec) {
	bs, err := os.ReadFile(fp)
	if err != nil {
		utils.ExitError(fmt.Sprintf("read file '%s' failed: %s", fp, err.Error()))
	}
	bs = utils.YamlToJson(bs)
	spec := new(workflowv1.WorkflowSpec)
	if err = json.Unmarshal(bs, spec); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal failed: %s", err.Error()))
	}
	if spec.Engine == "" {
		utils.ExitError("'engine' cannot be empty")
	}
	if spec.Name == "" {
		utils.ExitError("'name' cannot be empty")
	}
	if spec.Project == "" {
		utils.ExitError("'project' cannot be empty")
	}
	return bs, spec
}

// Create the workflow
func (h *Handler) Create(ctx context.Context, fp string) {
	_, spec := h.readSpec(fp)
	wf := workflowv1.Workflow{
		Spec: *spec,
	}
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   createPath,
		Method: http.MethodPost,
		Body:   wf,
	})
	fmt.Println(string(respBody))
}

type workflowID struct {
	ID string `json:"id"`
}

// Update update workflow
func (h *Handler) Update(ctx context.Context, fp string) {
	bs, spec := h.readSpec(fp)
	wfid := new(workflowID)
	if err := json.Unmarshal(bs, wfid); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal workflow id failed: %s", err.Error()))
	}
	if wfid.ID == "" {
		utils.ExitError("'id' of workflow should define when update")
	}
	wf := workflowv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: wfid.ID,
		},
		Spec: *spec,
	}
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(updatePath, wf.Name),
		Method: http.MethodPut,
		Body:   wf,
	})
	fmt.Println(string(respBody))
}

// GetDetail return the workflow detail by id
func (h *Handler) GetDetail(ctx context.Context, id string) {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(getDetailPath, id),
		Method: http.MethodGet,
	})
	wf := new(workflowv1.Workflow)
	if err := json.Unmarshal(respBody, wf); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal failed: %s", err.Error()))
	}
	bs, err := json.Marshal(wf.Spec)
	if err != nil {
		utils.ExitError(fmt.Sprintf("marshal failed: %s", err.Error()))
	}

	color.Green("## CreateUser: %s\n", wf.Annotations[workflowv1.WorkflowAnnotationCreateUser])
	color.Green("## UpdateUser: %s\n", wf.Annotations[workflowv1.WorkflowAnnotationUpdateUser])
	color.Green("## LastUpdateTime: %s\n", wf.Status.LastUpdateTime.Format("2006-01-02 15:04:05"))
	if wf.Status.PipelineID != "" {
		color.Green("## PipelineID: %s\n", wf.Status.PipelineID)
	}
	if wf.Status.Phase == "Error" {
		color.Red("## Status: %s\n", wf.Status.Phase)
		if wf.Status.Message != "" {
			color.Red("## Message: %s\n", wf.Status.Message)
		}
	} else {
		color.Green("## Status: %s\n", wf.Status.Phase)
		if wf.Status.Message != "" {
			color.Green("## Message: %s\n", wf.Status.Message)
		}
	}
	fmt.Println("---")
	fmt.Printf("id: %s\n", id)
	result := string(utils.JsonToYaml(bs))
	lines := strings.Split(result, "\n")
	needDelete2Indents := false
	for i := range lines {
		lineResult := lines[i]
		if strings.HasPrefix(lineResult, "  ") {
			lineResult = strings.TrimPrefix(lineResult, "  ")
		}
		if needDelete2Indents && strings.HasPrefix(lineResult, "    ") {
			lineResult = strings.TrimPrefix(lineResult, "  ")
		}
		if strings.HasPrefix(lineResult, "stages:") {
			needDelete2Indents = true
		}
		if strings.HasPrefix(lineResult, "stages:") || strings.HasPrefix(lineResult, "stepTemplates") {
			fmt.Println()
			fmt.Println()
		}
		fmt.Println(lineResult)
	}
}

// Delete delete the workflow by id
func (h *Handler) Delete(ctx context.Context, id string) {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(deletePath, id),
		Method: http.MethodDelete,
	})
	fmt.Println(string(respBody))
}

// Execute the workflow by workflow id
func (h *Handler) Execute(ctx context.Context, id string) {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(executePath, id),
		Method: http.MethodPost,
	})
	fmt.Println(string(respBody))
}

// ListHistories list the histories by workflow id
func (h *Handler) ListHistories(ctx context.Context, id string) {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(listHistoryPath, id),
		Method: http.MethodGet,
	})
	histories := make([]workflowv1.WorkflowHistory, 0)
	if err := json.Unmarshal(respBody, &histories); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response failed: %s", err.Error()))
	}

	tw := utils.DefaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NUM", "STATUS", "WORKFLOW-TRIGGER", "TRIGGER", "JOB-ID",
		}
	}())
	for i := range histories {
		history := histories[i]
		tw.Append(func() []string {
			return []string{
				history.Name, strconv.Itoa(int(history.Status.HistoryNum)), string(history.Status.Phase),
				fmt.Sprintf("%v", history.Spec.TriggerByWorkflow), history.Spec.TriggerType,
				history.Status.HistoryID,
			}
		}())
	}
	tw.Render()
}

// HistoryDetail return the detail of history
func (h *Handler) HistoryDetail(ctx context.Context, history string, showDetails bool) {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(historyDetailPath, history),
		Method: http.MethodGet,
	})
	wfHistory := new(workflowv1.WorkflowHistory)
	if err := json.Unmarshal(respBody, wfHistory); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response body failed: %s", err.Error()))
	}
	color.Green("## CreateUser: %s\n", wfHistory.Annotations[workflowv1.WorkflowAnnotationCreateUser])
	if wfHistory.Status.HistoryID != "" {
		color.Green("## HistoryID: %s\n", wfHistory.Status.HistoryID)
	}
	if wfHistory.Status.StartedAt != nil {
		color.Green("## StartedAt: %s\n", wfHistory.Status.StartedAt.Format("2006-01-02 15:04:05"))
	}
	if wfHistory.Status.FinishedAt != nil {
		color.Green("## FinishedAt: %s\n", wfHistory.Status.FinishedAt.Format("2006-01-02 15:04:05"))
	}
	if wfHistory.Status.CostTime != "" {
		color.Green("## CostTime: %s\n", wfHistory.Status.CostTime)
	}
	if wfHistory.Status.Phase == "Error" {
		color.Red("## Status: %s\n", wfHistory.Status.Phase)
		if wfHistory.Status.Message != "" {
			color.Red("## Message: %s\n", wfHistory.Status.Message)
		}
	} else {
		color.Green("## Status: %s\n", wfHistory.Status.Phase)
		if wfHistory.Status.Message != "" {
			color.Green("## Message: %s\n", wfHistory.Status.Message)
		}
	}
}
