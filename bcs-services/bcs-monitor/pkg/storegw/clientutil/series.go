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
 *
 */

package clientutil

import (
	"sort"

	"github.com/prometheus/prometheus/prompb"
	"github.com/thanos-io/thanos/pkg/store/storepb"
)

// TimeSeries 时间序列封装
type TimeSeries struct {
	*prompb.TimeSeries
}

// AddLabels 添加 Labels
func (s *TimeSeries) AddLabels(promLabels []prompb.Label) *TimeSeries {
	s.Labels = append(s.Labels, promLabels...)
	return s
}

// AddLabel :
func (s *TimeSeries) AddLabel(name, value string) *TimeSeries {
	s.Labels = append(s.Labels, prompb.Label{Name: name, Value: value})
	return s
}

// RenameLabel 重命名一个label, 如果有重复label，以后面一个为准
func (s *TimeSeries) RenameLabel(oldname, newName string) *TimeSeries {
	for _, label := range s.TimeSeries.Labels {
		if label.Name == oldname {
			s.Labels = append(s.Labels, prompb.Label{Name: newName, Value: label.Value})
		}
	}
	return s
}

// ToThanosSeries 转换为
func (s *TimeSeries) ToThanosSeries(skipChunks bool) (*storepb.Series, error) {
	// 返回的点需要按时间排序
	sort.Slice(s.Samples, func(i, j int) bool {
		return s.Samples[i].Timestamp < s.Samples[j].Timestamp
	})

	lset := make([]storepb.Label, 0, len(s.Labels))

	for _, v := range s.Labels {
		lset = append(lset, storepb.Label{Name: v.Name, Value: v.Value})
	}

	series := &storepb.Series{Labels: lset}
	// 不需要 chunks 数据, series 接口场景
	if !skipChunks {
		aggregatedChunks, err := ChunkSamples(s.TimeSeries)
		if err != nil {
			return nil, err
		}
		series.Chunks = aggregatedChunks
	}

	return series, nil
}
