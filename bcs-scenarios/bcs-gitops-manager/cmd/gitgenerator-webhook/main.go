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

// Package main xxx
package main

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/gitgenerator-webhook/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/gitgenerator-webhook/webhook"
)

func main() {
	cfg := options.Parse()
	blog.InitLogs(cfg.LogConfig)
	defer blog.CloseLogs()

	server := webhook.NewAdmissionWebhookServer(cfg)
	if err := server.Init(); err != nil {
		blog.Fatalf("init webhook server failed: %s", err.Error())
	}
	if err := server.Run(); err != nil {
		blog.Fatalf(err.Error())
	}
}
