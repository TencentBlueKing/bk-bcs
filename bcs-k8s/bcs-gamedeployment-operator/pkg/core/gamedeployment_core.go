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
	"strconv"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
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
	*gdv1alpha1.GameDeployment
}

var _ Control = &commonControl{}

func (c *commonControl) IsInitializing() bool {
	return false
}

func (c *commonControl) SetRevisionTemplate(revisionSpec map[string]interface{}, template map[string]interface{}) {
	revisionSpec["template"] = template
	template["$patch"] = "replace"
}

func (c *commonControl) ApplyRevisionPatch(patched []byte) (*gdv1alpha1.GameDeployment, error) {
	restoredDeploy := &gdv1alpha1.GameDeployment{}
	if err := json.Unmarshal(patched, restoredDeploy); err != nil {
		return nil, err
	}
	return restoredDeploy, nil
}

func (c *commonControl) IsReadyToScale() bool {
	return true
}

func (c *commonControl) NewVersionedPods(currentGD, updateGD *gdv1alpha1.GameDeployment,
	currentRevision, updateRevision string,
	expectedCreations, expectedCurrentCreations int,
	availableIDs []string, availableIndex []int,
) ([]*v1.Pod, error) {
	var newPods []*v1.Pod
	if expectedCreations <= expectedCurrentCreations {
		newPods = c.newVersionedPods(currentGD, currentRevision, expectedCreations, &availableIDs, availableIndex)
	} else {
		newPods = c.newVersionedPods(currentGD, currentRevision, expectedCurrentCreations, &availableIDs, availableIndex)
		newPods = append(newPods,
			c.newVersionedPods(updateGD, updateRevision, expectedCreations-expectedCurrentCreations, &availableIDs, availableIndex)...)
	}
	return newPods, nil
}

func (c *commonControl) newVersionedPods(cs *gdv1alpha1.GameDeployment, revision string, replicas int,
	availableIDs *[]string, availableIndex []int) []*v1.Pod {
	var newPods []*v1.Pod
	for i := 0; i < replicas; i++ {
		if len(*availableIDs) == 0 {
			return newPods
		}
		id := (*availableIDs)[0]
		*availableIDs = (*availableIDs)[1:]

		index := availableIndex[0]
		availableIndex =availableIndex[1:]

		pod, _ := kubecontroller.GetPodFromTemplate(&cs.Spec.Template, cs, metav1.NewControllerRef(cs, gdutil.ControllerKind))
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		pod.Labels[appsv1.ControllerRevisionHashLabelKey] = revision

		pod.Name = fmt.Sprintf("%s-%s", cs.Name, id)
		pod.Namespace = cs.Namespace
		pod.Labels[gdv1alpha1.GameDeploymentInstanceID] = id
		pod.Annotations[gdv1alpha1.GameDeploymentIndexID] = strconv.Itoa(index)

		injectDeploymentPodIndexToEnv(pod, strconv.Itoa(index))
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
	//TODO (by bryanhe) consider some pods maybe crashed or status changed, then the pods order to be PreDeleteHook maybe
	// change, maybe we should use a simple alphabetical sort
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

func (c *commonControl) ValidateGameDeploymentUpdate(oldGD, newGD *gdv1alpha1.GameDeployment) error {
	if newGD.Spec.UpdateStrategy.Type != gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType {
		return nil
	}

	oldTempJSON, _ := json.Marshal(oldGD.Spec.Template.Spec)
	newTempJSON, _ := json.Marshal(newGD.Spec.Template.Spec)
	patches, err := jsonpatch.CreatePatch(oldTempJSON, newTempJSON)
	if err != nil {
		return fmt.Errorf("failed calculate patches between old/new template spec")
	}

	for _, p := range patches {
		if p.Operation != "replace" || !inPlaceUpdateTemplateSpecPatchRexp.MatchString(p.Path) {
			return fmt.Errorf("only allowed to update images in spec for %s, but found %s %s",
				gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType, p.Operation, p.Path)
		}
	}
	return nil
}

func injectDeploymentPodIndexToEnv(pod *v1.Pod, index string) {
	if pod == nil {
		return
	}

	for i := range pod.Spec.Containers {
		pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env,
			v1.EnvVar{
				Name:      gdv1alpha1.GameDeploymentIndexEnv,
				Value:     index,
			})
	}
	return
}
