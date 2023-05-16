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

package model

// Metadata k8s 资源基础信息
type Metadata struct {
	APIVersion  string       `structs:"apiVersion"`
	Kind        string       `structs:"kind"`
	Name        string       `structs:"name"`
	Namespace   string       `structs:"namespace"`
	Labels      []Label      `structs:"labels"`
	Annotations []Annotation `structs:"annotations"`
	ResVersion  string       `structs:"resVersion"`
}

// Label k8s 资源标签
type Label struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}

// Annotation k8s 资源注解
type Annotation struct {
	Key   string `structs:"key"`
	Value string `structs:"value"`
}
