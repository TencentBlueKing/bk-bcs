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

package service

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/feed-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/grpcgw"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// newFeedServerMux new config server mux.
func newFeedServerMux() (*runtime.ServeMux, error) {
	opts := make([]grpc.DialOption, 0)

	network := cc.FeedServer().Network
	tls := network.TLS
	if !tls.Enable() {
		// dial without ssl
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// dial with ssl.
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return nil, fmt.Errorf("init grpc tls config failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	// build conn.
	addr := net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.RpcPort)))
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		logs.Errorf("dial config server failed, err: %v", err)
		return nil, err
	}

	// new grpc mux.
	mux := newGrpcMux()

	// register client to mux.
	if err = pbfs.RegisterUpstreamHandler(context.Background(), mux, conn); err != nil {
		logs.Errorf("register config server handler client failed, err: %v", err)
		return nil, err
	}

	return mux, nil
}

// newGrpcMux new grpc mux that has some processing of built-in http request to grpc request.
func newGrpcMux() *runtime.ServeMux {
	return runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		if key == "Authorization" {
			return key, true
		}
		return runtime.DefaultHeaderMatcher(key)
	}), grpcgw.JsonMarshalerOpt)
}
