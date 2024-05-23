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

package grpc

import (
	"time"

	"github.com/shimingyah/pool"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// ClientConn is a wrapper around a *grpc.ClientConn with closable
type ClientConn interface {
	Value() *grpc.ClientConn
	Close() error
}

// Director is a function that returns a gRPC ClientConn to be used to forward the call to.
type Director interface {

	// Director returns a gRPC ClientConn to be used to forward the call to.
	//
	// The presence of the `Context` allows for rich filtering, e.g. based on Metadata (headers).
	// If no handling is meant to be done, a `codes.NotImplemented` gRPC error should be returned.
	//
	// The context returned from this function should be the context for the *outgoing* (to backend) call. In case you want
	// to forward any Metadata between the inbound request and outbound requests, you should do it manually. However, you
	// *must* propagate the cancel function (`context.WithCancel`) of the inbound context to the one returned.
	//
	// It is worth noting that the Director will be fired *after* all server-side stream interceptors
	// are invoked. So decisions around authorization, monitoring etc. are better to be handled there.
	Director(ctx context.Context, fullMethodName string) (context.Context, ClientConn, error)
}

// TransparentHandler returns a handler that attempts to proxy all requests that are not registered in the server.
// The indented use here is as a transparent proxy, where the server doesn't know about the services implemented by the
// backends. It should be used as a `grpc.UnknownServiceHandler`.
//
// This can *only* be used if the `server` also uses grpcproxy.CodecForServer() ServerOption.
func TransparentHandler(director Director) grpc.StreamHandler {
	streamer := &handler{director}
	return streamer.handler
}

// FeedServerDirector is a Director implementation that routes all requests to the upstream feed server.
type FeedServerDirector struct {
	grpcPool pool.Pool
}

// NewFeedServerDirector returns a new FeedServerDirector
func NewFeedServerDirector() (*FeedServerDirector, error) {
	feedHost := cc.FeedProxy().Upstream.FeedServerHost
	grpcPool, err := pool.New(feedHost, pool.Options{
		Dial: func(address string) (*grpc.ClientConn, error) {
			timeoutCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
			return grpc.DialContext(timeoutCtx, feedHost,
				grpc.WithBlock(),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:                30 * time.Second,
					Timeout:             3 * time.Second,
					PermitWithoutStream: true,
				}))
		},
		MaxIdle:              8,
		MaxActive:            64,
		MaxConcurrentStreams: 64,
		Reuse:                true,
	})
	if err != nil {
		return nil, err
	}
	return &FeedServerDirector{grpcPool}, nil
}

// Director returns a gRPC ClientConn to be used to forward the call to.
func (d *FeedServerDirector) Director(ctx context.Context, fullMethodName string) (context.Context, ClientConn, error) {

	md, _ := metadata.FromIncomingContext(ctx)
	// Copy the bound metadata explicitly.
	outCtx, _ := context.WithCancel(ctx)
	outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())

	conn, err := d.grpcPool.Get()
	if err != nil {
		return outCtx, nil, err
	}

	return outCtx, conn, nil
}

// Terminate closes the grpc pool
func (d *FeedServerDirector) Terminate() error {
	if err := d.grpcPool.Close(); err != nil {
		logs.Warnf("close grpc pool failed, err %s", err.Error())
		return err
	}
	return nil
}
