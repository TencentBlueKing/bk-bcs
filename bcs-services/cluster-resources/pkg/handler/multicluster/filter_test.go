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

package multicluster

import (
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/storage"
)

func newPodResource(name, phase string, containerState map[string]interface{}) *storage.Resource {
	return &storage.Resource{
		ClusterID:    "BCS-K8S-00000",
		ResourceType: resCsts.Po,
		Data: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       resCsts.Po,
			"metadata": map[string]interface{}{
				"name":              name,
				"namespace":         "default",
				"uid":               name,
				"creationTimestamp": "2026-10-27T00:00:00Z",
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{"name": "main", "image": "nginx"},
				},
			},
			"status": map[string]interface{}{
				"phase": phase,
				"containerStatuses": []interface{}{
					map[string]interface{}{
						"name":         "main",
						"ready":        false,
						"restartCount": int64(0),
						"state":        containerState,
					},
				},
			},
		},
	}
}

func podNames(resources []*storage.Resource) []string {
	names := []string{}
	for _, r := range resources {
		names = append(names, r.Data["metadata"].(map[string]interface{})["name"].(string))
	}
	return names
}

func TestStatusFilterPodUnknownGroup(t *testing.T) {
	resources := []*storage.Resource{
		newPodResource("container-status-unknown", "Failed", map[string]interface{}{
			"terminated": map[string]interface{}{"reason": "ContainerStatusUnknown", "exitCode": int64(137)},
		}),
		newPodResource("phase-unknown", "Unknown", map[string]interface{}{}),
		newPodResource("crash", "Running", map[string]interface{}{
			"waiting": map[string]interface{}{"reason": "CrashLoopBackOff"},
		}),
	}

	f := &QueryFilter{Status: []string{"Unknown"}}
	assert.ElementsMatch(t, []string{"container-status-unknown", "phase-unknown"}, podNames(f.StatusFilter(resources)))

	f = &QueryFilter{Status: []string{"ContainerStatusUnknown"}}
	assert.ElementsMatch(t, []string{"container-status-unknown"}, podNames(f.StatusFilter(resources)))

	f = &QueryFilter{Status: []string{"CrashLoopBackOff"}}
	assert.ElementsMatch(t, []string{"crash"}, podNames(f.StatusFilter(resources)))
}
