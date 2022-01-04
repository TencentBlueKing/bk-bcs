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

package util

import (
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Unstructured2pbStruct unstructured.Unstructured => structpb.Struct
func Unstructured2pbStruct(u *unstructured.Unstructured) *structpb.Struct {
	fields := map[string]*structpb.Value{}
	for k, v := range u.UnstructuredContent() {
		val, _ := structpb.NewValue(v)
		fields[k] = val
	}
	return &structpb.Struct{Fields: fields}
}
