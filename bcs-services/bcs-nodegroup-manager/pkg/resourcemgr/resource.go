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
	// LabelKeyDrainDelay 节点抽离时间
	LabelKeyDrainDelay = "nodeDrainDelay"
	// LabelKeyDeadline 节点抽离ddl
	LabelKeyDeadline = "nodeDeadline"
)

// ListOptions options for list resource pools
type ListOptions struct {
	PageSize int
}

// GetOptions options for list resource pools
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

// Client for resource-manager
type Client interface {
	// ListResourcePools list all RessourcePool in resource-manager
	ListResourcePools(option *ListOptions) ([]*storage.ResourcePool, error)
	// GetResourcePool get specified ResourcePool according poolID
	GetResourcePool(poolID string, option *GetOptions) (*storage.ResourcePool, error)
	// GetResourcePoolByCondition get resource by condition, poolID is essential
	GetResourcePoolByCondition(poolID, consumerID, deviceRecord string, option *GetOptions) (*storage.ResourcePool, error)
	// ListTasks list scale down tasks from resource manager
	ListTasks(poolID, consumerID string, option *ListOptions) ([]*storage.ScaleDownTask, error)
	// GetTaskByID get task by record id
	GetTaskByID(recordID string, opt *GetOptions) (*storage.ScaleDownTask, error)
}

// New create resource-manager client instance
func New(opt *ClientOptions) Client {
	// init go-micro v2 client instance
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
	// need to do: uselessness, evaluate that clean these codes in future.
	return nil, fmt.Errorf("Not Implemented")
}

// GetResourcePool get detail information of resource pool
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
	// convert details to local ResourcePool definition
	if len(resp.Data) == 0 {
		blog.Errorf("resource-manager response empty Resource from ResourcePool %s", poolID)
		return nil, fmt.Errorf("empty resources response")
	}
	pool := convertToResourcePool(poolID, resp.Data)
	return pool, nil
}

// GetResourcePoolByCondition get detail information of resource pool
func (c *innerClient) GetResourcePoolByCondition(poolID, consumerID, deviceRecord string,
	option *GetOptions) (*storage.ResourcePool, error) {
	if len(poolID) == 0 {
		return nil, fmt.Errorf("lost resource pool ID")
	}
	req := &impl.ListResourceReq{
		PoolID: &poolID,
	}
	if consumerID != "" {
		req.MatchConsumerID = &consumerID
	}
	if deviceRecord != "" {
		req.MatchDeviceRecordID = &deviceRecord
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
		blog.Errorf("resource-manager response empty Resource from ResourcePool %s, consumerID:%s, "+
			"deviceRecord:%s", poolID, consumerID, deviceRecord)
		return nil, fmt.Errorf("empty resources response")
	}
	pool := convertToResourcePool(poolID, resp.Data)
	return pool, nil
}

func (c *innerClient) ListTasks(poolID, consumerID string, opt *ListOptions) ([]*storage.ScaleDownTask, error) {
	localTasks := make([]*storage.ScaleDownTask, 0)
	req := &impl.ListDeviceRecordByPoolReq{
		PoolID: &poolID,
	}
	resp, err := c.client.ListDeviceRecordByPool(context.Background(), req)
	if err != nil {
		blog.Errorf("get device records from resource-manager failed, %s", err.Error())
		return nil, err
	}
	if resp.GetCode() != success {
		blog.Errorf("get device records failed, resource-manager logic err: %s", resp.GetMessage())
		return nil, fmt.Errorf("resource-manager logic failure, %s", resp.GetMessage())
	}
	//convert details to local ResourcePool definition
	if len(resp.Data) == 0 {
		blog.Infof("resource-manager response empty device records from ResourcePool %s", poolID)
		return nil, fmt.Errorf("empty resources response")
	}
	for _, deviceRecord := range resp.Data {
		blog.Infof("recordId: %s, consumerID:%s", deviceRecord.GetId(), consumerID)
		_, err := c.GetResourcePoolByCondition(poolID, consumerID, deviceRecord.GetId(), nil)
		if err != nil {
			continue
		}
		// deviceList 是所有可以缩容的机器
		task := convertTaskToLocal(deviceRecord)
		blog.Infof("task:%v", task)
		localTasks = append(localTasks, task)
	}
	return localTasks, nil
}

func (c *innerClient) GetTaskByID(recordID string, opt *GetOptions) (*storage.ScaleDownTask, error) {
	req := &impl.GetDeviceRecordReq{DeviceRecordID: &recordID}
	resp, err := c.client.GetDeviceRecord(context.Background(), req)
	if err != nil {
		blog.Errorf("get device record by id %s from resource-manager failed, %s", recordID, err.Error())
		return nil, err
	}
	if resp.GetCode() != success {
		blog.Errorf("get device records failed, resource-manager logic err: %s", resp.GetMessage())
		return nil, fmt.Errorf("resource-manager logic failure, %s", resp.GetMessage())
	}
	return convertTaskToLocal(resp.Data), nil
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

func convertTaskToLocal(deviceRecord *impl.DeviceRecord) *storage.ScaleDownTask {
	deadline, _ := time.Parse(time.RFC3339, deviceRecord.GetConsumerLabels()[LabelKeyDeadline])
	return &storage.ScaleDownTask{
		TaskID:     deviceRecord.GetId(),
		TotalNum:   int(deviceRecord.GetNum()),
		Status:     deviceRecord.GetStatus(),
		DrainDelay: deviceRecord.GetConsumerLabels()[LabelKeyDrainDelay],
		Deadline:   deadline,
	}
}
