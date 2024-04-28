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

// Package tresource xxx
package tresource

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var (
	// ErrNotInited err server not init
	ErrNotInited = errors.New("server not init")
	// ErrNotImplement err fun not implement
	ErrNotImplement = errors.New("func not implement")
)

// Options for rm client
type Options struct {
	// Enable enable discovery
	Enable bool
	// Module module name
	Module string
	// other configInfo
	TLSConfig *tls.Config
}

// Config for connect rm server
type Config struct {
	// Hosts host
	Hosts []string
	// AuthToken token
	AuthToken string
	// TLSConfig cert
	TLSConfig *tls.Config
}

// NewResourceManager create ResourceManager SDK implementation
func NewResourceManager(config *Config) (ResourceManagerClient, func()) {
	rand.Seed(time.Now().UnixNano()) // nolint
	if len(config.Hosts) == 0 {
		//! pay more attention for nil return
		return nil, nil
	}
	// create grpc connection
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(config.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", config.AuthToken)
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	} else {
		// nolint
		opts = append(opts, grpc.WithInsecure())
	}
	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := rand.Intn(1024) % len(config.Hosts) // nolint
		addr := config.Hosts[selected]
		conn, err = grpc.Dial(addr, opts...)
		if err != nil {
			blog.Errorf("Create resource manager grpc client with %s error: %s", addr, err.Error())
			continue
		}
		if conn != nil {
			break
		}
	}
	if conn == nil {
		blog.Errorf("create no resource manager client after all instance tries")
		return nil, nil
	}

	// init cluster manager client
	return NewResourceManagerClient(conn), func() { conn.Close() } // nolint
}

// OrderState xxx
type OrderState string

var (
	// OrderFinished finish state
	OrderFinished OrderState = "FINISHED"
	// OrderFailed failed state
	OrderFailed OrderState = "FAILED"
	// OrderRequested requested state
	OrderRequested OrderState = "REQUESTED"
)

// String toString
func (os OrderState) String() string {
	return string(os)
}

const (
	statusOnSale  = "ONSALE"  // nolint
	statusNotSale = "NOTSALE" // nolint
)

// PoolLabel xxx
type PoolLabel string

var (
	// AllowedBusinessIDs businessID
	AllowedBusinessIDs PoolLabel = "allowedBusinessIDs"
	// AvailableQuota quota
	AvailableQuota PoolLabel = "availableQuota"
	// InstanceSpecs instance specs
	InstanceSpecs PoolLabel = "instanceSpecs"
	// ResourceType resource type(online/offline)
	ResourceType PoolLabel = "resourceType"
)

// String toString
func (pl PoolLabel) String() string {
	return string(pl)
}
