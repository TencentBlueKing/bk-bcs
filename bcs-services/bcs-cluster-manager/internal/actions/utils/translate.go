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

package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Translate 处理任务名称  // 特殊内容进行替换处理
func Translate(ctx context.Context, taskMethod, taskName string, translate string) (content string) {
	language := i18n.LanguageFromCtx(ctx)

	if translate != "" {
		taskMethod = translate
	}
	arr := strings.Split(taskMethod, "-")
	if len(arr) > 1 {
		content = i18n.T(ctx, arr[len(arr)-1])
	} else {
		content = i18n.T(ctx, taskMethod)
	}

	blog.Infof("Translate %s %s %s %s", language, taskMethod, taskName, content)
	if len(content) == 0 || content == taskMethod {
		return taskName
	}
	return content
}

// TranslateMsg 处理任务返回的msg
func TranslateMsg(ctx context.Context, resourceType, taskType, message string, t *cmproto.Task) string {
	// 获取语言
	lang := i18n.LanguageFromCtx(ctx)
	if lang == "zh" {
		return message
	}
	arr := strings.Split(taskType, "-")
	if len(arr) > 1 {
		taskType = arr[1]
	}
	if resourceType == "nodegroup" {
		switch taskType {
		case "SwitchNodeGroupAutoScaling":
			msg, ok := getTranslateFormat(ctx, "{{.SwitchNodeGroupAutoScalingOpenMsg}}",
				message, t.GetNodeGroupID())
			if ok {
				return msg
			}
			msg, ok = getTranslateFormat(ctx, "{{.SwitchNodeGroupAutoScalingCloseMsg}}",
				message, t.GetNodeGroupID())
			if ok {
				return msg
			}
		case "UpdateNodeGroupDesiredNode":
			msg, ok := getTranslateFormat(ctx, "{{.UpdateNodeGroupDesiredNodeMsg}}",
				message,
				t.GetClusterID(),
				t.GetNodeGroupID(),
				extractLastNumber(message))
			if ok {
				return msg
			}
		default:
			key := fmt.Sprintf("{{.%sMsg}}", taskType)
			msg, ok := getTranslateFormat(ctx, key,
				message,
				t.GetClusterID(),
				t.GetNodeGroupID())
			if ok {
				return msg
			}
		}
	}
	return message
}

func getTranslateFormat(ctx context.Context, key, message string, values ...interface{}) (string, bool) {
	msg := i18n.Tf(i18n.WithLanguage(context.Background(), "zh"), key, values...)
	if msg == message {
		return i18n.Tf(ctx, key, values...), true
	}
	return message, false
}

// 匹配以数字结尾的部分
func extractLastNumber(input string) string {
	re := regexp.MustCompile(`\d+$`)
	match := re.FindString(input)
	return match
}
