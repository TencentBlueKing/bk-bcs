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
	"flag"
	"fmt"

	"gomicro-discovery-operator/registry"

	frame "github.com/TencentBlueKing/blueking-apigateway-operator/pkg/discovery-operator-frame"
	"github.com/TencentBlueKing/blueking-apigateway-operator/pkg/discovery-operator-frame/options"

	"go.uber.org/zap/zapcore"
)

func main() {
	namespace := flag.String("namespace", "default", "operator's namespace")
	flag.Parse()
	opts := options.DefaultOptions()
	reg := &registry.MicroRegistry{}
	reg.Init()
	opts.Registry = reg
	opts.ConfigSchema = make(map[string]interface{})
	opts.RegisterNamespace = *namespace
	opts.ZapOpts.Level = zapcore.Level(-4)
	operator, err := frame.NewDiscoveryOperator(opts)
	if err != nil {
		fmt.Printf("Build operator failed: %v\n", err)
		return
	}
	reg.Client = operator.GetKubeClient()
	if err = operator.Run(); err != nil {
		fmt.Printf("Start operator failed: %v\n", err)
		return
	}
}
