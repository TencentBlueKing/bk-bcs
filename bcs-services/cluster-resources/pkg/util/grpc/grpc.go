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

	microMetadata "go-micro.dev/v4/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// NewGrpcConn 新建 Grpc 连接
func NewGrpcConn(address string, tlsConf *tls.Config) (conn *grpc.ClientConn, err error) {
	// 组装配置信息
	md := metadata.New(map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	})
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	if tlsConf != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// 尝试建立 grpc 连接
	return grpc.Dial(address, opts...)
}

// SetMD4CTX 为调用 Grpc 的 Context 设置 Metadata
func SetMD4CTX(ctx context.Context) context.Context {
	// 若存在 jwtToken 则透传到依赖服务
	rawMetadata, ok := microMetadata.FromContext(ctx)
	if ok {
		authorization, exists := rawMetadata.Get("Authorization")
		if exists {
			md := metadata.New(map[string]string{
				"Authorization": authorization,
			})
			return metadata.NewOutgoingContext(ctx, md)
		}
	}
	return ctx
}
