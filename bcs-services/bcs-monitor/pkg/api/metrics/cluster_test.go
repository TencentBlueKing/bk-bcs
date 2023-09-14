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

package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	bcstesting "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/testing"
)

func TestGetClusterOverview(t *testing.T) {
	c := &rest.Context{
		ClusterId: bcstesting.GetTestProjectId(),
	}

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			vector, err := GetClusterOverview(c)
			assert.NoError(t, err)
			fmt.Println(vector)
		}()
	}
	wg.Wait()
}

// QueryRange的基准测试，需要启动query, storegw
// 需要配置环境变量 eg:export TEST_CONFIG_FILE=D:/project_config/bcs-monitor/bcs_monitor.yml
func BenchmarkQueryRange(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		header := http.Header{}
		// query配置的启动端口
		rawURL := "http://127.0.0.1:10902"
		// 前提是先把数据打入prometheus，才能查得到
		promql := "myMetric{cluster_id=\"1\", provider=\"PROMETHEUS\", bcs_cluster_id=\"1\"}"
		endTime := time.Now()
		startTime := endTime.Add(-time.Hour)
		// 默认只返回 60 个点
		stepTime := endTime.Sub(startTime) / 60
		_, err := promclient.QueryRange(ctx, rawURL, header, promql, startTime, endTime, stepTime)
		if err != nil {
			fmt.Println(err)
		}
	}
}
