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

package k8sclient

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetBCSConfByClusterId 通过集群ID, 获取不同admin token 信息
func GetBCSConfByClusterId(clusterId string) *config.BCSConf {
	// 只保留prod环境,不再通过集群ID选取不同环境配置
	// 默认返回bcs配置
	return config.G.BCS
}

// GetK8SConfigByClusterId 通过集群 ID 获取 K8S Rest Config
func GetK8SConfigByClusterId(clusterId string) *rest.Config {
	bcsConf := GetBCSConfByClusterId(clusterId)
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.InnerHost, clusterId)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsConf.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: bcsConf.InsecureSkipVerify,
		},
	}
	return config
}

// GetK8SClientByClusterId 通过集群 ID 获取 k8s client 对象
func GetK8SClientByClusterId(clusterId string) (*kubernetes.Clientset, error) {
	config := GetK8SConfigByClusterId(clusterId)
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}
