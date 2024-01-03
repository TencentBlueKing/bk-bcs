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

package brpc

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// PoolInterface defines the gRPC client pool supported operations.
type PoolInterface interface {
	// Pick one gRPC client from the gRPC client pool
	Pick() interface{}
}

// NewClientPool create an gRPC client pool instance.
func NewClientPool(opt PoolOption) (PoolInterface, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	p := &pool{
		maxIndex: opt.PoolSize - 1,
		curIndex: 0,
		cons:     make([]interface{}, 0),
	}

	for i := 0; i < opt.PoolSize; i++ {
		one, err := newOneClient(opt)
		if err != nil {
			return nil, err
		}

		p.cons = append(p.cons, one)
	}

	return p, nil
}

type pool struct {
	lock     sync.Mutex
	maxIndex int
	curIndex int
	cons     []interface{}
}

// Pick one gRPC client from the gRPC client pool
func (p *pool) Pick() interface{} {
	p.lock.Lock()
	defer p.lock.Unlock()

	picked := p.cons[p.curIndex]

	if p.curIndex == p.maxIndex {
		p.curIndex = 0
	} else {
		p.curIndex++
	}

	return picked
}

func newOneClient(opt PoolOption) (interface{}, error) {

	kpOpt := keepalive.ClientParameters{
		Time:                30 * time.Second,
		PermitWithoutStream: true,
	}

	opts := make([]grpc.DialOption, 0)
	opts = append(opts, opt.SvrDiscover.LBRoundRobin(),
		grpc.WithWriteBufferSize(opt.WriteBufferSizeMB*1024*1024),
		grpc.WithReadBufferSize(opt.ReadBufferSizeMB*1024*1024),
		grpc.WithKeepaliveParams(kpOpt))

	tls := opt.TLS

	if !opt.TLS.Enable() {
		// dial without ssl
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// dial with ssl.
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return nil, fmt.Errorf("init tls failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	conn, err := grpc.Dial(serviced.GrpcServiceDiscoveryName(opt.ServiceName), opts...)
	if err != nil {
		return nil, fmt.Errorf("dial service %s failed, err: %v", opt.ServiceName, err)
	}

	// Note: add ping test and wait for service ready.

	return opt.NewClient(conn), nil
}
