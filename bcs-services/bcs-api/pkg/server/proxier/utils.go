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

package proxier

import (
	"crypto/tls"
	restclient "k8s.io/client-go/rest"
	"net/url"
	"strings"

	"fmt"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
)

const (
	bcsK8sClusterDomain = "kubernetes"
)

func ExtractIpAddress(serverAddress string) (*url.URL, error) {
	if !strings.HasSuffix(serverAddress, "/") {
		serverAddress = serverAddress + "/"
	}
	ipAddress, err := url.Parse(serverAddress)
	if err != nil {
		return nil, err
	}
	return ipAddress, nil
}

func TurnCredentialsIntoConfig(clusterCredentials *m.ClusterCredentials) (*restclient.Config, error) {

	tlsClientConfig := restclient.TLSClientConfig{
		ServerName: bcsK8sClusterDomain,
		CAData:     []byte(clusterCredentials.CaCertData),
	}
	return &restclient.Config{
		Host:            clusterCredentials.ServerAddresses,
		BearerToken:     clusterCredentials.UserToken,
		TLSClientConfig: tlsClientConfig,
	}, nil
}

// check tcp connection to addr
func CheckTcpConn(addr string) error {
	checkUrl, err := url.Parse(addr)
	if err != nil {
		return err
	}
	err = dialTls(checkUrl.Host)
	if err != nil {
		return fmt.Errorf("connection to "+addr+" failed: %s", err.Error())
	}
	return nil
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
