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

func TestNamespaceDayImplementPolicy(t *testing.T) {
	storeServer := mock.NewMockStore()
	minutePolicy := NewNamespaceMinutePolicy(mock.NewMockMetric(), storeServer)
	hourPolicy := NewNamespaceHourPolicy(mock.NewMockMetric(), storeServer)
	dayPolicy := NewNamespaceDayPolicy(mock.NewMockMetric(), storeServer)
	ctx := context.Background()
	minuteOpts := &types.JobCommonOpts{
		ObjectType:  types.NamespaceType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		Namespace:   "testNs",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionMinute),
	}
	hourOpts := &types.JobCommonOpts{
		ObjectType:  types.NamespaceType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		Namespace:   "testNs",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionHour,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionHour),
	}
	dayOpts := &types.JobCommonOpts{
		ObjectType:  types.NamespaceType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		Namespace:   "testNs",
		ClusterType: types.Kubernetes,
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
	bucket, err := utils.GetBucketTime(dayOpts.CurrentTime, types.DimensionDay)
	assert.Nil(t, err)
	namespace, err := storeServer.GetRawNamespaceInfo(ctx, dayOpts, bucket)
	assert.Nil(t, err)
	assert.NotNil(t, namespace)
}

func TestNamespaceHourImplementPolicy(t *testing.T) {
	storeServer := mock.NewMockStore()
	hourPolicy := NewNamespaceHourPolicy(mock.NewMockMetric(), storeServer)
	minutePolicy := NewNamespaceMinutePolicy(mock.NewMockMetric(), storeServer)
	ctx := context.Background()
	minuteOpts := &types.JobCommonOpts{
		ObjectType:  types.NamespaceType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		Namespace:   "testNs",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionMinute),
	}
	hourOpts := &types.JobCommonOpts{
		ObjectType:  types.NamespaceType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		Namespace:   "testNs",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionHour,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionHour),
	}

	client := &types.Clients{
		MonitorClient:   nil,
		K8sStorageCli:   mock.NewMockStorage(),
		MesosStorageCli: mock.NewMockStorage(),
	}
	minutePolicy.ImplementPolicy(ctx, minuteOpts, client)
	hourPolicy.ImplementPolicy(ctx, hourOpts, client)
	bucket, err := utils.GetBucketTime(hourOpts.CurrentTime, types.DimensionHour)
	assert.Nil(t, err)
	namespace, err := storeServer.GetRawNamespaceInfo(ctx, hourOpts, bucket)
	assert.Nil(t, err)
	assert.NotNil(t, namespace)
	assert.NotEqual(t, 0, len(namespace))
}

func TestNamespaceMinuteImplementPolicy(t *testing.T) {
	storeServer := mock.NewMockStore()
	minutePolicy := NewNamespaceMinutePolicy(mock.NewMockMetric(), storeServer)
	ctx := context.Background()
	opts := &types.JobCommonOpts{
		ObjectType:  types.NamespaceType,
		ProjectID:   "testProject",
		ClusterID:   "testCluster",
		Namespace:   "testNs",
		ClusterType: types.Kubernetes,
		Dimension:   types.DimensionMinute,
		CurrentTime: utils.FormatTime(time.Now(), types.DimensionMinute),
	}

	client := &types.Clients{
		MonitorClient:   nil,
		K8sStorageCli:   mock.NewMockStorage(),
		MesosStorageCli: mock.NewMockStorage(),
	}
	minutePolicy.ImplementPolicy(ctx, opts, client)
	bucket, err := utils.GetBucketTime(opts.CurrentTime, types.DimensionMinute)
	assert.Nil(t, err)
	namespace, err := storeServer.GetRawNamespaceInfo(ctx, opts, bucket)
	assert.Nil(t, err)
	assert.NotNil(t, namespace)
	assert.NotEqual(t, 0, len(namespace))
}
