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

// Package bkdevops defines the handler of bkdevops
package bkdevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/utils/httputils"
	v1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/apis/gitopsworkflow/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/option"
	thirdengine "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/third_engine"
)

type handler struct {
	op *option.ControllerOption
}

// NewBKDevOpsHandler create the instance of bkdevops engine handler
func NewBKDevOpsHandler() thirdengine.EngineInterface {
	return &handler{
		op: option.GlobalOption(),
	}
}

// Validate the workflow request
func (h *handler) Validate(workflow *v1.Workflow) ([]string, []string) {
	validator := &workflowValidator{
		workflow: workflow,
	}
	return validator.Validate()
}

// CreateOrUpdatePipeline create or update the pipeline with workflow defines
func (h *handler) CreateOrUpdatePipeline(ctx context.Context, user string, workflow *v1.Workflow) (string, error) {
	transfer := &workflowTransfer{workflow: workflow}
	pp, err := transfer.transToPipeline()
	if err != nil {
		return "", errors.Wrapf(err, "transfer workflow to pipeline failed")
	}
	ppID := workflow.Status.PipelineID
	if ppID == "" {
		return h.createPipeline(ctx, workflow.Spec.Project, user, pp)
	}
	return "", h.updatePipeline(ctx, workflow.Spec.Project, user, ppID, pp)
}

// DeletePipeline delete pipeline
func (h *handler) DeletePipeline(ctx context.Context, user string, workflow *v1.Workflow) error {
	return h.deletePipeline(ctx, workflow.Spec.Project, user, workflow.Status.PipelineID)
}

// ExecutePipeline execute the pipeline
func (h *handler) ExecutePipeline(ctx context.Context, user string, workflow *v1.Workflow,
	history *v1.WorkflowHistory) (int64, string, error) {
	params := make(map[string]string)
	for i := range history.Spec.Params {
		p := &history.Spec.Params[i]
		params[p.Name] = p.Value
	}
	return h.executePipeline(ctx, workflow.Spec.Project, user, workflow.Status.PipelineID, params)
}

// GetPipelineHistoryStatus return the pipeline history status
func (h *handler) GetPipelineHistoryStatus(ctx context.Context, proj, user, historyID string) (
	*thirdengine.PipelineHistoryStatus, error) {
	return h.queryPipelineHistoryStatus(ctx, proj, user, historyID)
}

const (
	createPath  = "/prod/v4/apigw-app/projects/%s/pipelines/pipeline"
	updatePath  = "/prod/v4/apigw-app/projects/%s/pipelines/pipeline?pipelineId=%s"
	deletePath  = "/prod/v4/apigw-app/projects/%s/pipelines/pipeline?pipelineId=%s"
	executePath = "/prod/v4/apigw-app/projects/%s/build_start?pipelineId=%s"
	detailPath  = "/prod/v4/apigw-app/projects/%s/build_detail?buildId=%s"

	headerDevopsUID           = "X-DEVOPS-UID"
	headerAuthorization       = "X-Bkapi-Authorization"
	headerAuthorizationFormat = `{"bk_app_code":"%s","bk_app_secret":"%s","bk_username":"%s"}`
)

func (h *handler) createPipeline(ctx context.Context, proj, user string, pp *pipeline) (string, error) {
	respBody, err := httputils.Send(ctx, &httputils.HTTPRequest{
		Url:    h.op.BKDevOpsUrl + fmt.Sprintf(createPath, proj),
		Method: http.MethodPost,
		Header: map[string]string{
			headerDevopsUID: user,
			headerAuthorization: fmt.Sprintf(headerAuthorizationFormat,
				h.op.BKDevOpsAppCode, h.op.BKDevOpsAppSecret, user),
		},
		Body: pp,
	})
	if err != nil {
		return "", errors.Wrapf(err, "do request to create pipeline failed")
	}
	resp := new(createResp)
	if err = json.Unmarshal(respBody, resp); err != nil {
		return "", errors.Wrapf(err, "unmarshal response body failed")
	}
	if resp.Status != 0 {
		return "", errors.Errorf("create pipeline resp code not 0 but %d: %s", resp.Status, resp.Message)
	}
	return resp.Data.ID, nil
}

func (h *handler) updatePipeline(ctx context.Context, proj, user, ppID string, pp *pipeline) error {
	respBody, err := httputils.Send(ctx, &httputils.HTTPRequest{
		Url:    h.op.BKDevOpsUrl + fmt.Sprintf(updatePath, proj, ppID),
		Method: http.MethodPut,
		Header: map[string]string{
			headerDevopsUID: user,
			headerAuthorization: fmt.Sprintf(headerAuthorizationFormat,
				h.op.BKDevOpsAppCode, h.op.BKDevOpsAppSecret, user),
		},
		Body: pp,
	})
	if err != nil {
		return errors.Wrapf(err, "do request to update pipeline failed")
	}
	resp := new(updateOrDeleteResp)
	if err = json.Unmarshal(respBody, resp); err != nil {
		return errors.Wrapf(err, "unmarshal response body failed")
	}
	if resp.Status != 0 || !resp.Data {
		return errors.Errorf("update pipeline resp code not 0 but %d: %s", resp.Status, resp.Message)
	}
	return nil
}

func (h *handler) deletePipeline(ctx context.Context, proj, user, ppID string) error {
	respBody, err := httputils.Send(ctx, &httputils.HTTPRequest{
		Url:    h.op.BKDevOpsUrl + fmt.Sprintf(deletePath, proj, ppID),
		Method: http.MethodDelete,
		Header: map[string]string{
			headerDevopsUID: user,
			headerAuthorization: fmt.Sprintf(headerAuthorizationFormat,
				h.op.BKDevOpsAppCode, h.op.BKDevOpsAppSecret, user),
		},
	})
	if err != nil {
		return errors.Wrapf(err, "do request to delete pipeline failed")
	}
	resp := new(updateOrDeleteResp)
	if err = json.Unmarshal(respBody, resp); err != nil {
		return errors.Wrapf(err, "unmarshal response body failed")
	}
	if resp.Status != 0 || !resp.Data {
		return errors.Errorf("delete pipeline resp code not 0 but %d: %s", resp.Status, resp.Message)
	}
	return nil
}

func (h *handler) executePipeline(ctx context.Context, proj, user, ppID string,
	params map[string]string) (int64, string, error) {
	respBody, err := httputils.Send(ctx, &httputils.HTTPRequest{
		Url:    h.op.BKDevOpsUrl + fmt.Sprintf(executePath, proj, ppID),
		Method: http.MethodPost,
		Header: map[string]string{
			headerDevopsUID: user,
			headerAuthorization: fmt.Sprintf(headerAuthorizationFormat,
				h.op.BKDevOpsAppCode, h.op.BKDevOpsAppSecret, user),
		},
		Body: params,
	})
	if err != nil {
		return 0, "", errors.Wrapf(err, "do request to update pipeline failed")
	}
	resp := new(executeResp)
	if err = json.Unmarshal(respBody, resp); err != nil {
		return 0, "", errors.Wrapf(err, "unmarshal response body failed")
	}
	if resp.Status != 0 {
		return 0, "", errors.Errorf("update pipeline resp code not 0 but %d: %s",
			resp.Status, resp.Message)
	}
	return resp.Data.Num, resp.Data.ID, nil
}

// queryPipelineHistoryStatus query the bkdevops pipeline history status
func (h *handler) queryPipelineHistoryStatus(ctx context.Context, proj, user,
	historyID string) (*thirdengine.PipelineHistoryStatus, error) {
	respBody, err := httputils.Send(ctx, &httputils.HTTPRequest{
		Url:    h.op.BKDevOpsUrl + fmt.Sprintf(detailPath, proj, historyID),
		Method: http.MethodGet,
		Header: map[string]string{
			headerDevopsUID: user,
			headerAuthorization: fmt.Sprintf(headerAuthorizationFormat,
				h.op.BKDevOpsAppCode, h.op.BKDevOpsAppSecret, user),
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "do request to update pipeline failed")
	}
	resp := new(executeStatusResp)
	if err = json.Unmarshal(respBody, resp); err != nil {
		return nil, errors.Wrapf(err, "unmarshal response body failed")
	}
	if resp.Status != 0 {
		return nil, errors.Errorf("update pipeline resp code not 0 but %d: %s", resp.Status, resp.Message)
	}
	historyStatus := &thirdengine.PipelineHistoryStatus{
		Status: transferStatus(ctx, resp.Data.Status),
	}
	if resp.Data.StartTime != 0 {
		start := time.UnixMilli(resp.Data.StartTime)
		historyStatus.StartTime = &start
	}
	if resp.Data.EndTime != 0 {
		end := time.UnixMilli(resp.Data.EndTime)
		historyStatus.StartTime = &end
	}
	for i := range resp.Data.Model.Stages {
		stage := &resp.Data.Model.Stages[i]
		stageStatus := thirdengine.StageStatus{
			Status: transferStatus(ctx, stage.Status),
		}
		for j := range stage.Containers {
			ct := stage.Containers[j]
			jobStatus := thirdengine.JobStatus{
				Status: transferStatus(ctx, ct.Status),
			}
			for k := range ct.Elements {
				ele := ct.Elements[k]
				jobStatus.StepStatuses = append(jobStatus.StepStatuses, thirdengine.StepStatus{
					Status: transferStatus(ctx, ele.Status),
				})
			}
			stageStatus.JobStatuses = append(stageStatus.JobStatuses, jobStatus)
		}
		historyStatus.StageStatuses = append(historyStatus.StageStatuses, stageStatus)
	}
	return historyStatus, nil
}

func transferStatus(ctx context.Context, status string) v1.HistoryStatus {
	switch status {
	case "FAILED":
		return v1.HistoryFailed
	case "SUCCEED":
		return v1.HistorySuccess
	case "RUNNING":
		return v1.HistoryRunning
	case "":
		return ""
	default:
		logctx.Warnf(ctx, "unknown bkdevops status '%s'", status)
		return v1.HistoryRunning
	}
}
