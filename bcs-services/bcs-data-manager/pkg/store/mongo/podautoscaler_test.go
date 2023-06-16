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

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

func Test_InsertPodAutoscalerInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		opts    *types.JobCommonOpts
		metric  *types.PodAutoscalerMetrics
		wantErr bool
	}{
		{name: "test1",
			opts: &types.JobCommonOpts{
				ObjectType:        types.WorkloadType,
				ProjectID:         "testproject1",
				ClusterID:         "testcluster1",
				Namespace:         "testnamespace1",
				WorkloadType:      types.DeploymentType,
				WorkloadName:      "testdeploy1",
				ClusterType:       types.Kubernetes,
				Dimension:         types.DimensionMinute,
				PodAutoscalerType: types.HPAType,
				PodAutoscalerName: "testhpa",
				CurrentTime:       utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute),
			},
			metric: &types.PodAutoscalerMetrics{
				Time: primitive.NewDateTimeFromTime(utils.FormatTime(time.Now().Add((-10)*time.Minute),
					types.DimensionMinute)),
				TotalSuccessfulRescale: 1,
			}},
		{name: "test2",
			opts: &types.JobCommonOpts{
				ObjectType:        types.ClusterType,
				ProjectID:         "testproject1",
				ClusterID:         "testcluster1",
				Namespace:         "testnamespace1",
				WorkloadType:      types.DeploymentType,
				WorkloadName:      "testdeploy1",
				ClusterType:       types.Kubernetes,
				Dimension:         types.DimensionMinute,
				PodAutoscalerType: types.HPAType,
				PodAutoscalerName: "testhpa",
				CurrentTime:       utils.FormatTime(time.Now(), types.DimensionMinute),
			},
			metric: &types.PodAutoscalerMetrics{
				Time:                   primitive.NewDateTimeFromTime(utils.FormatTime(time.Now(), types.DimensionMinute)),
				TotalSuccessfulRescale: 2,
			}},
		{name: "test3",
			opts: &types.JobCommonOpts{
				ObjectType:        types.ClusterType,
				ProjectID:         "testproject2",
				ClusterID:         "testcluster2",
				Namespace:         "testnamespace1",
				WorkloadName:      "testdeploy2",
				WorkloadType:      types.DeploymentType,
				ClusterType:       types.Kubernetes,
				Dimension:         types.DimensionMinute,
				PodAutoscalerType: types.GPAType,
				PodAutoscalerName: "testgpa",
				CurrentTime:       utils.FormatTime(time.Now().Add((-10)*time.Minute), types.DimensionMinute),
			},
			metric: &types.PodAutoscalerMetrics{
				Time: primitive.NewDateTimeFromTime(utils.FormatTime(time.Now().Add((-10)*time.Minute),
					types.DimensionMinute)),
				TotalSuccessfulRescale: 1,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.InsertPodAutoscalerInfo(ctx, tt.metric, tt.opts)
			assert.Nil(t, err)
		})
	}
}

func Test_GetPodAutoscalerInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		req     *bcsdatamanager.GetPodAutoscalerRequest
		want    *bcsdatamanager.PodAutoscaler
		wantErr bool
	}{
		{
			name: "test1",
			req: &bcsdatamanager.GetPodAutoscalerRequest{
				ClusterID:         "testcluster1",
				Namespace:         "testnamespace1",
				Dimension:         types.DimensionMinute,
				PodAutoscalerType: types.HPAType,
				PodAutoscalerName: "testhpa",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.GetPodAutoscalerInfo(ctx, tt.req)
			assert.Nil(t, err)
		})
	}
}

func Test_GetPodAutoscalerList(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		req     *bcsdatamanager.GetPodAutoscalerListRequest
		want    []*bcsdatamanager.PodAutoscaler
		wantErr bool
	}{
		{
			name: "test1",
			req:  &bcsdatamanager.GetPodAutoscalerListRequest{},
		},
		{
			name: "test2",
			req: &bcsdatamanager.GetPodAutoscalerListRequest{
				Project:   "testproject2",
				Dimension: types.DimensionMinute,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := store.GetPodAutoscalerList(ctx, tt.req)
			assert.Nil(t, err)
		})
	}
}

func Test_GetRawPodAutoscalerInfo(t *testing.T) {

}

func Test_generateAutoscalerResponse(t *testing.T) {

}

func TestNewModelPodAutoscaler(t *testing.T) {

}

func Test_genPodAutoscalerListCond(t *testing.T) {

}
