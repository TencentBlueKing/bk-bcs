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

// Package store is storage for cloud netservice
package store

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
)

// Interface for store data
type Interface interface {
	CreateSubnet(ctx context.Context, subnet *types.CloudSubnet) error
	DeleteSubnet(ctx context.Context, subnetID string) error
	UpdateSubnetState(ctx context.Context, subnetID string, state, minIPNumPerEni int32) error
	UpdateSubnetAvailableIP(ctx context.Context, subnetID string, availableIP int64) error
	ListSubnet(ctx context.Context, labelsMap map[string]string) ([]*types.CloudSubnet, error)
	GetSubnet(ctx context.Context, subnetID string) (*types.CloudSubnet, error)

	CreateIPObject(ctx context.Context, obj *types.IPObject) error
	UpdateIPObject(ctx context.Context, obj *types.IPObject) (*types.IPObject, error)
	DeleteIPObject(ctx context.Context, ip string) error
	GetIPObject(ctx context.Context, ip string) (*types.IPObject, error)
	ListIPObject(ctx context.Context, labelsMap map[string]string) ([]*types.IPObject, error)
	ListIPObjectByField(ctx context.Context, fieldKey string, fieldValue string) ([]*types.IPObject, error)

	GetIPQuota(ctx context.Context, cluster string) (*types.IPQuota, error)
	CreateIPQuota(ctx context.Context, quota *types.IPQuota) error
	UpdateIPQuota(ctx context.Context, quota *types.IPQuota) error
	DeleteIPQuota(ctx context.Context, cluster string) error
	ListIPQuota(ctx context.Context) ([]*types.IPQuota, error)
}
