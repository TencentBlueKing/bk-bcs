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

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

func Test_GetRawWorkloadInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		opts    *types.JobCommonOpts
		bucket  string
		want    int
		wantErr bool
	}{
		{
			name: "test1",
			opts: &types.JobCommonOpts{
				ObjectType:   types.ClusterType,
				ProjectID:    "testproject1",
				ClusterID:    "testcluster1",
				ClusterType:  types.Kubernetes,
				Dimension:    types.DimensionMinute,
				Namespace:    "testnamespace1",
				WorkloadType: types.DeploymentType,
				WorkloadName: "testdeploy1",
			},
			bucket: "2022-03-16 14:00:00",
		},
		{
			name: "test2",
			opts: &types.JobCommonOpts{
				ObjectType:   types.ClusterType,
				ProjectID:    "testproject1",
				ClusterID:    "testcluster1",
				ClusterType:  types.Kubernetes,
				Dimension:    types.DimensionMinute,
				Namespace:    "testnamespace1",
				WorkloadType: types.DeploymentType,
				WorkloadName: "testdeploy1",
			},
		},
		{
			name: "test2",
			opts: &types.JobCommonOpts{
				ObjectType:   types.ClusterType,
				ProjectID:    "testproject2",
				ClusterID:    "testcluster3",
				ClusterType:  types.Kubernetes,
				Dimension:    types.DimensionMinute,
				Namespace:    "testnamespace1",
				WorkloadType: types.DeploymentType,
				WorkloadName: "testdeploy1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.GetRawWorkloadInfo(ctx, tt.opts, tt.bucket)
			assert.Nil(t, err)
		})
	}
}

func Test_GetWorkloadInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		req     *bcsdatamanager.GetWorkloadInfoRequest
		want    *bcsdatamanager.Workload
		wantErr bool
	}{
		{
			name: "test1",
			req: &bcsdatamanager.GetWorkloadInfoRequest{
				ClusterID:    "testcluster1",
				Dimension:    types.DimensionMinute,
				Namespace:    "testnamespace1",
				WorkloadType: types.DeploymentType,
				WorkloadName: "testdeploy1",
			},
		},
		{
			name: "test2",
			req: &bcsdatamanager.GetWorkloadInfoRequest{
				ClusterID:    "testcluster3",
				Dimension:    types.DimensionMinute,
				Namespace:    "testnamespace1",
				WorkloadType: types.DeploymentType,
				WorkloadName: "testdeploy1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.GetWorkloadInfo(ctx, tt.req)
			assert.Nil(t, err)
		})
	}
}

func Test_GetWorkloadInfoList(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		req     *bcsdatamanager.GetWorkloadInfoListRequest
		want    *bcsdatamanager.Workload
		wantErr bool
	}{
		{
			name: "test1",
			req: &bcsdatamanager.GetWorkloadInfoListRequest{
				ClusterID:    "testcluster1",
				Dimension:    types.DimensionMinute,
				Namespace:    "testnamespace1",
				WorkloadType: types.DeploymentType,
			},
		},
		{
			name: "test2",
			req: &bcsdatamanager.GetWorkloadInfoListRequest{
				ClusterID:    "testcluster3",
				Dimension:    types.DimensionMinute,
				Namespace:    "testnamespace1",
				WorkloadType: types.DeploymentType,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := store.GetWorkloadInfoList(ctx, tt.req)
			assert.Nil(t, err)
		})
	}
}

func Test_InsertWorkloadInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		opts    *types.JobCommonOpts
		metric  *types.WorkloadMetrics
		wantErr bool
	}{
		{name: "test1",
			opts: &types.JobCommonOpts{
				ObjectType:   types.WorkloadType,
				ProjectID:    "testproject1",
				ClusterID:    "testcluster1",
				Namespace:    "testnamespace1",
				WorkloadType: types.DeploymentType,
				WorkloadName: "testdeploy1",
				ClusterType:  types.Kubernetes,
				Dimension:    types.DimensionMinute,
				CurrentTime:  utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute),
			},
			metric: &types.WorkloadMetrics{
				Time: primitive.NewDateTimeFromTime(utils.FormatTime(time.Now().Add((-10)*time.Minute),
					types.DimensionMinute)),
				CPUUsage:          0.2,
				MemoryRequest:     20,
				MemoryUsage:       0.2,
				InstanceCount:     20,
				CPURequest:        10,
				CPUUsageAmount:    5,
				MemoryUsageAmount: 10,
				MaxCPUUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxCPUUsage",
					MetricName: "MaxCPUUsage",
					Value:      0.2,
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinCPUUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinCPUUsage",
					MetricName: "MinCPUUsage",
					Value:      0.2,
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MaxMemoryUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxMemoryUsage",
					MetricName: "MaxMemoryUsage",
					Value:      0.2,
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinMemoryUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinMemoryUsage",
					MetricName: "MinMemoryUsage",
					Value:      0.2,
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinInstanceTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinInstance",
					MetricName: "MinInstance",
					Value:      float64(20),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MaxInstanceTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxInstance",
					MetricName: "MaxInstance",
					Value:      float64(20),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MaxCPUTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxCPU",
					MetricName: "MaxCPU",
					Value:      float64(5),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinCPUTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinCPU",
					MetricName: "MinCPU",
					Value:      float64(5),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MaxMemoryTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxMemory",
					MetricName: "MaxMemory",
					Value:      float64(10),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinMemoryTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinMemory",
					MetricName: "MinMemory",
					Value:      float64(10),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
			}},
		{name: "test2",
			opts: &types.JobCommonOpts{
				ObjectType:   types.ClusterType,
				ProjectID:    "testproject1",
				ClusterID:    "testcluster1",
				Namespace:    "testnamespace1",
				WorkloadType: types.DeploymentType,
				WorkloadName: "testdeploy1",
				ClusterType:  types.Kubernetes,
				Dimension:    types.DimensionMinute,
				CurrentTime:  utils.FormatTime(time.Now(), types.DimensionMinute),
			},
			metric: &types.WorkloadMetrics{
				Time:              primitive.NewDateTimeFromTime(utils.FormatTime(time.Now(), types.DimensionMinute)),
				CPUUsage:          0.3,
				MemoryRequest:     20,
				MemoryUsage:       0.3,
				InstanceCount:     20,
				CPURequest:        10,
				CPUUsageAmount:    5,
				MemoryUsageAmount: 10,
				MaxCPUUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxCPUUsage",
					MetricName: "MaxCPUUsage",
					Value:      0.3,
					Period:     time.Now().String(),
				},
				MinCPUUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinCPUUsage",
					MetricName: "MinCPUUsage",
					Value:      0.3,
					Period:     time.Now().String(),
				},
				MaxMemoryUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxMemoryUsage",
					MetricName: "MaxMemoryUsage",
					Value:      0.3,
					Period:     time.Now().String(),
				},
				MinMemoryUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinMemoryUsage",
					MetricName: "MinMemoryUsage",
					Value:      0.3,
					Period:     time.Now().String(),
				},
				MinInstanceTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinInstance",
					MetricName: "MinInstance",
					Value:      float64(18),
					Period:     time.Now().String(),
				},
				MaxInstanceTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxInstance",
					MetricName: "MaxInstance",
					Value:      float64(22),
					Period:     time.Now().String(),
				},
				MaxCPUTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxCPU",
					MetricName: "MaxCPU",
					Value:      float64(22),
					Period:     time.Now().String(),
				},
				MinCPUTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinCPU",
					MetricName: "MinCPU",
					Value:      float64(10),
					Period:     time.Now().String(),
				},
				MaxMemoryTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxMemory",
					MetricName: "MaxMemory",
					Value:      float64(12),
					Period:     time.Now().String(),
				},
				MinMemoryTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinMemory",
					MetricName: "MinMemory",
					Value:      float64(12),
					Period:     time.Now().String(),
				},
			}},
		{name: "test3",
			opts: &types.JobCommonOpts{
				ObjectType:   types.ClusterType,
				ProjectID:    "testproject2",
				ClusterID:    "testcluster2",
				Namespace:    "testnamespace1",
				WorkloadName: "testdeploy2",
				WorkloadType: types.DeploymentType,
				ClusterType:  types.Kubernetes,
				Dimension:    types.DimensionMinute,
				CurrentTime:  utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute),
			},
			metric: &types.WorkloadMetrics{
				Time: primitive.NewDateTimeFromTime(utils.FormatTime(time.Now().Add((-10)*time.Minute),
					types.DimensionMinute)),
				CPUUsage:          0.2,
				MemoryRequest:     20,
				MemoryUsage:       0.2,
				InstanceCount:     20,
				CPURequest:        10,
				CPUUsageAmount:    5,
				MemoryUsageAmount: 10,
				MaxCPUUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxCPUUsage",
					MetricName: "MaxCPUUsage",
					Value:      0.2,
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinCPUUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinCPUUsage",
					MetricName: "MinCPUUsage",
					Value:      0.2,
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MaxMemoryUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxMemoryUsage",
					MetricName: "MaxMemoryUsage",
					Value:      0.2,
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinMemoryUsageTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinMemoryUsage",
					MetricName: "MinMemoryUsage",
					Value:      0.2,
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinInstanceTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinInstance",
					MetricName: "MinInstance",
					Value:      float64(20),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MaxInstanceTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxInstance",
					MetricName: "MaxInstance",
					Value:      float64(20),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MaxCPUTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxCPU",
					MetricName: "MaxCPU",
					Value:      float64(10),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinCPUTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinCPU",
					MetricName: "MinCPU",
					Value:      float64(10),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MaxMemoryTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxMemory",
					MetricName: "MaxMemory",
					Value:      float64(12),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
				MinMemoryTime: &bcsdatamanager.ExtremumRecord{
					Name:       "MinMemory",
					MetricName: "MinMemory",
					Value:      float64(12),
					Period:     time.Now().Add((-10) * time.Minute).String(),
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.InsertWorkloadInfo(ctx, tt.metric, tt.opts)
			assert.Nil(t, err)
		})
	}
}

//
// func TestModelWorkload_generateCond(t *testing.T) {
//
// }
//
// func TestModelWorkload_generateWorkloadResponse(t *testing.T) {
//
// }
//
// func TestModelWorkload_preAggregate(t *testing.T) {
//
// }
//
// func TestNewModelWorkload(t *testing.T) {
//
// }
