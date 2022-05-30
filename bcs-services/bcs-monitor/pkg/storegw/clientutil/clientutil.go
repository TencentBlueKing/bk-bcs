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
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/tsdb/chunkenc"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ChunkSamples 按120个点，分割为chunk
func ChunkSamples(series *prompb.TimeSeries, maxSamplesPerChunk int) (chks []storepb.AggrChunk, err error) {
	samples := series.Samples

	for len(samples) > 0 {
		chunkSize := len(samples)
		if chunkSize > maxSamplesPerChunk {
			chunkSize = maxSamplesPerChunk
		}

		enc, cb, err := EncodeChunk(samples[:chunkSize])
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}

		chks = append(chks, storepb.AggrChunk{
			MinTime: int64(samples[0].Timestamp),
			MaxTime: int64(samples[chunkSize-1].Timestamp),
			Raw:     &storepb.Chunk{Type: enc, Data: cb},
		})

		samples = samples[chunkSize:]
	}
	return chks, nil
}

// EncodeChunk :
func EncodeChunk(ss []prompb.Sample) (storepb.Chunk_Encoding, []byte, error) {
	c := chunkenc.NewXORChunk()

	a, err := c.Appender()
	if err != nil {
		return 0, nil, err
	}
	for _, s := range ss {
		a.Append(s.Timestamp, s.Value)
	}
	return storepb.Chunk_XOR, c.Bytes(), nil
}

// DumpPromQL 组装matcherpromql，打印日志用
func DumpPromQL(req *storepb.SeriesRequest) string {
	return LabelMatcherToString(req.Matchers)
}

// LabelMatcherToString LabelMatcher 组装字符串
func LabelMatcherToString(matchers []storepb.LabelMatcher) string {
	b := strings.Builder{}
	for i, m := range matchers {
		if i != 0 {
			b.WriteRune(',')
		}
		t := "="
		switch m.Type {
		case storepb.LabelMatcher_NEQ:
			t = "!="
		case storepb.LabelMatcher_RE:
			t = "=~"
		case storepb.LabelMatcher_NRE:
			t = "!~"
		}
		fmt.Fprintf(&b, "%s%s%q", m.Name, t, m.Value)
	}
	return fmt.Sprintf("{%v}", &b)
}

// GetCluterID : 获取集群id value
func GetCluterID(series *prompb.TimeSeries) string {
	for _, v := range series.Labels {
		if v.Name == "cluster_id" {
			return v.Value
		}
	}
	return ""
}

// GetClusterList : 获取去除后的集群 ID 列表
func GetClusterList(series []*prompb.TimeSeries) []string {
	// 构造 map, 获取去除后的集群 ID
	clusterMap := map[string]struct{}{}
	for _, s := range series {
		clusterID := GetCluterID(s)
		if clusterID == "" {
			continue
		}
		clusterMap[clusterID] = struct{}{}
	}

	// map 转换为列表返回
	clusterList := make([]string, 0, len(clusterMap))
	for k := range clusterMap {
		clusterList = append(clusterList, k)
	}

	return clusterList
}

func matchMetricNames(m *labels.Matcher, metricNames []string) bool {
	for _, name := range metricNames {
		if m.Matches(name) {
			return true
		}
	}
	return false
}

// GetEQMatcherMap 获取 = 匹配关系
func GetEQMatcherMap(matchers []storepb.LabelMatcher) map[string]string {
	eqMatcherMap := map[string]string{}
	for _, m := range matchers {
		if m.Type == storepb.LabelMatcher_EQ {
			eqMatcherMap[m.Name] = m.Value
		}
	}
	return eqMatcherMap
}

// LabelsToSignature 签名, 拼接为 a=b&c=d 的格式
func LabelsToSignature(labels map[string]string) string {
	labelNames := make([]string, 0, len(labels))
	for labelName := range labels {
		labelNames = append(labelNames, labelName)
	}
	sort.Strings(labelNames)

	b := strings.Builder{}
	for i, labelName := range labelNames {
		if i != 0 {
			b.WriteRune('&')
		}
		fmt.Fprintf(&b, "%s=%s", labelName, labels[labelName])
	}

	return fmt.Sprintf("%v", &b)
}

// CopyMap 复制 map 对象
func CopyMap(m map[string]string) map[string]string {
	targetM := map[string]string{}
	for k, v := range m {
		targetM[k] = v
	}
	return targetM
}

// CleanStrList 去掉空字符串, 去掉字符串2边的空格
func CleanStrList(strList []string) []string {
	targetStrList := []string{}
	for _, str := range strList {
		s := strings.TrimSpace(str)
		if s == "" {
			continue
		}
		targetStrList = append(targetStrList, s)
	}
	return targetStrList
}

// ValidateSeries series是否合法
func ValidateSeries(s *prompb.TimeSeries, minTime int64, maxTime int64) bool {
	// 任何为空, 都不符合
	if len(s.Samples) == 0 || len(s.Labels) == 0 {
		return false
	}

	return true
}

// ExtendLabelSetByNames 通过 metrics 名称生成 store 的 labelsSet,
func ExtendLabelSetByNames(externalLabels labels.Labels, metricNames []string) []storepb.LabelSet {
	labelSets := make([]storepb.LabelSet, 0, len(metricNames))

	for _, name := range metricNames {
		lbSet := storepb.LabelSet{Labels: []storepb.Label{
			{Name: labels.MetricName, Value: name},
		}}

		for _, l := range externalLabels {
			lbSet.Labels = append(lbSet.Labels, storepb.Label{Name: l.Name, Value: l.Value})
		}

		labelSets = append(labelSets, lbSet)
	}

	return labelSets
}

// TimeSeriesToHash 计算hash签名
func TimeSeriesToHash(series *prompb.TimeSeries) uint64 {
	lb := map[string]string{}
	for _, v := range series.Labels {
		lb[v.Name] = v.Value
	}
	return labels.FromMap(lb).Hash()
}

// QueryResultToHashMap Result 转换为 hashMap
func QueryResultToHashMap(result *prompb.QueryResult) map[uint64]*prompb.TimeSeries {
	hashMap := map[uint64]*prompb.TimeSeries{}
	for _, series := range result.Timeseries {
		hash := TimeSeriesToHash(series)
		hashMap[hash] = series
	}
	return hashMap
}

// MergeTimeSeriesMap 合并2个时间序列 seriesMap 会被修改的变量
func MergeTimeSeriesMap(seriesMap map[uint64]*prompb.TimeSeries, toBeMerged map[uint64]*prompb.TimeSeries) {
	for hash, series := range toBeMerged {
		_, ok := seriesMap[hash]
		if !ok {
			seriesMap[hash] = series
		} else {
			seriesMap[hash].Samples = append(seriesMap[hash].Samples, series.Samples...)
		}
	}
	return
}

// GetLabelMatch
func GetLabelMatch(name string, matchers []storepb.LabelMatcher) *storepb.LabelMatcher {
	// 可能存在多个名称相同的 LabelMatch, prom解析为且的关系, 因为这里只支持=, =~, 可忽略这种 case
	for _, m := range matchers {
		if m.Name == name {
			return &m
		}
	}
	return nil
}

// GetLabelMatchValues
func GetLabelMatchValues(name string, matchers []storepb.LabelMatcher) ([]string, error) {
	m := GetLabelMatch(name, matchers)
	if m == nil {
		return []string{}, nil
	}

	if m.Type == storepb.LabelMatcher_EQ {
		return []string{m.Value}, nil
	}

	if m.Type == storepb.LabelMatcher_RE {
		return strings.Split(m.Value, "|"), nil
	}

	// 不支持 "不等于", "正则不等于" 2 个匹配规则
	return []string{}, errors.Errorf("Not support match type: %s", m.Type)
}

// GetLabelMatchValue
func GetLabelMatchValue(name string, matchers []storepb.LabelMatcher) (string, error) {
	m := GetLabelMatch(name, matchers)
	if m == nil {
		return "", nil
	}

	if m.Type == storepb.LabelMatcher_EQ {
		return m.Value, nil
	}

	if m.Type == storepb.LabelMatcher_RE {
		return m.Value, nil
	}

	// 不支持 "不等于", "正则不等于" 2 个匹配规则
	return "", errors.Errorf("Not support match type: %s", m.Type)
}
