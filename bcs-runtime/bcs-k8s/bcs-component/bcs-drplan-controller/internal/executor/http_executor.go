/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executor

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

// HTTPActionExecutor implements ActionExecutor for HTTP actions
type HTTPActionExecutor struct{}

// NewHTTPActionExecutor creates a new HTTP action executor
func NewHTTPActionExecutor() *HTTPActionExecutor {
	return &HTTPActionExecutor{}
}

// Execute executes an HTTP action
func (e *HTTPActionExecutor) Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Executing HTTP action: %s", action.Name)
	startTime := time.Now()
	status := e.newHTTPActionStatus(action.Name, startTime)

	if action.HTTP == nil {
		e.setHTTPActionStatusFailed(status, "HTTP configuration is nil")
		return status, fmt.Errorf("HTTP configuration is required")
	}

	url, body, method, headers, retryConfig, err := e.prepareHTTPRequestParams(action, params)
	if err != nil {
		e.setHTTPActionStatusFailed(status, err.Error())
		return status, err
	}

	httpResp, err := e.executeHTTPWithRetry(ctx, action, status, url, body, method, headers, retryConfig)
	if err != nil {
		e.setHTTPActionStatusFailed(status, err.Error())
		return status, err
	}

	e.setHTTPActionStatusSuccess(status, method, httpResp)
	klog.Infof("HTTP action %s completed successfully", action.Name)
	return status, nil
}

func (e *HTTPActionExecutor) newHTTPActionStatus(name string, startTime time.Time) *drv1alpha1.ActionStatus {
	return &drv1alpha1.ActionStatus{
		Name:      name,
		Phase:     drv1alpha1.PhaseRunning,
		StartTime: &metav1.Time{Time: startTime},
	}
}

func (e *HTTPActionExecutor) setHTTPActionStatusFailed(status *drv1alpha1.ActionStatus, message string) {
	status.Phase = drv1alpha1.PhaseFailed
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = message
}

// prepareHTTPRequestParams renders URL, body, method, headers and parses retry policy
func (e *HTTPActionExecutor) prepareHTTPRequestParams(action *drv1alpha1.Action, params map[string]interface{}) (
	url, body, method string, headers map[string]string, retryConfig *utils.RetryConfig, err error) {
	templateData := &utils.TemplateData{Params: params}

	url, err = utils.RenderTemplate(action.HTTP.URL, templateData)
	if err != nil {
		return "", "", "", nil, nil, fmt.Errorf("failed to render URL: %w", err)
	}
	if action.HTTP.Body != "" {
		body, err = utils.RenderTemplate(action.HTTP.Body, templateData)
		if err != nil {
			return "", "", "", nil, nil, fmt.Errorf("failed to render body: %w", err)
		}
	}
	method = action.HTTP.Method
	if method != "" {
		method, err = utils.RenderTemplate(method, templateData)
		if err != nil {
			return "", "", "", nil, nil, fmt.Errorf("failed to render method: %w", err)
		}
	}
	headers = make(map[string]string)
	for k, v := range action.HTTP.Headers {
		rendered, rerr := utils.RenderTemplate(v, templateData)
		if rerr != nil {
			return "", "", "", nil, nil, fmt.Errorf("failed to render header %q: %w", k, rerr)
		}
		headers[k] = rendered
	}
	retryConfig, err = utils.ParseRetryPolicy(action.RetryPolicy)
	if err != nil {
		return "", "", "", nil, nil, fmt.Errorf("invalid retry policy: %w", err)
	}
	return url, body, method, headers, retryConfig, nil
}

func (e *HTTPActionExecutor) executeHTTPWithRetry(ctx context.Context, action *drv1alpha1.Action, status *drv1alpha1.ActionStatus,
	url, body, method string, headers map[string]string, retryConfig *utils.RetryConfig) (*http.Response, error) {
	var httpResp *http.Response
	err := utils.RetryWithBackoff(ctx, retryConfig, func(ctx context.Context, attempt int32) error {
		var reqErr error
		httpResp, reqErr = e.executeHTTPRequest(ctx, action, url, body, method, headers)
		if reqErr != nil {
			klog.V(4).Infof("HTTP request attempt %d failed: %v", attempt, reqErr)
			return reqErr
		}
		status.RetryCount = attempt
		return nil
	})
	if err != nil {
		return nil, err
	}
	return httpResp, nil
}

func (e *HTTPActionExecutor) setHTTPActionStatusSuccess(status *drv1alpha1.ActionStatus, method string, httpResp *http.Response) {
	respBody, _ := io.ReadAll(httpResp.Body)
	_ = httpResp.Body.Close()
	bodyStr := string(respBody)
	if len(bodyStr) > 1000 {
		bodyStr = bodyStr[:1000]
	}
	status.Outputs = &drv1alpha1.ActionOutputs{
		HTTPResponse: &drv1alpha1.HTTPResponse{StatusCode: httpResp.StatusCode, Body: bodyStr},
	}
	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = fmt.Sprintf("HTTP %s succeeded with status %d", method, httpResp.StatusCode)
}

// executeHTTPRequest executes a single HTTP request (method and headers are already rendered)
func (e *HTTPActionExecutor) executeHTTPRequest(ctx context.Context, action *drv1alpha1.Action, url, body, method string, headers map[string]string) (*http.Response, error) {
	if method == "" {
		method = "GET"
	}

	klog.V(4).Infof("HTTP %s %s", method, url)

	// Create request
	var req *http.Request
	var err error
	if body != "" {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if action.HTTP.InsecureSkipVerify {
		client.Transport = &http.Transport{
			// #nosec G402 - InsecureSkipVerify is acceptable for internal test environments
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	// Check success codes (from original action, not parameterized)
	successCodes := action.HTTP.SuccessCodes
	if len(successCodes) == 0 {
		// Default: 200-299
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return resp, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	} else {
		success := false
		for _, code := range successCodes {
			if resp.StatusCode == code {
				success = true
				break
			}
		}
		if !success {
			return resp, fmt.Errorf("status code %d not in success codes", resp.StatusCode)
		}
	}

	return resp, nil
}

// Rollback rolls back an HTTP action (no-op for HTTP unless custom rollback defined)
func (e *HTTPActionExecutor) Rollback(ctx context.Context, action *drv1alpha1.Action, actionStatus *drv1alpha1.ActionStatus, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("HTTP action %s rollback", action.Name)

	// Create rollback status object
	rollbackStatus := &drv1alpha1.ActionStatus{
		Name:      actionStatus.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: time.Now()},
	}

	// Execute custom rollback if defined
	if action.Rollback != nil {
		klog.V(4).Infof("Executing custom rollback for HTTP action %s", action.Name)
		customStatus, err := e.Execute(ctx, action.Rollback, params)
		if err != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Custom rollback failed: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, err
		}
		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = "Rolled back: executed custom HTTP rollback action"
		rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		rollbackStatus.Outputs = customStatus.Outputs
		return rollbackStatus, nil
	}

	// No default rollback for HTTP actions
	rollbackStatus.Phase = drv1alpha1.PhaseSkipped
	rollbackStatus.Message = "No rollback defined for HTTP action"
	rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	klog.V(4).Infof("HTTP action %s rollback skipped (no default rollback)", action.Name)
	return rollbackStatus, nil
}

// Type returns the action type
func (e *HTTPActionExecutor) Type() string {
	return "HTTP"
}
