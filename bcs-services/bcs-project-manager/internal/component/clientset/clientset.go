/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clientset

import (
	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

// Client get client from client group by clusterID
func (cg *ClientGroup) Client(clusterID string) (*kubernetes.Clientset, error) {
	restConfig := &rest.Config{
		Host:        cg.GatewayHost + "/clusters/" + clusterID,
		BearerToken: cg.AuthToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
		QPS:   100,
		Burst: 100,
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: ", clusterID)
		return nil, err
	}
	return clientset, nil
}
