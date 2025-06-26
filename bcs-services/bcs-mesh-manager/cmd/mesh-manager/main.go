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

// Package main 提供mesh manager的入口
package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	opts := options.NewMeshManagerOptions()
	if err := options.Parse(opts); err != nil {
		log.Fatalf("parse options failed, err %s", err.Error())
	}

	if err := opts.Validate(); err != nil {
		log.Fatalf("validate options failed, err %s", err.Error())
	}

	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()

	server := app.NewServer(opts)
	if err := server.Init(); err != nil {
		blog.Fatalf("init mesh manager failed, err %s", err.Error())
	}
	if err := server.Run(); err != nil {
		blog.Fatalf("run mesh manager failed, %s", err.Error())
	}
}
