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
	"context"
	"flag"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/pkg/config"
)

var configFile = flag.String("conf", "", "path to sync apisix config")

func main() {
	flag.Parse()
	if configFile == nil || *configFile == "" {
		panic("config_file is required")
	}
	syncConf, err := config.Parse(*configFile)
	if err != nil {
		panic("parse config file failed: " + err.Error())
	}
	// blog初始化
	blog.InitLogs(*syncConf.Logging)

	syncResources := pkg.NewSyncResources(syncConf)
	err = syncResources.SyncGatewayResources(context.Background())
	if err != nil {
		blog.Errorf("sync apisix gateway resources failed: %s ", err)
		os.Exit(1)
		return
	}
}
