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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/cmd"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/cmd/config"
	_ "github.com/Tencent/bk-bcs/bcs-k8s/bcs-apiserver-proxy/pkg/utils/metrics"
)

func main() {
	// init apiServer-proxy options config
	apiServerProxyOptions := config.NewProxyAPIServerOptions()
	conf.Parse(apiServerProxyOptions)

	// init log config
	blog.InitLogs(apiServerProxyOptions.LogConfig)
	defer blog.CloseLogs()

	// set proxyManager
	proxyManager, err := cmd.NewProxyManager(apiServerProxyOptions)
	if err != nil {
		blog.Fatalf("init NewProxyManager failed: %v", err)
		return
	}

	// init proxyManager
	err = proxyManager.Init(apiServerProxyOptions)
	if err != nil {
		blog.Fatalf("init proxyManager failed: %v", err)
		return
	}

	// run proxyManager
	if err := proxyManager.Run(); err != nil {
		blog.Infof("proxyManager run quit: %v", err)
	}

	return
}
