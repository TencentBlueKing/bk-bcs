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

package hook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/typed/tkex/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultArgs = [...]string{"PodName", "PodNamespace", "PodIP", "PodContainer", "HostIP"}

const (
	WorkloadRevisionUniqueLabel string = "workload-revision"
	HookRunTypeLabel                   = "hookrun-type"
	HookRunTypeCanaryStepLabel         = "canary-step"
	HookRunTypePreDeleteLabel          = "pre-delete-step"
	HookRunTypePreInplaceLabel         = "pre-inplace-step"
	HookRunTypePostInplaceLabel        = "post-inplace-step"
	HookRunCanaryStepIndexLabel        = "canary-step-index"
	PodControllerRevision              = "pod-controller-revision"
	PodInstanceID                      = "instance-id"
)

const (
	CancelHookRun = `{
		"spec": {
			"terminate": true
		}
	}`
)

func StepLabels(index int32, revision string) map[string]string {
	indexStr := strconv.Itoa(int(index))
	return map[string]string{
		WorkloadRevisionUniqueLabel: revision,
		HookRunTypeLabel:            HookRunTypeCanaryStepLabel,
		HookRunCanaryStepIndexLabel: indexStr,
	}
}

func NewHookRunFromTemplate(template *hookv1alpha1.HookTemplate, args []hookv1alpha1.Argument, name, generateName,
	namespace string) (*hookv1alpha1.HookRun, error) {
	newArgs, err := MergeArgs(args, template.Spec.Args)
	if err != nil {
		return nil, err
	}
	ar := hookv1alpha1.HookRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:         name,
			GenerateName: generateName,
			Namespace:    namespace,
		},
		Spec: hookv1alpha1.HookRunSpec{
			Metrics: template.Spec.Metrics,
			Args:    newArgs,
		},
	}
	return &ar, nil
}

func MergeArgs(incomingArgs, templateArgs []hookv1alpha1.Argument) ([]hookv1alpha1.Argument, error) {
	newArgs := append(templateArgs[:0:0], templateArgs...)
	for _, arg := range incomingArgs {
		i := findArg(arg.Name, newArgs)
		if i >= 0 && arg.Value != nil {
			newArgs[i].Value = arg.Value
		} else if findDefaultArgs(arg.Name) {
			newArgs = append(newArgs, arg)
		}
	}
	for _, arg := range newArgs {
		if arg.Value == nil {
			return nil, fmt.Errorf("args.%s was not resolved", arg.Name)
		}
	}
	return newArgs, nil
}

func findArg(name string, args []hookv1alpha1.Argument) int {
	for i, arg := range args {
		if arg.Name == name {
			return i
		}
	}
	return -1
}

func findDefaultArgs(name string) bool {
	for _, argName := range defaultArgs {
		if argName == name {
			return true
		}
	}
	return false
}

func CreateWithCollisionCounter(hookRunIf tkexclientset.HookRunInterface, run hookv1alpha1.HookRun) (*hookv1alpha1.HookRun, error) {
	newControllerRef := metav1.GetControllerOf(&run)
	if newControllerRef == nil {
		return nil, errors.New("Supplied run does not have an owner reference")
	}
	collisionCount := 1
	baseName := run.Name
	for {
		createdRun, err := hookRunIf.Create(context.TODO(), &run, metav1.CreateOptions{})
		if err == nil {
			return createdRun, nil
		}
		if !k8serrors.IsAlreadyExists(err) {
			return nil, err
		}
		// TODO(jessesuen): switch from Get to List so that there's no guessing about which collision counter to use.
		existingRun, err := hookRunIf.Get(context.TODO(), run.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		existingEqual := IsSemanticallyEqual(run.Spec, existingRun.Spec)
		controllerRef := metav1.GetControllerOf(existingRun)
		controllerUIDEqual := controllerRef != nil && controllerRef.UID == newControllerRef.UID
		if !existingRun.Status.Phase.Completed() && existingEqual && controllerUIDEqual {
			// If we get here, the existing run has been determined to be our hooks run and we
			// likely reconciled the rollout with a stale cache (quite common).
			return existingRun, nil
		}
		run.Name = fmt.Sprintf("%s.%d", baseName, collisionCount)
		collisionCount++
	}
}

func IsSemanticallyEqual(left, right hookv1alpha1.HookRunSpec) bool {
	leftBytes, err := json.Marshal(left)
	if err != nil {
		panic(err)
	}
	rightBytes, err := json.Marshal(right)
	if err != nil {
		panic(err)
	}
	return string(leftBytes) == string(rightBytes)
}
