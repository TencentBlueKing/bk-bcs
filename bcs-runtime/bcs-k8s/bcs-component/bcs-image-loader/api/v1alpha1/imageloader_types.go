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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImageLoaderSpec defines the desired state of ImageLoader
type ImageLoaderSpec struct {
	// Images is the image list to be pulled by the job
	Images []string `json:"images"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling the image.
	// If specified, these secrets will be passed to individual puller implementations for them to use.  For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// ImagePullPolicy is the image pull policy for the job
	// +kubebuilder:default=Always
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// NodeSelector is a query over nodes that should match the job.
	// nil to match all nodes.
	// +optional
	NodeSelector *ImageLoaderNodeSelector `json:"nodeSelector,omitempty"`

	// PodSelector is a query over pods that should pull image on nodes of these pods.
	// Mutually exclusive with Selector.
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`

	// JobTimeout is the timeout for the job
	// defaults to 10 minutes
	// +kubebuilder:default=600
	JobTimeout int64 `json:"jobTimeout"`

	// BackoffLimit is the backoff limit for the job
	// defaults to 3
	// +kubebuilder:default=3
	BackoffLimit int32 `json:"backoffLimit"`
}

// ImageLoaderNodeSelector is a selector over nodes
type ImageLoaderNodeSelector struct {
	// Names specify a set of nodes to execute the job.
	// +optional
	Names []string `json:"names,omitempty"`

	// LabelSelector is a label query over nodes that should match the job.
	// +optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// ImageLoaderStatus defines the observed state of ImageLoader
type ImageLoaderStatus struct {
	// ObservedGeneration is the most recent generation observed for this ImageLoader. It corresponds to the
	// ImageLoader's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents time when the job was acknowledged by the job controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the all the image pull job was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Revision is the revision of the imageloader
	Revision string `json:"revision"`

	// Desired is the desired number of ImagePullJobs, this is typically equal to the number of len(spec.Images).
	Desired int32 `json:"desired"`

	// Active is the number of running ImagePullJobs which are acknowledged by the imagepulljob controller.
	// +optional
	Active int32 `json:"active"`

	// Completed is the number of ImagePullJobs which are finished
	// +optional
	Completed int32 `json:"completed"`

	// Succeeded is the number of image pull job which are finished and status.Succeeded==status.Desired.
	// +optional
	Succeeded int32 `json:"succeeded"`

	// FailedStatuses is the status of ImagePullJob which has the failed nodes(status.Failed>0) .
	// +optional
	FailedStatuses []*FailedStatus `json:"failedStatuses,omitempty"`
}

// FailedStatus the state of ImagePullJob which has the failed nodes(status.Failed>0)
type FailedStatus struct {
	// JobName is the name of ImagePullJob which has the failed nodes(status.Failed>0)
	// +optional
	JobName string `json:"imagePullJob,omitempty"`

	// Name of the image
	// +optional
	Name string `json:"name,omitempty"`

	// Message is the text prompt for job running status.
	// +optional
	Message string `json:"message,omitempty"`
}

// KindImageLoader is the Schema for the imageloaders API
const KindImageLoader = "ImageLoader"

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=.status.desired,name=Desired,type=integer,description=The desired number of jobs.
// +kubebuilder:printcolumn:JSONPath=.status.active,name=Active,type=integer,description=The active number of jobs.
// +kubebuilder:printcolumn:JSONPath=.status.completed,name=Completed,type=integer,description=The complete number of jobs.
// +kubebuilder:printcolumn:JSONPath=.status.succeeded,name=Succeeded,type=integer,description=The succeeded number of jobs.

// ImageLoader is the Schema for the imageloaders API
type ImageLoader struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageLoaderSpec   `json:"spec,omitempty"`
	Status ImageLoaderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ImageLoaderList contains a list of ImageLoader
type ImageLoaderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageLoader `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageLoader{}, &ImageLoaderList{})
}
