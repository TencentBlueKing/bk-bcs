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

package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-image-loader/api/v1alpha1"
)

func (r *ImageLoaderReconciler) reconcileImageLoader(ctx context.Context,
	imageLoader *tkexv1alpha1.ImageLoader) (
	*tkexv1alpha1.ImageLoaderStatus, *time.Duration, error) {
	var requeue time.Duration
	var err error
	newStatus := imageLoader.Status.DeepCopy()
	// 1. check if the spec is changed
	newRevision := getRevisionHash(&imageLoader.Spec)
	if newRevision == newStatus.Revision &&
		newStatus.Desired != int32(0) &&
		newStatus.Completed == newStatus.Desired {
		logger.Info("ImageLoader complete, skip reconcile")
		return newStatus, nil, nil
	}
	if newStatus.Revision == "" {
		r.resetStatus(imageLoader, newStatus)
		now := metav1.Now()
		newStatus.StartTime = &now
		newStatus.Revision = newRevision
		r.Recorder.Eventf(imageLoader, corev1.EventTypeNormal, "Start", "start to preload images")
	}
	if newRevision != newStatus.Revision {
		logger.Info("ImageLoader spec changed")
		r.resetStatus(imageLoader, newStatus)
		finished, cleanErr := r.cleanJobs(ctx, imageLoader)
		if cleanErr != nil {
			return newStatus, nil, cleanErr
		}
		if !finished {
			requeue = time.Second
			return newStatus, &requeue, nil
		}
		logger.Info("finish clean previous jobs")
		now := metav1.Now()
		newStatus.StartTime = &now
		newStatus.Revision = newRevision
	}

	// 2. create jobs if need
	baseJob := newJob(imageLoader)
	err = r.handleSelector(ctx, imageLoader, baseJob)
	if err != nil {
		return newStatus, nil, err
	}
	if baseJob.Spec.Completions == nil || *baseJob.Spec.Completions == 0 {
		r.resetStatus(imageLoader, newStatus)
		newStatus.ObservedGeneration = imageLoader.Generation
		newStatus.Completed = newStatus.Desired
		newStatus.Succeeded = newStatus.Desired
		logger.Info("no node need to preload image")
		r.Recorder.Eventf(imageLoader, corev1.EventTypeWarning, "Complete", "no node need to preload image")
		return newStatus, nil, nil
	}

	err = r.createJobsIfNeed(ctx, imageLoader, baseJob)
	if err != nil {
		return newStatus, nil, err
	}

	// 3. update status
	err = r.updateStatus(ctx, imageLoader, newStatus)
	if err != nil {
		return newStatus, nil, err
	}

	// 4. clean up jobs if all complete succeed
	if newStatus.Completed == newStatus.Desired && newStatus.Succeeded == newStatus.Desired {
		logger.Info("imagerloader's all jobs complete successfully, clean up jobs")
		_, err = r.cleanJobs(ctx, imageLoader)
		if err != nil {
			return newStatus, nil, err
		}
	}

	return newStatus, nil, nil
}

func (r *ImageLoaderReconciler) cleanJobs(ctx context.Context,
	loader *tkexv1alpha1.ImageLoader) (bool, error) {
	jobList := &batchv1.JobList{}
	if err := r.List(ctx, jobList, client.MatchingLabels{
		ImageLoaderNameKey: loader.Name,
	}, client.InNamespace(loader.Namespace)); err != nil {
		return false, err
	}
	if len(jobList.Items) == 0 {
		return true, nil
	}
	for i := range jobList.Items {
		logger.Info("delete job", "job", fmt.Sprintf(jobList.Items[i].Namespace, jobList.Items[i].Name))
		if err := r.Delete(ctx, &jobList.Items[i],
			client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
			return false, err
		}
	}
	return false, nil
}

func (r *ImageLoaderReconciler) createJobsIfNeed(ctx context.Context,
	loader *tkexv1alpha1.ImageLoader, baseJob *batchv1.Job) error {
	for i := range loader.Spec.Images {
		job := &batchv1.Job{}
		err := r.Get(ctx, types.NamespacedName{Namespace: loader.Namespace,
			Name: getJobName(loader, i)}, job)
		if err != nil && errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("create job for %d image", i))
			err = r.createJob(ctx, loader, baseJob, i)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ImageLoaderReconciler) createJob(ctx context.Context, loader *tkexv1alpha1.ImageLoader,
	baseJob *batchv1.Job, index int) error {
	// some field may be nil after deepcopy
	job := baseJob.DeepCopy()
	modifyJob(job, loader, index)

	err := r.Client.Create(ctx, job)
	if err != nil {
		logger.Error(err, "failed to create job", "job", fmt.Sprintf("%s/%s", job.Namespace, job.Name))
		return err
	}
	logger.Info("create job successfully", "job", fmt.Sprintf("%s/%s", job.Namespace, job.Name))
	return nil
}

func (r *ImageLoaderReconciler) resetStatus(loader *tkexv1alpha1.ImageLoader,
	newStatus *tkexv1alpha1.ImageLoaderStatus) {
	newStatus.ObservedGeneration = loader.Generation
	newStatus.Desired = int32(len(loader.Spec.Images))
	newStatus.Active = 0
	newStatus.Active = 0
	newStatus.Completed = 0
	newStatus.Succeeded = 0
	newStatus.FailedStatuses = make([]*tkexv1alpha1.FailedStatus, 0)
	newStatus.CompletionTime = nil
}

func (r *ImageLoaderReconciler) updateStatus(ctx context.Context, loader *tkexv1alpha1.ImageLoader,
	newStatus *tkexv1alpha1.ImageLoaderStatus) error {
	r.resetStatus(loader, newStatus)

	jobList := &batchv1.JobList{}
	if err := r.List(ctx, jobList, client.MatchingLabels{
		ImageLoaderNameKey: loader.Name,
	}, client.InNamespace(loader.Namespace)); err != nil {
		return err
	}
	for i := range jobList.Items {
		job := &jobList.Items[i]
		if len(job.Status.Conditions) == 0 {
			// running
			newStatus.Active++
			continue
		}
		// succeed
		if job.Status.Conditions[0].Type == batchv1.JobComplete &&
			job.Status.Conditions[0].Status == corev1.ConditionTrue {
			newStatus.Completed++
			if job.Status.Succeeded == *job.Spec.Completions {
				logger.Info("job complete successfully", "job", fmt.Sprintf("%s/%s", job.Namespace, job.Name))
				newStatus.Succeeded++
			}
			continue
		}
		// failed
		if job.Status.Conditions[0].Type == batchv1.JobFailed &&
			job.Status.Conditions[0].Status == corev1.ConditionTrue {
			newStatus.Completed++
			newStatus.FailedStatuses = append(newStatus.FailedStatuses,
				&tkexv1alpha1.FailedStatus{
					JobName: job.Name,
					Name:    job.Spec.Template.Spec.Containers[0].Image,
					Message: job.Status.Conditions[0].Message,
				})
			r.Recorder.Eventf(loader, corev1.EventTypeWarning, "Failed", "preload image %s failed",
				job.Spec.Template.Spec.Containers[0].Image)
			logger.Error(fmt.Errorf(job.Status.Conditions[0].Message), "job failed", "job", fmt.Sprintf("%s/%s", job.Namespace,
				job.Name))
		}

	}
	if newStatus.Desired == newStatus.Completed {
		now := metav1.Now()
		newStatus.CompletionTime = &now
		if newStatus.Succeeded == newStatus.Desired {
			logger.Info("imageloader completed successfully")
			r.Recorder.Eventf(loader, corev1.EventTypeNormal, "Succeed", "All imageloader jobs succeeded")
		} else {
			logger.Info("imageloader completed with partial jobs succeed", "succeeded", newStatus.Succeeded, "desired",
				newStatus.Desired)
			r.Recorder.Eventf(loader, corev1.EventTypeWarning, "Completed", "Some imageloader jobs failed")
		}
		return nil
	}
	logger.Info("waiting for job done")
	return nil
}
