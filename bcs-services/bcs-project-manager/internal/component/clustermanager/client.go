/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clustermanager

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/discovery"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var (
	// ErrNotInited err server not init
	ErrNotInited = errors.New("server not init")
)

// ClsManClient xxx
type ClsManClient struct {
	tlsConfig *tls.Config
	disc      *discovery.ModuleDiscovery
}

// ClusterClient xxx
var ClusterClient *ClsManClient

// SetClusterManagerClient set cluster manager client config
func SetClusterManagerClient(tlsConfig *tls.Config, disc *discovery.ModuleDiscovery) {
	ClusterClient = &ClsManClient{
		tlsConfig: tlsConfig,
		disc:      disc,
	}
}

// GetClusterManagerClient get cm client by discovery
func GetClusterManagerClient() (ClusterManagerClient, func(), error) {
	if ClusterClient == nil {
		return nil, nil, ErrNotInited
	}

	if ClusterClient.disc == nil {
		return nil, nil, fmt.Errorf("resourceManager module not enable dsicovery")
	}

	nodeServer, err := ClusterClient.disc.GetRandomServiceNode()
	if err != nil {
		return nil, nil, err
	}

	blog.Infof("ResManClient get node[%s] from disc", nodeServer.Address)
	conf := &Config{
		Hosts:     []string{nodeServer.Address},
		TLSConfig: ClusterClient.tlsConfig,
	}
	cli, closeCon := NewClusterManager(conf)

	return cli, closeCon, nil
}

// Config for connect cm server
type Config struct {
	Hosts     []string
	AuthToken string
	TLSConfig *tls.Config
}

// NewClusterManager create ResourceManager SDK implementation
func NewClusterManager(config *Config) (ClusterManagerClient, func()) {
	rand.Seed(time.Now().UnixNano())
	if len(config.Hosts) == 0 {
		// ! pay more attention for nil return
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
		opts = append(opts, grpc.WithInsecure())
	}
	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := rand.Intn(1024) % len(config.Hosts)
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
	return NewClusterManagerClient(conn), func() { conn.Close() }
}
