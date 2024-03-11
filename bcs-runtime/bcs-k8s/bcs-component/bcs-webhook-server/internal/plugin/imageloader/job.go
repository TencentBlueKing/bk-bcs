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

package imageloader

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
)

const (
	// DOTO combine all annotations and labels together
	// workloadNameAnno is the label key to identify the corresponding workload
	workloadNameAnno = "workloadName"
	// workloadInsNameLabel is the label key to identify the name of the workload instance
	workloadInsNameLabel = "workloadInsName"
	// workloadInsNamespaceLabel is the label key to identify the namespace of the workload instance
	workloadInsNamespaceLabel = "workloadInsNamespace"
)

var (
	// backoff limit of a job
	backoffLimit int32 = 1
	// active time seconds of a job
	activeTimeSecondsOfJob int64 = 900
	// delete pod after job is deleting
	jobDeletePropagationPolicy = metav1.DeletePropagationBackground
)

func (i *imageLoader) addJob(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("cound't get key for object %+v: %v", obj, err))
		return
	}
	i.queue.Add(key)

}

func (i *imageLoader) updateJob(old, cur interface{}) {
	oldJob, ok := old.(*batchv1.Job)
	if !ok {
		blog.Errorf("old job(%v) type assertion failed", old)
		return
	}
	if oldJob.Namespace != pluginName {
		return
	}

	curJob, ok := cur.(*batchv1.Job)
	if !ok {
		blog.Errorf("old job(%v) type assertion failed", cur)
		return
	}
	if curJob.Namespace != pluginName {
		return
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cur)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("cound't get key for object %+v: %v", cur, err))
		return
	}

	if !reflect.DeepEqual(oldJob, curJob) {
		i.queue.Add(key)
	}
}

func (i *imageLoader) sync(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	if namespace != pluginName {
		return nil
	}
	job, err := i.jobLister.Jobs(namespace).Get(name)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if job.DeletionTimestamp != nil {
		return nil
	}

	blog.V(3).Infof("start sync job %s/%s", namespace, name)
	// do nothing if job is not done, wait next status change
	event, done := i.isJobDone(job)
	if !done {
		return nil
	}

	// get workload to trigger the update
	workloadName, ok := job.Annotations[workloadNameAnno]
	if !ok {
		blog.Errorf("job %s has no workload name label, it should not in the namespace", job.Name)
		return fmt.Errorf("job %s has no workload name label, it should not in the namespace", job.Name)
	}
	workload, ok := i.workloads[workloadName]
	if !ok {
		blog.Errorf("job %s belongs to workload %s which is not supported in imageloader",
			job.Name, workloadName)
		return fmt.Errorf("job %s belongs to workload %s which is not supported in imageloader",
			job.Name, workloadName)
	}
	err = workload.JobDoneHook(
		job.Labels[workloadInsNamespaceLabel], job.Labels[workloadInsNameLabel], event)
	if err != nil {
		blog.Errorf("finish job %s of workload %s-%s-%s failed: %v",
			job.Name, workloadName, job.Labels[workloadInsNamespaceLabel], job.Labels[workloadInsNameLabel], err)
		return fmt.Errorf("finish job %s of workload %s-%s-%s failed: %v",
			job.Name, workloadName, job.Labels[workloadInsNamespaceLabel], job.Labels[workloadInsNameLabel], err)
	}

	// attach event to workload instance
	// do not finish the job if failed
	// DOTO retry on conflict
	// DOTO create or update
	if event != nil {
		_, err := i.k8sClient.CoreV1().Events(event.Namespace).Create(context.Background(), event, metav1.CreateOptions{})
		if err != nil {
			blog.Errorf("attach event(%v) to workload instance failed: %v", event, err)
			return fmt.Errorf("attach event(%v) to workload instance failed: %v", event, err)
		}
	}

	// delete the job
	return i.deleteJob(job.Namespace, job.Name)
}

func (i *imageLoader) isJobDone(job *batchv1.Job) (*corev1.Event, bool) {
	// job.Status.CompletionTime do not be set when reachs active deadline
	// job.Status.Failed do not be set when image pull failed
	if job.Status.Succeeded+job.Status.Failed < *job.Spec.Completions &&
		(len(job.Status.Conditions) == 0 || (job.Status.Conditions[0].Type != batchv1.JobFailed)) {
		// wait timeout
		// DOTO make sure CompletionTime will be setup when timeout
		return nil, false
	}

	if job.Status.Active == 0 && job.Status.Failed == 0 &&
		job.Status.Succeeded == *job.Spec.Completions {
		// all pods of the job success, no need to check node
		// reduce check time when job complete before node report image
		collectJobDuration(job.Name, actionRun, statusSuccess, time.Since(job.CreationTimestamp.Time))
		collectJobStatus(job.Name, actionRun, statusSuccess)
		return nil, true
	}

	// get nodes of the job and check the images
	nodes := strings.Split(job.Annotations[jobOnNodeAnno], ",")
	if len(nodes) == 0 {
		blog.Errorf("job %s had no nodes on, nodes anno: %s", job.Name, job.Annotations[jobOnNodeAnno])
		return nil, false
	}

	// get images of the job
	images := make([]string, len(job.Spec.Template.Spec.Containers))
	imageCount := 0
	for _, c := range job.Spec.Template.Spec.Containers {
		images[imageCount] = c.Image
		imageCount++
	}

	// check if all images is on all nodes
	for _, nodeName := range nodes {
		node, err := i.nodeLister.Get(nodeName)
		if err != nil {
			// node may not exist, ignore it
			// DOTO maybe network problem
			blog.Errorf("get node %s failed of job %s: %v, ignore the failure", nodeName, job.Name, err)
			continue
		}
		// build image index of node
		imagesOnNode := make(map[string]struct{})
		for _, i := range node.Status.Images {
			for _, n := range i.Names {
				imagesOnNode[n] = struct{}{}
			}
		}
		for _, i := range images {
			if _, ok := imagesOnNode[i]; !ok {
				errMsg := fmt.Sprintf("image %s is not on node %s", i, nodeName)
				nowTime := metav1.Now()
				event := &corev1.Event{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: pluginName + "-",
					},
					Reason:         "ImageNotFound",
					Message:        errMsg,
					Type:           corev1.EventTypeWarning,
					FirstTimestamp: nowTime,
					LastTimestamp:  nowTime,
				}
				blog.Error(errMsg)
				collectJobDuration(job.Name, actionRun, statusFailure, time.Since(job.CreationTimestamp.Time))
				collectJobStatus(job.Name, actionRun, statusFailure)
				return event, true
			}
		}
	}

	collectJobDuration(job.Name, actionRun, statusSuccess, time.Since(job.CreationTimestamp.Time))
	collectJobStatus(job.Name, actionRun, statusSuccess)
	return nil, true
}

func (i *imageLoader) createJob(job *batchv1.Job) error {
	start := time.Now()
	// create job with retry
	createFunc := func() error {
		_, createErr := i.k8sClient.BatchV1().Jobs(pluginName).Create(context.Background(), job, metav1.CreateOptions{})
		return createErr
	}
	isRetryable := func(err error) bool { // nolint
		return errors.IsAlreadyExists(err)
	}
	err := retry.OnError(retry.DefaultBackoff, isRetryable, createFunc)
	if err != nil {
		blog.Errorf("create job(%v) failed: %v", job, err)
		collectJobDuration(job.Name, actionCreate, statusFailure, time.Since(start))
		collectJobStatus(job.Name, actionCreate, statusFailure)
	} else {
		collectJobDuration(job.Name, actionCreate, statusSuccess, time.Since(start))
		collectJobStatus(job.Name, actionCreate, statusSuccess)
	}
	return err
}

func (i *imageLoader) deleteJob(namespace, name string) error {
	start := time.Now()
	err := i.k8sClient.BatchV1().Jobs(namespace).Delete(
		context.Background(),
		name,
		metav1.DeleteOptions{
			PropagationPolicy: &jobDeletePropagationPolicy,
		})
	if err != nil {
		blog.Errorf("delete job %s/%s failed: %v", namespace, name, err)
		if !errors.IsNotFound(err) {
			collectJobDuration(name, actionDelete, statusFailure, time.Since(start))
			collectJobStatus(name, actionDelete, statusFailure)
		}
	} else {
		collectJobDuration(name, actionDelete, statusSuccess, time.Since(start))
		collectJobStatus(name, actionDelete, statusSuccess)
	}
	return err
}

func (i *imageLoader) createJobIfNeed(job *batchv1.Job) error {
	currentJob, err := i.k8sClient.BatchV1().Jobs(pluginName).Get(context.Background(), job.Name, metav1.GetOptions{})
	// create job if not exist
	if errors.IsNotFound(err) {
		return i.createJob(job)
	}
	if err != nil {
		return err
	}
	// recreate job if failed
	failed := false
	for _, condition := range currentJob.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			failed = true
			break
		}
	}
	// recreate job if different
	diff := false
	if !reflect.DeepEqual(currentJob.Spec, job.Spec) {
		diff = true
	}
	if failed || diff {
		err = i.deleteJob(job.Namespace, job.Name)
		if err != nil {
			return err
		}
		err = i.createJob(job)
		if err != nil {
			return err
		}
	}
	return err
}

func newJob(cs []corev1.Container) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   pluginName,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: batchv1.JobSpec{
			// ensure every pod of the job has been scheduled before backofflimit
			BackoffLimit: &backoffLimit,
			// job timeout
			ActiveDeadlineSeconds: &activeTimeSecondsOfJob,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers:    cs,
					RestartPolicy: "Never",
				},
			},
		},
	}
	return job
}
