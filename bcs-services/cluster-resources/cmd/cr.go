/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package cmd 执行 ClusterResources 服务初始化
package cmd

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache/redis"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version"
)

var (
	showVersion  = flag.Bool("version", false, "show version info only")
	checkService = flag.Bool("checkService", false, "check dependency service status (redis, clusterManager...)")
	confFilePath = flag.String("conf", conf.DefaultConfPath, "config file path")
)

// Start 初始化并启动 ClusterResources 服务
func Start() {
	flag.Parse()

	// 若指定仅展示版本信息，则打印后退出
	if *showVersion {
		version.ShowVersionAndExit()
	}

	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	crConf, err := config.LoadConf(*confFilePath)
	if err != nil {
		panic(errorx.New(errcode.General, "load cluster resources config failed: %v", err))
	}
	// 初始化日志相关配置
	logging.InitLogger(&crConf.Log)
	logger := logging.GetLogger()
	defer logger.Sync()

	logger.Info(fmt.Sprintf("ConfigFilePath: %s", *confFilePath))
	logger.Info(fmt.Sprintf("VersionBuildInfo: {%s}", version.GetVersion()))

	// 若指定了只检查依赖服务，则检查通过后以零值退出，否则以非零值退出
	if *checkService {
		NewDependencyServiceChecker(crConf).DoAndExit()
	}

	// 初始化 Redis 客户端
	redis.InitRedisClient(&crConf.Redis)

	crSvc := newClusterResourcesService(crConf)
	if err := crSvc.Init(); err != nil {
		panic(errorx.New(errcode.General, "init cluster resources svc failed: %v", err))
	}
	if err := crSvc.Run(); err != nil {
		panic(errorx.New(errcode.General, "run cluster resources svc failed: %v", err))
	}
}
