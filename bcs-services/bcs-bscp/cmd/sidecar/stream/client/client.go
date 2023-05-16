/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package client NOTES
package client

import (
	"context"
	"fmt"
	"time"

	"bscp.io/cmd/sidecar/stream/types"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbfs "bscp.io/pkg/protocol/feed-server"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/version"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

// Interface implement all the client which is used to connect with upstream
// servers.
// Note: if the client is reconnecting to the upstream servers, it will block
// all the requests with a timeout so that these requests can use the new connection
// to connect with the upstream server.
type Interface interface {
	ReconnectUpstreamServer() error
	Handshake(vas *kit.Vas, spec *types.SidecarSpec) (*pbfs.HandshakeResp, error)
	Watch(vas *kit.Vas, payload []byte) (pbfs.Upstream_WatchClient, error)
	Messaging(vas *kit.Vas, typ sfs.MessagingType, payload []byte) (*pbfs.MessagingResp, error)
	EnableBounce(bounceIntervalHour uint)
}

// New create a rolling client instance.
func New(opt cc.SidecarUpstream) (Interface, error) {

	lb, err := newBalancer(opt.Endpoints)
	if err != nil {
		return nil, err
	}

	dialOpts := make([]grpc.DialOption, 0)
	// blocks until the connection is established.
	dialOpts = append(dialOpts, grpc.WithBlock())
	// TODO: confirm this
	dialOpts = append(dialOpts, grpc.WithUserAgent("bscp-sidecar"))

	if !opt.TLS.Enable() {
		// dial without ssl
		dialOpts = append(dialOpts, grpc.WithInsecure())
	} else {
		// dial with ssl.
		tlsC, err := tools.ClientTLSConfVerify(opt.TLS.InsecureSkipVerify, opt.TLS.CAFile, opt.TLS.CertFile,
			opt.TLS.KeyFile, opt.TLS.Password)
		if err != nil {
			return nil, fmt.Errorf("init upsteram client tls config failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(cred))
	}

	ver := version.SemanticVersion()
	rc := &rollingClient{
		options: &opt,
		sidecarVer: &pbbase.Versioning{
			Major: ver[0],
			Minor: ver[1],
			Patch: ver[2],
		},
		dialOpts: dialOpts,
		lb:       lb,
		wait:     initBlocker(),
	}

	rc.bounce = initBounce(rc.ReconnectUpstreamServer)

	if err := rc.dial(); err != nil {
		return nil, err
	}

	go rc.waitForStateChange()

	return rc, nil
}

// rollingClient is an implementation of the upstream server's client, it sends to and receive messages from
// the upstream feed server.
// Note:
// 1. it also hijacked the connections to upstream server so that it can
// do reconnection, bounce work and so on.
// 2. it blocks the request until the connections to the upstream server go back to normal when the connection
// is unavailable.
type rollingClient struct {
	options    *cc.SidecarUpstream
	sidecarVer *pbbase.Versioning
	dialOpts   []grpc.DialOption
	// cancelCtx cancel ctx is used to cancel the upstream connection.
	cancelCtx context.CancelFunc
	lb        *balancer
	bounce    *bounce

	wait     *blocker
	conn     *grpc.ClientConn
	upstream pbfs.UpstreamClient
}

// dial blocks until the connection is established.
func (rc *rollingClient) dial() error {

	if rc.conn != nil {
		if err := rc.conn.Close(); err != nil {
			logs.Errorf("close the previous connection failed, err: %v", err)
			// do not return here, the new connection will be established.
		}
	}

	timeout := rc.options.DialTimeoutMS
	if timeout == 0 {
		// set the default timeout time is 2 second.
		timeout = 2000
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Millisecond)
	endpoint := rc.lb.PickOne()
	conn, err := grpc.DialContext(ctx, endpoint, rc.dialOpts...)
	if err != nil {
		cancel()
		rc.cancelCtx = nil
		return fmt.Errorf("dial upstream grpc server failed, err: %v", err)
	}

	logs.Infof("dial upstream server %s success.", endpoint)

	rc.cancelCtx = cancel
	rc.conn = conn
	rc.upstream = pbfs.NewUpstreamClient(conn)

	return nil
}

// ReconnectUpstreamServer blocks until the new connection is established with dial again.
func (rc *rollingClient) ReconnectUpstreamServer() error {
	if !rc.wait.TryBlock() {
		logs.Warnf("received reconnect to upstream server request, but another reconnect is processing, ignore this")
		return nil
	}
	// got the block lock for now.

	defer rc.wait.Unblock()
	if err := rc.dial(); err != nil {
		return fmt.Errorf("reconnect upstream server failed because of %v", err)
	}

	return nil
}

// EnableBounce set conn reconnect interval, and start loop wait connect bounce. call multiple times,
// you need to wait for the last bounce interval to arrive, the bounce interval of set this time
// will take effect.
func (rc *rollingClient) EnableBounce(bounceIntervalHour uint) {
	rc.bounce.updateInterval(bounceIntervalHour)

	if !rc.bounce.state() {
		go rc.bounce.enableBounce()
	}

	return
}

// waitForStateChange use the connection state to determine what to do next.
func (rc *rollingClient) waitForStateChange() {
	for {
		if rc.conn.WaitForStateChange(context.TODO(), connectivity.Ready) {
			// TODO: loop and wait and then determine whether we need to create a
			// new connection
		}
	}

}
