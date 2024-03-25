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

// Package calculator 计算迁移计划的接口
package calculator

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// CalculateInterface defines the interface of calculator
type CalculateInterface interface {
	Calculate(ctx context.Context, req *CalculateConvergeRequest) (ResultPlan, error)
}

// CalculateConvergeRequest defines the request of calculator
type CalculateConvergeRequest struct {
	AuthenticationMethod string                 `json:"bkdata_authentication_method"`
	Token                string                 `json:"bkdata_data_token"`
	AppCode              string                 `json:"bk_app_code"`
	AppSecret            string                 `json:"bk_app_secret"`
	Data                 *CalculateData         `json:"data"`
	Config               *CalculateConfig       `json:"config"`
	Original             *CalculateOriginalData `json:"-"`
}

// String request to string
func (c *CalculateConvergeRequest) String() string {
	bs, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("unmarshal request failed: %s", err.Error())
	}
	return string(bs)
}

// String response to string
func (c *CalculateConvergeResponse) String() string {
	bs, err := json.Marshal(c)
	if err != nil {
		return fmt.Sprintf("unmarshal response failed: %s", err.Error())
	}
	return string(bs)
}

// CalculateOriginalData defines the original data of calculate
type CalculateOriginalData struct {
	Pods  []*PodItem  `json:"pods"`
	Nodes []*NodeItem `json:"nodes"`
}

// CalculateData the data item of request
type CalculateData struct {
	Inputs []RequestInputs `json:"inputs"`
}

// CalculateConfig the config item of request, containers affinity.
type CalculateConfig struct {
	PredictArgs PredictArgs `json:"predict_args"`
}

// RequestInputs defines the input items of request
type RequestInputs struct {
	Pod  string `json:"pod"`
	Node string `json:"node"`
	Time int64  `json:"time"`
}

// PodItem defines every pod
type PodItem struct {
	Item              string      `json:"item"`
	Index1            float64     `json:"index1"`
	Index2            float64     `json:"index2"`
	Container         string      `json:"container"`
	IsAllowMigrate    int32       `json:"is_allow_migrate"`
	MigrationPriority int32       `json:"migration_priority"`
	OriginalPod       *corev1.Pod `json:"-"`
}

// NodeItem defines every node
type NodeItem struct {
	Container      string       `json:"container"`
	Index1         float64      `json:"index1"`
	Index2         float64      `json:"index2"`
	ItemNums       int32        `json:"item_nums"`
	IsAllowMigrate int32        `json:"is_allow_migrate"`
	OriginalNode   *corev1.Node `json:"-"`
}

// PredictArgs defines the args of request
type PredictArgs struct {
	Scope              PredictScope     `json:"scope"`
	OptimizeTarget     []OptimizeTarget `json:"optimize_target"`
	IterationLimit     int32            `json:"iteration_limit"`
	PopulationSize     int32            `json:"population_size"`
	MigrationCostLimit float32          `json:"migration_cost_limit"`
	MigrationWaterline float32          `json:"migration_waterline"`
	MigrationDegree    string           `json:"migration_degree"`
	IsCompressed       int32            `json:"is_compressed"`
}

// PredictScope defines all the affinity container pod and pod, pod and node
type PredictScope struct {
	ItemAffinity          []Affinity `json:"item_affinity,omitempty"`
	ItemAntiAffinity      []Affinity `json:"item_anti_affinity,omitempty"`
	ContainerAffinity     []Affinity `json:"container_affinity,omitempty"`
	ContainerAntiAffinity []Affinity `json:"container_anti_affinity,omitempty"`
}

// Affinity defines the affinity
type Affinity struct {
	ItemCondition      []Condition `json:"item_condition,omitempty"`
	ContainerCondition []Condition `json:"container_condition,omitempty"`

	IsForced bool `json:"is_forced"`
}

// Condition defines the condition of affinity
type Condition struct {
	Table         string      `json:"table"`
	Col           string      `json:"col"`
	ConditionType string      `json:"condition_type"`
	Value         interface{} `json:"value"`
}

// OptimizeTarget defines the optimize target of calculator
type OptimizeTarget struct {
	Name              string `json:"name"`
	OptimizeDirection string `json:"optimize_direction"`
	FieldType         string `json:"field_type"`
}

// CalculateConvergeResponse defines the response of calculate.
type CalculateConvergeResponse struct {
	Result  bool          `json:"result"`
	Errors  []interface{} `json:"errors"`
	Message string        `json:"message"`
	Code    string        `json:"code"`
	Data    ResponseData  `json:"data"`
}

// ResponseData defines the data of response
type ResponseData struct {
	Errors         []interface{}        `json:"errors"`
	Message        string               `json:"message"`
	Code           string               `json:"code"`
	Result         bool                 `json:"result"`
	PredictTime    string               `json:"predict_time"`
	APIServingTime string               `json:"api_serving_time"`
	Data           APIServeResponseData `json:"data"`
}

// APIServeResponseData defines the response data of api served
type APIServeResponseData struct {
	Status string       `json:"status"`
	Data   []ResultData `json:"data"`
}

// ResultData defines the result data of calculator
type ResultData struct {
	Output []ResultOutput `json:"output"`
}

// ResultOutput defines the output of calculator
type ResultOutput struct {
	Timestamp int64      `json:"timestamp"`
	Plan      ResultPlan `json:"plan"`
}

// ResultPlan defines the plan of calculator
type ResultPlan struct {
	PlanCount int32          `json:"plan_count"`
	Plans     []ResponsePlan `json:"plans"`
}

// ResponsePlan defines all the migrate plans that calculator response
type ResponsePlan struct {
	PlanTags    []string              `json:"plan_tags"`
	MigratePlan []ResponseMigratePlan `json:"migrate_plan"`
}

// ResponseMigratePlan defines the plan detail for every item
type ResponseMigratePlan struct {
	Item     string `json:"item"`
	From     string `json:"from"`
	To       string `json:"to"`
	Priority int32  `json:"priority"`
}
