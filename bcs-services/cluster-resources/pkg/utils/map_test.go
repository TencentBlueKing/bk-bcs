/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"fmt"
	"testing"
)

func TestGetItems(t *testing.T) {
	deploySpec := map[string]interface{}{
		"replicas":             3,
		"revisionHistoryLimit": 10,
		"selector": map[string]interface{}{
			"matchLabels": map[string]interface{}{
				"app": "nginx",
			},
		},
		"strategy": map[string]interface{}{
			"rollingUpdate": map[string]interface{}{
				"maxSurge":       "25%",
				"maxUnavailable": "25%",
			},
			"type": "RollingUpdate",
		},
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"creationTimestamp": nil,
				"labels": map[string]interface{}{
					"app": "nginx",
				},
			},
			"spec": map[string]interface{}{
				"containers": []map[string]interface{}{
					{
						"image":           "nginx:latest",
						"imagePullPolicy": "IfNotPresent",
						"name":            "nginx",
						"ports": map[string]interface{}{
							"containerPort": 80,
							"protocol":      "TCP",
						},
						"resources": map[string]interface{}{},
					},
				},
				"dnsPolicy":                     "ClusterFirst",
				"restartPolicy":                 "Always",
				"schedulerName":                 "default-scheduler",
				"securityContext":               map[string]interface{}{},
				"terminationGracePeriodSeconds": 30,
			},
		},
	}

	// Success Case
	// depth 1，val type int
	if ret, _ := GetItems(deploySpec, []string{"replicas"}); ret != 3 {
		t.Errorf("Spec.replicas, Excepted: 3, Result: %s", ret)
	}
	// depth 2，val type map[string]interface{}
	ret, _ := GetItems(deploySpec, []string{"selector", "matchLabels"})
	if _, ok := ret.(map[string]interface{}); !ok {
		t.Errorf("Spec.selector.matchLabels not map[string]interface{} type")
	}
	// depth 2, val type string
	if ret, _ := GetItems(deploySpec, []string{"strategy", "type"}); ret != "RollingUpdate" {
		t.Errorf("Spec.strategy.type, Excepted: RollingUpdate, Result: %s", ret)
	}
	// depth 3, val type nil
	if ret, _ := GetItems(deploySpec, []string{"template", "metadata", "creationTimestamp"}); ret != nil {
		t.Errorf("Spec.template.metadata.creationTimestamp, Excepted: nil, Result: %s", ret)
	}
	// depth 3, val type string
	if ret, _ := GetItems(deploySpec, []string{"template", "spec", "restartPolicy"}); ret != "Always" {
		t.Errorf("Spec.template.spec.restartPolicy, Excepted: Always, Result: %s", ret)
	}

	// Error Case
	// not items error
	if ret, err := GetItems(deploySpec, []string{}); ret != nil || err != nil {
		fmt.Println(err)
		t.Errorf("Items is empty list, must raise error")
	}
	// not map[string]interface{} type error
	if ret, err := GetItems(deploySpec, []string{"replicas", "testKey"}); ret != nil || err != nil {
		fmt.Println(err)
		t.Errorf("Key spec.replicas, Value type not map[string]interface{}, must raise error")
	}
	if ret, err := GetItems(deploySpec, []string{"template", "spec", "containers", "image"}); ret != nil || err != nil {
		fmt.Println(err)
		t.Errorf("Key spec.template.spec.containers, Value type not map[string]interface{}, must raise error")
	}
	// key not exist
	if ret, err := GetItems(deploySpec, []string{"templateKey", "spec"}); ret != nil || err != nil {
		fmt.Println(err)
		t.Errorf("Key spec.templateKey not exists, must raise error")
	}
	if ret, err := GetItems(deploySpec, []string{"selector", "spec"}); ret != nil || err != nil {
		fmt.Println(err)
		t.Errorf("Key spec.selector.spec not exists, must raise error")
	}
}
