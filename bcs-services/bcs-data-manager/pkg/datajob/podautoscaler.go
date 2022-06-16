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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// PodAutoscalerDayPolicy podAutoscaler day policy
type PodAutoscalerDayPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// PodAutoscalerHourPolicy podAutoscaler hour policy
type PodAutoscalerHourPolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// PodAutoscalerMinutePolicy podAutoscaler minute policy
type PodAutoscalerMinutePolicy struct {
	MetricGetter metric.Server
	store        store.Server
}

// NewPodAutoscalerDayPolicy init day policy
func NewPodAutoscalerDayPolicy(getter metric.Server, store store.Server) *PodAutoscalerDayPolicy {
	return &PodAutoscalerDayPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// NewPodAutoscalerHourPolicy init hour policy
func NewPodAutoscalerHourPolicy(getter metric.Server, store store.Server) *PodAutoscalerHourPolicy {
	return &PodAutoscalerHourPolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// NewPodAutoscalerMinutePolicy init minute policy
func NewPodAutoscalerMinutePolicy(getter metric.Server, store store.Server) *PodAutoscalerMinutePolicy {
	return &PodAutoscalerMinutePolicy{
		MetricGetter: getter,
		store:        store,
	}
}

// ImplementPolicy implement PodAutoscalerDayPolicy
func (p *PodAutoscalerDayPolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts,
	clients *types.Clients) {
	hourOpts := &types.JobCommonOpts{
		ProjectID:         opts.ProjectID,
		ClusterID:         opts.ClusterID,
		Namespace:         opts.Namespace,
		Dimension:         types.DimensionHour,
		PodAutoscalerName: opts.PodAutoscalerName,
		PodAutoscalerType: opts.PodAutoscalerType,
	}
	bucket, _ := utils.GetBucketTime(opts.CurrentTime.AddDate(0, 0, -1), types.DimensionHour)
	hourMetrics, err := p.store.GetRawPodAutoscalerInfo(ctx, hourOpts, bucket)
	if err != nil {
		blog.Errorf("do pod autoscaler day policy failed, get metrics err:%v", err)
		return
	} else if len(hourMetrics) != 1 {
		blog.Errorf("do pod autoscaler day policy failed, metric length not equal 1, metrics:%v", hourMetrics)
		return
	}
	hourMetric := hourMetrics[0]
	dayMetric := &types.PodAutoscalerMetrics{
		Index:                  utils.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:                   primitive.NewDateTimeFromTime(utils.FormatTime(opts.CurrentTime, opts.Dimension)),
		TotalSuccessfulRescale: hourMetric.Total,
	}
	if err := p.store.InsertPodAutoscalerInfo(ctx, dayMetric, opts); err != nil {
		blog.Errorf("do pod autoscaler day policy error, opts: %v, err: %v", opts, err)
	}
}

// ImplementPolicy implement PodAutoscalerHourPolicy
func (p *PodAutoscalerHourPolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts,
	clients *types.Clients) {
	minuteOpts := &types.JobCommonOpts{
		ProjectID:         opts.ProjectID,
		ClusterID:         opts.ClusterID,
		Namespace:         opts.Namespace,
		Dimension:         types.DimensionMinute,
		PodAutoscalerName: opts.PodAutoscalerName,
		PodAutoscalerType: opts.PodAutoscalerType,
	}
	bucket, _ := utils.GetBucketTime(opts.CurrentTime.Add((-1)*time.Hour), types.DimensionMinute)
	minuteMetrics, err := p.store.GetRawPodAutoscalerInfo(ctx, minuteOpts, bucket)
	if err != nil {
		blog.Errorf("do pod autoscaler hour policy failed, get metrics err:%v", err)
		return
	} else if len(minuteMetrics) != 1 {
		blog.Errorf("do pod autoscaler hour policy failed, get metrics err, length not equal 1, metrics:%v", minuteMetrics)
		return
	}
	minuteMetric := minuteMetrics[0]
	hourMetric := &types.PodAutoscalerMetrics{
		Index:                  utils.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:                   primitive.NewDateTimeFromTime(utils.FormatTime(opts.CurrentTime, opts.Dimension)),
		TotalSuccessfulRescale: minuteMetric.Total,
	}
	if err := p.store.InsertPodAutoscalerInfo(ctx, hourMetric, opts); err != nil {
		blog.Errorf("do pod autoscaler hour policy error, opts: %v, err: %v", opts, err)
	}
}

// ImplementPolicy implement PodAutoscalerMinutePolicy
func (p *PodAutoscalerMinutePolicy) ImplementPolicy(ctx context.Context, opts *types.JobCommonOpts,
	clients *types.Clients) {
	blog.Infof("pa opts:%v", opts)
	count, err := p.MetricGetter.GetPodAutoscalerCount(opts, clients)
	if err != nil {
		blog.Error("do PodAutoscalerMinutePolicy failed, get pod autoscaler count error:%v", err)
		return
	}
	blog.Infof("autoscaler count:%d", count)
	minuteMetric := &types.PodAutoscalerMetrics{
		Index:                  utils.GetIndex(opts.CurrentTime, opts.Dimension),
		Time:                   primitive.NewDateTimeFromTime(utils.FormatTime(opts.CurrentTime, opts.Dimension)),
		TotalSuccessfulRescale: count,
	}
	if err := p.store.InsertPodAutoscalerInfo(ctx, minuteMetric, opts); err != nil {
		blog.Errorf("do pod autoscaler minute policy error, opts: %v, err: %v", opts, err)
	}
}
