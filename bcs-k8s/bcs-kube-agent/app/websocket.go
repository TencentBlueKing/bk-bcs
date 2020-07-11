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

package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
)

const (
	kubernetesServiceHost = "KUBERNETES_SERVICE_HOST"
	kubernetesServicePort = "KUBERNETES_SERVICE_PORT"

	Module        = "BCS-API-Tunnel-Module"
	RegisterToken = "BCS-API-Tunnel-Token"
	Params        = "BCS-API-Tunnel-Params"
	Cluster       = "BCS-API-Tunnel-ClusterId"

	ModuleName = "kube-agent"
)

func getenv(env string) (string, error) {
	value := os.Getenv(env)
	if value == "" {
		return "", fmt.Errorf("%s is empty", env)
	}
	return value, nil
}

func buildWebsocketToBke(cfg *rest.Config) error {
	bkeServerAddress := viper.GetString("bke.serverAddress")
	clusterId := viper.GetString("cluster.id")
	registerToken := os.Getenv("REGISTER_TOKEN")

	bkeServerUrl, err := url.Parse(bkeServerAddress)
	if err != nil {
		return err
	}

	if err := populateCAData(cfg); err != nil {
		return fmt.Errorf("error populating ca data: %s", err.Error())
	}

	kubernetesServiceHost, err := getenv(kubernetesServiceHost)
	if err != nil {
		return err
	}
	kubernetesServicePort, err := getenv(kubernetesServicePort)
	if err != nil {
		return err
	}
	params := map[string]interface{}{
		"address":   fmt.Sprintf("https://%s:%s", kubernetesServiceHost, kubernetesServicePort),
		"userToken": cfg.BearerToken,
		"caCert":    base64.StdEncoding.EncodeToString(cfg.CAData),
	}
	bytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	headers := map[string][]string{
		Module:        {ModuleName},
		Cluster:       {clusterId},
		RegisterToken: {registerToken},
		Params:        {base64.StdEncoding.EncodeToString(bytes)},
	}

	var tlsConfig *tls.Config
	insecureSkipVerify := viper.GetBool("agent.insecureSkipVerify")
	if insecureSkipVerify {
		tlsConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	} else {
		pool := x509.NewCertPool()
		caCrtStr := os.Getenv("SERVER_CERT")
		caCrt := []byte(caCrtStr)
		pool.AppendCertsFromPEM(caCrt)
		tlsConfig = &tls.Config{RootCAs: pool}
	}

	go func() {
		for {
			wsURL := fmt.Sprintf("wss://%s/bcsapi/v4/usermanager/v1/websocket/connect", bkeServerUrl.Host)
			blog.Infof("Connecting to %s with token %s", wsURL, registerToken)

			websocketDialer.ClientConnect(context.Background(), wsURL, headers, tlsConfig, nil, func(proto, address string) bool {
				switch proto {
				case "tcp":
					return true
				case "unix":
					return address == "/var/run/docker.sock"
				}
				return false
			})
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}
