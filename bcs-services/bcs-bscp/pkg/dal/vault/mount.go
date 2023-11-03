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

package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

// CreateMountPath 创建挂载目录
func (s *set) CreateMountPath(path string, config *api.MountInput) error {
	return s.cli.Sys().Mount(path, config)
}

// IsMountPathExists 挂载目录是否存在
func (s *set) IsMountPathExists(path string) (bool, error) {
	// 列出所有的挂载路径
	mounts, err := s.cli.Sys().ListMounts()
	if err != nil {
		return false, err
	}

	// 检查要创建的挂载路径是否已存在
	_, exists := mounts[fmt.Sprintf("%s/", path)]
	return exists, nil
}
