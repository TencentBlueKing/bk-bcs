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
	"time"

	"github.com/robfig/cron"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
)

var _ Scaler = &CronScaler{}
var recordScheduleName = ""

// CronScaler is a crontab GPA
type CronScaler struct {
	ranges []v1alpha1.TimeRange
	name   string
	now    time.Time
}

// NewCronScaler initializer crontab GPA
func NewCronScaler(ranges []v1alpha1.TimeRange) Scaler {
	return &CronScaler{ranges: ranges, name: Cron, now: time.Now()}
}

// GetReplicas return replicas  recommend by crontab GPA
func (s *CronScaler) GetReplicas(gpa *v1alpha1.GeneralPodAutoscaler, currentReplicas int32) (int32, error) {
	var max int32 = -1
	for _, t := range s.ranges {
		misMatch, finalMatch, err := s.getFinalMatchAndMisMatch(gpa, t.Schedule)
		if err != nil {
			klog.Error(err)
			return currentReplicas, nil
		}
		klog.Infof("firstMisMatch: %v, finalMatch: %v", misMatch, finalMatch)
		if finalMatch == nil {
			continue
		}
		if max < t.DesiredReplicas {
			max = t.DesiredReplicas
			recordScheduleName = t.Schedule
		}
		klog.Infof("Schedule %v recommend %v replicas, desire: %v", t.Schedule, max, t.DesiredReplicas)
	}
	if max == -1 {
		klog.Info("Now is not in any time range")
	}
	return max, nil
}

// ScalerName returns scaler name
func (s *CronScaler) ScalerName() string {
	return s.name
}

func (s *CronScaler) getFinalMatchAndMisMatch(gpa *v1alpha1.GeneralPodAutoscaler,
	schedule string) (*time.Time, *time.Time, error) {
	sched, err := cron.ParseStandard(schedule)
	if err != nil {
		return nil, nil, err
	}
	// lastTime := gpa.Status.LastCronScheduleTime.DeepCopy()
	// if recordScheduleName != schedule {
	// 	lastTime = nil
	// }
	// if lastTime == nil || lastTime.IsZero() {
	// 	lastTime = gpa.CreationTimestamp.DeepCopy()
	// }
	// match := lastTime.Time
	// misMatch := lastTime.Time
	// klog.Infof("Init time: %v, now: %v", lastTime, s.now)
	// t := lastTime.Time
	// for {
	// 	if !t.After(s.now) {
	// 		misMatch = t
	// 		t = sched.Next(t)
	// 		continue
	// 	}
	// 	match = t
	// 	break
	// }
	// if s.now.Sub(misMatch).Minutes() < 1 && s.now.After(misMatch) {
	// 	return &misMatch, &match, nil
	// }

	lastTime := s.now.Add(-2 * time.Minute)
	match := lastTime
	misMatch := lastTime
	t := lastTime
	for {
		if !t.After(s.now) {
			misMatch = t
			t = sched.Next(t)
			continue
		}
		match = t
		break
	}
	klog.Infof("mismatch: %v, match: %v, now: %v", misMatch, match, s.now)
	if s.now.Sub(misMatch).Minutes() <= 1 {
		return &misMatch, &match, nil
	}

	return nil, nil, nil
}
