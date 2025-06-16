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

package scaler

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	autoscaling "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
)

var (
	CPU     = "CPU"
	Webhook = "Webhook"
	Cron    = "Cron"
)

func Test_sortResults(t *testing.T) {
	type args struct {
		results []result
	}
	tests := []struct {
		name string
		args args
		want result
	}{
		{
			name: "test1",
			args: args{
				results: []result{
					generateResult(2, CPU, 0),
					generateResult(1, Webhook, 0),
					generateResult(3, Cron, 0),
				},
			},
			want: generateResult(3, Cron, 0),
		},
		{
			name: "test2",
			args: args{
				results: []result{
					generateResult(2, CPU, 0),
					generateResult(1, Webhook, 0),
					generateResult(1, Cron, 0),
				},
			},
			want: generateResult(2, CPU, 0),
		},
		{
			name: "test3",
			args: args{
				results: []result{
					generateResult(2, CPU, 1),
					generateResult(1, Webhook, 0),
					generateResult(3, Cron, 0),
				},
			},
			want: generateResult(2, CPU, 1),
		},
		{
			name: "test4",
			args: args{
				results: []result{
					generateResult(2, CPU, 2),
					generateResult(1, Webhook, 1),
					generateResult(3, Cron, 3),
				},
			},
			want: generateResult(3, Cron, 3),
		},
		{
			name: "test4",
			args: args{
				results: []result{
					generateResult(5, CPU, 2),
					generateResult(5, Webhook, 1),
					generateResult(2, Cron, 3),
				},
			},
			want: generateResult(2, Cron, 3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortResults(tt.args.results)
			if got[0].replicas != tt.want.replicas || got[0].metric != tt.want.metric {
				t.Errorf("sortResults() = %v, want %v", got[0], tt.want)
			}
		})
	}
}

func generateResult(replicas int32, metric string, priority int32) result {
	res := result{
		replicas:  replicas,
		metric:    metric,
		timestamp: time.Now(),
		priority:  priority,
	}
	switch metric {
	case CPU:
		res.statuses = []autoscaling.MetricStatus{
			{
				Type: autoscaling.ResourceMetricSourceType,
				Resource: &autoscaling.ResourceMetricStatus{
					Name: v1.ResourceCPU,
					Current: autoscaling.MetricValueStatus{
						Value: resource.NewMilliQuantity(100, resource.DecimalSI),
					},
				},
			},
		}
	case Webhook, Cron:
		res.statuses = []autoscaling.MetricStatus{}
	}
	return res
}
