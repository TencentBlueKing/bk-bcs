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

package app

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/collector"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/output"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/rdiscover"
)

// BcsMetricCollector BcsMetricCollector管理器定义
type BcsMetricCollector struct {
	mgr manager.Manager
}

// Run 执行启动逻辑
func (cli *BcsMetricCollector) Run(cfg *config.Config) error {
	return cli.mgr.Run(cfg)
}

// Run 实例化进程配置
func Run(cfg *config.Config) error {

	bcsCollector := &BcsMetricCollector{}

	cfg.Rd = rdiscover.NewRDiscover(cfg.ZKServerAddress)
	// 启动服务发现模块
	if err := cfg.Rd.Start(); nil != err {
		blog.Error("failed to start discovery service, error information is %s", err.Error())
		return err
	}

	out, outErr := output.New(context.TODO(), cfg)
	if nil != outErr {
		blog.Error("failed to create output object, error info is %s", outErr.Error())
		return outErr
	}

	c, collectorErr := collector.New(cfg, out)
	if nil != collectorErr {
		return collectorErr
	}

	bcsCollector.mgr = manager.New(c)

	if err := bcsCollector.Run(cfg); nil != err {
		return err
	}

	return fmt.Errorf("can not run here")
}
