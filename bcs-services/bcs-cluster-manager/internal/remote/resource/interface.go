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
 *
 */

package resource

import "context"

// ManagerResource interface shield cloud resource manager, user can use different resource management systems
type ManagerResource interface {
	// CreateResourcePool create resource pool for resource manager
	CreateResourcePool(ctx context.Context, info ResourcePoolInfo) (string, error)
	// DeleteResourcePool delete resource pool for resource manager
	DeleteResourcePool(ctx context.Context, poolID string) error
	// ApplyInstances apply instances
	ApplyInstances(ctx context.Context, instanceCount int, paras *ApplyInstanceReq) (*ApplyInstanceResp, error)
	// DestroyInstances destroy instances
	DestroyInstances(ctx context.Context, paras *DestroyInstanceReq) (*DestroyInstanceResp, error)
	// CheckOrderStatus check instance status by orderID
	CheckOrderStatus(ctx context.Context, orderID string) (*OrderInstanceList, error)
	// CheckInstanceStatus check instanceStatus by instanceID
	CheckInstanceStatus(ctx context.Context, instanceIDs []string) (*OrderInstanceList, error)
	// GetInstanceTypes get region instance types
	GetInstanceTypes(ctx context.Context, region string, spec InstanceSpec) ([]InstanceType, error)
	// GetDeviceInfoByDeviceID get device detailed info
	GetDeviceInfoByDeviceID(ctx context.Context, deviceID string) (*DeviceInfo, error)
}
