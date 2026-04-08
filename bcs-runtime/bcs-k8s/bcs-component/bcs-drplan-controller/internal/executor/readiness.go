/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executor

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	defaultWaitTimeout   = 5 * time.Minute
	waitPollInterval     = 5 * time.Second
	conditionStatusTrue  = "True"
	jobConditionComplete = "Complete"
	jobConditionFailed   = "Failed"
)

func parseActionTimeout(raw string) (time.Duration, error) {
	if raw == "" {
		return defaultWaitTimeout, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	if d <= 0 {
		return 0, fmt.Errorf("timeout must be positive")
	}
	return d, nil
}

func parseBindingCluster(binding string) (string, string, error) {
	parts := strings.SplitN(binding, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid binding cluster %q, expected namespace/name", binding)
	}
	return parts[0], parts[1], nil
}

func evaluateResourceReadiness(obj *unstructured.Unstructured) (bool, string, error) {
	switch strings.ToLower(obj.GetKind()) {
	case "deployment":
		return deploymentReady(obj)
	case "statefulset":
		return statefulSetReady(obj)
	case "daemonset":
		return daemonSetReady(obj)
	case "job":
		return jobReady(obj)
	default:
		return true, "resource exists", nil
	}
}

func deploymentReady(obj *unstructured.Unstructured) (bool, string, error) {
	desired, found, err := unstructured.NestedInt64(obj.Object, "spec", "replicas")
	if err != nil {
		return false, "", err
	}
	if !found {
		desired = 1
	}
	available, _, err := unstructured.NestedInt64(obj.Object, "status", "availableReplicas")
	if err != nil {
		return false, "", err
	}
	updated, _, err := unstructured.NestedInt64(obj.Object, "status", "updatedReplicas")
	if err != nil {
		return false, "", err
	}
	if available >= desired && updated >= desired {
		return true, "deployment ready", nil
	}
	return false, fmt.Sprintf("available=%d updated=%d desired=%d", available, updated, desired), nil
}

func statefulSetReady(obj *unstructured.Unstructured) (bool, string, error) {
	desired, found, err := unstructured.NestedInt64(obj.Object, "spec", "replicas")
	if err != nil {
		return false, "", err
	}
	if !found {
		desired = 1
	}
	ready, _, err := unstructured.NestedInt64(obj.Object, "status", "readyReplicas")
	if err != nil {
		return false, "", err
	}
	if ready >= desired {
		return true, "statefulset ready", nil
	}
	return false, fmt.Sprintf("ready=%d desired=%d", ready, desired), nil
}

func daemonSetReady(obj *unstructured.Unstructured) (bool, string, error) {
	desired, _, err := unstructured.NestedInt64(obj.Object, "status", "desiredNumberScheduled")
	if err != nil {
		return false, "", err
	}
	ready, _, err := unstructured.NestedInt64(obj.Object, "status", "numberReady")
	if err != nil {
		return false, "", err
	}
	if desired > 0 && ready >= desired {
		return true, "daemonset ready", nil
	}
	return false, fmt.Sprintf("ready=%d desired=%d", ready, desired), nil
}

func jobReady(obj *unstructured.Unstructured) (bool, string, error) {
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil {
		return false, "", err
	}
	if !found {
		return false, "job has no status.conditions yet", nil
	}
	for _, raw := range conditions {
		cond, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _, _ := unstructured.NestedString(cond, "type")
		condStatus, _, _ := unstructured.NestedString(cond, "status")
		if condStatus != conditionStatusTrue {
			continue
		}
		if condType == jobConditionComplete {
			return true, "job complete", nil
		}
		if condType == jobConditionFailed {
			return false, "", fmt.Errorf("job failed")
		}
	}
	return false, "job not completed yet", nil
}
