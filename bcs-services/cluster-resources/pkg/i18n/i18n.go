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
	"net/http"
	"strings"

	"go-micro.dev/v4/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// GetLangFromCookies 从 Cookies 中获取语言版本
func GetLangFromCookies(md metadata.Metadata) string {
	cookies, ok := md.Get(MetadataCookiesKey)
	if !ok {
		return DefaultLang
	}
	for _, c := range stringx.Split(cookies) {
		k, v := stringx.Partition(c, "=")
		if k != conf.LangCookieName {
			continue
		}
		if lang, ok := langMap[strings.ToLower(v)]; ok {
			return lang
		}
	}
	return DefaultLang
}

// GetLangFromReqCookies 从 Cookies 中获取语言版本
func GetLangFromReqCookies(req *http.Request) string {
	cookie, err := req.Cookie(LangCookieName)
	if err != nil {
		return DefaultLang
	}
	if lang, ok := langMap[strings.ToLower(cookie.Value)]; ok {
		return lang
	}
	return DefaultLang
}

// GetLangFromContext 从 Context 中获取语言版本
func GetLangFromContext(ctx context.Context) string {
	if lang := ctx.Value(ctxkey.LangKey); lang != nil {
		return lang.(string)
	}
	return DefaultLang
}

// GetMsg 获取国际化文本
func GetMsg(ctx context.Context, msgID string) string {
	return GetMsgWithLang(msgID, GetLangFromContext(ctx)) // nolint:contextcheck
}

// GetMsgWithLang 获取国际化文本
func GetMsgWithLang(msgID, lang string) string {
	if m, exists := i18nMsgMap[msgID]; exists {
		if msg, ok := m[lang]; ok {
			return msg
		}
	} else if runtime.RunMode != runmode.UnitTest {
		// NOTE 单元测试可能未初始化 MsgMap，忽略告警日志
		log.Warn(context.TODO(), "msgID `%s` not exists", msgID)
	}
	return msgID
}
