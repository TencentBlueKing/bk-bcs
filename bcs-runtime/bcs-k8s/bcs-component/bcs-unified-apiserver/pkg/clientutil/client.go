/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clientutil

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/clusternet/clusternet/pkg/wrappers/clientgo"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/config"
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

// GetKubeConfByClusterId 通过集群 ID 获取 k8s client 对象
func GetKubeConfByClusterId(clusterId string) (*rest.Config, error) {
	bcsConf := GetBCSConfByClusterId(clusterId)
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsConf.Token,
	}

	return config, nil
}

// GetKubeClientByClusterId 通过集群 ID 获取 k8s client 对象
func GetKubeClientByClusterId(clusterId string) (*kubernetes.Clientset, error) {
	bcsConf := GetBCSConfByClusterId(clusterId)
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsConf.Token,
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return k8sClient, nil
}

// GetClusternetClientByClusterId 通过集群 ID 获取 k8s client 对象, 使用clusternet shadow apis
func GetClusternetClientByClusterId(clusterId string) (*kubernetes.Clientset, error) {
	bcsConf := GetBCSConfByClusterId(clusterId)
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsConf.Token,
	}

	// This is the ONLY place you need to wrap for Clusternet
	config.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return clientgo.NewClusternetTransport(config.Host, rt)
	})

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return k8sClient, nil
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
