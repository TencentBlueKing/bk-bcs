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

package bcs

import (
	"context"
	"fmt"
	"testing"
	"time"

	cachenum30 "github.com/num30/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

func TestGetProject(t *testing.T) {
	initConf()
	ctx := context.Background()

	project, err := GetProject(ctx, config.G.BCS, getTestProjectId())
	assert.NoError(t, err)
	assert.Equal(t, project.ProjectId, getTestProjectId())
}

func BenchmarkCacheNum30Set(b *testing.B) {
	b.StopTimer()
	cacheKey := fmt.Sprintf("bcs.GetProject:%s", "bluking")
	project := Project{
		Name:          "blueking",
		ProjectId:     "blueking",
		Code:          "1",
		CcBizID:       "1",
		Creator:       "joker",
		Kind:          "project",
		RawCreateTime: "20240220",
	}
	goCache := cachenum30.New[*Project](5*time.Second, 10*time.Second)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		goCache.Set(cacheKey, &project, cachenum30.DefaultExpiration)
	}
}

func BenchmarkCacheNum30Get(b *testing.B) {
	b.StopTimer()
	cacheKey := fmt.Sprintf("bcs.GetProject:%s", "bluking")
	project := Project{
		Name:          "blueking",
		ProjectId:     "blueking",
		Code:          "1",
		CcBizID:       "1",
		Creator:       "joker",
		Kind:          "project",
		RawCreateTime: "20240220",
	}
	goCache := cachenum30.New[*Project](5*time.Second, 10*time.Second)
	goCache.Set(cacheKey, &project, cachenum30.DefaultExpiration)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		goCache.Get(cacheKey)
	}
}
