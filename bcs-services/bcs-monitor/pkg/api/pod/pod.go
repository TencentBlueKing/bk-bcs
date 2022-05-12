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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// 获取 Pod 容器列表
func GetContainerList(c *rest.Context) (interface{}, error) {
	clusterId := c.Param("clusterId")
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	containers, err := k8sclient.GetContainerNames(c.Request.Context(), clusterId, namespace, pod)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// 获取 容器日志
func GetContainerLog(c *rest.Context) (interface{}, error) {
	clusterId := c.Param("clusterId")
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	logQuery := &k8sclient.LogQuery{}
	if err := c.BindQuery(logQuery); err != nil {
		return nil, err
	}

	logs, err := k8sclient.GetContainerLog(c.Request.Context(), clusterId, namespace, pod, logQuery)
	return logs, err
}

// 下载日志
func DownloadContainerLog(c *rest.Context) {
	clusterId := c.Param("clusterId")
	namespace := c.Param("namespace")
	pod := c.Param("pod")
	logQuery := &k8sclient.LogQuery{}
	if err := c.BindQuery(logQuery); err != nil {
		rest.AbortWithBadRequestError(c, err)
		return
	}

	logs, err := k8sclient.GetContainerLogByte(c.Request.Context(), clusterId, namespace, pod, logQuery)
	if err != nil {
		rest.AbortWithBadRequestError(c, err)
		return
	}

	ts := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s-%s-%s.log", pod, logQuery.ContainerName, ts)

	c.WriteAttachment(logs, filename)
}

func Ws() error {
	return nil
}
