/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bcsmonitor

// Path of bcs monitor api
const (
	QueryPath       = "/api/v1/query"
	QueryRangePath  = "/api/v1/query_range"
	LabelValuesPath = "/api/v1/label/%s/values"
	LabelsPath      = "/api/v1/labels"
	SeriesPath      = "/api/v1/series"
)

// CommonResponse common response of prometheus
type CommonResponse struct {
	Status    string   `json:"status"`
	ErrorType string   `json:"errorType"`
	Error     string   `json:"error"`
	Warnings  []string `json:"warnings"`
}

// LabelResponse response of label api
type LabelResponse struct {
	CommonResponse
	Data []string `json:"data"`
}

// SeriesResponse response of series api
type SeriesResponse struct {
	CommonResponse
	Data []interface{} `json:"data"`
}

// QueryResponse response of query api
type QueryResponse struct {
	CommonResponse
	Data QueryData `json:"data"`
}

// QueryRangeResponse response of query_range api
type QueryRangeResponse struct {
	CommonResponse
	Data QueryRangeData `json:"data"`
}

// QueryData data struct of QueryResponse
type QueryData struct {
	ResultType string         `json:"resultType"`
	Result     []VectorResult `json:"result"`
}

// QueryRangeData data struct of QueryRangeResponse
type QueryRangeData struct {
	ResultType string         `json:"resultType"`
	Result     []MatrixResult `json:"result"`
}

// MatrixResult matrix result type
type MatrixResult struct {
	Metrics map[string]string `json:"metrics"`
	Values  [][]interface{}   `json:"values"`
}

// VectorResult vector result type
type VectorResult struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}
