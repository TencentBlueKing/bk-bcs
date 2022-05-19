/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metric

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/stretchr/testify/assert"
)

func newMonitor() bcsmonitor.ClientInterface {
	opt := bcsmonitor.BcsMonitorClientOpt{
		Schema:    "",
		Endpoint:  "",
		UserName:  "",
		Password:  "",
		AppCode:   "",
		AppSecret: "",
	}
	requester := bcsmonitor.NewRequester()
	monitorCli := bcsmonitor.NewBcsMonitorClient(opt, requester)
	monitorCli.SetCompleteEndpoint()
	header := http.Header{}
	monitorCli.SetDefaultHeader(header)
	return monitorCli
}

func Test_GetClusterCPUMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &common.JobCommonOpts{
		ObjectType:  common.ClusterType,
		ClusterID:   "BCS-MESOS-10039",
		ClusterType: common.Mesos,
		Dimension:   common.DimensionMinute,
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := common.NewClients(monitorCli, nil, nil, nil)
	_, _, _, _, err := getter.GetClusterCPUMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetClusterCPUMetrics(opts, clients))
}

func Test_GetClusterMemoryMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &common.JobCommonOpts{
		ObjectType:  common.ClusterType,
		ClusterID:   "BCS-K8S-15202",
		ClusterType: common.Kubernetes,
		Dimension:   common.DimensionMinute,
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := common.NewClients(monitorCli, nil, nil, nil)
	_, _, _, _, err := getter.GetClusterMemoryMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetClusterMemoryMetrics(opts, clients))
}

func Test_GetClusterNodeCount(t *testing.T) {

}

func Test_GetClusterNodeMetrics(t *testing.T) {

}

func Test_GetInstanceCount(t *testing.T) {
	monitorCli := newMonitor()
	opts := &common.JobCommonOpts{
		ObjectType:  common.ClusterType,
		ClusterID:   "BCS-MESOS-20042",
		ClusterType: common.Mesos,
		Dimension:   common.DimensionMinute,
		Namespace:   "bcs-system",
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := common.NewClients(monitorCli, nil, nil, nil)
	_, err := getter.GetInstanceCount(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetInstanceCount(opts, clients))
	opts2 := &common.JobCommonOpts{
		ObjectType:  common.ClusterType,
		ClusterID:   "BCS-K8S-15171",
		ClusterType: common.Kubernetes,
		Dimension:   common.DimensionMinute,
		Namespace:   "thanos",
		CurrentTime: time.Time{},
	}
	_, err = getter.GetInstanceCount(opts2, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetInstanceCount(opts2, clients))
	opts3 := &common.JobCommonOpts{
		ObjectType:   common.WorkloadType,
		ClusterID:    "BCS-K8S-15171",
		ClusterType:  common.Kubernetes,
		Dimension:    common.DimensionMinute,
		Namespace:    "thanos",
		WorkloadType: common.DeploymentType,
		Name:         "testdeploy",
		CurrentTime:  time.Time{},
	}
	_, err = getter.GetInstanceCount(opts3, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetInstanceCount(opts3, clients))
}

func Test_GetNamespaceCPUMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &common.JobCommonOpts{
		ObjectType:  common.NamespaceType,
		ClusterID:   "BCS-MESOS-20042",
		ClusterType: common.Mesos,
		Dimension:   common.DimensionMinute,
		Namespace:   "cq-loadtest",
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := common.NewClients(monitorCli, nil, nil, nil)
	_, _, _, err := getter.GetNamespaceCPUMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetNamespaceCPUMetrics(opts, clients))
}

func Test_GetNamespaceInstanceCount(t *testing.T) {
	monitorCli := newMonitor()
	opts := &common.JobCommonOpts{
		ObjectType:  common.NamespaceType,
		ClusterID:   "BCS-MESOS-20042",
		ClusterType: common.Mesos,
		Dimension:   common.DimensionMinute,
		Namespace:   "cq-loadtest",
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := common.NewClients(monitorCli, nil, nil, nil)
	_, err := getter.GetInstanceCount(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetInstanceCount(opts, clients))
}

func Test_GetNamespaceMemoryMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &common.JobCommonOpts{
		ObjectType:  common.NamespaceType,
		ClusterID:   "BCS-MESOS-20042",
		ClusterType: common.Mesos,
		Dimension:   common.DimensionHour,
		Namespace:   "cq-loadtest",
		CurrentTime: time.Time{},
	}
	getter := &MetricGetter{}
	clients := common.NewClients(monitorCli, nil, nil, nil)
	_, _, _, err := getter.GetNamespaceMemoryMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetNamespaceMemoryMetrics(opts, clients))
}

func Test_GetNamespaceResourceLimit(t *testing.T) {

}

func Test_GetWorkloadCPUMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &common.JobCommonOpts{
		ObjectType:   common.WorkloadType,
		ClusterID:    "BCS-K8S-15171",
		ClusterType:  common.Kubernetes,
		Dimension:    common.DimensionMinute,
		Namespace:    "bcs-system",
		WorkloadType: common.DeploymentType,
		Name:         "event-exporter",
		CurrentTime:  time.Time{},
	}
	getter := &MetricGetter{}
	clients := common.NewClients(monitorCli, nil, nil, nil)
	_, _, _, err := getter.GetWorkloadCPUMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetWorkloadCPUMetrics(opts, clients))
}

func Test_GetWorkloadMemoryMetrics(t *testing.T) {
	monitorCli := newMonitor()
	opts := &common.JobCommonOpts{
		ObjectType:   common.WorkloadType,
		ClusterID:    "BCS-K8S-15091",
		ClusterType:  common.Kubernetes,
		Dimension:    common.DimensionMinute,
		Namespace:    "mars-test",
		WorkloadType: common.DeploymentType,
		Name:         "mars-test-test1-micro-gateway-operator",
		CurrentTime:  time.Now(),
	}
	getter := &MetricGetter{}
	clients := common.NewClients(monitorCli, nil, nil, nil)
	_, _, _, err := getter.GetWorkloadMemoryMetrics(opts, clients)
	assert.Nil(t, err)
	fmt.Println(getter.GetWorkloadMemoryMetrics(opts, clients))
	fmt.Println(getter.GetInstanceCount(opts, clients))
	podCondition := generatePodCondition(opts.ClusterID, opts.Namespace, opts.WorkloadType, opts.Name)
	test, err := monitorCli.QueryByPost(fmt.Sprintf("sum(container_spec_memory_limit_bytes{%s})by(pod_name)", podCondition), opts.CurrentTime)
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
