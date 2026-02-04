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
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

// JobActionExecutor implements ActionExecutor for Job actions
type JobActionExecutor struct {
	client client.Client
}

// NewJobActionExecutor creates a new Job action executor
func NewJobActionExecutor(client client.Client) *JobActionExecutor {
	return &JobActionExecutor{client: client}
}

// Execute executes a Job action
func (e *JobActionExecutor) Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Executing Job action: %s", action.Name)
	startTime := time.Now()

	status := &drv1alpha1.ActionStatus{
		Name:      action.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: startTime},
	}

	if action.Job == nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = "Job configuration is nil"
		return status, fmt.Errorf("job configuration is required")
	}

	templateData := &utils.TemplateData{Params: params}
	jobNamespace, err := utils.RenderTemplate(action.Job.Namespace, templateData)
	if err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to render Job namespace: %v", err)
		return status, err
	}
	if jobNamespace == "" {
		jobNamespace = "default"
	}

	// Create Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", action.Name),
			Namespace:    jobNamespace,
		},
		Spec: action.Job.Template.Spec,
	}

	// Set TTL if specified
	if action.Job.TTLSecondsAfterFinished != nil {
		job.Spec.TTLSecondsAfterFinished = action.Job.TTLSecondsAfterFinished
	}

	klog.V(4).Infof("Creating Job in namespace %s", job.Namespace)
	if err := e.client.Create(ctx, job); err != nil {
		status.Phase = drv1alpha1.PhaseFailed
		status.CompletionTime = &metav1.Time{Time: time.Now()}
		status.Message = fmt.Sprintf("Failed to create Job: %v", err)
		return status, err
	}

	// Store job reference
	status.Outputs = &drv1alpha1.ActionOutputs{
		JobRef: &corev1.ObjectReference{
			Kind:      "Job",
			Namespace: job.Namespace,
			Name:      job.Name,
			UID:       job.UID,
		},
	}

	status.Phase = drv1alpha1.PhaseSucceeded
	status.CompletionTime = &metav1.Time{Time: time.Now()}
	status.Message = fmt.Sprintf("Job %s/%s created successfully", job.Namespace, job.Name)

	klog.Infof("Job action %s completed, created Job %s/%s", action.Name, job.Namespace, job.Name)
	return status, nil
}

// Rollback rolls back a Job action by deleting the job
func (e *JobActionExecutor) Rollback(ctx context.Context, action *drv1alpha1.Action, actionStatus *drv1alpha1.ActionStatus, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	klog.Infof("Rolling back Job action: %s", action.Name)

	// Create rollback status object
	rollbackStatus := &drv1alpha1.ActionStatus{
		Name:      actionStatus.Name,
		Phase:     "Running",
		StartTime: &metav1.Time{Time: time.Now()},
	}

	// Execute custom rollback if defined
	if action.Rollback != nil {
		klog.V(4).Infof("Executing custom rollback for Job action %s", action.Name)
		customStatus, err := e.Execute(ctx, action.Rollback, params)
		if err != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Custom rollback failed: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, err
		}
		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = "Rolled back: executed custom rollback action"
		rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
		rollbackStatus.Outputs = customStatus.Outputs
		return rollbackStatus, nil
	}

	// Automatic rollback: delete the job
	if actionStatus.Outputs != nil && actionStatus.Outputs.JobRef != nil {
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      actionStatus.Outputs.JobRef.Name,
				Namespace: actionStatus.Outputs.JobRef.Namespace,
			},
		}

		klog.V(4).Infof("Deleting Job %s/%s", job.Namespace, job.Name)
		if err := e.client.Delete(ctx, job); client.IgnoreNotFound(err) != nil {
			rollbackStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStatus.Message = fmt.Sprintf("Failed to delete Job: %v", err)
			rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			return rollbackStatus, fmt.Errorf("failed to delete Job: %w", err)
		}

		klog.Infof("Job %s/%s deleted successfully", job.Namespace, job.Name)
		rollbackStatus.Phase = drv1alpha1.PhaseSucceeded
		rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted Job %s/%s", job.Namespace, job.Name)
	} else {
		// No job to delete
		rollbackStatus.Phase = drv1alpha1.PhaseSkipped
		rollbackStatus.Message = "No Job to rollback"
	}

	rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	return rollbackStatus, nil
}

// Type returns the action type
func (e *JobActionExecutor) Type() string {
	return "Job"
}
