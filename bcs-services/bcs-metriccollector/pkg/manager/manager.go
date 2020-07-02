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

package manager

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/collector"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/register"
)

type manager struct {
	collector collector.Collector
}

func (cli *manager) Run(cfg *config.Config) error {
	if cfg.RunMode == config.TraditionalType {
		scheme := "http"
		if len(cfg.ServerCertFile) != 0 && len(cfg.ServerKeyFile) != 0 {
			scheme = "https"
		}

		err := register.Register(cfg.LocalIP, scheme, 0, cfg.MetricPort, cfg.BCSZk, "")
		if err != nil {
			return fmt.Errorf("register collector to zk failed, err: %v", err)
		}
	}

	return cli.collector.Run(context.TODO(), cfg)
}
