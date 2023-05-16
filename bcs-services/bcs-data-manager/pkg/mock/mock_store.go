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
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

var (
	clusterMap   = map[string]*types.ClusterData{}
	namespaceMap = map[string]*types.NamespaceData{}
	workloadMap  = map[string]*types.WorkloadData{}
	projectMap   = map[string]*types.ProjectData{}
	publicMap    = map[string]*types.PublicData{}
)

// MockStore mock store
type MockStore struct {
	mock.Mock
}

func NewMockStore() store.Server {
	mockStore := &MockStore{}
	return mockStore
}

func (m *MockStore) GetProjectList(ctx context.Context, req *datamanager.GetAllProjectListRequest) ([]*datamanager.Project, int64, error) {
	projectList := make([]*datamanager.Project, 0)
	m.On("GetProjectList").Return(projectList, int64(len(projectList)), nil)
	args := m.Called()
	return args.Get(0).([]*datamanager.Project), args.Get(1).(int64), args.Error(2)
}

func (m *MockStore) GetProjectInfo(ctx context.Context, req *datamanager.GetProjectInfoRequest) (*datamanager.Project, error) {
	var project *datamanager.Project
	testProject := req.GetProject()
	m.On("GetProjectInfo", "testProject").Return(project, nil)
	m.On("GetProjectInfo", "testErr").Return(project, fmt.Errorf("get project err"))
	args := m.Called(testProject)
	return args.Get(0).(*datamanager.Project), args.Error(1)
}

func (m *MockStore) GetClusterInfoList(ctx context.Context, req *datamanager.GetClusterListRequest) ([]*datamanager.Cluster, int64, error) {
	var clusterList []*datamanager.Cluster
	testProject := req.GetProject()
	m.On("GetClusterInfoList", "testProject").Return(clusterList, int64(len(clusterList)), nil)
	m.On("GetClusterInfoList", "testErr").Return(clusterList, int64(0), fmt.Errorf("get cluster list err"))
	args := m.Called(testProject)
	return args.Get(0).([]*datamanager.Cluster), args.Get(1).(int64), args.Error(2)
}
func (m *MockStore) GetClusterInfo(ctx context.Context, req *datamanager.GetClusterInfoRequest) (*datamanager.Cluster, error) {
	var cluster *datamanager.Cluster
	testCluster := req.GetClusterID()
	m.On("GetClusterInfo", "testCluster").Return(cluster, nil)
	m.On("GetClusterInfo", "testErr").Return(cluster, fmt.Errorf("get cluster err"))
	args := m.Called(testCluster)
	return args.Get(0).(*datamanager.Cluster), args.Error(1)
}

func (m *MockStore) GetNamespaceInfoList(ctx context.Context, req *datamanager.GetNamespaceInfoListRequest) ([]*datamanager.Namespace, int64, error) {
	var nsList []*datamanager.Namespace
	testCluster := req.GetClusterID()
	m.On("GetNamespaceInfoList", "testCluster").Return(nsList, int64(len(nsList)), nil)
	m.On("GetNamespaceInfoList", "testErr").Return(nsList, int64(0), fmt.Errorf("get ns list err"))
	args := m.Called(testCluster)
	return args.Get(0).([]*datamanager.Namespace), args.Get(1).(int64), args.Error(2)
}

func (m *MockStore) GetNamespaceInfo(ctx context.Context, req *datamanager.GetNamespaceInfoRequest) (*datamanager.Namespace, error) {
	var ns *datamanager.Namespace
	testCluster := req.GetClusterID()
	testNamespace := req.GetNamespace()
	m.On("GetNamespaceInfo", "testCluster", "testNs").Return(ns, nil)
	m.On("GetNamespaceInfo", "testErr", "testErr").Return(ns, fmt.Errorf("get ns list err"))
	args := m.Called(testCluster, testNamespace)
	return args.Get(0).(*datamanager.Namespace), args.Error(1)
}
func (m *MockStore) GetWorkloadInfoList(ctx context.Context, req *datamanager.GetWorkloadInfoListRequest) ([]*datamanager.Workload, int64, error) {
	var workloadList []*datamanager.Workload
	testCluster := req.GetClusterID()
	testNs := req.GetNamespace()
	testType := req.GetWorkloadType()
	m.On("GetWorkloadInfoList", "testCluster", "testNs", "testType").Return(workloadList, int64(len(workloadList)), nil)
	m.On("GetWorkloadInfoList", "testErr", "testErr", "testErr").Return(workloadList, int64(0), fmt.Errorf("get wl list err"))
	args := m.Called(testCluster, testNs, testType)
	return args.Get(0).([]*datamanager.Workload), args.Get(1).(int64), args.Error(2)
}
func (m *MockStore) GetWorkloadInfo(ctx context.Context, req *datamanager.GetWorkloadInfoRequest) (*datamanager.Workload, error) {
	var workload *datamanager.Workload
	testCluster := req.GetClusterID()
	testNs := req.GetNamespace()
	testType := req.GetWorkloadType()
	testName := req.GetWorkloadName()
	m.On("GetWorkloadInfo", "testCluster", "testNs", "testType", "testName").Return(workload, nil)
	m.On("GetWorkloadInfo", "testErr", "testErr", "testErr", "testErr").Return(workload, fmt.Errorf("get wl err"))
	args := m.Called(testCluster, testNs, testType, testName)
	return args.Get(0).(*datamanager.Workload), args.Error(1)
}
func (m *MockStore) GetRawPublicInfo(ctx context.Context, opts *types.JobCommonOpts) ([]*types.PublicData, error) {
	var publicData []*types.PublicData
	m.On("GetRawPublicInfo").Return(publicData, nil)
	args := m.Called()
	return args.Get(0).([]*types.PublicData), args.Error(1)
}
func (m *MockStore) GetRawProjectInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.ProjectData, error) {
	var projectData []*types.ProjectData
	projectData = append(projectData, projectMap[opts.Dimension])
	testProject := opts.ProjectID
	m.On("GetRawProjectInfo", "testProject").Return(projectData, nil)
	m.On("GetRawProjectInfo", "testErr").Return(nil, fmt.Errorf("get data err"))
	args := m.Called(testProject)
	return args.Get(0).([]*types.ProjectData), args.Error(1)
}
func (m *MockStore) GetRawClusterInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.ClusterData, error) {
	var clusterData []*types.ClusterData
	testCluster := opts.ClusterID
	clusterData = append(clusterData, clusterMap[opts.Dimension])
	m.On("GetRawClusterInfo", "testCluster").Return(clusterData, nil)
	m.On("GetRawClusterInfo", "").Return(clusterData, nil)
	m.On("GetRawClusterInfo", "testErr").Return(nil, fmt.Errorf("get data err"))
	args := m.Called(testCluster)
	return args.Get(0).([]*types.ClusterData), args.Error(1)
}
func (m *MockStore) GetRawNamespaceInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.NamespaceData, error) {
	var nsData []*types.NamespaceData
	testNs := opts.Namespace
	nsData = append(nsData, namespaceMap[opts.Dimension])
	m.On("GetRawNamespaceInfo", "testNs").Return(nsData, nil)
	m.On("GetRawNamespaceInfo", "testErr").Return(nil, fmt.Errorf("get data err"))
	args := m.Called(testNs)
	return args.Get(0).([]*types.NamespaceData), args.Error(1)
}
func (m *MockStore) GetRawWorkloadInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.WorkloadData, error) {
	var workloadData []*types.WorkloadData
	testWl := opts.WorkloadName
	workloadData = append(workloadData, workloadMap[opts.Dimension])
	m.On("GetRawWorkloadInfo", "testWorkload").Return(workloadData, nil)
	m.On("GetRawWorkloadInfo", "").Return(workloadData, nil)
	m.On("GetRawWorkloadInfo", "testErr").Return(nil, fmt.Errorf("get data err"))
	args := m.Called(testWl)
	return args.Get(0).([]*types.WorkloadData), args.Error(1)
}

func (m *MockStore) InsertProjectInfo(ctx context.Context, metrics *types.ProjectMetrics, opts *types.JobCommonOpts) error {
	testProject := opts.ProjectID
	m.On("InsertProjectInfo", "testProject").Return(nil)
	m.On("InsertProjectInfo", "testErr").Return(fmt.Errorf("insert data err"))
	if nil == projectMap[opts.Dimension] {
		newMetrics := make([]*types.ProjectMetrics, 0)
		newMetrics = append(newMetrics, metrics)
		newProjectBucket := &types.ProjectData{
			CreateTime: primitive.NewDateTimeFromTime(time.Now()),
			UpdateTime: primitive.NewDateTimeFromTime(time.Now()),
			BucketTime: "mockBucketTime",
			Dimension:  opts.Dimension,
			ProjectID:  opts.ProjectID,
			Metrics:    newMetrics,
		}
		projectMap[opts.Dimension] = newProjectBucket
	} else {
		data := projectMap[opts.Dimension]
		data.Metrics = append(data.Metrics, metrics)
		data.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
		projectMap[opts.Dimension] = data
	}
	args := m.Called(testProject)
	return args.Error(0)
}
func (m *MockStore) InsertClusterInfo(ctx context.Context, metrics *types.ClusterMetrics, opts *types.JobCommonOpts) error {
	testCluster := opts.ClusterID
	m.On("InsertClusterInfo", "testCluster").Return(nil)
	m.On("InsertClusterInfo", "testErr").Return(fmt.Errorf("insert data err"))
	if nil == clusterMap[opts.Dimension] {
		newMetrics := make([]*types.ClusterMetrics, 0)
		newMetrics = append(newMetrics, metrics)
		newClusterBucket := &types.ClusterData{
			CreateTime:  primitive.NewDateTimeFromTime(time.Now()),
			UpdateTime:  primitive.NewDateTimeFromTime(time.Now()),
			BucketTime:  "mockBucketTime",
			Dimension:   opts.Dimension,
			ProjectID:   opts.ProjectID,
			ClusterID:   opts.ClusterID,
			ClusterType: opts.ClusterType,
			Metrics:     newMetrics,
		}
		clusterMap[opts.Dimension] = newClusterBucket
	} else {
		data := clusterMap[opts.Dimension]
		data.Metrics = append(data.Metrics, metrics)
		data.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
		clusterMap[opts.Dimension] = data
	}
	args := m.Called(testCluster)
	return args.Error(0)
}

func (m *MockStore) InsertNamespaceInfo(ctx context.Context, metrics *types.NamespaceMetrics, opts *types.JobCommonOpts) error {
	testNamespace := opts.Namespace
	m.On("InsertNamespaceInfo", "testNs").Return(nil)
	if nil == namespaceMap[opts.Dimension] {
		newMetrics := make([]*types.NamespaceMetrics, 0)
		newMetrics = append(newMetrics, metrics)
		newNamespaceBucket := &types.NamespaceData{
			CreateTime:  primitive.NewDateTimeFromTime(time.Now()),
			UpdateTime:  primitive.NewDateTimeFromTime(time.Now()),
			BucketTime:  "mockBucketTime",
			Dimension:   opts.Dimension,
			ProjectID:   opts.ProjectID,
			ClusterID:   opts.ClusterID,
			ClusterType: opts.ClusterType,
			Namespace:   opts.Namespace,
			Metrics:     newMetrics,
		}
		namespaceMap[opts.Dimension] = newNamespaceBucket
	} else {
		data := namespaceMap[opts.Dimension]
		data.Metrics = append(data.Metrics, metrics)
		data.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
		namespaceMap[opts.Dimension] = data
	}
	args := m.Called(testNamespace)
	return args.Error(0)
}

func (m *MockStore) InsertWorkloadInfo(ctx context.Context, metrics *types.WorkloadMetrics, opts *types.JobCommonOpts) error {
	testWorkload := opts.WorkloadName
	m.On("InsertWorkloadInfo", "testWorkload").Return(nil)
	m.On("InsertWorkloadInfo", "testErr").Return(fmt.Errorf("insert data err"))
	if nil == workloadMap[opts.Dimension] {
		newMetrics := make([]*types.WorkloadMetrics, 0)
		newMetrics = append(newMetrics, metrics)
		newWorkloadBucket := &types.WorkloadData{
			CreateTime:   primitive.NewDateTimeFromTime(time.Now()),
			UpdateTime:   primitive.NewDateTimeFromTime(time.Now()),
			BucketTime:   "mockBucketTime",
			Dimension:    opts.Dimension,
			ProjectID:    opts.ProjectID,
			ClusterID:    opts.ClusterID,
			ClusterType:  opts.ClusterType,
			Namespace:    opts.Namespace,
			WorkloadType: opts.WorkloadType,
			Name:         opts.WorkloadName,
			Metrics:      newMetrics,
		}
		workloadMap[opts.Dimension] = newWorkloadBucket
	} else {
		data := workloadMap[opts.Dimension]
		data.Metrics = append(data.Metrics, metrics)
		data.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
		workloadMap[opts.Dimension] = data
	}
	args := m.Called(testWorkload)
	return args.Error(0)
}
func (m *MockStore) GetWorkloadCount(ctx context.Context, opts *types.JobCommonOpts, bucket string, after time.Time) (int64, error) {
	testCount := opts.ClusterID
	m.On("GetWorkloadCount", "testCluster").Return(int64(5), nil)
	m.On("GetWorkloadCount", "testErr").Return(int64(0), fmt.Errorf("get count err"))
	args := m.Called(testCount)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockStore) InsertPublicInfo(ctx context.Context, metrics *types.PublicData, opts *types.JobCommonOpts) error {
	testCluster := opts.ClusterID
	m.On("InsertPublicInfo", "testCluster").Return(nil)
	m.On("InsertPublicInfo", "testErr").Return(fmt.Errorf("insert data err"))
	args := m.Called(testCluster)
	return args.Error(0)
}

func (m *MockStore) InsertPodAutoscalerInfo(ctx context.Context, metrics *types.PodAutoscalerMetrics, opts *types.JobCommonOpts) error {
	return nil
}
func (m *MockStore) GetPodAutoscalerList(ctx context.Context, request *datamanager.GetPodAutoscalerListRequest) ([]*datamanager.PodAutoscaler, int64, error) {
	return nil, 0, nil
}
func (m *MockStore) GetPodAutoscalerInfo(ctx context.Context, request *datamanager.GetPodAutoscalerRequest) (*datamanager.PodAutoscaler, error) {
	return nil, nil
}
func (m *MockStore) GetRawPodAutoscalerInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.PodAutoscalerData, error) {
	return nil, nil
}
