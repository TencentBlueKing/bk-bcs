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

package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg/client"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

func newClientWithConfiguration() pkg.HelmClient {
	return client.New(client.Config{
		APIServer: viper.GetString("config.apiserver"),
		AuthToken: viper.GetString("config.token"),
	})
}

func newGRPCClientWithConfiguration() helmmanager.HelmManagerClient { // nolint
	cli, _ := newHelmClient(&bcsapi.Config{
		Hosts:     []string{viper.GetString("config.apiserver")},
		AuthToken: viper.GetString("config.token"),
	})
	return cli
}

// newHelmClient create HelmManager SDK implementation
func newHelmClient(config *bcsapi.Config) (helmmanager.HelmManagerClient, func()) { // nolint
	rand.Seed(time.Now().UnixNano()) // nolint
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
	for k, v := range config.Header {
		header[k] = v
	}
	md := metadata.New(header)
	auth := &bcsapi.Authentication{InnerClientName: config.InnerClientName}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		auth.Insecure = true
	}
	opts = append(opts, grpc.WithPerRPCCredentials(auth))
	if config.AuthToken != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(bcsapi.NewTokenAuth(config.AuthToken)))
	}

	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := rand.Intn(1024) % len(config.Hosts) // nolint
		addr := config.Hosts[selected]
		conn, err = grpc.Dial(addr, opts...)
		if err != nil {
			blog.Errorf("Create helm manager grpc client with %s error: %s", addr, err.Error())
			continue
		}
		if conn != nil {
			break
		}
	}
	if conn == nil {
		blog.Errorf("create no helm manager client after all instance tries")
		return nil, nil
	}
	return helmmanager.NewHelmManagerClient(conn), func() { conn.Close() } // nolint
}

func getInputData() ([]byte, error) {
	if jsonData != "" {
		return []byte(jsonData), nil
	}

	if jsonFile != "" {
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	return nil, fmt.Errorf("empty param data")
}
