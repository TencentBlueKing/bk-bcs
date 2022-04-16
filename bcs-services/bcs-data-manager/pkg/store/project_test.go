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

package store

import (
	"context"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
				ProjectID: "testproject1",
				Dimension: common.DimensionMinute,
			},
		},
		{
			name: "test2",
			req: &bcsdatamanager.GetProjectInfoRequest{
				ProjectID: "testproject2",
				Dimension: common.DimensionMinute,
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
		opts    *common.JobCommonOpts
		bucket  string
		want    int
		wantErr bool
	}{
		{
			name: "test1",
			opts: &common.JobCommonOpts{
				ProjectID: "testproject1",
				Dimension: common.DimensionMinute,
			},
			bucket: "2022-03-16 11:00:00",
		},
		{
			name: "test2",
			opts: &common.JobCommonOpts{
				ObjectType: common.ClusterType,
				ProjectID:  "testproject1",
				Dimension:  common.DimensionMinute,
			},
		},
		{
			name: "test3",
			opts: &common.JobCommonOpts{
				ObjectType: common.ClusterType,
				ProjectID:  "testproject2",
				Dimension:  common.DimensionMinute,
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
		opts    *common.JobCommonOpts
		metric  *common.ProjectMetrics
		wantErr bool
	}{
		{name: "test1",
			opts: &common.JobCommonOpts{
				ObjectType:  common.ClusterType,
				ProjectID:   "testproject1",
				Dimension:   common.DimensionMinute,
				CurrentTime: common.FormatTime(time.Now().AddDate(0, 0, -1), common.DimensionMinute),
			},
			metric: &common.ProjectMetrics{
				Time: primitive.NewDateTimeFromTime(
					common.FormatTime(time.Now().AddDate(0, 0, -1), common.DimensionDay)),
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
			opts: &common.JobCommonOpts{
				ObjectType:  common.ClusterType,
				ProjectID:   "testproject1",
				ClusterType: common.Kubernetes,
				Dimension:   common.DimensionMinute,
				CurrentTime: common.FormatTime(time.Now(), common.DimensionMinute),
			},
			metric: &common.ProjectMetrics{
				Time:               primitive.NewDateTimeFromTime(common.FormatTime(time.Now(), common.DimensionDay)),
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
			opts: &common.JobCommonOpts{
				ObjectType:  common.ClusterType,
				ProjectID:   "testproject2",
				ClusterType: common.Kubernetes,
				Dimension:   common.DimensionMinute,
				CurrentTime: common.FormatTime(time.Now().Add((-10)*time.Minute), common.DimensionMinute),
			},
			metric: &common.ProjectMetrics{
				Time:               primitive.NewDateTimeFromTime(common.FormatTime(time.Now(), common.DimensionDay)),
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
