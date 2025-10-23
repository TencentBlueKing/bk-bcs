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

package bkmonitor

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thanos-io/thanos/pkg/store/storepb"

	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/testing"
)

func TestMarshalJSON(t *testing.T) {
	s := Sample{Timestamp: 1653811620000, Value: 771479351.231}
	b, err := json.Marshal(s)
	assert.NoError(t, err)
	assert.True(t, len(b) > 1)
}

func TestQueryByPromQL(t *testing.T) {
	ctx := context.Background()

	matchers := []storepb.LabelMatcher{
		{Type: storepb.LabelMatcher_EQ, Name: "bcs_cluster_id", Value: "BCS-K8S-00000"},
		{Type: storepb.LabelMatcher_EQ, Name: "__name__", Value: "container_network_receive_bytes_total"},
		{Type: storepb.LabelMatcher_RE, Name: "pod_name", Value: "kube-apiserver.*"},
	}

	step := int64(60)
	end := time.Now().Unix()
	start := end - 60

	rawURL := os.Getenv("BK_MONITOR_URL")

	series, err := QueryByPromQL(ctx, rawURL, "2", "system", start, end, step, matchers, "")
	assert.NoError(t, err)
	assert.True(t, len(series) > 1)
}
