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

package hotpatchupdate

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// HotPatchUpdateStateKey records the state of hotpatch-update.
	// The value of annotation is HotPatchUpdateState.
	HotPatchUpdateStateKey string = "hotpatch-update-state"

	// PodHotpatchContainerKey hot-patch annotation
	PodHotpatchContainerKey = "io.kubernetes.hotpatch.container"
)

// HotPatchUpdateState records latest hotpatch-update state, including old statuses of containers.
// nolint
type HotPatchUpdateState struct {
	// Revision is the updated revision hash.
	Revision string `json:"revision"`

	// UpdateTimestamp is the time when the hot-patch update happens.
	UpdateTimestamp metav1.Time `json:"updateTimestamp"`

	// LastContainerStatuses records the before-hot-patch-update container statuses. It is a map from ContainerName
	// to HotPatchUpdateContainerStatus
	LastContainerStatuses map[string]HotPatchUpdateContainerStatus `json:"lastContainerStatuses"`
}

// HotPatchUpdateContainerStatus records the statuses of the container that are mainly used
// to determine whether the HotPatchUpdate is completed.
// nolint
type HotPatchUpdateContainerStatus struct {
	ImageID string `json:"imageID,omitempty"`
}
