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

	status := &drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: startTime},
	}

	if action.HTTP == nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = "HTTP configuration is nil"
		return status, fmt.Errorf("HTTP configuration is required")
	}

	// Render URL with parameters
	templateData := &utils.TemplateData{Params: params}
	url, err := utils.RenderTemplate(action.HTTP.URL, templateData)
	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to render URL: %v", err)
		return status, err
	}

	// Render body if present
	var body string
	if action.HTTP.Body != "" {
		body, err = utils.RenderTemplate(action.HTTP.Body, templateData)
		if err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to render body: %v", err)
			return status, err
		}
	}

	// Render method and headers with parameters
	method := action.HTTP.Method
	if method != "" {
		method, err = utils.RenderTemplate(method, templateData)
		if err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to render method: %v", err)
			return status, err
		}
	}
	headers := make(map[string]string)
	for k, v := range action.HTTP.Headers {
		rendered, err := utils.RenderTemplate(v, templateData)
		if err != nil {
			status.Phase = drv1alpha1.PhaseFailed
			status.CompletionTime = &metav1.Time{Time: time.Now()}
			status.Message = fmt.Sprintf("Failed to render header %q: %v", k, err)
			return status, err
		}
		headers[k] = rendered
	}

	// Parse retry policy
	retryConfig, err := utils.ParseRetryPolicy(action.RetryPolicy)
	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Invalid retry policy: %v", err)
		return status, err
	}

	// Execute with retry
	var httpResp *http.Response
	err = utils.RetryWithBackoff(ctx, retryConfig, func(ctx context.Context, attempt int32) error {
		httpResp, err = e.executeHTTPRequest(ctx, action, url, body, method, headers)
		if err != nil {
			klog.V(4).Infof("HTTP request attempt %d failed: %v", attempt, err)
			return err
		}
		status.RetryCount = attempt
		return nil
	})

	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = err.Error()
		return status, err
	}

	// Store response
	respBody, _ := io.ReadAll(httpResp.Body)
	_ = httpResp.Body.Close()

	status.Outputs = &drv1alpha1.ActionOutputs{
		HTTPResponse: &drv1alpha1.HTTPResponse{
			StatusCode: httpResp.StatusCode,
			Body:       string(respBody)[:min(len(respBody), 1000)], // Truncate to 1000 chars
		},
	}

	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = fmt.Sprintf("HTTP %s succeeded with status %d", method, httpResp.StatusCode)

	klog.Infof("HTTP action %s completed successfully", action.Name)
	return status, nil
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
