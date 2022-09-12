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

package core

import (
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/inplaceupdate"

	v1 "k8s.io/api/core/v1"
)

// Control xxx
type Control interface {
	// IsInitializing xxx
	// common
	IsInitializing() bool
	SetRevisionTemplate(revisionSpec map[string]interface{}, template map[string]interface{})
	ApplyRevisionPatch(patched []byte) (*gdv1alpha1.GameDeployment, error)

	// IsReadyToScale xxx
	// scale
	IsReadyToScale() bool
	NewVersionedPods(currentCS, updateCS *gdv1alpha1.GameDeployment,
		currentRevision, updateRevision string,
		expectedCreations, expectedCurrentCreations int,
		availableIDs []string, availableIndex []int,
	) ([]*v1.Pod, error)

	// IsPodUpdatePaused xxx
	// update
	IsPodUpdatePaused(pod *v1.Pod) bool
	IsPodUpdateReady(pod *v1.Pod, minReadySeconds int32) bool
	GetPodsSortFunc(pods []*v1.Pod, waitUpdateIndexes []int) func(i, j int) bool
	GetUpdateOptions() *inplaceupdate.UpdateOptions

	// ValidateGameDeploymentUpdate xxx
	// validation
	ValidateGameDeploymentUpdate(oldCS, newCS *gdv1alpha1.GameDeployment) error
}

// New xxx
func New(gd *gdv1alpha1.GameDeployment) Control {
	return &commonControl{GameDeployment: gd}
}
