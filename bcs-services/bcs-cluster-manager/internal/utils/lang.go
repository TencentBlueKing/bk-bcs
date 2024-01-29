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

package utils

const (
	// DefaultLang 默认语言
	DefaultLang = ZH
	// ZH 中文
	ZH = "zh"
	// EN 英文
	EN = "en"

	// RU 俄语
	// RU = "ru"
	// JA 日语
	// JA = "ja"
)

// 语言版本简写映射表
// nolint
var langMap = map[string]string{
	"zh":      ZH,
	"zh-cn":   ZH,
	"zh-hans": ZH,
	"zh-hant": ZH,
	"en":      EN,
	"en-us":   EN,
	"en-gb":   EN,
	// "ru":      RU,
	// "ru-RU":   RU,
	// "ja":      JA,
	// "ja-JP":   JA,
}
