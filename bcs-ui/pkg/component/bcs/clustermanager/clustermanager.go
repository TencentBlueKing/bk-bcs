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

// Package clustermanager xxx
package clustermanager

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"github.com/patrickmn/go-cache"
	microRgt "go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/constants"
)

const (
	// ClusterManagerServiceName cluster manager service name
	ClusterManagerServiceName = "clustermanager.bkbcs.tencent.com"

	// cache key
	cacheClusterIDKeyPrefix = "cluster_%s"

	// defaultExpiration
	defaultExpiration = 10 * time.Minute
)

// Client xxx
type Client struct {
	Cache *cache.Cache
}

var client *Client

// NewClient create cluster manager service client
func NewClient(tlsConfig *tls.Config, microRgt microRgt.Registry) error {
	if !discovery.UseServiceDiscovery() {
		dis := discovery.NewModuleDiscovery(ClusterManagerServiceName, microRgt)
		err := dis.Start()
		if err != nil {
			return err
		}
		clustermanager.SetClientConfig(tlsConfig, dis)
	} else {
		clustermanager.SetClientConfig(tlsConfig, nil)
	}
	client = &Client{
		Cache: cache.New(defaultExpiration, cache.NoExpiration),
	}
	return nil
}

// GetCluster get cluster from cluster manager
func GetCluster(ctx context.Context, clusterID string) (*clustermanager.Cluster, error) {
	key := fmt.Sprintf(cacheClusterIDKeyPrefix, clusterID)
	v, ok := client.Cache.Get(key)
	if ok {
		if cluster, ok := v.(*clustermanager.Cluster); ok {
			return cluster, nil
		}
	}
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.GetCluster(ctx, &clustermanager.GetClusterReq{ClusterID: clusterID})
	if err != nil {
		return nil, fmt.Errorf("GetCluster error: %s", err)
	}
	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("GetCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	client.Cache.Set(key, p.Data, defaultExpiration)
	return p.Data, nil
}
