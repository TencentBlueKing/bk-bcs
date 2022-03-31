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
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

// makeAceptLanguage : 合法的语言列表
func makeAcceptLanguage() (acceptLanguage []language.Tag) {
	langMap := map[string]language.Tag{}
	for _, v := range availableLanguage {
		langMap[v.String()] = v
	}
	for _, v := range langMap {
		acceptLanguage = append(acceptLanguage, v)
	}
	return
}

// getLangHandler ...
func getLangHandler(c *gin.Context, defaultLng string) string {
	if c == nil || c.Request == nil {
		return defaultLng
	}

	// lang参数 -> cookie -> accept-language -> 配置文件中的language
	lng := c.Query("lang")
	if lng != "" {
		return lng
	}

	lng, err := c.Cookie("blueking_language")
	if err == nil && lng != "" {
		return lng
	}

	lng, err = getMatchLangByHeader(c.GetHeader("accept-language"))
	if err == nil && lng != "" {
		return lng
	}

	return defaultLng
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
	language, ok := availableLanguage[tag]
	if !ok {
		return "", errors.Errorf("not found %s", lng)
	}

	return language.String(), nil
}

// GetMessage accepts values in following formats:
//   - GetMessage("MessageID")
//   - GetMessage("MessageID", error)
//   - GetMessage("MessageID", "value")
//   - GetMessage("MessageID",map[string]string{}{"key1": "value1", "key2": "value2"})
func GetMessage(messageID string, values ...interface{}) string {
	// 如果messageID 没有国际化, 默认原样返回
	if _, err := ginI18n.GetMessage(messageID); err != nil {
		return messageID
	}

	if values == nil {
		return ginI18n.MustGetMessage(messageID)
	}

	switch param := values[0].(type) {
	case error:
		// - Must("MessageID", error)
		return ginI18n.MustGetMessage(&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: map[string]string{"err": param.Error()},
		})
	case string:
		// - Must("MessageID", "value")
		return ginI18n.MustGetMessage(&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: param,
		})
	case map[string]string:
		// - Must("MessageID",map[string]string{}{"key1": "value1", "key2": "value2"})
		return ginI18n.MustGetMessage(&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: param,
		})
	default:
		return ginI18n.MustGetMessage(&i18n.LocalizeConfig{
			MessageID: messageID,
		})
	}

}

// Localize 国际化
func Localize() gin.HandlerFunc {
	bundle := ginI18n.WithBundle(defaultBundleConfig)
	handle := ginI18n.WithGetLngHandle(getLangHandler)

	return ginI18n.Localize(bundle, handle)
}
