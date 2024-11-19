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

package cmdb

var (
	fieldBizL2Info = []string{
		"businessDepartmentId",
		"businessDepartmentName",
		"businessLevel1Id",
		"businessLevel1Name",
		"businessLevel1Operator",
		"businessLevel2Id",
		"businessLevel2Name",
		"businessLevel2Operator",
		"initialOperationDepartmentId",
		"initialOperationDepartmentName",
		"operationProductId",
		"operationProductName",
		"planProductId",
		"planProductName"}

	fieldServerInfo = []string{"serverId", "serverAssetId", "serverSn", "serverSourceTypeId", "maintainer",
		"maintainerBak", "maintenanceDepartmentName", "parentServerAssetId", "parentServerId"}

	defaultSize = 10
	maxSize     = 50
)

const (
	businessLevel2Id = "businessLevel2Id"
	serverIp         = "serverIp"
)

// QueryBusinessLeven2InfoReq query business level2 request
type QueryBusinessLeven2InfoReq struct {
	ResultColumn []string               `json:"resultColumn"`
	Size         int                    `json:"size"`
	ScrollId     string                 `json:"scrollId"`
	Condition    map[string]interface{} `json:"condition"`
}

func buildQueryCondition(key, operator string, values []interface{}) map[string]interface{} {
	return map[string]interface{}{
		key: operatorValues{
			Operator: operator,
			Value:    values,
		},
	}
}

type operatorValues struct {
	Operator string        `json:"operator"`
	Value    []interface{} `json:"value"`
}

// BasicResp basic resp
type BasicResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"traceId"`
}

// QueryBusinessL2InfoResp query business resp
type QueryBusinessL2InfoResp struct {
	BasicResp
	Data BusinessL2Resp `json:"data"`
}

// BusinessL2Resp business level2 resp
type BusinessL2Resp struct {
	List     []BusinessL2Info `json:"list"`
	ScrollId string           `json:"scrollId"`
	HasNext  bool             `json:"hasNext"`
}

// BusinessL2Info business level2 info
type BusinessL2Info struct {
	BizLevel1Id   int    `json:"businessLevel1Id"`
	BizLevel1Name string `json:"businessLevel1Name"`
	BizLevel2Id   int    `json:"businessLevel2Id"`
	BizLevel2Name string `json:"businessLevel2Name"`
	// 运营产品名称和ID
	BsiProductName string `json:"operationProductName"`
	BsiProductId   int    `json:"operationProductId"`
	// 规划产品名称和ID
	PlanProductName string `json:"planProductName"`
	PlanProductId   int    `json:"planProductId"`
}

// QueryServerInfoReq query server info request
type QueryServerInfoReq struct {
	ResultColumn []string               `json:"resultColumn"`
	Condition    map[string]interface{} `json:"condition"`
}

// QueryServerInfoResp query server info resp
type QueryServerInfoResp struct {
	BasicResp
	Data ServerListResp `json:"data"`
}

// ServerListResp server info resp
type ServerListResp struct {
	List []Server `json:"list"`
}

// Server server info
type Server struct {
	ServerAssetId string `json:"serverAssetId"`
	SvrOperator   string `json:"maintainer"`
	BakOperator   string `json:"maintainerBak"`
}
