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

	// Note that Kubernetes registers workqueue metrics to default prometheus Registry. And the registry will be
	// initialized by the package 'k8s.io/apiserver/pkg/server'.
	// See https://github.com/kubernetes/kubernetes/blob/f61ed439882e34d9dad28b602afdc852feb2337a/staging/src/k8s.io/component-base/metrics/prometheus/workqueue/metrics.go#L25
	// But the controller-runtime registers workqueue metrics to its own Registry instead of default prometheus Registry.
	// See https://github.com/kubernetes-sigs/controller-runtime/blob/4d10a0615b11507451ecb58bfd59f0f6ef313a29/pkg/metrics/workqueue.go#L24-L26
	// However, global workqueue metrics factory will be only initialized once.
	// See https://github.com/kubernetes/kubernetes/blob/f61ed439882e34d9dad28b602afdc852feb2337a/staging/src/k8s.io/client-go/util/workqueue/metrics.go#L257-L261
	// So this package should be initialized before 'k8s.io/apiserver/pkg/server', thus the internal registry of
	// controller-runtime could be set first.
	_ "sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/cmd/mcs-agent/app"
	apiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	ctx := apiserver.SetupSignalContext()

	if err := app.NewAgentCommand(ctx).Execute(); err != nil {
		os.Exit(1)
	}
}
