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
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/prompb"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"github.com/thanos-io/thanos/pkg/testutil"
)

func TestGetCluterID(t *testing.T) {
	sr := prompb.TimeSeries{
		Labels: []prompb.Label{
			{Name: "test", Value: "test_1"},
		},
	}
	clusterID := GetCluterID(&sr)
	testutil.Assert(t, clusterID == "", "cluster_id not exist")

	sr = prompb.TimeSeries{
		Labels: []prompb.Label{
			{Name: "test", Value: "test_1"},
			{Name: "cluster_id", Value: "test_1"},
		},
	}
	clusterID = GetCluterID(&sr)
	testutil.Assert(t, clusterID == "test_1", "get cluster_id failed")
}

func TestGetCluterList(t *testing.T) {
	sr := prompb.TimeSeries{
		Labels: []prompb.Label{
			{Name: "test", Value: "test_1"},
		},
	}

	sr1 := prompb.TimeSeries{
		Labels: []prompb.Label{
			{Name: "test", Value: "test_1"},
			{Name: "cluster_id", Value: "test_1"},
		},
	}

	sr2 := prompb.TimeSeries{
		Labels: []prompb.Label{
			{Name: "test", Value: "test_1"},
			{Name: "cluster_id", Value: "test_2"},
		},
	}

	clusterList := GetClusterList([]*prompb.TimeSeries{&sr, &sr1, &sr2})

	testutil.Assert(t, len(clusterList) == 2, "get cluster_id length")
	t.Log(clusterList)
}

func TestTimeSeriesToHash(t *testing.T) {
	series := prompb.TimeSeries{
		Labels: []prompb.Label{
			{Name: "test", Value: "test_1"},
			{Name: "cluster_id", Value: "test_1"},
		},
	}
	hash := TimeSeriesToHash(&series)
	t.Log(hash)
	testutil.Assert(t, hash == 1705576646752493466)

	hash2 := labels.FromMap(map[string]string{
		"cluster_id": "test_1",
		"test":       "test_1",
	}).Hash()
	t.Log(hash2)

	testutil.Assert(t, hash == hash2, "not equal")
}

func TestMergeTimeSeriesMap(t *testing.T) {
	seriesA := prompb.TimeSeries{
		Labels: []prompb.Label{
			{Name: "test", Value: "test_1"},
			{Name: "cluster_id", Value: "test_1"},
		},
		Samples: []prompb.Sample{
			{Timestamp: 123, Value: 23},
			{Timestamp: 1234, Value: 234},
		},
	}
	seriesB := prompb.TimeSeries{
		Labels: []prompb.Label{
			{Name: "test", Value: "test_1"},
			{Name: "cluster_id", Value: "test_1"},
		},
		Samples: []prompb.Sample{
			{Timestamp: 1231, Value: 23},
			{Timestamp: 12342, Value: 234},
		},
	}
	seriesMap := map[uint64]*prompb.TimeSeries{
		1: &seriesA,
	}

	toBeMerged := map[uint64]*prompb.TimeSeries{
		2: &seriesB,
	}
	MergeTimeSeriesMap(seriesMap, toBeMerged)
	t.Log(seriesMap)
	testutil.Assert(t, len(seriesMap) == 2)
	testutil.Assert(t, len(seriesMap[1].Samples) == 2)

	toBeMerged = map[uint64]*prompb.TimeSeries{
		1: &seriesB,
	}
	MergeTimeSeriesMap(seriesMap, toBeMerged)
	t.Log(seriesMap)
	testutil.Assert(t, len(seriesMap) == 2)
	testutil.Assert(t, len(seriesMap[1].Samples) == 4)
}

func TestGetLabelMatchValues(t *testing.T) {
	matchers := []storepb.LabelMatcher{
		{Name: "namespace", Value: "gameai", Type: storepb.LabelMatcher_EQ},
		{Name: "pod_name", Value: "test1|ldi", Type: storepb.LabelMatcher_EQ},
	}
	values, err := GetLabelMatchValues("pod_name", matchers)
	testutil.Ok(t, err)
	testutil.Assert(t, len(values) == 1)
	testutil.Assert(t, values[0] == "test1|ldi")

	matchers = []storepb.LabelMatcher{
		{Name: "namespace", Value: "gameai", Type: storepb.LabelMatcher_EQ},
		{Name: "pod_name", Value: "test1|ldi", Type: storepb.LabelMatcher_RE},
	}
	values, err = GetLabelMatchValues("pod_name", matchers)
	testutil.Ok(t, err)
	testutil.Assert(t, len(values) == 2)
	testutil.Assert(t, values[0] == "test1" && values[1] == "ldi")

	_, err = GetLabelMatchValue("pod_name", matchers)
	testutil.NotOk(t, err)
}
