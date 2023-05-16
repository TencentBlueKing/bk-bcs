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
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

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
