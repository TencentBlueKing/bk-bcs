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
	"encoding/json"
	"fmt"
	"regexp"

	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/update/inplaceupdate"
	"github.com/mattbaird/jsonpatch"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubecontroller "k8s.io/kubernetes/pkg/controller"
)

var (
	inPlaceUpdateTemplateSpecPatchRexp = regexp.MustCompile("^/containers/([0-9]+)/image$")
)

type commonControl struct {
	*tkexv1alpha1.GameDeployment
}

var _ Control = &commonControl{}

func (c *commonControl) IsInitializing() bool {
	return false
}

func (c *commonControl) SetRevisionTemplate(revisionSpec map[string]interface{}, template map[string]interface{}) {
	revisionSpec["template"] = template
	template["$patch"] = "replace"
}

func (c *commonControl) ApplyRevisionPatch(patched []byte) (*tkexv1alpha1.GameDeployment, error) {
	restoredDeploy := &tkexv1alpha1.GameDeployment{}
	if err := json.Unmarshal(patched, restoredDeploy); err != nil {
		return nil, err
	}
	return restoredDeploy, nil
}

func (c *commonControl) IsReadyToScale() bool {
	return true
}

func (c *commonControl) NewVersionedPods(currentGD, updateGD *tkexv1alpha1.GameDeployment,
	currentRevision, updateRevision string,
	expectedCreations, expectedCurrentCreations int,
	availableIDs []string,
) ([]*v1.Pod, error) {
	var newPods []*v1.Pod
	if expectedCreations <= expectedCurrentCreations {
		newPods = c.newVersionedPods(currentGD, currentRevision, expectedCreations, &availableIDs)
	} else {
		newPods = c.newVersionedPods(currentGD, currentRevision, expectedCurrentCreations, &availableIDs)
		newPods = append(newPods, c.newVersionedPods(updateGD, updateRevision, expectedCreations-expectedCurrentCreations, &availableIDs)...)
	}
	return newPods, nil
}

func (c *commonControl) newVersionedPods(cs *tkexv1alpha1.GameDeployment, revision string, replicas int, availableIDs *[]string) []*v1.Pod {
	var newPods []*v1.Pod
	for i := 0; i < replicas; i++ {
		if len(*availableIDs) == 0 {
			return newPods
		}
		id := (*availableIDs)[0]
		*availableIDs = (*availableIDs)[1:]

		pod, _ := kubecontroller.GetPodFromTemplate(&cs.Spec.Template, cs, metav1.NewControllerRef(cs, gdutil.ControllerKind))
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		pod.Labels[appsv1.ControllerRevisionHashLabelKey] = revision

		pod.Name = fmt.Sprintf("%s-%s", cs.Name, id)
		pod.Namespace = cs.Namespace
		pod.Labels[tkexv1alpha1.GameDeploymentInstanceID] = id

		inplaceupdate.InjectReadinessGate(pod)

		newPods = append(newPods, pod)
	}
	return newPods
}

func (c *commonControl) IsPodUpdatePaused(pod *v1.Pod) bool {
	return false
}

func (c *commonControl) IsPodUpdateReady(pod *v1.Pod, minReadySeconds int32) bool {
	if !gdutil.IsRunningAndAvailable(pod, minReadySeconds) {
		return false
	}
	condition := inplaceupdate.GetCondition(pod)
	if condition != nil && condition.Status != v1.ConditionTrue {
		return false
	}
	return true
}

func (c *commonControl) GetPodsSortFunc(pods []*v1.Pod, waitUpdateIndexes []int) func(i, j int) bool {
	// not-ready < ready, unscheduled < scheduled, and pending < running
	return func(i, j int) bool {
		return kubecontroller.ActivePods(pods).Less(waitUpdateIndexes[i], waitUpdateIndexes[j])
	}
}

func (c *commonControl) GetUpdateOptions() *inplaceupdate.UpdateOptions {
	opts := &inplaceupdate.UpdateOptions{}
	if c.Spec.UpdateStrategy.InPlaceUpdateStrategy != nil {
		opts.GracePeriodSeconds = c.Spec.UpdateStrategy.InPlaceUpdateStrategy.GracePeriodSeconds
	}
	return opts
}

func (c *commonControl) ValidateGameDeploymentUpdate(oldCS, newCS *tkexv1alpha1.GameDeployment) error {
	if newCS.Spec.UpdateStrategy.Type != tkexv1alpha1.InPlaceGameDeploymentUpdateStrategyType {
		return nil
	}

	oldTempJSON, _ := json.Marshal(oldCS.Spec.Template.Spec)
	newTempJSON, _ := json.Marshal(newCS.Spec.Template.Spec)
	patches, err := jsonpatch.CreatePatch(oldTempJSON, newTempJSON)
	if err != nil {
		return fmt.Errorf("failed calculate patches between old/new template spec")
	}

	for _, p := range patches {
		if p.Operation != "replace" || !inPlaceUpdateTemplateSpecPatchRexp.MatchString(p.Path) {
			return fmt.Errorf("only allowed to update images in spec for %s, but found %s %s",
				tkexv1alpha1.InPlaceGameDeploymentUpdateStrategyType, p.Operation, p.Path)
		}
	}
	return nil
}
