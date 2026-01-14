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
 */

// Package cluster xxx
package cluster

import (
	"fmt"
	"regexp"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	clusternet "github.com/clusternet/clusternet/pkg/generated/clientset/versioned"
)

// get kube client for request user cluster by bcs gateway
func (h *clusterClient) getKubeClientByClusterId(clusterId string) (*kubernetes.Clientset, error) {
	host := fmt.Sprintf("%s/clusters/%s", h.opt.Endpoint, clusterId)
	config := &rest.Config{
		Host:        host,
		BearerToken: h.opt.Token,
	}

	return kubernetes.NewForConfig(config)
}

func (h *clusterClient) getClusternetClientByClusterId(clusterId string) (*clusternet.Clientset, error) {
	host := fmt.Sprintf("%s/clusters/%s", h.opt.Endpoint, clusterId)
	config := &rest.Config{
		Host:        host,
		BearerToken: h.opt.Token,
	}

	return clusternet.NewForConfig(config)
}

// getDynamicClientByClusterId 获取管理器 cli
func (h *clusterClient) getDynamicClientByClusterId(clusterId string) (*dynamic.DynamicClient, error) {
	host := fmt.Sprintf("%s/clusters/%s", h.opt.Endpoint, clusterId)
	config := &rest.Config{
		Host:            host,
		BearerToken:     h.opt.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// 获取客户端
	return dynamicClient, nil
}

// isMatchPattern
func (h *clusterClient) isMatchPattern(pattern, subnetName string) bool {
	re, _ := regexp.Compile(pattern)
	return re.MatchString(subnetName)
}
