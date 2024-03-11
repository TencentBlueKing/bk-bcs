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

package server

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

const (
	gracefulexit = 5
)

// Options for proxy
type Options struct {
	conf.LogConfig
	common.ServerConfig
	common.ClientConfig
	Registry         *common.Registry `json:"registry,omitempty"`
	PeerConnectURL   string           `json:"peerconnecturl,omitempty"`
	PeerConnectToken string           `json:"peerconnecttoken,omitempty"`
	ServiceName      string           `json:"servicename,omitempty"`
}

// DefaultOptions for proxy
func DefaultOptions() *Options {
	return &Options{
		LogConfig: conf.LogConfig{
			LogDir:       "/data/bcs/logs/bcs",
			Verbosity:    3,
			AlsoToStdErr: true,
		},
		ServerConfig: common.ServerConfig{
			Address:    "127.0.0.1",
			Port:       8081,
			HTTPPort:   8080,
			MetricPort: 8082,
			ServerCa:   "/data/bcs/cert/bcs/bcs-ca.crt",
			ServerCert: "/data/bcs/cert/bcs/bcs-server.crt",
			ServerKey:  "/data/bcs/cert/bcs/bcs-server.key",
		},
		ClientConfig: common.ClientConfig{
			ClientCa:   "/data/bcs/cert/bcs/bcs-ca.crt",
			ClientCert: "/data/bcs/cert/bcs/bcs-client.crt",
			ClientKey:  "/data/bcs/cert/bcs/bcs-client.key",
		},
		Registry: &common.Registry{
			Endpoints: "127.0.0.1",
			CA:        "/data/bcs/cert/etcd/etcd-ca.pem",
			Key:       "/data/bcs/cert/etcd/etcd-key.pem",
			Cert:      "/data/bcs/cert/etcd/etcd.pem",
		},
		PeerConnectURL:   "",
		PeerConnectToken: "",
		ServiceName:      common.ProxyName,
	}
}

// Complete all unset item
func (o *Options) Complete() error {
	if len(o.ServiceName) == 0 {
		o.ServiceName = common.ProxyName
	}
	if len(o.PeerConnectURL) == 0 {
		o.PeerConnectURL = common.ConnectURI
	}
	if err := o.ClientConfig.Complete(); err != nil {
		return err
	}
	if err := o.ServerConfig.Complete(); err != nil {
		return err
	}
	if err := o.Registry.Complete(); err != nil {
		return err
	}
	return nil
}

// Validate all config item
func (o *Options) Validate() error {
	if len(o.PeerConnectToken) == 0 {
		return fmt.Errorf("lost proxy tunnel token")
	}
	if o.ServerTLS == nil {
		return fmt.Errorf("lost server side TLS config")
	}
	if o.ClientTLS == nil {
		return fmt.Errorf("lost client side TLS config")
	}
	if o.Registry.TLSConfig == nil {
		return fmt.Errorf("lost registry TLS config")
	}
	return nil
}
