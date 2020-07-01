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

package kubedriver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/options"
)

const (
	Module        = "BCS-API-Tunnel-Module"
	RegisterToken = "BCS-API-Tunnel-Token"
	Params        = "BCS-API-Tunnel-Params"
	Cluster       = "BCS-API-Tunnel-ClusterId"
)

func buildWebsocketToApi(o *options.KubeDriverServerOptions) error {
	if o.RegisterUrl == "" {
		return errors.New("register url is empty")
	}
	bcsApiUrl, err := url.Parse(o.RegisterUrl)
	if err != nil {
		return err
	}

	if o.RegisterToken == "" {
		return errors.New("register token is empty")
	}
	if o.CustomClusterID == "" {
		return errors.New("custom clusterid is empty")
	}

	var serverAddress string
	if o.SecureServerConfigured() {
		serverAddress = fmt.Sprintf("https://%s:%d", o.BindAddress.String(), o.SecurePort)
	} else {
		serverAddress = fmt.Sprintf("http://%s:%d", o.BindAddress.String(), o.InsecurePort)
	}

	params := map[string]interface{}{
		"address": serverAddress,
	}
	bytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	headers := map[string][]string{
		Module:        {ModuleName},
		Cluster:       {o.CustomClusterID},
		RegisterToken: {o.RegisterToken},
		Params:        {base64.StdEncoding.EncodeToString(bytes)},
	}

	var tlsConfig *tls.Config
	if o.InsecureSkipVerify {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		// use bcs cacert
		pool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(o.ServerTLS.CAFile)
		if err != nil {
			return err
		}
		if ok := pool.AppendCertsFromPEM(ca); ok != true {
			return fmt.Errorf("append ca cert failed")
		}
		tlsConfig = &tls.Config{RootCAs: pool}
	}

	go func() {
		for {
			wsURL := fmt.Sprintf("wss://%s/bcsapi/v1/websocket/connect", bcsApiUrl.Host)
			blog.Infof("Connecting to %s with token %s", wsURL, o.RegisterToken)

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
