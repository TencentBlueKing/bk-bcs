/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-监控平台 (Blueking - Monitor) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pod

import (
	"io"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// PodLogStream Server Sent Events Handler 连接处理函数
// @Summary  SSE 实时日志流
// @Tags     Pod
// @Param    container_name  query  string  true  "容器名称"
// @Produce  text/event-stream
// @Success  200  {string}  string
// @Router   /namespaces/:namespace/pods/:pod/logs/stream [get]
func PodLogStream(c *rest.Context) {
	clusterId := c.Param("clusterId")
	namespace := c.Param("namespace")
	pod := c.Param("pod")

	logQuery := &k8sclient.LogQuery{}
	if err := c.BindQuery(logQuery); err != nil {
		rest.AbortWithBadRequestError(c, err)
		return
	}

	logChan, err := k8sclient.GetPodLogStream(c.Request.Context(), clusterId, namespace, pod, logQuery)
	if err != nil {
		rest.AbortWithBadRequestError(c, err)
		return
	}

	c.Stream(func(w io.Writer) bool {
		for log := range logChan {
			c.SSEvent("message", log)
		}
		return true
	})
}
