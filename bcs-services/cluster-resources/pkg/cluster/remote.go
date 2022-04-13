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

package cluster

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	bcsapicm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
)

// TODO 梳理整块逻辑，合并到 client.go 中
// NOTE 因 Go，Grpc，Micro 版本冲突，这里从原文件拷贝并改造:
// github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/client.go
// github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager.go

// Config ...
type Config struct {
	Hosts     []string
	TLSConfig *tls.Config
	AuthToken string
}

// NewClusterManager ...
func NewClusterManager(config *Config) bcsapicm.ClusterManagerClient {
	ctx := context.TODO()
	rand.Seed(time.Now().UnixNano())
	if len(config.Hosts) == 0 {
		return nil
	}
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
			log.Error(ctx, "Create cluster manager grpc client with %s error: %v", addr, err)
			continue
		}
		if conn != nil {
			break
		}
	}
	if conn == nil {
		log.Error(ctx, "create no cluster manager client after all instance tries")
		return nil
	}
	return bcsapicm.NewClusterManagerClient(conn)
}
