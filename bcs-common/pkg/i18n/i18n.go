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

// Package i18n xxx
package i18n

import (
	"context"
	"embed"
)

// SetPath sets the directory path storing i18n files.
func SetPath(path []embed.FS) {
	Instance().SetPath(path)
}

// SetLanguage sets the language for translator.
func SetLanguage(language string) {
	Instance().SetLanguage(language)
}

// SetDelimiters sets the delimiters for translator.
func SetDelimiters(left, right string) {
	Instance().SetDelimiters(left, right)
}

// T is alias of Translate for convenience.
func T(ctx context.Context, content string) string {
	return Instance().T(ctx, content)
}

// Tf is alias of TranslateFormat for convenience.
func Tf(ctx context.Context, format string, values ...interface{}) string {
	return Instance().TranslateFormat(ctx, format, values...)
}

// TranslateFormat translates, formats and returns the `format` with configured language
// and given `values`.
func TranslateFormat(ctx context.Context, format string, values ...interface{}) string {
	return Instance().TranslateFormat(ctx, format, values...)
}

// Translate translates `content` with configured language and returns the translated content.
func Translate(ctx context.Context, content string) string {
	return Instance().Translate(ctx, content)
}

// GetContent retrieves and returns the configured content for given key and specified language.
// It returns an empty string if not found.
func GetContent(ctx context.Context, key string) string {
	return Instance().GetContent(ctx, key)
}
