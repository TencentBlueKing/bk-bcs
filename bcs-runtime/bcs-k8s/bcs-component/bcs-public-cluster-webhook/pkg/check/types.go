/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package check

// BlackListConfig 黑名单的配置
type BlackListConfig struct {
	ExcludeNamespace []string         `yaml:"excludeNamespace"`
	List             []*MatchResource `yaml:"list"`
}

// MatchResource 匹配的资源信息
type MatchResource struct {
	ExcludeNamespace []string      `yaml:"excludeNamespace"`
	Message          string        `yaml:"message"`
	ResourceType     []string      `yaml:"resourceType"`
	MatchQuery       []*MatchQuery `yaml:"matchQuery"`
}

// MatchQuery 具体匹配项
type MatchQuery struct {
	JSONPath  string   `yaml:"jsonPath"`
	Operation string   `yaml:"operation"`
	Value     []string `yaml:"value"`
}

// RequestCheck request
type RequestCheck struct {
	Kind      string
	Namespace string
	Name      string
	Object    []byte
}

// ResponseCheck response
type ResponseCheck struct {
	Allowed bool
	Message string
}
