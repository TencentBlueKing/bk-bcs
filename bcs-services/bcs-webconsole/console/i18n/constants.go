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
	"embed"

	ginI18n "github.com/gin-contrib/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

//go:embed localize
var fs embed.FS

const (
	defaultFormatBundleFile = "yaml"
	defaultRootPath         = "localize"
)

var (
	defaultLanguage      = language.SimplifiedChinese
	defaultUnmarshalFunc = yaml.Unmarshal

	availableLanguage = map[string]language.Tag{
		"en":         language.English,
		"zh":         language.SimplifiedChinese,
		"zh-hans-cn": language.SimplifiedChinese,
		"zh-hans":    language.SimplifiedChinese,
		"zh-cn":      language.SimplifiedChinese,
	}

	defaultAcceptLanguage = makeAcceptLanguage()

	defaultBundleConfig = &ginI18n.BundleCfg{
		RootPath:         defaultRootPath,
		AcceptLanguage:   defaultAcceptLanguage,
		DefaultLanguage:  defaultLanguage,
		UnmarshalFunc:    defaultUnmarshalFunc,
		FormatBundleFile: defaultFormatBundleFile,
		Loader:           &ginI18n.EmbedLoader{FS: fs},
	}
)
