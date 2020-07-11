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

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
)

func main() {
	dir, _ := os.Getwd()

	cfg := conf.LogConfig{
		LogDir:       dir + "/blog",
		LogMaxSize:   500,
		ToStdErr:     true,
		AlsoToStdErr: true,
		Verbosity:    6,
	}
	blog.InitLogs(cfg)

	c := metric.Config{
		ModuleName:          "fake_module",
		IP:                  "0.0.0.0",
		MetricPort:          9089,
		DisableGolangMetric: true,
		ClusterID:           "breeze-demo-clusterid",
	}

	healthz := func() metric.HealthMeta {
		return metric.HealthMeta{
			CurrentRole: "Master",
			IsHealthy:   true,
		}
	}

	demo := new(DemoMetric)
	numeric := metric.MetricContructor{
		GetMeta:   demo.GetNumericMeta,
		GetResult: demo.GetNumericResult,
	}

	sm := metric.MetricContructor{
		GetMeta:   demo.GetStringMeta,
		GetResult: demo.GetStringResult,
	}

	if err := metric.NewMetricController(c, healthz, &numeric, &sm); err != nil {
		fmt.Printf("new metric collector failed. err: %v\n", err)
		return
	}
	fmt.Println("start running")
	select {}
}

type DemoMetric struct{}

func (DemoMetric) GetNumericMeta() *metric.MetricMeta {
	return &metric.MetricMeta{
		Name: "timenow_seconds",
		Help: "show current time in unix time.",
		ConstLables: map[string]string{
			"c_time": "c_demolable",
		},
	}
}

func (DemoMetric) GetNumericResult() (*metric.MetricResult, error) {
	v, err := metric.FormFloatOrString(time.Now().Unix())
	if err != nil {
		return nil, err
	}
	return &metric.MetricResult{
		Value: v,
		VariableLabels: map[string]string{
			"var_key": "var_value",
		},
	}, nil
}

func (DemoMetric) GetStringMeta() *metric.MetricMeta {
	return &metric.MetricMeta{
		Name: "birthday_string",
		Help: "show current time in string.",
		ConstLables: map[string]string{
			"c_time": "c_demolable",
		},
	}
}

func (DemoMetric) GetStringResult() (*metric.MetricResult, error) {
	v, err := metric.FormFloatOrString(time.Now().String())
	if err != nil {
		return nil, err
	}
	return &metric.MetricResult{
		Value: v,
		VariableLabels: map[string]string{
			"var_key": "var_value",
		},
	}, nil
}
