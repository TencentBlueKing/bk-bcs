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
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
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

// mkSameNameRes 构造同名、不同命名空间/集群的资源，模拟多集群聚合场景
func mkSameNameRes(clusterID, ns string) *storage.Resource {
	return &storage.Resource{
		ClusterID:    clusterID,
		ResourceType: "Deployment",
		Data: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":              "order-test",
				"namespace":         ns,
				"creationTimestamp": "2026-01-01T00:00:00Z",
			},
		},
	}
}

func namespacesOf(res []*storage.Resource) []string {
	out := make([]string, 0, len(res))
	for _, r := range res {
		out = append(out, mapx.GetStr(r.Data, "metadata.namespace"))
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestQueryFilterSortStableWithSameName 同名资源（不同命名空间/集群）在任意输入顺序下，
// 排序结果必须唯一且确定，不能随并发聚合的完成顺序变化。
func TestQueryFilterSortStableWithSameName(t *testing.T) {
	base := []*storage.Resource{
		mkSameNameRes("BCS-K8S-00000", "restart-verify-ns"),
		mkSameNameRes("BCS-K8S-00000", "default"),
		mkSameNameRes("BCS-K8S-00000", "zz"),
		mkSameNameRes("BCS-K8S-00000", "aa"),
		mkSameNameRes("BCS-K8S-00001", "aa"), // 同名同命名空间、不同集群
	}
	wantAsc := []string{"aa", "aa", "default", "restart-verify-ns", "zz"}
	wantDesc := []string{"zz", "restart-verify-ns", "default", "aa", "aa"}

	cases := []struct {
		name   string
		filter QueryFilter
		want   []string
	}{
		{"byName asc", QueryFilter{Limit: len(base)}, wantAsc},
		{"byName desc", QueryFilter{Order: OrderDesc, Limit: len(base)}, wantDesc},
		{"byNamespace asc", QueryFilter{SortBy: SortByNamespace, Limit: len(base)}, wantAsc},
		{"byAge asc", QueryFilter{SortBy: SortByAge, Limit: len(base)}, wantAsc},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for i := 0; i < 200; i++ {
				res := append([]*storage.Resource{}, base...)
				rand.Shuffle(len(res), func(a, b int) { res[a], res[b] = res[b], res[a] })

				got := namespacesOf(c.filter.Page(res))
				if !equalStrings(got, c.want) {
					t.Fatalf("第 %d 次排序结果不稳定: got %v, want %v", i, got, c.want)
				}
			}
		})
	}
}
