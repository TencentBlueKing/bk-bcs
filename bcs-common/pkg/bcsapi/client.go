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

package bcsapi

import (
	"crypto/tls"
	"encoding/json"

	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
)

//! v4 version binding~

const (
	gatewayPrefix   = "/bcsapi/v4/"
	clusterIDHeader = "BCS-ClusterID"
)

// Config for bcsapi
type Config struct {
	// bcsapi host, available like 127.0.0.1:8080
	Hosts []string
	// tls configuratio
	TLSConfig *tls.Config
	// AuthToken for permission verification
	AuthToken string
	// clusterID for Kubernetes/Mesos operation
	ClusterID string
	// proxy flag for go through bcs-api-gateway
	Gateway bool
	// etcd registry config for bcs modules
	Etcd registry.CMDOptions
}

// BasicResponse basic http response for bkbcs
type BasicResponse struct {
	Code    int             `json:"code"`
	Result  bool            `json:"result"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// NewClient create new bcsapi instance
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
	}
}

// Client all module client api composition
type Client struct {
	config *Config
}

// UserManager client interface
func (c *Client) UserManager() UserManager {
	return NewUserManager(c.config)
}

// MesosDriver client interface
// func (c *Client) MesosDriver() MesosDriver {
// 	return &MesosDriverClient{}
// }

// Storage client interface
func (c *Client) Storage() Storage {
	return NewStorage(c.config)
}

// ClusterManager grpc cluster manager client
func (c *Client) ClusterManager() cm.ClusterManagerClient {
	return NewClusterManager(c.config)
}
