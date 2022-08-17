/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resourcemgr

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	grpc "github.com/asim/go-micro/plugins/client/grpc/v4"
	etcd "github.com/asim/go-micro/plugins/registry/etcd/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	impl "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/resourcemgr/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

const (
	// success definition for service calls
	success = 0
)

// ListOptions options for list resoruce pools
type ListOptions struct {
	PageSize int
}

// GetOptions options for list resoruce pools
type GetOptions struct {
	GetCacheIfEmpty bool
}

// ClientOptions for local resource-manager client
type ClientOptions struct {
	// Name for resource-manager registry
	Name string
	// Etcd endpoints information
	Etcd []string
	// EtcdConfig tls config for etcd
	EtcdConfig *tls.Config
	// ClientConfig tls config
	ClientConfig *tls.Config
	// Cache for store ResourcePool information
	Cache storage.Storage
}

//Client for resource-manager
type Client interface {
	// ListResourcePools list all RessourcePool in resource-manager
	ListResourcePools(option *ListOptions) ([]*storage.ResourcePool, error)
	// GetResourcePool get specified ResourcePool according poolID
	GetResourcePool(poolID string, option *GetOptions) (*storage.ResourcePool, error)
}

// New create resource-manager client instance
func New(opt *ClientOptions) Client {
	//init go-micro v2 client instance
	c := grpc.NewClient(
		client.Registry(etcd.NewRegistry(
			registry.Addrs(opt.Etcd...),
			registry.TLSConfig(opt.EtcdConfig)),
		),
		grpc.AuthTLS(opt.ClientConfig),
	)
	// create resource-manager go-micro client api
	return &innerClient{
		client: impl.NewResourceManagerService(opt.Name, c),
	}
}

// innerClient
type innerClient struct {
	client impl.ResourceManagerService
}

// ListResourcePools list all resource pool in resource-manager
func (c *innerClient) ListResourcePools(option *ListOptions) ([]*storage.ResourcePool, error) {
	req := &impl.ListResourcePoolReq{}
	resp, err := c.client.ListResourcePool(context.Background(), req)
	if err != nil {
		blog.Errorf("list all resource pool from resource-manager failed, %s", err.Error())
		return nil, err
	}
	if resp.GetCode() != success {
		blog.Errorf("list all resource pool failed, resource-manager logic err: %s", resp.GetMessage())
		return nil, fmt.Errorf("resource-manager list resource pool err, %s", resp.GetMessage())
	}
	// convert to local ResourcePool definition
	// todo(DeveloperJim): uselessness, evaluate that clean these codes in future.
	return nil, fmt.Errorf("Not Implemented")
}

// GetResourcePool get detail information of resourcepool
func (c *innerClient) GetResourcePool(poolID string, option *GetOptions) (*storage.ResourcePool, error) {
	if len(poolID) == 0 {
		return nil, fmt.Errorf("lost resource pool ID")
	}
	req := &impl.ListResourceReq{
		PoolID: &poolID,
	}
	resp, err := c.client.ListResource(context.Background(), req)
	if err != nil {
		blog.Errorf("get resource pool details from resource-manager failed, %s", err.Error())
		return nil, err
	}
	if resp.GetCode() != success {
		blog.Errorf("get resource pool details failed, resource-manager logic err: %s", resp.GetMessage())
		return nil, fmt.Errorf("resource-manager logic failure, %s", resp.GetMessage())
	}
	//convert details to local ResourcePool definition
	if len(resp.Data) == 0 {
		blog.Errorf("resource-manager response empty Resource from ResourcePool %s", poolID)
		return nil, fmt.Errorf("empty resources response")
	}
	pool := convertToResourcePool(poolID, resp.Data)
	return pool, nil
}

func convertToResourcePool(poolID string, res []*impl.Resource) *storage.ResourcePool {
	pool := &storage.ResourcePool{
		ID:   poolID,
		Name: poolID,
		// UpdatedTime must compare with local cache
		// UpdatedTime: time.Now(),
	}
	var latest int64
	for _, r := range res {
		resource, updated := convertResourceToLocal(r)
		switch resource.Phase {
		case storage.NodeIdleState:
			pool.IdleNum++
		case storage.NodeInitState:
			pool.InitNum++
		case storage.NodeReturnState:
			pool.ReturnedNum++
		case storage.NodeConsumedState:
			pool.ConsumedNum++
		default:
			// unknown status set as Consumed
			pool.ConsumedNum++
		}
		if updated > latest {
			latest = updated
		}
	}
	pool.UpdatedTime = time.Unix(latest, 0)
	return pool
}

func convertResourceToLocal(r *impl.Resource) (*storage.Resource, int64) {
	resource := &storage.Resource{
		ID:               r.GetId(),
		InnerIP:          r.GetInnerIP(),
		InnerIPv6:        r.GetInnerIPv6(),
		ResourceType:     r.GetResourceType(),
		ResourceProvider: r.GetResourceProvider(),
		Labels:           r.GetLabels(),
		UpdatedTime:      time.Unix(*r.UpdateTime, 0),
	}
	if r.Status != nil {
		resource.Phase = r.Status.GetPhase()
		resource.Cluster = r.Status.GetClusterID()
		resource.DevicePool = r.Status.GetDevicePoolID()
	} else {
		// feature protection(DeveloperJim): consider that unknown status is Consumed
		resource.Phase = storage.NodeConsumedState
	}
	return resource, *r.UpdateTime
}
