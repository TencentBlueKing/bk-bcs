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

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// CreateRKv create released kv
func (s *set) CreateRKv(kit *kit.Kit, opt *types.CreateReleasedKvOption) (int, error) {
	if err := opt.Validate(); err != nil {
		return 0, err
	}

	data := map[string]interface{}{
		"kv_type": opt.KvType,
		"value":   opt.Value,
	}
	version, err := s.cli.KVv2(MountPath).Put(kit.Ctx,
		fmt.Sprintf(releasedKvPath, opt.BizID, opt.AppID, opt.ReleaseID, opt.Key), data)
	if err != nil {
		return 0, err
	}

	return version.VersionMetadata.Version, nil
}

// GetRKv Get Released kv by version
func (s *set) GetRKv(kit *kit.Kit, opt *types.GetRKvOption) (kvType table.DataType, value string, err error) {

	if err = opt.Validate(); err != nil {
		return
	}

	kv, err := s.cli.KVv2(MountPath).GetVersion(kit.Ctx, fmt.Sprintf(releasedKvPath, opt.BizID, opt.AppID,
		opt.ReleasedID, opt.Key), opt.Version)
	if err != nil {
		return
	}

	kvTypeStr, ok := kv.Data["kv_type"].(string)
	if !ok {
		return "", "", fmt.Errorf("failed to get 'kv_type' as a string from kv.Data,"+
			" err : %v", err)
	}
	kvType = table.DataType(kvTypeStr)

	value, ok = kv.Data["value"].(string)
	if !ok {
		return "", "", fmt.Errorf("value type assertion failed: err : %v", err)
	}

	return
}
