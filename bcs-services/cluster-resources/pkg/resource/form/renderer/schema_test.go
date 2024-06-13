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

package renderer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/validator"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/schema"
)

func TestSchemaRenderer(t *testing.T) {
	assert.Nil(t, i18n.InitMsgMap())

	// 默认版本（中文）
	for kind := range validator.FormSupportedResAPIVersion {
		result, err := NewSchemaRenderer(context.TODO(), envs.TestClusterID, "", kind, "default", "", false).Render()
		assert.Nil(t, err)

		// 验证 schema 的合法性
		loader := schema.NewGoLoader(result["schema"])

		jsonSchema, err := schema.NewSchema(loader)
		assert.Nil(t, err)
		assert.NotNil(t, jsonSchema)

		// 确保没有重要的修改建议
		suggestions, err := jsonSchema.Review()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(suggestions.Major()), "kind %s's schema have major suggestions", kind)

		// 确保原配置中，没有多余的配置
		diffRet := jsonSchema.Diff()
		assert.Equal(t, 0, len(diffRet))
	}

	// 英文版本
	ctx := context.WithValue(context.TODO(), ctxkey.LangKey, i18n.EN)
	for kind := range validator.FormSupportedResAPIVersion {
		_, err := NewSchemaRenderer(ctx, envs.TestClusterID, "", kind, "default", "", false).Render()
		assert.Nil(t, err)
	}

	// 中文版本
	ctx = context.WithValue(context.TODO(), ctxkey.LangKey, i18n.ZH)
	for kind := range validator.FormSupportedResAPIVersion {
		_, err := NewSchemaRenderer(ctx, envs.TestClusterID, "", kind, "default", "", false).Render()
		assert.Nil(t, err)
	}
}
