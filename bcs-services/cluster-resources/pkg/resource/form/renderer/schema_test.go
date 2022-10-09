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

package renderer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/validator"
)

func TestSchemaRenderer(t *testing.T) {
	assert.Nil(t, i18n.InitMsgMap())

	// 默认版本（中文）
	for kind := range validator.FormSupportedResAPIVersion {
		_, err := NewSchemaRenderer(context.TODO(), envs.TestClusterID, kind, "default", "").Render()
		assert.Nil(t, err)
		// TODO 如何在单元测试中验证 schema 的合法性？（非标准 schema）
	}

	// 英文版本
	ctx := context.WithValue(context.TODO(), ctxkey.LangKey, i18n.EN)
	for kind := range validator.FormSupportedResAPIVersion {
		_, err := NewSchemaRenderer(ctx, envs.TestClusterID, kind, "default", "").Render()
		assert.Nil(t, err)
	}

	// 中文版本
	ctx = context.WithValue(context.TODO(), ctxkey.LangKey, i18n.ZH)
	for kind := range validator.FormSupportedResAPIVersion {
		_, err := NewSchemaRenderer(ctx, envs.TestClusterID, kind, "default", "").Render()
		assert.Nil(t, err)
	}
}
