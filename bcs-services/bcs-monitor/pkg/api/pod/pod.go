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

func Ws() error {
	return nil
}

func DownloadLog() error {
	return nil
}
