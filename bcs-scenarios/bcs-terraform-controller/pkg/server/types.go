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

package server

// ApplyRequest apply请求
type ApplyRequest struct {
	// Name tf cr名称
	Name string `json:"name,omitempty"`
	// Namespace 名称空间
	Namespace string `json:"namespace,omitempty"`
}

// TaskApplyResult task apply结果
type TaskApplyResult struct {
	// Name tf cr名称
	Name string `json:"name,omitempty"`
	// Namespace 名称空间
	Namespace string `json:"namespace,omitempty"`
	// Result 结果
	Result bool `json:"result,omitempty"`
	// Message 消息
	Message string `json:"message,omitempty"`
}

// ApplyResponse apply响应
type ApplyResponse struct {
	// Code 状态码
	Code int `json:"statusCode,omitempty"`
	// Message 消息
	Message string `json:"message,omitempty"`
	// Result 结果
	Result bool `json:"result,omitempty"`
	// Data 详细结果
	Data []*TaskApplyResult `json:"data,omitempty"`
}

// CreatePlanRequest 创建plan
type CreatePlanRequest struct {
	// ID trace-id
	ID string `json:"id,omitempty"`
	// Name 名称
	Name string `json:"name,omitempty"`
	// Namespace 名称空间
	Namespace string `json:"namespace,omitempty"`
	// TargetRevision 目标对象
	TargetRevision string `json:"targetRevision,omitempty"`
	// Hook 是否进入hook逻辑
	Hook bool `json:"hook,omitempty"`
}

// GetTerraformRequest 根据repo url查询tf资源
type GetTerraformRequest struct {
	Url string `json:"url,omitempty"`
}

// BaseResponse 通用响应
type BaseResponse struct {
	// Code 状态码
	Code int `json:"statusCode,omitempty"`
	// Message 消息
	Message string `json:"message,omitempty"`
	// Result 结果
	Result bool `json:"result,omitempty"`
	// Data 数据
	Data interface{} `json:"data,omitempty"`
}
