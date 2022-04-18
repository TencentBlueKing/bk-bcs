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

package core

import (
	"reflect"
	"testing"
	"time"

	//testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs"
	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
)

func Test_parseTimeWithZone(t *testing.T) {
	utc, _ := time.LoadLocation("UTC")
	sh, _ := time.LoadLocation("Asia/Shanghai")
	time1UTC, _ := time.ParseInLocation(TIME_LAYOUT, "2022-02-28 00:00:00", utc)
	time1SH, _ := time.ParseInLocation(TIME_LAYOUT, "2022-02-28 08:00:00", sh)
	type args struct {
		currentTime time.Time
		zone        string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "0:00 of UTC, 8:00 of Asia/Shanghai",
			args: args{
				currentTime: time1UTC,
				zone:        "Asia/Shanghai",
			},
			want:    time1SH,
			wantErr: false,
		},
		{
			name: "error of Time Zone",
			args: args{
				currentTime: time1UTC,
				zone:        "Asia/Beijing",
			},
			want:    time.Time{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeWithZone(tt.args.currentTime, tt.args.zone)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeWithZone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTimeWithZone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFinalMatchAndMisMatch(t *testing.T) {
	utc, _ := time.LoadLocation("UTC")
	sh, _ := time.LoadLocation("Asia/Shanghai")
	timeutc, _ := time.ParseInLocation(TIME_LAYOUT, "2022-02-28 00:00:00", utc)
	timesh, _ := time.ParseInLocation(TIME_LAYOUT, "2022-02-28 08:00:00", sh)
	timesh2, _ := time.ParseInLocation(TIME_LAYOUT, "2022-02-28 08:01:00", sh)
	type args struct {
		schedule    string
		currentTime time.Time
		zone        string
	}
	tests := []struct {
		name    string
		args    args
		want    *time.Time
		want1   *time.Time
		wantErr bool
	}{
		{
			name: "current in range, hour",
			args: args{
				schedule:    "* 7-9 * * *",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    &timesh,
			want1:   &timesh2,
			wantErr: false,
		},
		{
			name: "current before of range, hour",
			args: args{
				schedule:    "* 9-10 * * *",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    nil,
			want1:   nil,
			wantErr: false,
		},
		{
			name: "current after of range, hour",
			args: args{
				schedule:    "* 6-7 * * *",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    nil,
			want1:   nil,
			wantErr: false,
		},
		{
			name: "current at the left edge of range, hour",
			args: args{
				schedule:    "* 8-10 * * *",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    &timesh,
			want1:   &timesh2,
			wantErr: false,
		},
		{
			name: "current at the right edge of range, hour",
			args: args{
				schedule:    "* 7-8 * * *",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    &timesh,
			want1:   &timesh2,
			wantErr: false,
		},
		{
			name: "current in range, day",
			args: args{
				schedule:    "* * 28 * *",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    &timesh,
			want1:   &timesh2,
			wantErr: false,
		},
		{
			name: "current out of range, day",
			args: args{
				schedule:    "* * 20-25 * *",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    nil,
			want1:   nil,
			wantErr: false,
		},
		{
			name: "current in range, weekday",
			args: args{
				schedule:    "* * * * 1",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    &timesh,
			want1:   &timesh2,
			wantErr: false,
		},
		{
			name: "current out of range, weekday",
			args: args{
				schedule:    "* * * * 5",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    nil,
			want1:   nil,
			wantErr: false,
		},
		{
			name: "wrong range",
			args: args{
				schedule:    "* * * *",
				currentTime: timeutc,
				zone:        "Asia/Shanghai",
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getFinalMatchAndMisMatch(tt.args.schedule, tt.args.currentTime, tt.args.zone)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFinalMatchAndMisMatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFinalMatchAndMisMatch() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getFinalMatchAndMisMatch() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func testProvider(t *testing.T) *bcs.Provider {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	autoOpts := config.AutoscalingOptions{
		CloudProviderName: "BCS",
	}
	discoveryOpts := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupSpecs: []string{"0:10:test-ng-1"},
	}
	provider := bcs.BuildCloudProvider(autoOpts, discoveryOpts, resourceLimiter)

	return provider.(*bcs.Provider)
}
func TestBufferedAutoscaler_doCron(t *testing.T) {
	utc, _ := time.LoadLocation("UTC")
	timeutc, _ := time.ParseInLocation(TIME_LAYOUT, "2022-02-28 00:00:00", utc)

	readyNodeLister := kubernetes.NewTestNodeLister(nil)
	allNodeLister := kubernetes.NewTestNodeLister(nil)
	scheduledPodMock := &podListerMock{}
	unschedulablePodMock := &podListerMock{}
	podDisruptionBudgetListerMock := &podDisruptionBudgetListerMock{}
	daemonSetListerMock := &daemonSetListerMock{}
	//onScaleUpMock := &onScaleUpMock{}
	//onScaleDownMock := &onScaleDownMock{}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Now())
	tn := BuildTestNode("tn", 1000, 1000)
	tni := schedulernodeinfo.NewNodeInfo()
	tni.SetNode(tn)
	provider := testProvider(t)
	// provider := testprovider.NewTestAutoprovisioningCloudProvider(
	// 	func(id string, delta int) error {
	// 		return onScaleUpMock.ScaleUp(id, delta)
	// 	}, func(id string, name string) error {
	// 		return onScaleDownMock.ScaleDown(id, name)
	// 	},
	// 	nil, nil,
	// 	nil, map[string]*schedulernodeinfo.NodeInfo{"ng1": tni, "ng2": tni})
	// provider.AddNodeGroup("ng1", 1, 10, 1)
	// provider.AddNode("ng1", n1)

	// Create context with mocked lister registry.
	options := config.AutoscalingOptions{
		EstimatorName:                       estimator.BinpackingEstimatorName,
		ScaleDownEnabled:                    true,
		ScaleDownUtilizationThreshold:       0.5,
		MaxNodesTotal:                       1,
		MaxCoresTotal:                       10,
		MaxMemoryTotal:                      100000,
		ScaleDownUnreadyTime:                time.Minute,
		ScaleDownUnneededTime:               time.Minute,
		FilterOutSchedulablePodsUsesPacking: true,
	}
	processorCallbacks := newBufferedAutoscalerProcessorCallbacks()
	context := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, processorCallbacks)
	listerRegistry := kube_util.NewListerRegistry(allNodeLister, readyNodeLister, scheduledPodMock,
		unschedulablePodMock, podDisruptionBudgetListerMock, daemonSetListerMock,
		nil, nil, nil, nil)
	context.ListerRegistry = listerRegistry

	clusterStateConfig := clusterstate.ClusterStateRegistryConfig{
		OkTotalUnreadyCount:  1,
		MaxNodeProvisionTime: 10 * time.Second,
	}

	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterStateConfig, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterState, 0)

	autoscaler := &BufferedAutoscaler{
		Context:               &context,
		clusterStateRegistry:  clusterState,
		lastScaleUpTime:       time.Now(),
		lastScaleDownFailTime: time.Now(),
		scaleDown:             sd,
		processors:            NewTestProcessors(),
		processorCallbacks:    processorCallbacks,
		initialized:           true,
	}

	type args struct {
		context              *contextinternal.Context
		currentTime          time.Time
		clusterStateRegistry *clusterstate.ClusterStateRegistry
	}
	tests := []struct {
		name string
		args args
		want errors.AutoscalerError
	}{
		// TODO: Add test cases.
		{
			name: "in range for one rule",
			args: args{
				context:              &context,
				currentTime:          timeutc,
				clusterStateRegistry: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := autoscaler
			if got := b.doCron(tt.args.context,
				tt.args.clusterStateRegistry, tt.args.currentTime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BufferedAutoscaler.doCron() = %v, want %v", got, tt.want)
			}
		})
	}
}
