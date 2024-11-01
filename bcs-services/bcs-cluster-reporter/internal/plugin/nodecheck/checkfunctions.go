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

// Package nodecheck xxx
package nodecheck

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/nodeagent/processcheck"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/types/process"
	"path/filepath"
	"strings"
)

func checkProcess(detail processcheck.Detail, nodeName string) []plugin_manager.CheckItem {
	result := make([]plugin_manager.CheckItem, 0, 0)

	// 做一些通用的检查
	// TODO 后续考虑支持在nodeagent中配置自定义参数检查项
	// 有问题的检查项添加到result
	for _, processInfo := range detail.ProcessInfo {
		switch filepath.Base(processInfo.BinaryPath) {
		case "dockerd":
			result = append(result, checkDocker(processInfo)...)
		case "kubelet":
			result = append(result, checkKubelet(processInfo)...)
		}
	}

	for key, val := range result {
		val.ItemTarget = nodeName
		result[key] = val
	}

	return result
}

func checkDocker(processInfo process.ProcessInfo) []plugin_manager.CheckItem {
	checkItem := plugin_manager.CheckItem{
		ItemName: processConfigCheckItem,
		Status:   normalStatus,
		Level:    plugin_manager.WARNLevel,
		Detail:   fmt.Sprintf(StringMap[ConfigFileDetail], "docker daemon.json"),
		Normal:   true,
	}

	result := make([]plugin_manager.CheckItem, 0, 0)
	for fileName, configfile := range processInfo.ConfigFiles {
		if filepath.Base(fileName) == "daemon.json" {
			if !strings.Contains(configfile, "data-root") && !strings.Contains(configfile, "\"graph\"") {
				checkItem.Status = configErrorStatus
				checkItem.Normal = false
				checkItem.Detail = fmt.Sprintf(StringMap[flagNotSetDetail], "data-root,graph")
			}
			result = append(result, checkItem)
		}
	}

	checkItem = plugin_manager.CheckItem{
		ItemName: processConfigCheckItem,
		Status:   normalStatus,
		Level:    plugin_manager.WARNLevel,
		Detail:   fmt.Sprintf(StringMap[ConfigFileDetail], "docker service"),
		Normal:   true,
	}
	for fileName, serviceFile := range processInfo.ServiceFiles {
		if strings.HasSuffix(filepath.Base(fileName), ".service") {
			if strings.Contains(serviceFile, "BindsTo") {
				checkItem.Status = configErrorStatus
				checkItem.Detail = fmt.Sprintf(StringMap[flagNotSetDetail], "BindsTo")
			}
			result = append(result, checkItem)
		}
	}

	return result
}

func checkKubelet(processInfo process.ProcessInfo) []plugin_manager.CheckItem {
	checkItem := plugin_manager.CheckItem{
		ItemName: processConfigCheckItem,
		Level:    plugin_manager.WARNLevel,
		Normal:   true,
	}

	result := make([]plugin_manager.CheckItem, 0, 0)

	flags := []string{"--root-dir", "--read-only-port=0"}
	for _, param := range processInfo.Params {
		for index, flag := range flags {
			if strings.Contains(param, flag) && flag != "" {
				flags[index] = ""
			}
		}
	}

	for _, flag := range flags {
		if flag != "" {
			checkItem.Status = ConfigNotFoundStatus
			checkItem.Detail = "kubelet " + fmt.Sprintf(StringMap[flagNotSetDetail], flag)
			result = append(result, checkItem)
		}
	}

	return result
}
