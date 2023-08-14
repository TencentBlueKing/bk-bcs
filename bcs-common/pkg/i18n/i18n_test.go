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
 *
 */

package i18n

import (
	"context"
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n/i18n/common"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n/i18n/testdata"
)

func Test_Basic(t *testing.T) {
	i18n := New(Options{
		Path: []embed.FS{testdata.Assets},
	})

	i18n.SetLanguage("none")
	assert.Equal(t, i18n.T(context.Background(), "{{.hello}}{{.world}}"), "{{.hello}}{{.world}}")

	i18n.SetLanguage("zh")
	assert.Equal(t, i18n.T(context.Background(), "hello"), "你好")
	assert.Equal(t, i18n.T(context.Background(), "{{.hello}}{{.world}}"), "你好世界")

	i18n.SetLanguage("en")
	assert.Equal(t, i18n.T(context.Background(), "hello"), "hello")
	assert.Equal(t, i18n.T(context.Background(), "{{.hello}} {{.world}}"), "hello world")

	i18n.SetLanguage("ja")
	assert.Equal(t, i18n.T(context.Background(), "hello"), "こんにちは")
	assert.Equal(t, i18n.T(context.Background(), "{{.hello}}{{.world}}"), "こんにちは世界")
}

func Test_TranslateFormat(t *testing.T) {
	i18n := New(Options{
		Path: []embed.FS{testdata.Assets, common.Assets},
	})

	i18n.SetLanguage("none")
	assert.Equal(t, i18n.Tf(context.Background(), "{{.hello}}{{.world}} %d", 2023), "{{.hello}}{{.world}} 2023")

	i18n.SetLanguage("zh")
	assert.Equal(t, i18n.Tf(context.Background(), "{{.OrderPay}}", 1691552860, 60.3), "您已成功完成订单号 1691552860 支付，支付金额￥60.30。")

	i18n.SetLanguage("en")
	assert.Equal(t, i18n.Tf(context.Background(), "{{.hello}} {{.world}} %d", 2023), "hello world 2023")

	i18n.SetLanguage("ja")
	assert.Equal(t, i18n.Tf(context.Background(), "{{.hello}}{{.world}} %d", 2023), "こんにちは世界 2023")
}

func Test_SetCtxLanguage(t *testing.T) {
	i18n := New(Options{
		Path: []embed.FS{testdata.Assets},
	})

	i18n.SetLanguage("en")
	assert.Equal(t, i18n.Tf(WithLanguage(context.Background(), "zh"), "{{.hello}} {{.world}} %d", 2023), "你好 世界 2023")
}

func Test_DefaultOptions(t *testing.T) {
	i18n := New()

	i18n.SetPath([]embed.FS{testdata.Assets})

	i18n.SetLanguage("en")
	assert.Equal(t, i18n.Tf(WithLanguage(context.Background(), "zh"), "{{.hello}} {{.world}} %d", 2023), "你好 世界 2023")

	i18n.SetDelimiters("{$", "}")
	assert.Equal(t, i18n.Tf(context.Background(), "{$hello} {$world} %d", 2023), "hello world 2023")
}

func Test_Instance(t *testing.T) {
	i18n := Instance("test")
	i18n.SetPath([]embed.FS{testdata.Assets})

	assert.Equal(t, i18n.Tf(context.Background(), "{{.hello}} {{.world}} %d", 2023), "hello world 2023")
}
