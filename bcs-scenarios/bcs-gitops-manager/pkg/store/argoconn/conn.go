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

// Package argoconn xxx
package argoconn

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	api "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	grpc_util "github.com/argoproj/argo-cd/v2/util/grpc"
	argoio "github.com/argoproj/argo-cd/v2/util/io"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type jwtCredentials struct {
	Token string
}

// RequireTransportSecurity mock function
func (c jwtCredentials) RequireTransportSecurity() bool {
	return false
}

// GetRequestMetadata return the metadata with token
func (c jwtCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		api.MetaDataTokenKey: c.Token,
	}, nil
}

func parseHeaders(headerStrings []string) (http.Header, error) {
	headers := http.Header{}
	for _, kv := range headerStrings {
		i := strings.IndexByte(kv, ':')
		// zero means meaningless empty header name
		if i <= 0 {
			return nil, fmt.Errorf("additional headers must be colon(:)-separated: %s", kv)
		}
		headers.Add(kv[0:i], kv[i+1:])
	}
	return headers, nil
}

// NewConn will create the connection to argocd server.
// Refer to: https://github.com/argoproj/argo-cd/blob/v2.8.1/pkg/apiclient/apiclient.go#L488
func NewConn(op *api.ClientOptions) (*grpc.ClientConn, io.Closer, error) {
	closers := make([]io.Closer, 0)
	serverAddr := op.ServerAddr
	network := "tcp"

	var creds = credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true, // nolint
	})

	endpointCredentials := jwtCredentials{
		Token: op.AuthToken,
	}
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithMax(3),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(1000 * time.Millisecond)),
	}
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(endpointCredentials))
	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(api.MaxGRPCMessageSize), grpc.MaxCallSendMsgSize(api.MaxGRPCMessageSize)))
	dialOpts = append(dialOpts, grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)))
	dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(
		grpc_middleware.ChainUnaryClient(grpc_retry.UnaryClientInterceptor(retryOpts...))))
	dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	dialOpts = append(dialOpts, grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))

	ctx := context.Background()

	headers, err := parseHeaders(op.Headers)
	if err != nil {
		return nil, nil, err
	}
	for k, vs := range headers {
		for _, v := range vs {
			ctx = metadata.AppendToOutgoingContext(ctx, k, v)
		}
	}

	conn, e := grpc_util.BlockingDial(ctx, network, serverAddr, creds, dialOpts...)
	closers = append(closers, conn)
	return conn, argoio.NewCloser(func() error {
		var firstErr error
		for i := range closers {
			err := closers[i].Close()
			if err != nil {
				firstErr = err
			}
		}
		return firstErr
	}), e
}
