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
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

var (
	availableLanguage = map[string]language.Tag{
		"en":         language.English,
		"en-us":      language.English,
		"en_US":      language.English,
		"zh":         language.SimplifiedChinese,
		"zh-hans-cn": language.SimplifiedChinese,
		"zh-hans":    language.SimplifiedChinese,
		"zh-Hans":    language.SimplifiedChinese,
		"zh-cn":      language.SimplifiedChinese,
		"zh_CN":      language.SimplifiedChinese,
	}
	defaultAcceptLanguage = makeAcceptLanguage()
)

// makeAcceptLanguage : 合法的语言列表
func makeAcceptLanguage() (acceptLanguage []language.Tag) {
	langMap := map[string]language.Tag{}
	for _, v := range availableLanguage {
		langMap[v.String()] = v
	}
	for _, v := range langMap {
		acceptLanguage = append(acceptLanguage, v)
	}
	return acceptLanguage
}

// getMatchLangByHeader 解析 header, 查找最佳匹配
func getMatchLangByHeader(lng string) (string, error) {
	if lng == "" {
		return "", errors.Errorf("not found accept-language header value")
	}

	// 用户接受的语言
	userAccept, _, err := language.ParseAcceptLanguage(lng)
	if err != nil {
		return "", err
	}

	// 系统中允许的语言
	matcher := language.NewMatcher(defaultAcceptLanguage)
	// 根据顺序优先级进行匹配
	matchedTag, _, _ := matcher.Match(userAccept...)

	// x/text/language: change of behavior for language matcher
	// https://github.com/golang/go/issues/24211
	var tag string
	if len(matchedTag.String()) < 2 {
		return "", errors.Errorf("not found %s", lng)
	}

	tag = matchedTag.String()[0:2]
	lang, ok := availableLanguage[tag]
	if !ok {
		return "", errors.Errorf("not found %s", lng)
	}

	return lang.String(), nil
}

// IsAvailableLanguage determine if the language is legal
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

	lng, err = getMatchLangByHeader(r.Header.Get("accept-language"))
	if err == nil && lng != "" {
		return lng
	}

	// default config
	return defaultLang
}
