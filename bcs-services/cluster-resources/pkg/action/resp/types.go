/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resp

// ListParams 资源 List 请求通用参数
type ListParams struct {
	ClusterID    string
	ResKind      string
	GroupVersion string
	Namespace    string
	Format       string
	Scene        string
}

// GetParams 资源 Get 请求通用参数
type GetParams struct {
	ClusterID    string
	ResKind      string
	GroupVersion string
	Namespace    string
	Name         string
	Format       string
}

// DataBuilder 接收不同类型的参数，转换成 List/Retrieve 请求的响应数据
type DataBuilder interface {
	// BuildList 构建 List API RespData
	BuildList() (map[string]interface{}, error)
	// Build 构建 Retrieve API RespData
	Build() (map[string]interface{}, error)
}

// DataBuilderParams 初始化 DataBuilder 参数
type DataBuilderParams struct {
	Manifest map[string]interface{}
	Kind     string
	Format   string
	Scene    string
}
