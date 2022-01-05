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

package cmd

import (
	"flag"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache/redis"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
)

var confFilePath = flag.String("conf", common.DefaultConfPath, "配置文件路径")

var globalConf *config.ClusterResourcesConf

// Start 初始化并启动 ClusterResources 服务
func Start() {
	flag.Parse()

	var loadConfErr error
	globalConf, loadConfErr = config.LoadConf(*confFilePath)
	if loadConfErr != nil {
		panic(fmt.Errorf("load cluster resources config failed: %w", loadConfErr))
	}
	// 初始化日志相关配置
	logging.InitLogger(&globalConf.Log)
	logger := logging.GetLogger()
	defer logger.Sync()

	// 初始化 Redis 客户端
	redis.InitRedisClient(&globalConf.Redis)

	crSvc := newClusterResourcesService(globalConf)
	if err := crSvc.Init(); err != nil {
		panic(fmt.Errorf("init cluster resources svc failed: %w", err))
	}
	if err := crSvc.Run(); err != nil {
		panic(fmt.Errorf("run cluster resources svc failed: %w", err))
	}
}
