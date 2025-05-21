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

// Package main
package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api/metrics/query"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	bcstesting "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/testing"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

func main() {
	ctx := context.Background()

	rawURL := config.G.BCS.Host + fmt.Sprintf("/bcsapi/v4/monitor/api/metrics/projects/%s/clusters/%s/overview",
		bcstesting.GetTestProjectId(), bcstesting.GetTestClusterId())

	var (
		count             int64
		errCount          int64
		respErrCount      int64
		unmarshalErrCount int64
		zeroValueErrCount int64
	)

	wg := &sync.WaitGroup{}

	c := 10
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				atomic.AddInt64(&count, 1)
				resp, err := component.GetClient().R().SetContext(ctx).SetHeaders(utils.GetLaneIDByCtx(ctx)).
					SetAuthToken(config.G.BCS.Token).Get(rawURL)
				if err != nil || resp.StatusCode() != http.StatusOK {
					atomic.AddInt64(&errCount, 1)
					atomic.AddInt64(&respErrCount, 1)
					continue
				}

				// 部分接口，如 usermanager 返回的content-type不是json, 需要手动Unmarshal
				result := new(query.ClusterOverviewMetric)
				if err := component.UnmarshalBKResult(resp, result); err != nil {
					atomic.AddInt64(&errCount, 1)
					atomic.AddInt64(&unmarshalErrCount, 1)
					continue
				}

				if result.CPUUsage.Total == "0" || result.CPUUsage.Used == "0" ||
					result.MemoryUsage.TotalByte == "0" || result.MemoryUsage.UsedByte == "0" ||
					result.DiskUsage.TotalByte == "0" || result.DiskUsage.UsedByte == "0" {
					atomic.AddInt64(&errCount, 1)
					atomic.AddInt64(&zeroValueErrCount, 1)
					continue
				}

				fmt.Println(
					"count", atomic.LoadInt64(&count),
					"errCount", atomic.LoadInt64(&errCount),
					"respErrCount", atomic.LoadInt64(&respErrCount),
					"unmarshalErrCount", atomic.LoadInt64(&unmarshalErrCount),
					"zeroValueErrCount", atomic.LoadInt64(&zeroValueErrCount),
				)
				time.Sleep(time.Millisecond * 100)
			}
		}()
	}
	wg.Wait()
}
