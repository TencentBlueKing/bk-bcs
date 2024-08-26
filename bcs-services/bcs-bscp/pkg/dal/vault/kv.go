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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

const (
	// MountPath mount path
	MountPath = "bk_bscp"
	// kvPath kv path
	kvPath = "biz/%d/apps/%d/kvs/%s"
	// releasedKvPath kv revision path
	releasedKvPath = "biz/%d/apps/%d/releases/%d/kvs/%s"
)

// UpsertKv 创建｜更新kv
func (s *set) UpsertKv(kit *kit.Kit, opt *types.UpsertKvOption) (int, error) {

	if err := opt.Validate(); err != nil {
		return 0, err
	}

	data := map[string]interface{}{
		"kv_type": opt.KvType,
		"value":   opt.Value,
	}
	secret, err := s.cli.KVv2(MountPath).Put(kit.Ctx, fmt.Sprintf(kvPath, opt.BizID, opt.AppID, opt.Key), data)
	if err != nil {
		return 0, err
	}

	return secret.VersionMetadata.Version, nil

}

// GetLastKv 获取最新的kv
func (s *set) GetLastKv(kit *kit.Kit, opt *types.GetLastKvOpt) (kvType table.DataType, value string, err error) {

	if err = opt.Validate(); err != nil {
		return
	}

	kv, err := s.cli.KVv2(MountPath).Get(kit.Ctx, fmt.Sprintf(kvPath, opt.BizID, opt.AppID, opt.Key))
	if err != nil {
		return
	}

	kvTypeStr, ok := kv.Data["kv_type"].(string)
	if !ok {
		// nolint goconst
		return "", "", fmt.Errorf("failed to get 'kv_type' as a string from kv.Data,"+
			" err : %v", err)
	}
	kvType = table.DataType(kvTypeStr)

	value, ok = kv.Data["value"].(string)
	if !ok {
		return "", "", fmt.Errorf("value type assertion failed, err : %v", err)
	}

	return kvType, value, nil
}

// GetKvByVersion 根据版本获取kv
func (s *set) GetKvByVersion(kit *kit.Kit, opt *types.GetKvByVersion) (kvType table.DataType, value string, err error) {

	if err = opt.Validate(); err != nil {
		return
	}

	kv, err := s.cli.KVv2(MountPath).GetVersion(kit.Ctx, fmt.Sprintf(kvPath, opt.BizID, opt.AppID, opt.Key), opt.Version)
	if err != nil {
		return
	}

	kvTypeStr, ok := kv.Data["kv_type"].(string)
	if !ok {
		return "", "", errf.Errorf(errf.InvalidRequest, i18n.T(kit, `get 'kv_type' as a string 
		from kv.Data failed, err: %v`, err))
	}
	kvType = table.DataType(kvTypeStr)

	value, ok = kv.Data["value"].(string)
	if !ok {
		return "", "", errf.Errorf(errf.InvalidRequest, i18n.T(kit, "value type assertion failed, err: %v", err))
	}

	return kvType, value, nil

}

// DeleteKv deletes specified key-value data from Vault.
func (s *set) DeleteKv(kit *kit.Kit, opt *types.DeleteKvOpt) error {

	if err := opt.Validate(); err != nil {
		return err
	}
	return s.cli.KVv2(MountPath).DeleteMetadata(kit.Ctx, fmt.Sprintf(kvPath, opt.BizID, opt.AppID, opt.Key))
}
