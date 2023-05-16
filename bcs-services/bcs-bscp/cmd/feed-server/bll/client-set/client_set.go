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

package clientset

import (
	"fmt"

	"bscp.io/pkg/cc"
	iamauth "bscp.io/pkg/iam/auth"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/cache-service"
	"bscp.io/pkg/runtime/brpc"
	"bscp.io/pkg/serviced"

	"google.golang.org/grpc"
)

// New create a client set instance.
func New(sd serviced.Discover, authorizer iamauth.Authorizer) (*ClientSet, error) {
	cs, err := newClientSet(sd, cc.FeedServer().Network.TLS, authorizer)
	if err != nil {
		return nil, fmt.Errorf("new client set failed, err: %v", err)
	}

	return cs, nil
}

func newClientSet(sd serviced.Discover, tls cc.TLSConfig, authorizer iamauth.Authorizer) (*ClientSet, error) {
	logs.Infof("start initialize the client set.")

	opt := brpc.PoolOption{
		PoolSize:          10,
		ReadBufferSizeMB:  16,
		WriteBufferSizeMB: 32,
		ServiceName:       cc.CacheServiceName,
		SvrDiscover:       sd,
		TLS:               tls,
		NewClient: func(conn *grpc.ClientConn) interface{} {
			return pbcs.NewCacheClient(conn)
		},
	}

	cachePool, err := brpc.NewClientPool(opt)
	if err != nil {
		return nil, fmt.Errorf("new cache service client pool failed, err: %v", err)
	}

	return &ClientSet{
		cachePool:  cachePool,
		authorizer: authorizer,
	}, nil

}

// ClientSet defines configure server's all the depends api client.
type ClientSet struct {
	cachePool  brpc.PoolInterface
	authorizer iamauth.Authorizer
}

// CS return one cache service client from the client pool
func (cs *ClientSet) CS() pbcs.CacheClient {
	return cs.cachePool.Pick().(pbcs.CacheClient)
}

// Authorizer return an authorization client
func (cs *ClientSet) Authorizer() iamauth.Authorizer {
	return cs.authorizer
}
