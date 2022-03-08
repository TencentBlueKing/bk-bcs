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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var atI18n GinI18n

// NewI18n ...
func NewI18n(opts ...Option) {
	// init default value
	ins := &ginI18nImpl{
		getLngHandler: defaultGetLngHandler,
	}

	// 设置默认语言
	lang := language.Make(config.G.Base.LanguageCode)
	if lang.String() != "und" {
		defaultBundleConfig.DefaultLanguage = lang
	} else {
		logger.Warnf("failed to set default language, unknown language code : %s", config.G.Base.LanguageCode)
	}

	defaultAcceptLanguage = append(defaultAcceptLanguage, language.Make(config.G.Base.LanguageCode))

	ins.setBundle(defaultBundleConfig)

	// overwrite default value by options
	for _, opt := range opts {
		opt(ins)
	}

	atI18n = ins
}

func NewLocalizeConfig(messageID string, templateData interface{}) *i18n.LocalizeConfig {
	return &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	}
}

// GetMessage accepts values in following formats:
//   - GetMessage("MessageID")
//   - GetMessage("MessageID", error)
//   - GetMessage("MessageID", "value")
//   - GetMessage("MessageID",map[string]string{}{"key1": "value1", "key2": "value2"})
func GetMessage(messageID string, values ...interface{}) string {

	if values == nil {
		return atI18n.mustGetMessage(&i18n.LocalizeConfig{
			MessageID: messageID,
		})
	}

	switch param := values[0].(type) {
	case error:
		// - Must("MessageID", error)
		return atI18n.mustGetMessage(&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: map[string]string{"err": param.Error()},
		})
	case string:
		// - Must("MessageID", "value")
		return atI18n.mustGetMessage(&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: param,
		})
	case map[string]string:
		// - Must("MessageID",map[string]string{}{"key1": "value1", "key2": "value2"})
		return atI18n.mustGetMessage(&i18n.LocalizeConfig{
			MessageID:    messageID,
			TemplateData: param,
		})
	default:
		return atI18n.mustGetMessage(&i18n.LocalizeConfig{
			MessageID: messageID,
		})
	}

}
