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

// Package storage xxx
package storage

import "time"

// CommonResponseHeader is common response for storage resource api
type CommonResponseHeader struct {
	Result   bool   `json:"result"`
	Code     int    `json:"code"`
	Message  string `json:"message"`
	PageSize int64  `json:"pageSize"`
	Offset   int64  `json:"offset"`
	Total    int64  `json:"total"`
}

// CommonDataHeader is common header for storage dynamic data api
type CommonDataHeader struct {
	Namespace          string    `json:"namespace"`
	ResourceName       string    `json:"resourceName"`
	UpdateTime         time.Time `json:"updateTime"`
	ID                 string    `json:"_id"`
	ResourceType       string    `json:"resourceType"`
	ClusterID          string    `json:"clusterId"`
	IsBcsObjectDeleted bool      `json:"_isBcsObjectDeleted"`
	CreateTime         time.Time `json:"createTime"`
}
