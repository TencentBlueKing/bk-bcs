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

// Package registryv4 xxx
package registryv4

import (
	"crypto/tls"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
)

// CMDOptions config item for etcd discovery
type CMDOptions struct {
	Feature bool   `json:"etcd_feature" value:"false" usage:"switch that turn on etcd registry feature"`
	Address string `json:"etcd_address" value:"[::1]:2379,127.0.0.1:2379" usage:"etcd registry feature, multiple ip addresses splited by comma"` // nolint
	CA      string `json:"etcd_ca" value:"" usage:"etcd registry CA"`
	Cert    string `json:"etcd_cert" value:"" usage:"etcd registry tls cert file"`
	Key     string `json:"etcd_key" value:"" usage:"etcd registry tls key file"`
}

// GetTLSConfig get tls config from configuration
func (co *CMDOptions) GetTLSConfig() (*tls.Config, error) {
	if !co.Feature {
		return nil, fmt.Errorf("no etcd registry feature on")
	}
	config, err := ssl.ClientTslConfVerity(co.CA, co.Cert, co.Key, "")
	if err != nil {
		return nil, err
	}
	return config, nil
}
