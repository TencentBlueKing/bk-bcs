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

// Package bcsapi xxx
package bcsapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	registry "github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
)

// ErrNotInited err server not init
var ErrNotInited = errors.New("server not init")

// ! v4 version binding~

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
	// Header for request header
	Header map[string]string
	// InnerClientName for bcs inner auth, like bcs-cluster-manager
	InnerClientName string
}

// BasicResponse basic http response for bkbcs
type BasicResponse struct {
	Code    int             `json:"code"`
	Result  bool            `json:"result"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Authentication defines the common interface for the credentials which need to
// attach auth info to every RPC
type Authentication struct {
	InnerClientName string
	Insecure        bool
}

// GetRequestMetadata gets the current request metadata
func (a *Authentication) GetRequestMetadata(context.Context, ...string) (
	map[string]string, error,
) {
	return map[string]string{"X-Bcs-Client": a.InnerClientName}, nil
}

// RequireTransportSecurity indicates whether the credentials requires
// transport security.
func (a *Authentication) RequireTransportSecurity() bool {
	return !a.Insecure
}

// NewTokenAuth implementations of grpc credentials interface
func NewTokenAuth(t string) *GrpcTokenAuth {
	return &GrpcTokenAuth{
		Token: t,
	}
}

// GrpcTokenAuth grpc token
type GrpcTokenAuth struct {
	Token string
}

// GetRequestMetadata convert http Authorization for grpc key
func (t GrpcTokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", t.Token),
	}, nil
}

// RequireTransportSecurity RequireTransportSecurity
func (t GrpcTokenAuth) RequireTransportSecurity() bool {
	return false
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

// ClientConfig cluster manager client config
type ClientConfig struct {
	TLSConfig *tls.Config
	Discovery *discovery.ModuleDiscovery
}
