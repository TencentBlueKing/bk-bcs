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

// Package i18n xxx
package i18n

import (
	"golang.org/x/text/language"
	"net/http"
)

var availableLanguage = map[string]language.Tag{
	"en":         language.English,
	"en-us":      language.English,
	"en_US":      language.English,
	"zh":         language.SimplifiedChinese,
	"zh-hans-cn": language.SimplifiedChinese,
	"zh-hans":    language.SimplifiedChinese,
	"zh-cn":      language.SimplifiedChinese,
	"zh_CN":      language.SimplifiedChinese,
}

func IsAvailableLanguage(s string) bool {
	if _, ok := availableLanguage[s]; ok {
		return true
	}
	return false
}

// GetAvailableLanguage get available language
func GetAvailableLanguage(s string, defaultLang string) language.Tag {
	if _, ok := availableLanguage[s]; ok {
		return availableLanguage[s]
	}
	// default config
	return availableLanguage[defaultLang]
}

// GetLangByRequest get the language by request
func GetLangByRequest(r *http.Request, defaultLang string) string {

	// lang参数 -> cookie -> accept-language -> 配置文件中的language
	lng := r.FormValue("lang")
	if lng != "" {
		return lng
	}

	cookie, err := r.Cookie("blueking_language")
	if err == nil && cookie != nil {
		return cookie.Value
	}

	lng = r.Header.Get("accept-language")
	if lng != "" {
		return lng
	}

	// default config
	return defaultLang
}
