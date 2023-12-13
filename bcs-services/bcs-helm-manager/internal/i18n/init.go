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

package i18n

import (
	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
)

// NOTE ClusterResources 国际化主要有三部分：
// 1. 错误信息、2. 表单模板（Schema）、3. 示例文件（example conf, references）
// 示例文件采用中、英两份配置来完成国际化，而错误信息和表单模板则使用自制的一套逻辑（MsgMap）
// 没有使用常用的 golang i18n 框架原因是为兼容模板国际化，如果有支持模板的 i18n 框架也可替换

// 国际化字典
var i18nMsgMap map[string]map[string]string

// InitMsgMap 服务启动时初始化 i18n 配置
func InitMsgMap() error {
	// 读取国际化配置文件
	yamlFile, err := Assets.ReadFile(common.LocalizeFileName)
	if err != nil {
		return err
	}
	rawMsgList := []map[string]string{}
	if err = yaml.Unmarshal(yamlFile, &rawMsgList); err != nil {
		return err
	}
	// 转换格式，填充中文默认值
	i18nMsgMap = map[string]map[string]string{}
	for _, msg := range rawMsgList {
		i18nMsgMap[msg["msgID"]] = map[string]string{
			ZH: msg["msgID"], EN: msg["en"], // RU: msg["ru"], JA: msg["ja"],
		}
	}
	return nil
}
