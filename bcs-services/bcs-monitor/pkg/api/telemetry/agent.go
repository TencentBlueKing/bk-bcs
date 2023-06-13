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

package telemetry

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// IsBKMonitorAgent 是否安装蓝鲸监控采集器, 灰度使用
func IsBKMonitorAgent(c *rest.Context) (interface{}, error) {
	clusterId := c.Param("clusterId")
	conf := k8sclient.GetBCSConf()

	cluster, err := bcs.GetCluster(clusterId)
	if err != nil {
		return nil, err
	}
	project, err := bcs.GetProject(c.Request.Context(), conf, cluster.ProjectId)
	if err != nil {
		return nil, err
	}
	createTime, err := project.CreateTime()
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"agent":  "bk_monitor",
		"enable": false,
	}
	if createTime.After(config.G.BKMonitor.AgentEnableAfterTime) {
		data["enable"] = true
	}
	return data, nil
}
