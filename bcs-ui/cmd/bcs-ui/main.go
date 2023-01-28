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
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

// isPrintVersion 是否是--version, -v命令
func isPrintVersion() bool {
	if len(os.Args) < 2 {
		return false
	}
	arg := os.Args[1]
	for _, name := range []string{"version", "v"} {
		if arg == "-"+name || arg == "--"+name {
			return true
		}
	}
	return false
}

func main() {
	// metrics 配置
	metrics := prometheus.NewRegistry()
	metrics.MustRegister(version.NewCollector("bcs_ui"))

	prometheus.DefaultRegisterer = metrics

	if isPrintVersion() {
		os.Exit(0)
	}

	Execute()
}
