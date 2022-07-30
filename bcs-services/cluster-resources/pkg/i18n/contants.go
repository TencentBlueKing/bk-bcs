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

const (
	// ZH 中文
	ZH = "zh"
	// EN 英文
	EN = "en"
	// DefaultLang 默认语言
	DefaultLang = ZH
)

// 语言版本简写映射表
var langMap = map[string]string{
	"zh":      ZH,
	"zh-cn":   ZH,
	"zh-hans": ZH,
	"zh-hant": ZH,
	"en":      EN,
	"en-us":   EN,
	"en-gb":   EN,
}

// MetadataCookiesKey 在 GoMicro Metadata 中，Cookie 的键名
const MetadataCookiesKey = "Grpcgateway-Cookie"
