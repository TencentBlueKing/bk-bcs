/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package validator xxx
package validator

import (
	"context"

	"github.com/TencentBlueKing/gopkg/collection/set"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// FormDataValidator 表单数据校验器
// 目前只做基础的数据校验，表单校验主要还是前端按 Schema 规则完成，后续可支持后台校验 Schema 规则
type FormDataValidator struct {
	ctx        context.Context
	formData   map[string]interface{}
	apiVersion string
	kind       string
}

// New xxx
func New(ctx context.Context, formData map[string]interface{}, apiVersion, kind string) *FormDataValidator {
	return &FormDataValidator{ctx: ctx, formData: formData, apiVersion: apiVersion, kind: kind}
}

// Validate xxx
func (v *FormDataValidator) Validate() error {
	for _, f := range []func() error{
		// 1. 检查资源版本
		v.validateAPIVersion,
		// 2. 检查 Metadata 表单数据
		v.validateMetadata,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (v *FormDataValidator) validateAPIVersion() error {
	supportedAPIVersions, ok := FormSupportedResAPIVersion[v.kind]
	if !ok {
		return errorx.New(errcode.Unsupported, i18n.GetMsg(v.ctx, "资源类型 `%s` 不支持表单化"), v.kind)
	}
	if !slice.StringInSlice(v.apiVersion, supportedAPIVersions) {
		return errorx.New(
			errcode.Unsupported,
			i18n.GetMsg(v.ctx, "资源类型 %s APIVersion %s 不在受支持的版本列表 %v 中，请改用 Yaml 模式而非表单化"),
			v.kind, v.apiVersion, supportedAPIVersions,
		)
	}
	return nil
}

func (v *FormDataValidator) validateMetadata() error {
	// 检查是否存在重复的 Label Key
	if err := checkDuplicateKey(
		mapx.GetList(v.formData, "metadata.labels"),
		i18n.GetMsg(v.ctx, "标签有重复的键，请检查"),
	); err != nil {
		return err
	}
	// 检查是否存在重复的 Annotations Key
	if err := checkDuplicateKey(
		mapx.GetList(v.formData, "metadata.annotations"),
		i18n.GetMsg(v.ctx, "注解有重复的键，请检查"),
	); err != nil {
		return err
	}
	return nil
}

// checkDuplicateKey 检查 k-v 列表中是否存在重复的 key
func checkDuplicateKey(kvList []interface{}, errMsg string) error {
	keys := set.NewStringSet()
	for _, kv := range kvList {
		k := kv.(map[string]interface{})["key"].(string)
		if keys.Has(k) {
			return errorx.New(errcode.ValidateErr, errMsg)
		}
		keys.Add(k)
	}
	return nil
}
