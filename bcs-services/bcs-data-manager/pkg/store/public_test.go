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
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestModelPublic_GetRawPublicInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		opts    *common.JobCommonOpts
		wantErr bool
	}{
		{
			name: "test1",
			opts: &common.JobCommonOpts{
				ObjectType:  common.NamespaceType,
				ProjectID:   "testproject1",
				ClusterID:   "testcluster1",
				Namespace:   "testnamespace1",
				ClusterType: common.Kubernetes,
				Dimension:   common.DimensionMinute,
				CurrentTime: common.FormatTime(time.Now().Add((-10)*time.Minute), common.DimensionMinute),
			},
		},
		{
			name: "test2",
			opts: &common.JobCommonOpts{
				ObjectType:   common.WorkloadType,
				ProjectID:    "testproject1",
				ClusterID:    "testcluster1",
				Namespace:    "testnamespace1",
				WorkloadType: common.DeploymentType,
				Name:         "testdeploy1",
				ClusterType:  common.Kubernetes,
				Dimension:    common.DimensionMinute,
				CurrentTime:  common.FormatTime(time.Now().Add((-10)*time.Minute), common.DimensionMinute),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.GetRawPublicInfo(ctx, tt.opts)
			assert.Nil(t, err)
		})
	}
}

func TestModelPublic_InsertPublicInfo(t *testing.T) {
	store := newTestMongo()
	ctx := context.Background()
	tests := []struct {
		name    string
		opts    *common.JobCommonOpts
		metric  *common.PublicData
		wantErr bool
	}{
		{
			name: "test1",
			opts: &common.JobCommonOpts{
				ObjectType:  common.NamespaceType,
				ProjectID:   "testproject1",
				ClusterID:   "testcluster1",
				Namespace:   "testnamespace1",
				ClusterType: common.Kubernetes,
				Dimension:   common.DimensionMinute,
				CurrentTime: common.FormatTime(time.Now().Add((-10)*time.Minute), common.DimensionMinute),
			},
			metric: &common.PublicData{
				CreateTime:  primitive.NewDateTimeFromTime(time.Now()),
				UpdateTime:  primitive.NewDateTimeFromTime(time.Now()),
				ObjectType:  common.NamespaceType,
				ProjectID:   "testproject1",
				ClusterID:   "testcluster1",
				ClusterType: common.Kubernetes,
				Namespace:   "testnamespace1",
				Metrics: &common.NamespacePublicMetrics{
					ResourceLimit: nil,
					SuggestCPU:    200,
					SuggestMemory: 400,
				},
			},
		},
		{
			name: "test2",
			opts: &common.JobCommonOpts{
				ObjectType:   common.WorkloadType,
				ProjectID:    "testproject1",
				ClusterID:    "testcluster1",
				Namespace:    "testnamespace1",
				WorkloadType: common.DeploymentType,
				Name:         "testdeploy1",
				ClusterType:  common.Kubernetes,
				Dimension:    common.DimensionMinute,
				CurrentTime:  common.FormatTime(time.Now().Add((-10)*time.Minute), common.DimensionMinute),
			},
			metric: &common.PublicData{
				CreateTime:   primitive.NewDateTimeFromTime(time.Now()),
				UpdateTime:   primitive.NewDateTimeFromTime(time.Now()),
				ObjectType:   common.NamespaceType,
				ProjectID:    "testproject1",
				ClusterID:    "testcluster1",
				ClusterType:  common.Kubernetes,
				Namespace:    "testnamespace1",
				WorkloadType: common.DeploymentType,
				WorkloadName: "testdeploy1",
				Metrics: &common.WorkloadPublicMetrics{
					SuggestCPU:    2,
					SuggestMemory: 4,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.InsertPublicInfo(ctx, tt.metric, tt.opts)
			assert.Nil(t, err)
		})
	}

}
