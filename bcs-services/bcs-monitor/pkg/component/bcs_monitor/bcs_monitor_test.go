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

package bcsmonitor

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	bcstesting "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/testing"
	"github.com/stretchr/testify/assert"
)

func TestQueryInstant(t *testing.T) {
	promql := fmt.Sprintf(`up{cluster_id="%s"}`, bcstesting.GetTestClusterId())
	vector, warnings, err := QueryInstantVector(context.Background(), bcstesting.GetTestProjectId(), promql, nil, time.Now())
	assert.NoError(t, err)
	fmt.Println(vector, warnings)
}

func TestQueryRange(t *testing.T) {
	promql := fmt.Sprintf(`up{cluster_id="%s"}`, bcstesting.GetTestClusterId())
	end := time.Now()
	start := end.Add(-time.Minute * 5)
	vector, warnings, err := QueryRangeMatrix(context.Background(), bcstesting.GetTestProjectId(), promql, nil, start, end, time.Minute)
	assert.NoError(t, err)
	fmt.Println(vector, warnings)
}

func TestQueryMultiValues(t *testing.T) {
	promqlMap := map[string]string{
		"up1": fmt.Sprintf(`up{cluster_id="%s"}`, bcstesting.GetTestClusterId()),
		"up2": fmt.Sprintf(`up{cluster_id="%s"}`, bcstesting.GetTestClusterId()),
	}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			vector, err := QueryMultiValues(context.Background(), bcstesting.GetTestProjectId(), promqlMap, nil, time.Now())
			assert.NoError(t, err)
			fmt.Println(vector)
		}()
	}
	wg.Wait()
}
