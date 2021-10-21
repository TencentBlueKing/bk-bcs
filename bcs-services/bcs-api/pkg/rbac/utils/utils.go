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

package utils

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func TurnCredentialsIntoConfig(clusterCredentials *m.ClusterCredentials) *restclient.Config {
	tlsClientConfig := restclient.TLSClientConfig{
		CAData: []byte(clusterCredentials.CaCertData),
	}

	return &restclient.Config{
		Host:            clusterCredentials.ServerAddresses,
		BearerToken:     clusterCredentials.UserToken,
		TLSClientConfig: tlsClientConfig,
	}
}

func GetKubeClient(clusterId string) (*kubernetes.Clientset, error) {
	clusterCredentials := sqlstore.GetCredentials(clusterId)
	if clusterCredentials == nil {
		return nil, fmt.Errorf("cluster credentias not found for cluster: %s", clusterId)
	}
	apiServerList := clusterCredentials.GetServerAddressesList()
	var kubeClient *kubernetes.Clientset
	for _, apiServer := range apiServerList {
		hostPort := strings.TrimPrefix(apiServer, "https://")
		if err := pingEndpoint(hostPort); err == nil {
			clusterCredentials.ServerAddresses = apiServer
			kubeClient, err = makeKubeclient(clusterCredentials)
			return kubeClient, err
		}
	}
	return nil, fmt.Errorf("couldn't find an available apiserver for cluster: %s", clusterId)
}

// probe the health of the apiserver address for 3 times
func pingEndpoint(host string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = dialTls(host)
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			blog.Errorf("Error connecting the apiserver %s. Retrying...: %s", host, err.Error())
			time.Sleep(time.Second * 3)
			continue
		} else if err != nil {
			blog.Errorf("Error connecting the apiserver %s: %s", host, err.Error())
			return err
		}
		return nil
	}
	return err
}

func makeKubeclient(clusterCredentials *m.ClusterCredentials) (*kubernetes.Clientset, error) {
	restConfig := TurnCredentialsIntoConfig(clusterCredentials)
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error when building kubeclient from restconfig: %s", err.Error())
	}
	return kubeClient, nil
}

func dialTls(host string) error {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", host, conf)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
