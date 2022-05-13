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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetEnvByClusterId 获取集群所属环境, 目前通过集群ID前缀判断
func GetEnvByClusterId(clusterId string) config.BCSClusterEnv {
	if strings.HasPrefix(clusterId, "BCS-K8S-1") {
		return config.UatCluster
	}
	if strings.HasPrefix(clusterId, "BCS-K8S-2") {
		return config.DebugCLuster
	}
	if strings.HasPrefix(clusterId, "BCS-K8S-4") {
		return config.ProdEnv
	}
	return config.ProdEnv
}

// GetBCSConfByClusterId 通过集群ID, 获取不同admin token 信息
func GetBCSConfByClusterId(clusterId string) *config.BCSConf {
	env := GetEnvByClusterId(clusterId)
	conf, ok := config.G.BCSEnvMap[env]
	if ok {
		return conf
	}
	// 默认返回bcs配置
	return config.G.BCS
}

// GetK8SConfigByClusterId 通过集群 ID 获取 K8S Rest Config
func GetK8SConfigByClusterId(clusterId string) *rest.Config {
	bcsConf := GetBCSConfByClusterId(clusterId)
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsConf.Token,
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
