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

package query

import (
	"math"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/promql"
	"github.com/thanos-io/thanos/pkg/compact/downsample"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/extprom"
	"github.com/thanos-io/thanos/pkg/query"
	"github.com/thanos-io/thanos/pkg/store"
)

// NewQueryableCreator xxx
func NewQueryableCreator(reg *prometheus.Registry, logKit blog.GlogKit,
	discoveryClient *DiscoveryClient) query.QueryableCreator {
	proxy := store.NewProxyStore(logKit, reg, discoveryClient.Endpoints().GetStoreClients, component.Query, nil,
		storeResponseTimeout)

	queryableCreator := query.NewQueryableCreator(
		logKit,
		extprom.WrapRegistererWithPrefix("bcs_monitor_query_", reg),
		proxy,
		maxConcurrentSelects,
		queryTimeout,
	)
	return queryableCreator
}

// NewQueryEngine xxx
func NewQueryEngine(reg *prometheus.Registry, logKit blog.GlogKit) func(int64) *promql.Engine {
	engineOpts := promql.EngineOpts{
		Logger:        logKit,
		Reg:           reg,
		MaxSamples:    math.MaxInt32,
		Timeout:       queryTimeout,
		LookbackDelta: lookbackDelta,
		NoStepSubqueryIntervalFn: func(int64) int64 {
			return defaultEvaluationInterval.Milliseconds()
		},
	}
	return engineFactory(promql.NewEngine, engineOpts, dynamicLookbackDelta)
}

// engineFactory creates from 1 to 3 promql.Engines depending on
// dynamicLookbackDelta and eo.LookbackDelta and returns a function
// that returns appropriate engine for given maxSourceResolutionMillis.
//
// instead of creating several Engines here.
func engineFactory(
	newEngine func(promql.EngineOpts) *promql.Engine,
	eo promql.EngineOpts,
	dynamicLookbackDelta bool,
) func(int64) *promql.Engine {
	resolutions := []int64{downsample.ResLevel0}
	if dynamicLookbackDelta {
		resolutions = []int64{downsample.ResLevel0, downsample.ResLevel1, downsample.ResLevel2}
	}
	var (
		engines = make([]*promql.Engine, len(resolutions))
		ld      = eo.LookbackDelta.Milliseconds()
	)
	wrapReg := func(engineNum int) prometheus.Registerer {
		return extprom.WrapRegistererWith(map[string]string{"engine": strconv.Itoa(engineNum)}, eo.Reg)
	}

	lookbackDelta := eo.LookbackDelta
	for i, r := range resolutions {
		if ld < r {
			lookbackDelta = time.Duration(r) * time.Millisecond
		}
		engines[i] = newEngine(promql.EngineOpts{
			Logger:                   eo.Logger,
			Reg:                      wrapReg(i),
			MaxSamples:               eo.MaxSamples,
			Timeout:                  eo.Timeout,
			ActiveQueryTracker:       eo.ActiveQueryTracker,
			LookbackDelta:            lookbackDelta,
			NoStepSubqueryIntervalFn: eo.NoStepSubqueryIntervalFn,
		})
	}
	return func(maxSourceResolutionMillis int64) *promql.Engine {
		for i := len(resolutions) - 1; i >= 1; i-- {
			left := resolutions[i-1]
			if resolutions[i-1] < ld {
				left = ld
			}
			if left < maxSourceResolutionMillis {
				return engines[i]
			}
		}
		return engines[0]
	}
}
