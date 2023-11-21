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

package mongo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

func TestModelCluster_GetClusterInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		req     *bcsdatamanager.GetClusterInfoRequest
		want    error
		wantErr bool
	}{
		{
			name: "test1",
			req: &bcsdatamanager.GetClusterInfoRequest{
				ClusterID: "testcluster1",
				Dimension: types.DimensionMinute,
			},
		},
		{
			name: "test2",
			req: &bcsdatamanager.GetClusterInfoRequest{
				ClusterID: "testcluster3",
				Dimension: types.DimensionMinute,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.GetClusterInfo(ctx, tt.req)
			assert.Nil(t, err)
		})
	}
}

func TestModelCluster_GetClusterInfoList(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		req     *bcsdatamanager.GetClusterListRequest
		want    *bcsdatamanager.Cluster
		wantErr bool
	}{
		{
			name: "test1",
			req: &bcsdatamanager.GetClusterListRequest{
				Project:   "",
				Dimension: types.DimensionMinute,
			},
		},
		{
			name: "test2",
			req: &bcsdatamanager.GetClusterListRequest{
				Project:   "testproject3",
				Dimension: types.DimensionMinute,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := store.GetClusterInfoList(ctx, tt.req)
			assert.Nil(t, err)
		})
	}
}

func TestModelCluster_GetRawClusterInfo(t *testing.T) {
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
				ObjectType:  types.ClusterType,
				ProjectID:   "testproject1",
				ClusterID:   "testcluster1",
				ClusterType: types.Kubernetes,
				Dimension:   types.DimensionMinute,
			},
			bucket: "2022-03-16 11:00:00",
		},
		{
			name: "test2",
			opts: &types.JobCommonOpts{
				ObjectType:  types.ClusterType,
				ProjectID:   "testproject1",
				ClusterID:   "testcluster1",
				ClusterType: types.Kubernetes,
				Dimension:   types.DimensionMinute,
			},
		},
		{
			name: "test3",
			opts: &types.JobCommonOpts{
				ObjectType:  types.ClusterType,
				ProjectID:   "testproject2",
				ClusterID:   "testcluster3",
				ClusterType: types.Kubernetes,
				Dimension:   types.DimensionMinute,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.GetRawClusterInfo(ctx, tt.opts, tt.bucket)
			assert.Nil(t, err)
		})
	}
}

func TestModelCluster_InsertClusterInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		opts    *types.JobCommonOpts
		metric  *types.ClusterMetrics
		wantErr bool
	}{
		{name: "test1",
			opts: &types.JobCommonOpts{
				ObjectType:  types.ClusterType,
				ProjectID:   "testproject1",
				ClusterID:   "testcluster1",
				BusinessID:  "testbusiness1",
				ClusterType: types.Kubernetes,
				Dimension:   types.DimensionMinute,
				CurrentTime: utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute),
			},
			metric: &types.ClusterMetrics{
				Time: primitive.NewDateTimeFromTime(utils.FormatTime(time.Now().Add((-10)*time.Minute),
					types.DimensionMinute)),
				TotalLoadCPU:       10,
				CPUUsage:           0.2,
				TotalLoadMemory:    50,
				MemoryRequest:      120,
				MemoryUsage:        0.2,
				InstanceCount:      20,
				CpuRequest:         60,
				AvgLoadMemory:      5,
				AvgLoadCPU:         1,
				NodeCount:          10,
				AvailableNodeCount: 10,
				WorkloadCount:      10,
				MinNode: &bcsdatamanager.ExtremumRecord{
					Name:       "MinNode",
					MetricName: "MaxNode",
					Value:      float64(8),
					Period:     utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute).String(),
				},
				MaxNode: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxNode",
					MetricName: "MaxNode",
					Value:      float64(12),
					Period:     utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute).String(),
				},
				MinInstance: &bcsdatamanager.ExtremumRecord{
					Name:       "MinInstance",
					MetricName: "MinInstance",
					Value:      float64(20),
					Period:     utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute).String(),
				},
				MaxInstance: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxInstance",
					MetricName: "MaxInstance",
					Value:      float64(22),
					Period:     utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute).String(),
				},
				MinUsageNode: "minUsageNode",
				NodeQuantile: []*bcsdatamanager.NodeQuantile{{
					Percentage:   "50",
					NodeCPUUsage: "0.1",
				}},
				TotalCPU:    100,
				TotalMemory: 200,
			}},
		{name: "test2",
			opts: &types.JobCommonOpts{
				ObjectType:  types.ClusterType,
				ProjectID:   "testproject1",
				ClusterID:   "testcluster1",
				BusinessID:  "testbusiness1",
				ClusterType: types.Kubernetes,
				Dimension:   types.DimensionMinute,
				CurrentTime: utils.FormatTime(time.Now(), types.DimensionMinute),
			},
			metric: &types.ClusterMetrics{
				Time:               primitive.NewDateTimeFromTime(utils.FormatTime(time.Now(), types.DimensionMinute)),
				TotalLoadCPU:       10,
				CPUUsage:           0.2,
				TotalLoadMemory:    50,
				MemoryRequest:      120,
				MemoryUsage:        0.2,
				InstanceCount:      20,
				CpuRequest:         60,
				AvgLoadMemory:      5,
				AvgLoadCPU:         1,
				NodeCount:          10,
				AvailableNodeCount: 10,
				WorkloadCount:      10,
				MinNode: &bcsdatamanager.ExtremumRecord{
					Name:       "MinNode",
					MetricName: "MinNode",
					Value:      float64(7),
					Period:     utils.FormatTime(time.Now(), types.DimensionMinute).String(),
				},
				MaxNode: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxNode",
					MetricName: "MaxNode",
					Value:      float64(13),
					Period:     utils.FormatTime(time.Now(), types.DimensionMinute).String(),
				},
				MinInstance: &bcsdatamanager.ExtremumRecord{
					Name:       "MinInstance",
					MetricName: "MinInstance",
					Value:      float64(18),
					Period:     utils.FormatTime(time.Now(), types.DimensionMinute).String(),
				},
				MaxInstance: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxInstance",
					MetricName: "MaxInstance",
					Value:      float64(23),
					Period:     utils.FormatTime(time.Now(), types.DimensionMinute).String(),
				},
				MinUsageNode: "1.1.1.1",
				NodeQuantile: []*bcsdatamanager.NodeQuantile{{
					Percentage:   "50",
					NodeCPUUsage: "0.1",
				}},
				TotalCPU:    100,
				TotalMemory: 200,
			}},
		{name: "test3",
			opts: &types.JobCommonOpts{
				ObjectType:  types.ClusterType,
				ProjectID:   "testproject2",
				ClusterID:   "testcluster2",
				BusinessID:  "testbusiness2",
				ClusterType: types.Kubernetes,
				Dimension:   types.DimensionMinute,
				CurrentTime: utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute),
			},
			metric: &types.ClusterMetrics{
				Time:               primitive.NewDateTimeFromTime(utils.FormatTime(time.Now(), types.DimensionMinute)),
				TotalLoadCPU:       10,
				CPUUsage:           0.2,
				TotalLoadMemory:    50,
				MemoryRequest:      120,
				MemoryUsage:        0.2,
				InstanceCount:      20,
				CpuRequest:         60,
				AvgLoadMemory:      5,
				AvgLoadCPU:         1,
				NodeCount:          10,
				AvailableNodeCount: 10,
				WorkloadCount:      10,
				MinNode: &bcsdatamanager.ExtremumRecord{
					Name:       "MinNode",
					MetricName: "MaxNode",
					Value:      float64(8),
					Period:     utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute).String(),
				},
				MaxNode: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxNode",
					MetricName: "MaxNode",
					Value:      float64(12),
					Period:     utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute).String(),
				},
				MinInstance: &bcsdatamanager.ExtremumRecord{
					Name:       "MinInstance",
					MetricName: "MinInstance",
					Value:      float64(20),
					Period:     utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute).String(),
				},
				MaxInstance: &bcsdatamanager.ExtremumRecord{
					Name:       "MaxInstance",
					MetricName: "MaxInstance",
					Value:      float64(22),
					Period:     utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute).String(),
				},
				MinUsageNode: "minUsageNode",
				NodeQuantile: []*bcsdatamanager.NodeQuantile{{
					Percentage:   "50",
					NodeCPUUsage: "0.1",
				}},
				TotalCPU:    100,
				TotalMemory: 200,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.InsertClusterInfo(ctx, tt.metric, tt.opts)
			assert.Nil(t, err)
		})
	}
}

// func TestModelCluster_generateClusterResponse(t *testing.T) {
//
// }
//
// func TestModelCluster_preAggregate(t *testing.T) {
//
// }
//
// func TestNewModelCluster(t *testing.T) {
//
// }

func newTestMongo() store.Server {
	mongoOptions := &mongo.Options{
		Hosts:                 []string{"127.0.0.1:27017"},
		ConnectTimeoutSeconds: 3,
		Database:              "datamanager_test",
		Username:              "data",
		Password:              "",
	}
	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		fmt.Println(err)
	}
	err = mongoDB.Ping()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("init mongo db successfully")

	bkbaseConfig := &types.BkbaseConfig{}
	return NewServer(mongoDB, bkbaseConfig)
}

func TestAggregate(t *testing.T) {
	pipeline := make([]map[string]interface{}, 0)
	pipeline = append(pipeline, map[string]interface{}{"$unwind": "$metrics"})
	pipeline = append(pipeline, map[string]interface{}{
		"$project": map[string]interface{}{
			"_id":     0,
			"metrics": 1,
		}})
	pipeline = append(pipeline, map[string]interface{}{"$match": map[string]interface{}{
		ClusterIDKey: "test",
		DimensionKey: "minute",
		MetricTimeKey: map[string]interface{}{
			"$gte": primitive.NewDateTimeFromTime(getStartTime("minute")),
		},
	}})
	fmt.Println(pipeline)
	fmt.Printf("%v", pipeline)
	fmt.Println(len(pipeline))
}
