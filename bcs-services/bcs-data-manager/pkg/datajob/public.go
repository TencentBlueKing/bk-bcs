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

package datajob

import (
	"context"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
)

// PublicDayPolicy public day
type PublicDayPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// NewPublicDayPolicy init public day policy
func NewPublicDayPolicy(getter metric.Server, store store.Server) *PublicDayPolicy {
	return &PublicDayPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// ImplementPolicy PublicDayPolicy implement
func (p *PublicDayPolicy) ImplementPolicy(ctx context.Context, opts *common.JobCommonOpts, clients *common.Clients) {
	bucket, _ := common.GetBucketTime(opts.CurrentTime.AddDate(0, 0, -1), common.DimensionDay)
	p.insertWorkloadPublic(ctx, opts, bucket)
}

func (p *PublicDayPolicy) insertWorkloadPublic(ctx context.Context, opts *common.JobCommonOpts, bucket string) {
	workloadList, err := p.store.GetRawWorkloadInfo(ctx, opts, bucket)
	if err != nil {
		blog.Errorf("do day public policy error:%v", err)
	}
	chPool := make(chan struct{}, 50)
	wg := sync.WaitGroup{}
	for key := range workloadList {
		wg.Add(1)
		chPool <- struct{}{}
		go func(key int) {
			defer wg.Done()
			maxCPUUsage := workloadList[key].MaxCPUUsageTime.Value
			maxMemoryUsage := workloadList[key].MaxMemoryUsageTime.Value
			maxCPU := workloadList[key].MaxCPUTime.Value
			maxMemory := workloadList[key].MaxMemoryTime.Value
			suggestCPU := maxCPUUsage * maxCPU * 2
			suggestMemory := maxMemoryUsage * maxMemory * 2
			workloadPublicMetric := &common.WorkloadPublicMetrics{
				SuggestCPU:    suggestCPU,
				SuggestMemory: suggestMemory,
			}
			workloadPublic := &common.PublicData{
				WorkloadName: workloadList[key].Name,
				WorkloadType: workloadList[key].WorkloadType,
				ObjectType:   common.WorkloadType,
				ClusterID:    workloadList[key].ClusterID,
				Namespace:    workloadList[key].Namespace,
				ClusterType:  workloadList[key].ClusterType,
				ProjectID:    workloadList[key].ProjectID,
				Metrics:      workloadPublicMetric,
			}
			workloadOpts := &common.JobCommonOpts{
				ObjectType:   common.WorkloadType,
				ProjectID:    workloadList[key].ProjectID,
				ClusterID:    workloadList[key].ClusterID,
				ClusterType:  workloadList[key].ClusterType,
				Namespace:    workloadList[key].Namespace,
				WorkloadType: workloadList[key].WorkloadType,
				Name:         workloadList[key].Name,
			}
			err := p.store.InsertPublicInfo(ctx, workloadPublic, workloadOpts)
			if err != nil {
				blog.Errorf("insert workload public data error: %v", err)
			}
			<-chPool
		}(key)
	}
	wg.Wait()
}
