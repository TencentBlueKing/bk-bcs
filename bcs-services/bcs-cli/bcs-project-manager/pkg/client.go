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

package pkg

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Config describe the options Client need
type Config struct {
	// APIServer for bcs-api-gateway address
	APIServer string
	// AuthToken for bcs permission token
	AuthToken string
	// Operator for the bk-repo operations
	Operator string
}

func NewGrpcClientConn(ctx context.Context) (*grpc.ClientConn, context.Context, error) {
	config := &Config{
		APIServer: viper.GetString("bcs.apiserver"),
		AuthToken: viper.GetString("bcs.token"),
		Operator:  viper.GetString("bcs.operator"),
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
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(config.APIServer, opts...)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "create grpc client with '%s' failed", config.APIServer)
	}
	if conn == nil {
		return nil, nil, fmt.Errorf("conn is nil")
	}
	return conn, metadata.NewOutgoingContext(ctx, md), nil
}
