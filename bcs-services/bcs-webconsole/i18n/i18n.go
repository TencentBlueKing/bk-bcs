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
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var atI18n GinI18n

// NewI18n ...
func NewI18n(opts ...Option) {
	// init default value
	ins := &ginI18nImpl{
		getLngHandler: defaultGetLngHandler,
	}
	ins.setBundle(defaultBundleConfig)

	// overwrite default value by options
	for _, opt := range opts {
		opt(ins)
	}

	atI18n = ins
}

// Localize ...
func Localize(opts ...Option) gin.HandlerFunc {
	NewI18n(opts...)
	return func(context *gin.Context) {
		atI18n.SetCurrentContext(context)
	}
}

func MustGetMessage(param interface{}) string {
	return atI18n.mustGetMessage(param)
}

func NewLocalizeConfig(messageID string, templateData interface{}) *i18n.LocalizeConfig {
	return &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	}
}
