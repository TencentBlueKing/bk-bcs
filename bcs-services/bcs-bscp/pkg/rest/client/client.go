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

// Package client NOTES
package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// NewClient new http client.
func NewClient(c *tools.TLSConfig) (*http.Client, error) {
	tlsConf := new(tls.Config)
	if nil != c {
		tlsConf.InsecureSkipVerify = c.InsecureSkipVerify
		if len(c.CAFile) != 0 && len(c.CertFile) != 0 && len(c.KeyFile) != 0 {
			var err error
			tlsConf, err = tools.ClientTLSConfVerify(c.InsecureSkipVerify, c.CAFile, c.CertFile, c.KeyFile, c.Password)
			if err != nil {
				return nil, err
			}
		}
	}

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig:     tlsConf,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		MaxIdleConnsPerHost:   1000,
		ResponseHeaderTimeout: 10 * time.Minute,
	}

	client := new(http.Client)
	client.Transport = transport
	return client, nil
}

// HTTPClient http client interface.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
