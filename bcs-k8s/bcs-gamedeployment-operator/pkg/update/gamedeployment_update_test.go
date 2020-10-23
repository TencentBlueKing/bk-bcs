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

package update

func getInt32Pointer(i int32) *int32 {
	return &i
}

/*func TestCalculateUpdateCount(t *testing.T) {
	readyPod := func() *v1.Pod {
		return &v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning, Conditions: []v1.PodCondition{{Type: v1.PodReady, Status: v1.ConditionTrue}}}}
	}
	cases := []struct {
		strategy          tkexv1alpha1.GameDeploymentUpdateStrategy
		totalReplicas     int
		waitUpdateIndexes []int
		pods              []*v1.Pod
		expectedResult    int
	}{
		{
			strategy:          tkexv1alpha1.GameDeploymentUpdateStrategy{},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{readyPod(), readyPod(), readyPod()},
			expectedResult:    1,
		},
		{
			strategy:          tkexv1alpha1.GameDeploymentUpdateStrategy{},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{readyPod(), {}, readyPod()},
			expectedResult:    0,
		},
		{
			strategy:          tkexv1alpha1.GameDeploymentUpdateStrategy{},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{{}, readyPod(), readyPod()},
			expectedResult:    1,
		},
		{
			strategy:          tkexv1alpha1.GameDeploymentUpdateStrategy{},
			totalReplicas:     10,
			waitUpdateIndexes: []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			pods:              []*v1.Pod{{}, readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), {}, readyPod()},
			expectedResult:    1,
		},
		{
			strategy:          tkexv1alpha1.GameDeploymentUpdateStrategy{Partition: getInt32Pointer(2), MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromInt(3))},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1},
			pods:              []*v1.Pod{{}, readyPod(), readyPod()},
			expectedResult:    0,
		},
		{
			strategy:          tkexv1alpha1.GameDeploymentUpdateStrategy{Partition: getInt32Pointer(2), MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromString("50%"))},
			totalReplicas:     8,
			waitUpdateIndexes: []int{0, 1, 2, 3, 4, 5, 6},
			pods:              []*v1.Pod{{}, readyPod(), {}, readyPod(), readyPod(), readyPod(), readyPod(), {}},
			expectedResult:    3,
		},
		{
			// maxUnavailable = 0 and maxSurge = 2, usedSurge = 1
			strategy: tkexv1alpha1.GameDeploymentUpdateStrategy{
				MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromInt(0)),
				MaxSurge:       intstrutil.ValueOrDefault(nil, intstrutil.FromInt(2)),
			},
			totalReplicas:     4,
			waitUpdateIndexes: []int{0, 1},
			pods:              []*v1.Pod{readyPod(), readyPod(), readyPod(), readyPod(), readyPod()},
			expectedResult:    1,
		},
		{
			// maxUnavailable = 1 and maxSurge = 2, usedSurge = 2
			strategy: tkexv1alpha1.GameDeploymentUpdateStrategy{
				MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromInt(1)),
				MaxSurge:       intstrutil.ValueOrDefault(nil, intstrutil.FromInt(2)),
			},
			totalReplicas:     4,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), readyPod()},
			expectedResult:    3,
		},
	}

	coreControl := gdcore.New(&tkexv1alpha1.GameDeployment{})
	for i, tc := range cases {
		res := calculateUpdateCount(coreControl, tc.strategy, 0, tc.totalReplicas, tc.waitUpdateIndexes, tc.pods)
		if res != tc.expectedResult {
			t.Fatalf("case #%d failed, expected %d, got %d", i, tc.expectedResult, res)
		}
	}
}*/
