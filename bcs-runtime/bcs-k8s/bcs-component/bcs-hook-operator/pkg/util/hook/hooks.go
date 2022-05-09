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

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/typed/tkex/v1alpha1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ValidateMetrics(metrics []v1alpha1.Metric) error {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics specified")
	}
	duplicateNames := make(map[string]bool)
	for i, metric := range metrics {
		if _, ok := duplicateNames[metric.Name]; ok {
			return fmt.Errorf("metrics[%d]: duplicate name '%s", i, metric.Name)
		}
		duplicateNames[metric.Name] = true
		if err := ValidateMetric(metric); err != nil {
			return fmt.Errorf("metrics[%d]: %v", i, err)
		}
	}
	return nil
}

// ValidateMetric validates a single metric spec
func ValidateMetric(metric v1alpha1.Metric) error {
	if metric.Count > 0 {
		if (metric.Count < metric.FailureLimit) || (metric.Count < metric.SuccessfulLimit) {
			return fmt.Errorf("count must be >= failureLimit && >= successfulLimit")
		}
		if metric.Count < metric.InconclusiveLimit {
			return fmt.Errorf("count must be >= inconclusiveLimit")
		}
	}
	if metric.Count > 1 && metric.Interval == "" {
		return fmt.Errorf("interval must be specified when count > 1")
	}
	if metric.Interval != "" {
		if _, err := metric.Interval.Duration(); err != nil {
			return fmt.Errorf("invalid interval string: %v", err)
		}
	}
	if metric.InitialDelay != "" {
		if _, err := metric.InitialDelay.Duration(); err != nil {
			return fmt.Errorf("invalid startDelay string: %v", err)
		}
	}

	if metric.FailureLimit < 0 {
		return fmt.Errorf("failureLimit must be >= 0")
	}

	if metric.SuccessfulLimit < 0 {
		return fmt.Errorf("successLimit must be >= 0")
	}

	if metric.InconclusiveLimit < 0 {
		return fmt.Errorf("inconclusiveLimit must be >= 0")
	}
	if metric.ConsecutiveErrorLimit != nil && *metric.ConsecutiveErrorLimit < 0 {
		return fmt.Errorf("consecutiveErrorLimit must be >= 0")
	}
	if metric.ConsecutiveSuccessfulLimit != nil && *metric.ConsecutiveSuccessfulLimit < 1 {
		return fmt.Errorf("consecutiveSuccessfulLimit must be >= 1")
	}
	numProviders := 0

	if metric.Provider.Web != nil {
		numProviders++
	}
	if metric.Provider.Prometheus != nil {
		numProviders++
	}
	if metric.Provider.Kubernetes != nil {
		numProviders++
	}

	if numProviders == 0 {
		return fmt.Errorf("no provider specified")
	}
	if numProviders > 1 {
		return fmt.Errorf("multiple providers specified")
	}
	return nil
}

func NewHookRunFromTemplate(template *v1alpha1.HookTemplate, args []v1alpha1.Argument, name, generateName, namespace string) (*v1alpha1.HookRun, error) {
	newArgs, err := MergeArgs(args, template.Spec.Args)
	if err != nil {
		return nil, err
	}
	ar := v1alpha1.HookRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:         name,
			GenerateName: generateName,
			Namespace:    namespace,
		},
		Spec: v1alpha1.HookRunSpec{
			Metrics: template.Spec.Metrics,
			Args:    newArgs,
		},
	}
	return &ar, nil
}

func MergeArgs(incomingArgs, templateArgs []v1alpha1.Argument) ([]v1alpha1.Argument, error) {
	newArgs := append(templateArgs[:0:0], templateArgs...)
	for _, arg := range incomingArgs {
		i := findArg(arg.Name, newArgs)
		if i >= 0 && arg.Value != nil {
			newArgs[i].Value = arg.Value
		}
	}
	for _, arg := range newArgs {
		if arg.Value == nil {
			return nil, fmt.Errorf("args.%s was not resolved", arg.Name)
		}
	}
	return newArgs, nil
}

func findArg(name string, args []v1alpha1.Argument) int {
	for i, arg := range args {
		if arg.Name == name {
			return i
		}
	}
	return -1
}

func CreateWithCollisionCounter(hookRunIf tkexclientset.HookRunInterface, run v1alpha1.HookRun) (*v1alpha1.HookRun, error) {
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

func IsSemanticallyEqual(left, right v1alpha1.HookRunSpec) bool {
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
