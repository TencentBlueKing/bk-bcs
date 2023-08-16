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
	"strings"
)

const (
	ctxLanguage = "I18nLanguage"
	// DefaultLang defaultLanguage  defines the default language if user does not specify in options.
	defaultLanguage = zh
	// zh 中文
	zh = "zh"
	// en 英文
	en = "en"
)

// 语言版本简写映射表
var langMap = map[string]string{
	"zh":      zh,
	"zh-cn":   zh,
	"zh-hans": zh,
	"zh-hant": zh,
	"en":      en,
	"en-us":   en,
	"en-gb":   en,
	// "ru":      RU,
	// "ru-RU":   RU,
	// "ja":      JA,
	// "ja-JP":   JA,
}

// WithLanguage append language setting to the context and returns a new context.
func WithLanguage(ctx context.Context, language string) context.Context {
	if ctx == nil {
		ctx = context.TODO()
	}

	return context.WithValue(ctx, ctxLanguage, toLanguage(language))
}

// LanguageFromCtx retrieves and returns language name from context.
// It returns an empty string if it is not set previously.
func LanguageFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v := ctx.Value(ctxLanguage)
	if v != nil {
		return v.(string)
	}
	return ""
}

func toLanguage(language string) string {
	if lang, ok := langMap[strings.ToLower(language)]; ok {
		return lang
	}
	return defaultLanguage
}
