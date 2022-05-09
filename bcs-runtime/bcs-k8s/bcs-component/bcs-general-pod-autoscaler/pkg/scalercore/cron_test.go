/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scalercore

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
)

func Test_GetReplicas(t *testing.T) {
	testTime1, _ := time.Parse("2006-01-02 15:04:05", "2020-12-18 09:04:41")
	testTime2, _ := time.Parse("2006-01-02 15:04:05", "2020-12-18 09:00:01")
	testTime3, _ := time.Parse("2006-01-02 15:04:05", "2020-12-18 12:59:59")
	testTime4, _ := time.Parse("2006-01-02 15:04:05", "2020-12-18 12:58:59")
	testTime5, err := time.Parse("2006-01-02 15:04:05", "2020-12-18 08:58:59")
	lastTime := metav1.Time{Time: testTime1.Add(-1 * time.Second)}
	gpa := &v1alpha1.GeneralPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: testTime1.Add(-60 * time.Minute)},
		},
		Status: v1alpha1.GeneralPodAutoscalerStatus{},
	}
	definedGPA := &v1alpha1.GeneralPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Time{Time: testTime1.Add(-60 * time.Minute)},
		},
		Status: v1alpha1.GeneralPodAutoscalerStatus{
			LastCronScheduleTime: &lastTime,
		},
	}
	cronGPA := gpa.DeepCopy()
	lastCronTime := metav1.Time{Time: testTime5}
	cronGPA.Status.LastCronScheduleTime = &lastCronTime
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name    string
		ranges  []v1alpha1.TimeRange
		desired int32
		gpa     *v1alpha1.GeneralPodAutoscaler
		time    time.Time
	}{
		{
			name: "single timeRange, out of range",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 10-12 * * *",
					DesiredReplicas: 1,
				},
			},
			desired: -1,
		},
		{
			name: "single timeRange, in range",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 9-12 * * *",
					DesiredReplicas: 1,
				},
			},
			desired: 1,
		},
		{
			name: "multi timeRange, none match",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 10-12 * * *",
					DesiredReplicas: 1,
				},
				{
					Schedule:        "*/1 13-16 * * *",
					DesiredReplicas: 1,
				},
			},
			desired: -1,
		},
		{
			name: "multi timeRange, one match",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 9-12 * * *",
					DesiredReplicas: 1,
				},
				{
					Schedule:        "*/1 13-16 * * *",
					DesiredReplicas: 3,
				},
			},
			desired: 1,
		},
		{
			name: "multi timeRange, all match",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 8-12 * * *",
					DesiredReplicas: 1,
				},
				{
					Schedule:        "*/1 9-10 * * *",
					DesiredReplicas: 3,
				},
			},
			desired: 3,
		},
		{
			name: "cross day, not match",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 1-3 * * *",
					DesiredReplicas: 1,
				},
			},
			desired: -1,
		},
		{
			name: "single timeRange, in range, in a minute",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 9-12 * * *",
					DesiredReplicas: 1,
				},
			},
			desired: 1,
			gpa:     definedGPA,
		},
		{
			name: "single timeRange, in range, in a minute",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 9-12 * * *",
					DesiredReplicas: 1,
				},
			},
			desired: 1,
			gpa:     definedGPA,
		},
		{
			name: "single timeRange, in range, start boundary",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 9-12 * * *",
					DesiredReplicas: 1,
				},
			},
			desired: 1,
			time:    testTime2,
		},
		{
			name: "single timeRange, in range, end boundary",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 9-12 * * *",
					DesiredReplicas: 1,
				},
			},
			desired: 1,
			time:    testTime3,
		},
		{
			name: "single timeRange, in range, end boundary, 1 minute left",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 9-12 * * *",
					DesiredReplicas: 1,
				},
			},
			desired: 1,
			time:    testTime4,
		},
		{
			name: "multi timeRange lastCronTime is not nil, lastCron out this range, in range, end boundary, 1 minute left",
			ranges: []v1alpha1.TimeRange{
				{
					Schedule:        "*/1 9-12 * * *",
					DesiredReplicas: 1,
				},
				{
					Schedule:        "*/1 5-8 * * *",
					DesiredReplicas: 2,
				},
			},
			desired: 1,
			time:    testTime4,
			gpa:     cronGPA,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			defaultGPA := gpa
			if c.gpa != nil {
				defaultGPA = c.gpa
			}
			testTime := testTime1
			if !c.time.IsZero() {
				testTime = c.time
			}
			cron := &CronScaler{ranges: c.ranges, name: Cron, now: testTime}
			actual, err := cron.GetReplicas(defaultGPA, 0)
			if err != nil {
				t.Error(err)
			}
			if actual != c.desired {
				t.Errorf("desired: %v, actual: %v", c.desired, actual)
			}
		})
	}
}
