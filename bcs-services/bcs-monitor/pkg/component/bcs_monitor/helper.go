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

package bcsmonitor

import (
	"net/http"

	"github.com/prometheus/common/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

func getQueryURL() (string, http.Header) {
	var url string
	header := http.Header{}
	if config.G.Web.QueryAuth {
		header.Add("Authorization", "Bearer admin")
	}

	if config.G.BCS.QueryURL != "" {
		url = config.G.BCS.QueryURL
	} else {
		// 集群内, 走 service 接口
		url = "http://bcs-monitor-query"

		// url = fmt.Sprintf("%s/bcsapi/v4/monitor/query", config.G.BCS.Host)
		// header.Add("Authorization", fmt.Sprintf("Bearer %s", config.G.BCS.Token))
	}
	return url, header
}

// GetFirstValue 获取第一个值
func GetFirstValue(vector model.Vector) string {
	if len(vector) == 0 {
		return ""
	}
	return vector[0].Value.String()
}

// GetLabelSet 获取第一个值的labels
func GetLabelSet(vector model.Vector) map[string]string {
	labelSet := map[string]string{}
	if len(vector) == 0 {
		return labelSet
	}
	for k, v := range vector[0].Metric {
		labelSet[string(k)] = string(v)
	}
	return labelSet
}
