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

package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/promclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	bcstesting "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/testing"
)

func main() {
	ctx := context.Background()
	header := http.Header{}
	header.Add("Authorization", "Bearer "+config.G.BCS.Token)
	rawURL := config.G.BCS.Host + "/bcsapi/v4/monitor/query"
	promql := fmt.Sprintf(`bcs:cluster:cpu:total{cluster_id="%s"}`, bcstesting.GetTestClusterId())

	var (
		count    int64
		errCount int64
	)

	wg := &sync.WaitGroup{}

	c := 10
	for i := 0; i < c; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				atomic.AddInt64(&count, 1)
				result, err := promclient.QueryInstant(ctx, rawURL, header, promql, time.Now())
				if err != nil || len(result.Warnings) > 0 {
					atomic.AddInt64(&errCount, 1)
				}
				fmt.Println("count", atomic.LoadInt64(&count), "errCount", atomic.LoadInt64(&errCount))
				time.Sleep(time.Millisecond * 100)
			}
		}()
	}
	wg.Wait()
}
