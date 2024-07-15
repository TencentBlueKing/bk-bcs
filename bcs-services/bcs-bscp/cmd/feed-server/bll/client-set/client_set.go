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

// Package clientset provides a collection of clients for communication with external services.
package clientset

import (
	"fmt"

	"google.golang.org/grpc"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/bedis"
	iamauth "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/brpc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
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

	// init redis client
	bds, err := bedis.NewRedisCache(cc.FeedServer().RedisCluster)
	if err != nil {
		return nil, fmt.Errorf("new redis cluster failed, err: %v", err)
	}

	return &ClientSet{
		cachePool:  cachePool,
		authorizer: authorizer,
		bds:        bds,
	}, nil

}

// ClientSet defines configure server's all the depends api client.
type ClientSet struct {
	cachePool  brpc.PoolInterface
	authorizer iamauth.Authorizer
	bds        bedis.Client
}

// CS return one cache service client from the client pool
func (cs *ClientSet) CS() pbcs.CacheClient {
	return cs.cachePool.Pick().(pbcs.CacheClient)
}

// Authorizer return an authorization client
func (cs *ClientSet) Authorizer() iamauth.Authorizer {
	return cs.authorizer
}

// Redis return redis client
func (cs *ClientSet) Redis() bedis.Client {
	return cs.bds
}
