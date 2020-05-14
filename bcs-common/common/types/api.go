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

package types

//RequestApi old version api request for bcs-client & bcs-api
type RequestApi struct {
	AppId    string      `json:"appid,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Request  interface{} `json:"request,omitempty"`
}

//BcsRequest request for bcs-api & bcs-client
type BcsRequest struct {
	AppId             string            `json:"appid,omitempty"`
	Operator          string            `json:"operator,omitempty"`
	DataType          BcsDataType       `json:"dataType,omitempty"`
	Pod               BcsTaskgroup      `json:"pod,omitempty"`
	ReplicaController ReplicaController `json:"replicaController,omitempty"`
	ConfigMap         BcsConfigMap      `json:"configMap,omitempty"`
	Service           BcsService        `json:"service,omitempty"`
	Secret            BcsSecret         `json:"secret,omitempty"`
	LoadBalance       BcsLoadBalance    `json:"loadBalance,omitempty"`
}

// bcs standard header keys for http api
const (
	BcsApiHeader_ClusterID = "BCS-ClusterID"
	BcsApiHeader_Operator  = "BCS-Operator"
	BcsApiHeader_UUID      = "BCS-UUID"
)

// bcs http method
const (
	HttpMethod_POST   = "POST"
	HttpMethod_PUT    = "PUT"
	HttpMethod_GET    = "GET"
	HttpMethod_DELETE = "DELETE"
	HttpMethod_PATCH  = "PATCH"
)

type APIResponse struct {
	Result  bool        `json:"result"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}
