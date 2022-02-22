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
	"embed"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed localize
var fs embed.FS

type ginI18nImpl struct {
	bundle          *i18n.Bundle
	currentContext  *gin.Context
	localizeByLng   map[string]*i18n.Localizer
	defaultLanguage language.Tag
	getLngHandler   GetLngHandler
	files           embed.FS
}

// getMessage get localize message by lng and messageID
func (i *ginI18nImpl) getMessage(localizeConfig *i18n.LocalizeConfig) (string, error) {
	lng := i.getLngHandler(i.currentContext, i.defaultLanguage.String())
	localizer := i.getLocalizeByLng(lng)

	message, err := localizer.Localize(localizeConfig)
	if err != nil {
		return "", err
	}

	return message, nil
}

// mustGetMessage ...
func (i *ginI18nImpl) mustGetMessage(localizeConfig *i18n.LocalizeConfig) string {
	message, err := i.getMessage(localizeConfig)
	// 如果没有翻译, 原样返回
	if err != nil {
		return localizeConfig.MessageID
	}
	return message
}

func (i *ginI18nImpl) setFiles() {
	i.files = fs
}

func (i *ginI18nImpl) SetCurrentContext(ctx context.Context) {
	i.currentContext = ctx.(*gin.Context)
}

func (i *ginI18nImpl) setBundle(cfg *BundleCfg) {
	bundle := i18n.NewBundle(cfg.DefaultLanguage)
	bundle.RegisterUnmarshalFunc(cfg.FormatBundleFile, cfg.UnmarshalFunc)

	i.bundle = bundle
	i.defaultLanguage = cfg.DefaultLanguage

	i.setFiles()
	i.loadMessageFiles(cfg)
	i.setLocalizeByLng(cfg.AcceptLanguage)
}

func (i *ginI18nImpl) setGetLngHandler(handler GetLngHandler) {
	i.getLngHandler = handler
}

// loadMessageFiles load all file localize to bundle
func (i *ginI18nImpl) loadMessageFiles(config *BundleCfg) {
	for _, lng := range config.AcceptLanguage {
		path := fmt.Sprintf("%s/%s.%s", config.RootPath, lng.String(), config.FormatBundleFile)

		buf, err := i.files.ReadFile(path)
		if err != nil {
			continue
		}
		_, err = i.bundle.ParseMessageFileBytes(buf, path)
		if err != nil {
			blog.Infof("Failed to load all file localize to bundle , err : %v", err)
		}
	}
}

// setLocalizeByLng set localize by language
func (i *ginI18nImpl) setLocalizeByLng(acceptLanguage []language.Tag) {
	i.localizeByLng = map[string]*i18n.Localizer{}
	for _, lng := range acceptLanguage {
		lngStr := lng.String()
		i.localizeByLng[lngStr] = i.newLocalize(lngStr)
	}

	// set defaultLanguage if it isn't exist
	defaultLng := i.defaultLanguage.String()
	if _, hasDefaultLng := i.localizeByLng[defaultLng]; !hasDefaultLng {
		i.localizeByLng[defaultLng] = i.newLocalize(defaultLng)
	}
}

// newLocalize create a Localize by language
func (i *ginI18nImpl) newLocalize(lng string) *i18n.Localizer {
	lngDefault := i.defaultLanguage.String()
	lngs := []string{
		lng,
	}

	if lng != lngDefault {
		lngs = append(lngs, lngDefault)
	}

	localizer := i18n.NewLocalizer(
		i.bundle,
		lngs...,
	)
	return localizer
}

// getLocalizeByLng get Localize by language
func (i *ginI18nImpl) getLocalizeByLng(lng string) *i18n.Localizer {
	localizer, hasValue := i.localizeByLng[lng]
	if hasValue {
		return localizer
	}

	return i.localizeByLng[i.defaultLanguage.String()]
}
