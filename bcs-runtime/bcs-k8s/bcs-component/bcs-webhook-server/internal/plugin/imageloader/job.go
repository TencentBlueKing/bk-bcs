/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package imageloader

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TODO combine all annotations and labels together
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
	activeTimeSecondsOfJob int64 = 300
	// delete pod when job is deleted
	jobDeletePropagationPolicy = metav1.DeletePropagationForeground
)

func (i *imageLoader) addJob(o interface{}) {
	// get the job
	job, ok := o.(*batchv1.Job)
	if !ok {
		blog.Errorf("job(%v) type assertion failed", o)
		return
	}
	go i.jobChanged(job)
	return
}

func (i *imageLoader) updateJob(o, n interface{}) {
	// get the job
	job, ok := n.(*batchv1.Job)
	if !ok {
		blog.Errorf("job(%v) type assertion failed", n)
		return
	}
	go i.jobChanged(job)
	return
}

func (i *imageLoader) jobChanged(job *batchv1.Job) {
	blog.V(3).Infof("job status changed: %v", job)
	// do nothing if job is not done, wait next status change
	event, done := i.isJobDone(job)
	if !done {
		return
	}

	// get workload to trigger the update
	workloadName, ok := job.Annotations[workloadNameAnno]
	if !ok {
		blog.Errorf("job %s has no workload name label, it should not in the namespace", job.Name)
		return
	}
	workload, ok := i.workloads[workloadName]
	if !ok {
		blog.Errorf("job %s belongs to workload %s which is not supported in imageloader",
			job.Name, workloadName)
		return
	}
	err := workload.JobDoneHook(
		job.Labels[workloadInsNamespaceLabel], job.Labels[workloadInsNameLabel], event)
	if err != nil {
		blog.Errorf("finish job %s of workload %s-%s-%s failed: %v",
			job.Name, workloadName, job.Labels[workloadInsNamespaceLabel], job.Labels[workloadInsNameLabel], err)
		return
	}

	// attach event to workload instance
	// do not finish the job if failed
	// TODO retry on conflict
	// TODO create or update
	if event != nil {
		_, err := i.k8sClient.CoreV1().Events(event.Namespace).Create(context.Background(), event, metav1.CreateOptions{})
		if err != nil {
			blog.Errorf("attach event(%v) to workload instance failed: %v", event, err)
			return
		}
	}

	// delete the job
	i.deleteJob(job)
}

func (i *imageLoader) isJobDone(job *batchv1.Job) (*corev1.Event, bool) {
	// job.Status.CompletionTime do not be set when reachs active deadline
	// job.Status.Failed do not be set when image pull failed
	if job.Status.Succeeded+job.Status.Failed < *job.Spec.Completions &&
		(len(job.Status.Conditions) == 0 || (job.Status.Conditions[0].Type != batchv1.JobFailed)) {
		// wait timeout
		// TODO make sure CompletionTime will be setup when timeout
		return nil, false
	}

	if job.Status.Active == 0 && job.Status.Failed == 0 &&
		job.Status.Succeeded == *job.Spec.Completions {
		// all pods of the job success, no need to check node
		// reduce check time when job complete before node report image
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
			// TODO maybe network problem
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
				nowTime := metav1.Time{Time: time.Now()}
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
				return event, true
			}
		}
	}

	return nil, true
}

func (i *imageLoader) createJob(job *batchv1.Job) error {
	// TODO retry on conflict
	_, err := i.k8sClient.BatchV1().Jobs(pluginName).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("create job(%v) failed: %v", job, err)
	}
	return err
}

func (i *imageLoader) deleteJob(job *batchv1.Job) error {
	err := i.k8sClient.BatchV1().Jobs(job.Namespace).Delete(
		context.Background(),
		job.Name,
		metav1.DeleteOptions{
			PropagationPolicy: &jobDeletePropagationPolicy,
		})
	if err != nil {
		blog.Errorf("delete job %s failed: %v", job.Name, err)
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
