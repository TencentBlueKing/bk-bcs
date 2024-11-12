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

// Package clientset xxx
package clientset

import (
	"crypto/tls"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
)

// ClientGroup ...
type ClientGroup struct {
	AuthToken   string
	GatewayHost string
	TLSConfig   *tls.Config
}

var group *ClientGroup

// SetClientGroup init client group config
func SetClientGroup(gatewayHost, authToken string) {
	group = &ClientGroup{
		AuthToken:   authToken,
		GatewayHost: gatewayHost,
	}
}

// GetClientGroup ...
func GetClientGroup() *ClientGroup {
	return group
}

func (cg *ClientGroup) getRestConfig(clusterID string) *rest.Config {
	return &rest.Config{
		Host:        cg.GatewayHost + "/clusters/" + clusterID,
		BearerToken: cg.AuthToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
		QPS:   100,
		Burst: 100,
	}
}

// Client get client from client group by clusterID
func (cg *ClientGroup) Client(clusterID string) (*kubernetes.Clientset, error) {
	restConfig := cg.getRestConfig(clusterID)
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: ", clusterID)
		return nil, err
	}
	return clientset, nil
}

// RuntimeClient get client from controller runtime client by clusterID
func (cg *ClientGroup) RuntimeClient(clusterID string) (client.Client, error) {
	restConfig := cg.getRestConfig(clusterID)

	// 创建 Controller-runtime 客户端
	runtimeClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		logging.Error("get runtime client for cluster %s failed, err: ", clusterID)
		return nil, err
	}

	return runtimeClient, nil
}
