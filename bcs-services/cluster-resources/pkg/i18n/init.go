/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package i18n

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
)

// 国际化字典
var i18nMsgMap map[string]map[string]string

// InitMsgMap 服务启动时初始化 i18n 配置
func InitMsgMap() error {
	// 读取国际化配置文件
	yamlFile, err := ioutil.ReadFile(envs.LocalizeFilePath)
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
			ZH: msg["msgID"], EN: msg["en"],
		}
	}
	return nil
}
