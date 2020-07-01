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

package ipscheduler

import (
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/pkg/netservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/pkg/netservice/types"

	"fmt"
	"strings"
)

//BcsConfig config item for ipscheduler
type BcsConfig struct {
	ZkHost   string         `json:"zkHost"`
	TLS      *types.SSLInfo `json:"tls,omitempty"`
	Interval int            `json:"interval,omitempty"`
}

func createNetSvcClient() (netservice.Client, error) {
	conf := newConf()

	var client netservice.Client
	var clientErr error
	if conf.TLS == nil {
		client, clientErr = netservice.NewClient()
	} else {
		client, clientErr = netservice.NewTLSClient(conf.TLS.CACert, conf.TLS.Key, conf.TLS.PubKey, conf.TLS.Passwd)
	}
	if clientErr != nil {
		return nil, clientErr
	}
	//client get bcs-netservice info
	hosts := strings.Split(conf.ZkHost, ";")
	if err := client.GetNetService(hosts); err != nil {
		return nil, fmt.Errorf("get netservice failed, %s", err.Error())
	}
	return client, nil
}

func newConf() BcsConfig {
	conf := BcsConfig{
		ZkHost: config.ZkHosts,
		TLS: &types.SSLInfo{
			CACert: config.ClientCert.CAFile,
			Key:    config.ClientCert.KeyFile,
			PubKey: config.ClientCert.CertFile,
			Passwd: config.ClientCert.CertPasswd,
		},
	}

	return conf
}
