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

// Package grpc NOTES
package grpc

import (
	"io"
	"net"
	"net/url"
	"path"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	httpproxy "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-proxy/proxy/http"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/feed-server"
)

var (
	clientStreamDescForProxying = &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}
)

type handler struct {
	director Director
}

// handler is where the real magic of proxying happens.
// It is invoked like any gRPC server stream and uses the gRPC server framing to get and receive bytes from the wire,
// forwarding it to a ClientStream established against the relevant ClientConn.
func (s *handler) handler(srv interface{}, serverStream grpc.ServerStream) error {
	// little bit of gRPC internals never hurt anyone
	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}
	// We require that the director's returned context inherits from the serverStream.Context().
	outgoingCtx, backendConn, err := s.director.Director(serverStream.Context(), fullMethodName)
	if err != nil {
		return err
	}
	clientCtx, clientCancel := context.WithCancel(outgoingCtx)
	defer func() {
		clientCancel()
		if e := backendConn.Close(); e != nil {
			logs.Errorf("close feedserver backend conn failed: %v", e)
		}
	}()

	// TODO: Add a `forwarded` header to metadata, https://en.wikipedia.org/wiki/X-Forwarded-For.
	clientStream, err := grpc.NewClientStream(clientCtx, clientStreamDescForProxying, backendConn.Value(), fullMethodName)
	if err != nil {
		return err
	}

	// Special case for GetDownloadURL, we need to replace the host in the response URL.
	if fullMethodName == pbfs.Upstream_GetDownloadURL_FullMethodName {
		f := &emptypb.Empty{}
		if err := serverStream.RecvMsg(f); err != nil {
			return err
		}
		if err := clientStream.SendMsg(f); err != nil {
			return err
		}

		resp := &pbfs.GetDownloadURLResp{}
		if err := clientStream.RecvMsg(resp); err != nil {
			return err
		}
		network := cc.FeedProxy().Network
		targetHost := net.JoinHostPort(network.BindIP, strconv.Itoa(int(network.HttpPort)))
		resp.Url = replaceHost(resp.Url, targetHost)
		if err := serverStream.SendMsg(resp); err != nil {
			return err
		}
		return nil
	}

	// Explicitly *do not close* s2cErrChan and c2sErrChan, otherwise the select below will not terminate.
	// Channels do not have to be closed, it is just a control flow mechanism, see
	// https://groups.google.com/forum/#!msg/golang-nuts/pZwdYRGxCIk/qpbHxRRPJdUJ
	s2cErrChan := s.forwardServerToClient(serverStream, clientStream)
	c2sErrChan := s.forwardClientToServer(clientStream, serverStream)
	// We don't know which side is going to stop sending first, so we need a select between the two.
	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if s2cErr != io.EOF {
				// we may have gotten a receive error (stream disconnected, a read error etc) in which case we need
				// to cancel the clientStream to the backend, let all of its goroutines be freed up by the CancelFunc and
				// exit with an error to the stack
				return status.Errorf(codes.Internal, "failed proxying s2c: %v", s2cErr)
			}
			// this is the happy case where the sender has encountered io.EOF, and won't be sending anymore./
			// the clientStream>serverStream may continue pumping though.
			_ = clientStream.CloseSend()
		case c2sErr := <-c2sErrChan:
			// This happens when the clientStream has nothing else to offer (io.EOF), returned a gRPC error. In those two
			// cases we may have received Trailers as part of the call. In case of other errors (stream closed) the trailers
			// will be nil.
			serverStream.SetTrailer(clientStream.Trailer())
			// c2sErr will contain RPC error from client code. If not io.EOF return the RPC error as server stream error.
			if c2sErr != io.EOF {
				return c2sErr
			}
			return nil
		}
	}

	return status.Errorf(codes.Internal, "gRPC proxying should never reach this stage.")
}

func (s *handler) forwardClientToServer(src grpc.ClientStream, dst grpc.ServerStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &emptypb.Empty{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if i == 0 {
				// This is a bit of a hack, but client to server headers are only readable after first client msg is
				// received but must be written to server stream before the first msg is flushed.
				// This is the only place to do it nicely.
				md, err := src.Header()
				if err != nil {
					ret <- err
					break
				}
				if err := dst.SendHeader(md); err != nil {
					ret <- err
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func (s *handler) forwardServerToClient(src grpc.ServerStream, dst grpc.ClientStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &emptypb.Empty{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func replaceHost(inputURL, targetHost string) string {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		// Handle error according to your needs, here we simply return the original URL
		return inputURL
	}

	// Replace Host
	parsedURL.Scheme = "http"
	parsedURL.Host = targetHost
	parsedURL.Path = path.Join(httpproxy.ProxyDownloadPrefix, parsedURL.Path)
	parsedURL.RawPath = path.Join(httpproxy.ProxyDownloadPrefix, parsedURL.RawPath)

	// Building URL back to string format and return
	return parsedURL.String()
}
