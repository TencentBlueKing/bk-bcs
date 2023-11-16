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

// Package vault NOTES
package vault

import (
	vault "github.com/hashicorp/vault/api"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// Set ...
type Set interface {
	// IsMountPathExists 挂载目录是否存在
	IsMountPathExists(path string) (bool, error)
	// CreateMountPath 创建挂载目录
	CreateMountPath(path string, config *vault.MountInput) error
	// UpsertKv 创建｜更新kv
	UpsertKv(kit *kit.Kit, opt *types.UpsertKvOption) (int, error)
	// GetLastKv 获取最新的kv
	GetLastKv(kit *kit.Kit, opt *types.GetLastKvOpt) (kvType types.KvType, value string, err error)
	// GetKvByVersion 根据版本获取kv
	GetKvByVersion(kit *kit.Kit, opt *types.GetKvByVersion) (kvType types.KvType, value string, err error)
	// DeleteKv deletes specified key-value data from Vault.
	DeleteKv(kit *kit.Kit, opt *types.DeleteKvOpt) error
}

type set struct {
	cli *vault.Client
}

// NewSet ...
func NewSet(opt cc.Vault) (Set, error) {

	config := vault.DefaultConfig()
	config.Address = opt.Address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	client.SetToken(opt.Token)

	s := &set{
		cli: client,
	}

	return s, nil
}
