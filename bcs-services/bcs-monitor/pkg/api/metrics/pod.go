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

package metrics

import (
	"strings"

	bcsmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// PodUsageQuery Pod 查询
type PodUsageQuery struct {
	UsageQuery  `json:",inline"`
	Namespace   string   `json:"namespace"`
	PodNameList []string `json:"pod_name_list"`
}

// PodCPUUsage :
func PodCPUUsage(c *rest.Context) (interface{}, error) {
	query := &PodUsageQuery{}
	if err := c.ShouldBindJSON(query); err != nil {
		return nil, err
	}

	queryTime, err := query.GetQueryTime()
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"clusterId":   c.ClusterId,
		"namespace":   query.Namespace,
		"podNameList": strings.Join(query.PodNameList, "|"),
	}

	promql := `bcs:pod:cpu_usage{cluster_id="%<clusterId>s", namespace="%<namespace>s", pod_name=~"%<podNameList>s"}`
	vector, _, err := bcsmonitor.QueryRangeF(c.Context, c.ProjectId, promql, params, queryTime.Start, queryTime.End, queryTime.Step)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return vector, nil
}
