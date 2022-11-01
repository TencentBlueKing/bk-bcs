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

func TestModelProject_GetProjectInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		req     *bcsdatamanager.GetProjectInfoRequest
		want    error
		wantErr bool
	}{
		{
			name: "test1",
			req: &bcsdatamanager.GetProjectInfoRequest{
				Project:   "testproject1",
				Dimension: types.DimensionMinute,
			},
		},
		{
			name: "test2",
			req: &bcsdatamanager.GetProjectInfoRequest{
				Project:   "testproject2",
				Dimension: types.DimensionMinute,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.GetProjectInfo(ctx, tt.req)
			assert.Nil(t, err)
		})
	}
}

func TestModelProject_GetRawProjectInfo(t *testing.T) {
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
				ProjectID: "testproject1",
				Dimension: types.DimensionMinute,
			},
			bucket: "2022-03-16 11:00:00",
		},
		{
			name: "test2",
			opts: &types.JobCommonOpts{
				ObjectType: types.ClusterType,
				ProjectID:  "testproject1",
				Dimension:  types.DimensionMinute,
			},
		},
		{
			name: "test3",
			opts: &types.JobCommonOpts{
				ObjectType: types.ClusterType,
				ProjectID:  "testproject2",
				Dimension:  types.DimensionMinute,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.GetRawProjectInfo(ctx, tt.opts, tt.bucket)
			assert.Nil(t, err)
		})
	}
}

func TestModelProject_InsertProjectInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		opts    *types.JobCommonOpts
		metric  *types.ProjectMetrics
		wantErr bool
	}{
		{name: "test1",
			opts: &types.JobCommonOpts{
				ObjectType:  types.ClusterType,
				ProjectID:   "testproject1",
				Dimension:   types.DimensionMinute,
				CurrentTime: utils.FormatTime(time.Now().AddDate(0, 0, -1), types.DimensionMinute),
			},
			metric: &types.ProjectMetrics{
				Time: primitive.NewDateTimeFromTime(
					utils.FormatTime(time.Now().AddDate(0, 0, -1), types.DimensionDay)),
				ClustersCount:      int64(10),
				TotalCPU:           300,
				TotalMemory:        600,
				TotalLoadCPU:       100,
				TotalLoadMemory:    200,
				AvgLoadCPU:         100 / 10,
				AvgLoadMemory:      200 / 10,
				CPUUsage:           float64(100) / float64(300),
				MemoryUsage:        float64(200) / float64(600),
				NodeCount:          100,
				AvailableNodeCount: 99,
				MinNode:            nil,
				MaxNode:            nil,
			}},
		{name: "test2",
			opts: &types.JobCommonOpts{
				ObjectType:  types.ClusterType,
				ProjectID:   "testproject1",
				ClusterType: types.Kubernetes,
				Dimension:   types.DimensionMinute,
				CurrentTime: utils.FormatTime(time.Now(), types.DimensionMinute),
			},
			metric: &types.ProjectMetrics{
				Time:               primitive.NewDateTimeFromTime(utils.FormatTime(time.Now(), types.DimensionDay)),
				ClustersCount:      int64(10),
				TotalCPU:           300,
				TotalMemory:        600,
				TotalLoadCPU:       100,
				TotalLoadMemory:    200,
				AvgLoadCPU:         100 / 10,
				AvgLoadMemory:      200 / 10,
				CPUUsage:           float64(100) / float64(300),
				MemoryUsage:        float64(200) / float64(600),
				NodeCount:          100,
				AvailableNodeCount: 99,
				MinNode:            nil,
				MaxNode:            nil,
			}},
		{name: "test3",
			opts: &types.JobCommonOpts{
				ObjectType:  types.ClusterType,
				ProjectID:   "testproject2",
				ClusterType: types.Kubernetes,
				Dimension:   types.DimensionMinute,
				CurrentTime: utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute),
			},
			metric: &types.ProjectMetrics{
				Time:               primitive.NewDateTimeFromTime(utils.FormatTime(time.Now(), types.DimensionDay)),
				ClustersCount:      int64(10),
				TotalCPU:           300,
				TotalMemory:        600,
				TotalLoadCPU:       100,
				TotalLoadMemory:    200,
				AvgLoadCPU:         100 / 10,
				AvgLoadMemory:      200 / 10,
				CPUUsage:           float64(100) / float64(300),
				MemoryUsage:        float64(200) / float64(600),
				NodeCount:          100,
				AvailableNodeCount: 99,
				MinNode:            nil,
				MaxNode:            nil,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.InsertProjectInfo(ctx, tt.metric, tt.opts)
			assert.Nil(t, err)
		})
	}
}

//
// func TestModelProject_generateProjectResponse(t *testing.T) {
//
// }
//
// func TestModelProject_preAggregate(t *testing.T) {
//
// }
//
// func TestNewModelProject(t *testing.T) {
//
// }
