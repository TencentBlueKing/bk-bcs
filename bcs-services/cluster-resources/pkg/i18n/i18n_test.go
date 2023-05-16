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

package i18n

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-micro.dev/v4/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
)

func TestGetLangFromCookies(t *testing.T) {
	md := metadata.Metadata{}
	assert.Equal(t, DefaultLang, GetLangFromCookies(md))

	md = metadata.Metadata{MetadataCookiesKey: "blueking_language=EN-US"}
	assert.Equal(t, EN, GetLangFromCookies(md))

	md = metadata.Metadata{MetadataCookiesKey: "blueking_language=zh"}
	assert.Equal(t, ZH, GetLangFromCookies(md))

	md = metadata.Metadata{MetadataCookiesKey: "blueking_language=ru"}
	assert.Equal(t, DefaultLang, GetLangFromCookies(md))
}

func TestGetMsg(t *testing.T) {
	// 初始化 i18n 字典
	assert.Nil(t, InitMsgMap())

	// 默认中文
	ctx := context.TODO()
	assert.Equal(t, "无指定操作权限", GetMsg(ctx, "无指定操作权限"))

	// 指定为英文
	ctx = context.WithValue(ctx, ctxkey.LangKey, EN)
	assert.Equal(t, "no operate permission!", GetMsg(ctx, "无指定操作权限"))

	// 指定为中文
	ctx = context.WithValue(ctx, ctxkey.LangKey, ZH)
	assert.Equal(t, "无指定操作权限", GetMsg(ctx, "无指定操作权限"))
}

func TestGetMsgWithLang(t *testing.T) {
	// 初始化 i18n 字典
	assert.Nil(t, InitMsgMap())

	// 指定为英文
	assert.Equal(t, "no operate permission!", GetMsgWithLang("无指定操作权限", EN))

	// 指定为中文
	assert.Equal(t, "无指定操作权限", GetMsgWithLang("无指定操作权限", ZH))
}
