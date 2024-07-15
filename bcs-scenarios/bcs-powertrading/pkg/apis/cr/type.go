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

package cr

const (
	getPerfDetailUrl = "%s/api/v1/performance/get_perf_detail"
)

// GetPerfDetailReq getPerfDetail req
type GetPerfDetailReq struct {
	Dsl    *GetPerfDetailDsl `json:"dsl"`
	Offset int               `json:"offset"`
	Limit  int               `json:"limit"`
}

// GetPerfDetailDsl getPerfDetail Dsl
type GetPerfDetailDsl struct {
	MatchExpr []GetPerfDetailMatchExpr `json:"matchExpr"`
}

// GetPerfDetailMatchExpr getPerfDetail matchExpr
type GetPerfDetailMatchExpr struct {
	Key      string   `json:"key"`
	Values   []string `json:"values"`
	Operator string   `json:"operator"`
}

// GetPerfDetailRsp getPerfDetail rsp
type GetPerfDetailRsp struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    GetPerfDetailData `json:"data"`
}

// GetPerfDetailData getPerfDetail data
type GetPerfDetailData struct {
	TotalCount int                  `json:"totalCount"`
	Items      []*GetPerfDetailItem `json:"items"`
}

// GetPerfDetailItem getPerfDetail item
type GetPerfDetailItem struct {
	IP               string  `json:"ip"`
	CpuPercent       float64 `json:"cpu_percent"`
	MemTotal         float64 `json:"mem_total"`
	Mem4Linux        float64 `json:"mem4linux"`
	MaxCpuCoreAmount float64 `json:"max_cpu_core_amount"`
}
