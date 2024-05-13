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

// Package secret xxx
package secret

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/options"
)

// SecretRequest 密钥结构以 project -> path -> data 为基准
// project 是项目根路径，作为项目入口，来实现不同项目的软隔离隔离，通过token-policy-project来保证项目的认证
// path 是密钥存储路径，方便区分管理密钥，实现不同环境不同路径等。 类似文件系统，每个path下有目录和文件，文件里保存具体data密钥
// data 是具体密钥数据，由多个键值对组成，结构是一个 map[string]interface{}
// nolint
type SecretRequest struct {
	Project string                 `json:"project"`
	Path    string                 `json:"path"`
	Data    map[string]interface{} `json:"data"`
}

// SecretMetadata secret metadata
// nolint
type SecretMetadata struct {
	CreateTime     time.Time                `json:"CreatedTime"`
	UpdatedTime    time.Time                `json:"UpdatedTime"`
	CurrentVersion int                      `json:"CurrentVersion"`
	Version        map[string]SecretVersion `json:"Versions"`
}

// SecretVersion secret version
// nolint
type SecretVersion struct {
	Version      int       `mapstructure:"version"`
	CreatedTime  time.Time `mapstructure:"created_time"`
	DeletionTime time.Time `mapstructure:"deletion_time"`
	Destroyed    bool      `mapstructure:"destroyed"`
}

// SecretManager secret interface
// nolint
type SecretManager interface {
	// Init init client
	Init() error
	Stop()

	// InitProject
	// 1. 初始化项目与权限，vault使用volume,policy和token
	// 2. 需要存储认证到secret中
	// NOTE 全部通过projectName来定义一些默认规则
	InitProject(project string) error
	ReverseInitProject(project string) []error
	// GetSecretAnnotation 通过initProject创建的secret信息需要传入project的annotations，为了不侵入gitops-manager需要从vault-plugin获取
	GetSecretAnnotation(project string) string

	// GetSecret
	// 只获取数据,不返回metadata，不对key进行操作只返回map。 无法对path执行 GetSecret 操作，需要使用 ListSecret
	GetSecret(ctx context.Context, req *SecretRequest) (map[string]interface{}, error)

	// GetMetadata 返回文件 metadata, 主要用于采集创建时间，更改时间，标签等
	GetMetadata(ctx context.Context, req *SecretRequest) (*SecretMetadata, error)

	// ListSecret 返回当前path下的所有文件和目录, 目录用以/结尾, 无法针对文件执行 ListSecret 操作
	ListSecret(ctx context.Context, req *SecretRequest) ([]string, error)

	// CreateSecret 创建文件,data为具体密钥， 幂等
	CreateSecret(ctx context.Context, req *SecretRequest) error

	// UpdateSecret 更新文件， 幂等      新建版本，会不会把metadata修改为本版本
	UpdateSecret(ctx context.Context, req *SecretRequest) error

	// DeleteSecret 完全移除文件或目录的data和metadata
	// nolint
	DeleteSecret(ctx context.Context, req *SecretRequest) error
}

// SecretManagerWithVersion 基于SecretManager实现版本控制管理
// nolint
type SecretManagerWithVersion interface {
	SecretManager

	// GetSecretWithVersion 获取具体某个版本的Secret
	GetSecretWithVersion(ctx context.Context, req *SecretRequest, version int) (map[string]interface{}, error)

	// GetVersionsAsList 列出Secret的所有版本
	GetVersionsAsList(ctx context.Context, req *SecretRequest) ([]*SecretVersion, error)

	// Rollback 回滚到指定版本
	Rollback(ctx context.Context, req *SecretRequest, version int) error

	// DeleteVersion 删除某个版本,这里的删除并没有直接移除这个值,只会标记一个deletion_time值,可通过undelete恢复
	DeleteVersion(ctx context.Context, req *SecretRequest, version []int) error

	// UnDeleteVersion 撤销某些版本的删除
	UnDeleteVersion(ctx context.Context, req *SecretRequest, version []int) error

	// DestroyVersion 永久销毁版本
	DestroyVersion(ctx context.Context, req *SecretRequest, version []int) error
}

// NewSecretManager create storage client
func NewSecretManager(opt *options.Options) SecretManagerWithVersion {
	return &VaultSecretManager{
		option: opt,
	}
}
