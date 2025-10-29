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

package metric

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"
)

func newMonitor() bcsmonitor.ClientInterface {
	opt := bcsmonitor.BcsMonitorClientOpt{
		Endpoint:  "",
		AppCode:   "bcs-data-manager",
		AppSecret: "",
	}
	request := requester.NewRequester()
	monitorCli := bcsmonitor.NewBcsMonitorClient(opt, request)
	defaultHeader := http.Header{}
	defaultHeader.Add("", "")
	monitorCli.SetDefaultHeader(defaultHeader)
	return monitorCli
}

func Test_GetClusterCPUMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ClusterID:   "BCS-K8S-25975",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		CurrentTime: time.Time{},
		IsBKMonitor: true,
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	_, err := getter.GetClusterCPUMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetClusterCPUMetrics(opts, clients))
}

func Test_GetClusterMemoryMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ClusterID:   "BCS-K8S-25975",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		CurrentTime: time.Time{},
		IsBKMonitor: true,
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	_, err := getter.GetClusterMemoryMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetClusterMemoryMetrics(opts, clients))
}

func Test_GetClusterNodeCount(t *testing.T) {

}

func Test_GetClusterNodeMetrics(t *testing.T) {

}

func Test_GetInstanceCount(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ClusterID:   "BCS-MESOS-20042",
		ClusterType: types.Mesos,
		Dimension:   types.DimensionMinute,
		Namespace:   "bcs-system",
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	_, err := getter.GetInstanceCount(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetInstanceCount(opts, clients))
	opts2 := &types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ClusterID:   "BCS-K8S-15171",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		Namespace:   "thanos",
		CurrentTime: time.Time{},
	}
	_, err = getter.GetInstanceCount(opts2, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetInstanceCount(opts2, clients))
	opts3 := &types.JobCommonOpts{
		ObjectType:   types.WorkloadType,
		ClusterID:    "BCS-K8S-15171",
		ClusterType:  types.Kubernetes,
		Dimension:    types.DimensionMinute,
		Namespace:    "thanos",
		WorkloadType: types.DeploymentType,
		WorkloadName: "testdeploy",
		CurrentTime:  time.Time{},
	}
	_, err = getter.GetInstanceCount(opts3, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetInstanceCount(opts3, clients))
}

func Test_GetNamespaceCPUMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:  types.NamespaceType,
		ClusterID:   "BCS-K8S-15202",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		Namespace:   "bcs-system",
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	_, err := getter.GetNamespaceCPUMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetNamespaceCPUMetrics(opts, clients))
}

func Test_GetNamespaceInstanceCount(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:  types.NamespaceType,
		ClusterID:   "BCS-MESOS-20042",
		ClusterType: types.Mesos,
		Dimension:   types.DimensionMinute,
		Namespace:   "cq-loadtest",
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	_, err := getter.GetInstanceCount(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetInstanceCount(opts, clients))
}

func Test_GetNamespaceMemoryMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:  types.NamespaceType,
		ClusterID:   "BCS-K8S-15202",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionHour,
		Namespace:   "bcs-system",
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	_, err := getter.GetNamespaceMemoryMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetNamespaceMemoryMetrics(opts, clients))
}

func Test_GetNamespaceResourceLimit(t *testing.T) {

}

func Test_GetWorkloadCPUMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:   types.WorkloadType,
		ClusterID:    "BCS-K8S-25975",
		ClusterType:  types.Kubernetes,
		Dimension:    types.DimensionMinute,
		Namespace:    "bk-system",
		WorkloadType: types.DeploymentType,
		WorkloadName: "bcs-k8s-watch",
		CurrentTime:  time.Time{},
		IsBKMonitor:  true,
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	_, err := getter.GetWorkloadCPUMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetWorkloadCPUMetrics(opts, clients))
}

func Test_GetWorkloadMemoryMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:   types.WorkloadType,
		ClusterID:    "BCS-K8S-15202",
		ClusterType:  types.Kubernetes,
		Dimension:    types.DimensionMinute,
		Namespace:    "default",
		WorkloadType: types.DeploymentType,
		WorkloadName: "event-exporter",
		CurrentTime:  time.Now(),
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	_, err := getter.GetWorkloadMemoryMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetWorkloadMemoryMetrics(opts, clients))
	fmt.Println(getter.GetInstanceCount(opts, clients))
	podCondition := generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.WorkloadName)
	test, err := monitorCli.QueryByPost(fmt.Sprintf("sum(container_spec_memory_limit_bytes{%s})by(pod_name)",
		podCondition), opts.CurrentTime)
	fmt.Println(err)
	fmt.Println(test)
}

func TestPodName(t *testing.T) {
	regx := regexp.MustCompile("mars-test-test1-micro-gateway-operator" + "-[0-9a-z]*-[0-9a-z]*$")
	fmt.Println(regx.MatchString("mars-test-test1-micro-gateway-operator-5bf9c8d6fb-2f8fz"))
	regx2 := regexp.MustCompile("prometheus-operator-prometheus-node-exporter" + "-[0-9a-z]*$")
	fmt.Println(regx2.MatchString("prometheus-operator-prometheus-node-exporter-thpmd"))
	regx3 := regexp.MustCompile("prometheus-prometheus-operator-prometheus" + "-[0-9a-z]*$")
	fmt.Println(regx3.MatchString("prometheus-prometheus-operator-prometheus-0"))
}

func TestHPA(t *testing.T) {
	// query := "kube_event_unique_events_total{cluster_id=\"BCS-K8S-15202\", source=\"/cluster-autoscaler\",
	// reason=~\"ScaledUpGroup|ScaleDown\"}"
	// query := "min_over_time(sum(kube_event_unique_events_total{cluster_id=\"BCS-K8S-15202\",
	// source=\"/cluster-autoscaler\",reason=~\"ScaleDown\"}))"
	query := "kube_event_unique_events_total{cluster_id=\"BCS-K8S-15202\", " +
		"involved_object_kind=\"GeneralPodAutoscaler\",involved_object_name=\"gpa-test\",namespace=\"default\"," +
		"source=\"/pod-autoscaler\",reason=\"SuccessfulRescale\"}"
	// query := "sum(kube_event_unique_events_total{cluster_id=\"BCS-K8S-15202\",
	// source=\"/cluster-autoscaler\",reason=~\"ScaleDown\"})"
	monitorCli := newMonitor()
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()
	fmt.Println(start)
	fmt.Println(end)
	rsp, err := monitorCli.QueryRangeByPost(query, start, end, 30*time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
	assert.NotEqual(t, 0, len(rsp.Data.Result))
	fillSlice := fillMetrics(float64(16593460), rsp.Data.Result[0].Values, 30)
	fmt.Println(fillSlice)
	total := getIncreasingIntervalDifference(fillSlice)
	assert.Equal(t, 0, total)
}

func TestGetPodAutoscalerCount(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:        types.PodAutoscalerType,
		ClusterID:         "BCS-K8S-15202",
		ClusterType:       types.Kubernetes,
		Dimension:         types.DimensionMinute,
		Namespace:         "default",
		PodAutoscalerType: types.GPAType,
		PodAutoscalerName: "gpa-test",
		CurrentTime:       time.Now().Add(-44 * time.Minute),
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	_, err := getter.GetPodAutoscalerCount(opts, clients)
	assert.Nil(t, err)
}

func TestGetCACount(t *testing.T) {
	monitorCli := newMonitor()
	opts := &types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ClusterID:   "BCS-K8S-15202",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		CurrentTime: utils.FormatTime(time.Now().Add(-102*time.Minute), types.DimensionMinute),
	}
	getter := &MetricGetter{}
	clients := types.NewClients(monitorCli, nil, nil)
	count, err := getter.GetCACount(opts, clients)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, count)
}
