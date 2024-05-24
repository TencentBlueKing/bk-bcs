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

// Package main is the entry point of the application and contains the main logic of the application.
package main

import (
	"fmt"
	"os"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-proxy/app"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-proxy/options"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

func main() {
	cc.InitService(cc.FeedProxyName)

	opts, err := options.InitOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "start feed proxy failed, err: %v", err)
		os.Exit(1)
	}

	if err := app.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "start feed proxy failed, err: %v", err)
		logs.CloseLogs()
		os.Exit(1)
	}
}
