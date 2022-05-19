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

package cmd

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/client/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/client/pkg"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/stretchr/testify/assert"
)

func TestGetCluster(t *testing.T) {
	client := pkg.NewDataManagerCli(&pkg.Config{
		APIServer: "",
		AuthToken: "",
	})
	rsp, err := client.GetClusterInfo(&bcsdatamanager.GetClusterInfoRequest{
		ClusterID: "BCS-K8S-15202",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func TestGetNamespace(t *testing.T) {
	client := pkg.NewDataManagerCli(&pkg.Config{
		APIServer: "",
		AuthToken: "",
	})
	rsp, err := client.GetNamespaceInfo(&bcsdatamanager.GetNamespaceInfoRequest{
		ClusterID: "BCS-K8S-15202",
		Dimension: "hour",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func TestGetProject(t *testing.T) {
	//GetProject(nil, []string{"111"})
	client := pkg.NewDataManagerCli(&pkg.Config{
		APIServer: "",
		AuthToken: "",
	})
	rsp, err := client.GetProjectInfo(&bcsdatamanager.GetProjectInfoRequest{
		ProjectID: "111",
	})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}

func TestGetWorkload(t *testing.T) {

}

func TestClusterPrint(t *testing.T) {
	metrics := make([]*bcsdatamanager.ClusterMetrics, 0)
	metrics1 := &bcsdatamanager.ClusterMetrics{
		Time:               "2022-03-10",
		NodeCount:          "10",
		AvailableNodeCount: "10",
		MinNode: &bcsdatamanager.ExtremumRecord{
			Name:       "minNode",
			MetricName: "minNode",
			Value:      8,
			Period:     "2022-03-10 05:00:00, 2022-03-10 07:00:00",
		},
		MaxNode: &bcsdatamanager.ExtremumRecord{
			Name:       "maxNode",
			MetricName: "maxNode",
			Value:      12,
			Period:     "2022-03-10 20:00:00, 2022-03-10 22:00:00",
		},
		NodeQuantile:         nil,
		MinUsageNode:         "2.2.2.2",
		TotalCPU:             "60",
		TotalMemory:          "120",
		TotalLoadCPU:         "30",
		TotalLoadMemory:      "50",
		AvgLoadCPU:           "0",
		AvgLoadMemory:        "0",
		CPUUsage:             "0.6",
		MemoryUsage:          "0.42",
		WorkloadCount:        "60",
		InstanceCount:        "120",
		MinInstance:          nil,
		MaxInstanceTime:      nil,
		CpuRequest:           "50",
		MemoryRequest:        "100",
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}
	metrics2 := &bcsdatamanager.ClusterMetrics{
		Time:               "2022-03-10",
		NodeCount:          "10",
		AvailableNodeCount: "10",
		MinNode: &bcsdatamanager.ExtremumRecord{
			Name:       "minNode",
			MetricName: "minNode",
			Value:      8,
			Period:     "2022-03-10 05:00:00, 2022-03-10 07:00:00",
		},
		MaxNode: &bcsdatamanager.ExtremumRecord{
			Name:       "maxNode",
			MetricName: "maxNode",
			Value:      12,
			Period:     "2022-03-10 20:00:00, 2022-03-10 22:00:00",
		},
		NodeQuantile:         nil,
		MinUsageNode:         "2.2.2.2",
		TotalCPU:             "60",
		TotalMemory:          "120",
		TotalLoadCPU:         "30",
		TotalLoadMemory:      "50",
		AvgLoadCPU:           "0",
		AvgLoadMemory:        "0",
		CPUUsage:             "0.6",
		MemoryUsage:          "0.42",
		WorkloadCount:        "60",
		InstanceCount:        "120",
		MinInstance:          nil,
		MaxInstanceTime:      nil,
		CpuRequest:           "50",
		MemoryRequest:        "100",
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}
	metrics = append(metrics, metrics1, metrics2)
	res := &bcsdatamanager.Cluster{
		ProjectID:            "testid",
		ClusterID:            "testcluster",
		Dimension:            "day",
		StartTime:            "2022-03-09",
		EndTime:              "2022-03-10",
		Metrics:              metrics,
		MinNode:              nil,
		MaxNode:              nil,
		MinInstance:          nil,
		MaxInstance:          nil,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}
	rsp := &bcsdatamanager.GetClusterInfoResponse{
		Code:                 0,
		Message:              "",
		Data:                 res,
		XXX_NoUnkeyedLiteral: struct{}{},
		XXX_unrecognized:     nil,
		XXX_sizecache:        0,
	}
	printer.PrintClusterInTable(false, rsp.Data)
	printer.PrintClusterInJSON(rsp.Data)
}
