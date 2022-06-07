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

package datajob

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/mock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClusterDayImplementPolicy(t *testing.T) {
	storeServer := mock.NewMockStore()
	metricGetter := mock.NewMockMetric()
	minutePolicy := NewClusterMinutePolicy(metricGetter, storeServer)
	hourPolicy := NewClusterHourPolicy(metricGetter, storeServer)
	dayPolicy := NewClusterDayPolicy(metricGetter, storeServer)
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
	cmCli := &cmanager.ClusterManagerClientWithHeader{
		Cli: mock.NewMockCm(),
		Ctx: ctx,
	}
	client := &types.Clients{
		MonitorClient:   nil,
		K8sStorageCli:   mock.NewMockStorage(),
		MesosStorageCli: mock.NewMockStorage(),
		CmCli:           cmCli,
	}
	minutePolicy.ImplementPolicy(ctx, minuteOpts, client)
	hourPolicy.ImplementPolicy(ctx, hourOpts, client)
	dayPolicy.ImplementPolicy(ctx, dayOpts, client)
	bucket, err := utils.GetBucketTime(dayOpts.CurrentTime, types.DimensionHour)
	assert.Nil(t, err)
	cluster, err := storeServer.GetRawClusterInfo(ctx, dayOpts, bucket)
	assert.Nil(t, err)
	assert.NotNil(t, cluster)
}

func TestClusterHourImplementPolicy(t *testing.T) {
	storeServer := mock.NewMockStore()
	hourPolicy := NewClusterHourPolicy(mock.NewMockMetric(), storeServer)
	minutePolicy := NewClusterMinutePolicy(mock.NewMockMetric(), storeServer)
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
	cmCli := &cmanager.ClusterManagerClientWithHeader{
		Cli: mock.NewMockCm(),
		Ctx: ctx,
	}
	client := &types.Clients{
		MonitorClient:   nil,
		K8sStorageCli:   mock.NewMockStorage(),
		MesosStorageCli: mock.NewMockStorage(),
		CmCli:           cmCli,
	}
	minutePolicy.ImplementPolicy(ctx, minuteOpts, client)
	hourPolicy.ImplementPolicy(ctx, hourOpts, client)
	bucket, err := utils.GetBucketTime(hourOpts.CurrentTime, types.DimensionHour)
	assert.Nil(t, err)
	cluster, err := storeServer.GetRawClusterInfo(ctx, hourOpts, bucket)
	assert.Nil(t, err)
	assert.NotNil(t, cluster)
}

func TestClusterMinuteImplementPolicy(t *testing.T) {
	storeServer := mock.NewMockStore()
	minutePolicy := NewClusterMinutePolicy(mock.NewMockMetric(), storeServer)
	ctx := context.Background()
	opts := &types.JobCommonOpts{
		ObjectType:  types.ClusterType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionMinute),
	}
	cmCli := &cmanager.ClusterManagerClientWithHeader{
		Cli: mock.NewMockCm(),
		Ctx: ctx,
	}
	client := &types.Clients{
		MonitorClient:   nil,
		K8sStorageCli:   mock.NewMockStorage(),
		MesosStorageCli: mock.NewMockStorage(),
		CmCli:           cmCli,
	}
	minutePolicy.ImplementPolicy(ctx, opts, client)
	bucket, err := utils.GetBucketTime(opts.CurrentTime, types.DimensionMinute)
	assert.Nil(t, err)
	cluster, err := storeServer.GetRawClusterInfo(ctx, opts, bucket)
	assert.Nil(t, err)
	assert.NotNil(t, cluster)
}
