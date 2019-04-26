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
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/metric"
	"bk-bcs/bcs-services/bcs-loadbalance/rdiscover"
	"strings"
)

func (lp *LBEventProcessor) metricRegister() error {
	runMode := metric.Master_Master_Mode
	if strings.ToLower(lp.config.Proxy) == "awselb" || strings.ToLower(lp.config.Proxy) == "qcloudclb" {
		runMode = metric.Master_Slave_Mode
	}

	c := metric.Config{
		ModuleName: "bcs-loadbalance",
		MetricPort: lp.config.MetricPort,
		IP:         rdiscover.GetAvailableIP(),
		ClusterID:  lp.config.ClusterID,
		RunMode:    runMode,
	}

	statData := metric.MetricContructor{
		GetMeta:   lp.cfgManager.GetMetricMeta,
		GetResult: lp.cfgManager.GetMetricResult,
	}

	if err := metric.NewMetricController(
		c,
		lp.cfgManager.GetHealthInfo,
		&statData,
	); err != nil {
		blog.Errorf("metric server error: %v", err)
		return err
	}
	blog.Infof("start metric server successfully, IP %s, metric port %d",
		rdiscover.GetAvailableIP(), lp.config.MetricPort)
	return nil
}
