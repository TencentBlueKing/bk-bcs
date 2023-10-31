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

	"bscp.io/pkg/kit"
)

const (
	// MountPath mount path
	MountPath = "bk_bscp"
	// kvPath kv path
	kvPath = "biz/%d/apps/%d/kv/key/%s"
)

// UpsertKv 创建｜更新kv
func (s *set) UpsertKv(kit kit.Kit, bizID, appID uint32, key, content string) (int, error) {

	data := map[string]interface{}{
		"data": content,
	}
	secret, err := s.cli.KVv2(MountPath).Put(kit.Ctx, fmt.Sprintf(kvPath, bizID, appID, key), data)
	if err != nil {
		return 0, err
	}

	return secret.VersionMetadata.Version, nil

}

// GetLastKv 获取最新的kv
func (s *set) GetLastKv(kit kit.Kit, bizID, appID uint32, key string) (string, error) {

	kv, err := s.cli.KVv2(MountPath).Get(kit.Ctx, fmt.Sprintf(kvPath, bizID, appID, key))
	if err != nil {
		return "", err
	}

	value, ok := kv.Data["data"].(string)
	if !ok {
		return "", fmt.Errorf("value type assertion failed: err : %v", err)
	}

	return value, nil

}

// GetKvByVersion 根据版本获取kv
func (s *set) GetKvByVersion(kit kit.Kit, bizID, appID uint32, key string, version int) (string, error) {

	kv, err := s.cli.KVv2(MountPath).GetVersion(kit.Ctx, fmt.Sprintf(kvPath, bizID, appID, key), version)
	if err != nil {
		return "", err
	}

	value, ok := kv.Data["data"].(string)
	if !ok {
		return "", fmt.Errorf("value type assertion failed: err : %v", err)
	}

	return value, nil

}
