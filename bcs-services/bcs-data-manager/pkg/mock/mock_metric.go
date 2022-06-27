/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package mock

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/stretchr/testify/mock"
)

type MockMetric struct {
	mock.Mock
}

func NewMockMetric() metric.Server {
	return &MockMetric{}
}

func (m *MockMetric) GetWorkloadCPUMetrics(opts *types.JobCommonOpts, clients *types.Clients) (float64, float64, float64, error) {
	testWorkload := opts.WorkloadName
	m.On("GetWorkloadCPUMetrics", "testWorkload").Return(2.00, 1.0, 1.0, nil)
	m.On("GetWorkloadCPUMetrics", "testErr").Return(0.0, 0.0, 0.0, fmt.Errorf("test err"))
	args := m.Called(testWorkload)
	return args.Get(0).(float64), args.Get(1).(float64), args.Get(2).(float64), args.Error(3)
}

func (m *MockMetric) GetWorkloadMemoryMetrics(opts *types.JobCommonOpts, clients *types.Clients) (int64, int64, float64, error) {
	testWorkload := opts.WorkloadName
	m.On("GetWorkloadMemoryMetrics", "testWorkload").Return(int64(200), int64(100), 0.5, nil)
	m.On("GetWorkloadMemoryMetrics", "testErr").Return(int64(0), int64(0), 0.0, fmt.Errorf("test err"))
	args := m.Called(testWorkload)
	return args.Get(0).(int64), args.Get(1).(int64), args.Get(2).(float64), args.Error(3)
}
func (m *MockMetric) GetNamespaceCPUMetrics(opts *types.JobCommonOpts, clients *types.Clients) (float64, float64, float64, error) {
	testNs := opts.Namespace
	m.On("GetNamespaceCPUMetrics", "testNs").Return(2.00, 1.0, 1.0, nil)
	m.On("GetNamespaceCPUMetrics", "testErr").Return(0.0, 0.0, 0.0, fmt.Errorf("test err"))
	args := m.Called(testNs)
	return args.Get(0).(float64), args.Get(1).(float64), args.Get(2).(float64), args.Error(3)
}
func (m *MockMetric) GetNamespaceMemoryMetrics(opts *types.JobCommonOpts, clients *types.Clients) (int64, int64, float64, error) {
	testNs := opts.Namespace
	m.On("GetNamespaceMemoryMetrics", "testNs").Return(int64(200), int64(100), 0.5, nil)
	m.On("GetNamespaceMemoryMetrics", "testErr").Return(int64(0), int64(0), 0.0, fmt.Errorf("test err"))
	args := m.Called(testNs)
	return args.Get(0).(int64), args.Get(1).(int64), args.Get(2).(float64), args.Error(3)
}
func (m *MockMetric) GetClusterCPUMetrics(opts *types.JobCommonOpts, clients *types.Clients) (float64,
	float64, float64, float64, error) {
	testCluster := opts.ClusterID
	m.On("GetClusterCPUMetrics", "testCluster").Return(200.00, 100.00, 10.0, 10.0/200.0, nil)
	m.On("GetClusterCPUMetrics", "testErr").Return(0.0, 0.0, 0.0, 0.0, fmt.Errorf("test err"))
	args := m.Called(testCluster)
	return args.Get(0).(float64), args.Get(1).(float64), args.Get(2).(float64), args.Get(3).(float64), args.Error(4)
}
func (m *MockMetric) GetClusterMemoryMetrics(opts *types.JobCommonOpts, clients *types.Clients) (int64,
	int64, int64, float64, error) {
	testCluster := opts.ClusterID
	m.On("GetClusterMemoryMetrics", "testCluster").Return(int64(200), int64(100), int64(10), 10.0/200.0, nil)
	m.On("GetClusterMemoryMetrics", "testErr").Return(int64(0), int64(0), int64(0), 0.0, fmt.Errorf("test err"))
	args := m.Called(testCluster)
	return args.Get(0).(int64), args.Get(1).(int64), args.Get(2).(int64), args.Get(3).(float64), args.Error(4)
}
func (m *MockMetric) GetInstanceCount(opts *types.JobCommonOpts, clients *types.Clients) (int64, error) {
	testCluster := opts.ClusterID
	m.On("GetInstanceCount", "testCluster").Return(int64(50), nil)
	m.On("GetInstanceCount", "testErr").Return(int64(0), fmt.Errorf("test err"))
	args := m.Called(testCluster)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockMetric) GetClusterNodeMetrics(opts *types.JobCommonOpts,
	clients *types.Clients) (string, []*bcsdatamanager.NodeQuantile, error) {
	testCluster := opts.ClusterID
	var node []*bcsdatamanager.NodeQuantile
	m.On("GetClusterNodeMetrics", "testCluster").Return("testNode", node, nil)
	m.On("GetClusterNodeMetrics", "testErr").Return("", nil, fmt.Errorf("test err"))
	args := m.Called(testCluster)
	return args.Get(0).(string), args.Get(1).([]*bcsdatamanager.NodeQuantile), args.Error(2)
}
func (m *MockMetric) GetClusterNodeCount(opts *types.JobCommonOpts, clients *types.Clients) (int64, int64, error) {
	testCluster := opts.ClusterID
	m.On("GetClusterNodeCount", "testCluster").Return(int64(20), int64(19), nil)
	m.On("GetClusterNodeCount", "testErr").Return(int64(0), int64(0), fmt.Errorf("test err"))
	args := m.Called(testCluster)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockMetric) GetPodAutoscalerCount(opts *types.JobCommonOpts, clients *types.Clients) (int64, error) {
	return 0, nil
}
func (m *MockMetric) GetCACount(opts *types.JobCommonOpts, clients *types.Clients) (int64, error) {
	return 0, nil
}
