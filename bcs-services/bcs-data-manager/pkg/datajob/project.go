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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProjectDayPolicy project day
type ProjectDayPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// NewProjectDayPolicy init day policy
func NewProjectDayPolicy(getter metric.Server, store store.Server) *ProjectDayPolicy {
	return &ProjectDayPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// ImplementPolicy project day implement
func (p *ProjectDayPolicy) ImplementPolicy(ctx context.Context, opts *common.JobCommonOpts, clients *common.Clients) {
	bucketTime, err := common.GetBucketTime(opts.CurrentTime.AddDate(0, 0, -1), common.DimensionHour)
	if err != nil {
		blog.Errorf("do day project policy error, get bucket err:%v", err)
		return
	}
	hourOpts := &common.JobCommonOpts{
		ObjectType: common.ClusterType,
		ProjectID:  opts.ProjectID,
		Dimension:  common.DimensionHour,
	}
	clusters, err := p.store.GetRawClusterInfo(ctx, hourOpts, bucketTime)
	if err != nil {
		blog.Errorf("do day project policy error, err:%v", err)
		return
	}
	if len(clusters) == 0 {
		blog.Errorf("do day project policy error, the length of  clusters is 0")
		return
	}
	nodeCount, availableNode := p.calculateProjectNodeCount(clusters)
	totalCPU, loadCPU := p.CalculateCpu(clusters)
	totalMemory, loadMemory := p.calculateMemory(clusters)
	projectMetric := &common.ProjectMetrics{
		Index:              common.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:               primitive.NewDateTimeFromTime(opts.CurrentTime),
		ClustersCount:      int64(len(clusters)),
		TotalCPU:           totalCPU,
		TotalMemory:        totalMemory,
		TotalLoadCPU:       loadCPU,
		TotalLoadMemory:    loadMemory,
		AvgLoadCPU:         loadCPU / float64(len(clusters)),
		AvgLoadMemory:      loadMemory / int64(len(clusters)),
		CPUUsage:           loadCPU / totalCPU,
		MemoryUsage:        float64(loadMemory) / float64(totalMemory),
		NodeCount:          nodeCount,
		AvailableNodeCount: availableNode,
		MinNode:            nil,
		MaxNode:            nil,
	}
	err = p.store.InsertProjectInfo(ctx, projectMetric, opts)
	if err != nil {
		blog.Errorf("do day project policy error, err:%v", err)
	}
}

// CalculateCpu calculate cpu
func (p *ProjectDayPolicy) CalculateCpu(clusters []*common.ClusterData) (float64, float64) {
	var total, load float64
	for key := range clusters {
		if len(clusters[key].Metrics) != 0 {
			length := len(clusters[key].Metrics)
			total += clusters[key].Metrics[length-1].TotalCPU
			load += clusters[key].Metrics[length-1].TotalLoadCPU
		}
	}
	return total, load
}

func (p *ProjectDayPolicy) calculateMemory(clusters []*common.ClusterData) (int64, int64) {
	var total, load int64
	for key := range clusters {
		if len(clusters[key].Metrics) != 0 {
			length := len(clusters[key].Metrics)
			total += clusters[key].Metrics[length-1].TotalMemory
			load += clusters[key].Metrics[length-1].TotalLoadMemory
		}
	}
	return total, load
}

func (p *ProjectDayPolicy) calculateProjectNodeCount(clusters []*common.ClusterData) (int64, int64) {
	var nodeCount, availableNode int64
	for key := range clusters {
		if len(clusters[key].Metrics) != 0 {
			length := len(clusters[key].Metrics)
			nodeCount += clusters[key].Metrics[length-1].NodeCount
			availableNode += clusters[key].Metrics[length-1].AvailableNodeCount
		}
	}
	return nodeCount, availableNode
}
