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

package datajob

import (
	"context"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/mock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestProjectDayPolicy_CalculateCpu(t *testing.T) {
	storeServer := mock.NewMockStore()
	metricGetter := mock.NewMockMetric()
	projectDayPolicy := NewProjectDayPolicy(metricGetter, storeServer)
	clusters := []*types.ClusterData{
		{
			Metrics: []*types.ClusterMetrics{{
				TotalCPU:        20,
				TotalLoadCPU:    10,
				TotalMemory:     50000,
				TotalLoadMemory: 5000,
			}},
		}, {
			Metrics: []*types.ClusterMetrics{{
				TotalCPU:        20,
				TotalLoadCPU:    10,
				TotalMemory:     50000,
				TotalLoadMemory: 5000,
			}, {
				TotalCPU:        30,
				TotalLoadCPU:    15,
				TotalMemory:     60000,
				TotalLoadMemory: 6000,
			}},
		},
	}
	total, load := projectDayPolicy.calculateCpu(clusters)
	assert.Equal(t, float64(50), total)
	assert.Equal(t, float64(25), load)
}

func Test_ProjectDayPolicy(t *testing.T) {
	storeServer := mock.NewMockStore()
	metricGetter := mock.NewMockMetric()
	minutePolicy := NewClusterMinutePolicy(metricGetter, storeServer)
	hourPolicy := NewClusterHourPolicy(metricGetter, storeServer)
	dayPolicy := NewClusterDayPolicy(metricGetter, storeServer)
	projectDayPolicy := NewProjectDayPolicy(metricGetter, storeServer)
	ctx := context.Background()
	minuteOpts := &types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionMinute),
	}
	hourOpts := &types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionHour,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionHour),
	}
	dayOpts := &types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionDay,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionDay),
	}
	projectOpts := &types.JobCommonOpts{
		ObjectType:  types.ProjectType,
		ProjectID:   "testProject",
		Dimension:   types.DimensionDay,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionDay),
	}

	client := &types.Clients{
		MonitorClient:   nil,
		K8sStorageCli:   mock.NewMockStorage(),
		MesosStorageCli: mock.NewMockStorage(),
	}
	minutePolicy.ImplementPolicy(ctx, minuteOpts, client)
	hourPolicy.ImplementPolicy(ctx, hourOpts, client)
	dayPolicy.ImplementPolicy(ctx, dayOpts, client)
	projectDayPolicy.ImplementPolicy(ctx, projectOpts, client)
	bucket, err := utils.GetBucketTime(projectOpts.CurrentTime, types.DimensionDay)
	assert.Nil(t, err)
	project, err := storeServer.GetRawProjectInfo(ctx, dayOpts, bucket)
	assert.Nil(t, err)
	assert.NotNil(t, project)
}

func Test_calculateMemory(t *testing.T) {
	storeServer := mock.NewMockStore()
	metricGetter := mock.NewMockMetric()
	projectDayPolicy := NewProjectDayPolicy(metricGetter, storeServer)
	clusters := []*types.ClusterData{
		{
			Metrics: []*types.ClusterMetrics{{
				TotalCPU:        20,
				TotalLoadCPU:    10,
				TotalMemory:     50000,
				TotalLoadMemory: 5000,
			}},
		}, {
			Metrics: []*types.ClusterMetrics{{
				TotalCPU:        20,
				TotalLoadCPU:    10,
				TotalMemory:     50000,
				TotalLoadMemory: 5000,
			}, {
				TotalCPU:        30,
				TotalLoadCPU:    15,
				TotalMemory:     60000,
				TotalLoadMemory: 6000,
			}},
		},
	}
	total, load := projectDayPolicy.calculateMemory(clusters)
	assert.Equal(t, int64(110000), total)
	assert.Equal(t, int64(11000), load)
}

func Test_calculateProjectNodeCount(t *testing.T) {
	storeServer := mock.NewMockStore()
	metricGetter := mock.NewMockMetric()
	projectDayPolicy := NewProjectDayPolicy(metricGetter, storeServer)
	clusters := []*types.ClusterData{
		{
			Metrics: []*types.ClusterMetrics{{
				NodeCount:          20,
				AvailableNodeCount: 19,
			}},
		}, {
			Metrics: []*types.ClusterMetrics{{
				NodeCount:          20,
				AvailableNodeCount: 19,
			}, {
				NodeCount:          20,
				AvailableNodeCount: 20,
			}},
		},
	}
	total, load := projectDayPolicy.calculateProjectNodeCount(clusters)
	assert.Equal(t, int64(40), total)
	assert.Equal(t, int64(39), load)
}
