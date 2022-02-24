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

package pbstruct

import (
	"encoding/base64"
	"unicode/utf8"

	"google.golang.org/protobuf/runtime/protoimpl"
	spb "google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Unstructured2pbStruct unstructured.Unstructured => structpb.Struct
func Unstructured2pbStruct(u *unstructured.Unstructured) *spb.Struct {
	fields := map[string]*spb.Value{}
	for k, v := range u.UnstructuredContent() {
		val, _ := spb.NewValue(v)
		fields[k] = val
	}
	return &spb.Struct{Fields: fields}
}

// MapSlice2ListValue []map[string]interface{} => structpb.ListValue
func MapSlice2ListValue(l []map[string]interface{}) (*spb.ListValue, error) {
	x := &spb.ListValue{Values: make([]*spb.Value, len(l))}
	for i, v := range l {
		var err error
		// 由于这里 v 的类型已确定，因此不使用 interface2pbValue，
		// 直接使用 map2StructValue，可减少一次类型检查
		x.Values[i], err = map2StructValue(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// map[string]interface{} => structpb.Value
func map2StructValue(m map[string]interface{}) (*spb.Value, error) {
	v2, err := Map2pbStruct(m)
	if err != nil {
		return nil, err
	}
	return spb.NewStructValue(v2), nil
}

// Map2pbStruct map[string]interface{} => structpb.Struct
// structpb.NewStruct 中对列表类型只支持 []interface{}，如 []string 类型则不受支持
// 解决思路有以下几个，这里采用了思路 1，若有更好的实现可以替换
// 1. 自定义 NewStruct，支持需要的类型，也就是 Map2PbStruct 存在的原因
// 2. 在使用 NewStruct 前逐层检查类型，并做一次转换，缺点是需要坐两次类型检查（检查，构建各一次）
// 3. json.Marshal + protojson.UnmarshalJSON，原理和 2 相似，但是会有性能瓶颈
func Map2pbStruct(m map[string]interface{}) (*spb.Struct, error) {
	x := &spb.Struct{Fields: make(map[string]*spb.Value, len(m))}
	for k, v := range m {
		if !utf8.ValidString(k) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", k)
		}
		var err error
		x.Fields[k], err = interface2pbValue(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// interface -> structpb.Value
// 参考 structpb.NewValue 实现，添加对 []string 类型的支持，若需要支持更多类型可按需添加
func interface2pbValue(v interface{}) (*spb.Value, error) {
	switch v := v.(type) {
	case nil:
		return spb.NewNullValue(), nil
	case bool:
		return spb.NewBoolValue(v), nil
	case int:
		return spb.NewNumberValue(float64(v)), nil
	case int32:
		return spb.NewNumberValue(float64(v)), nil
	case int64:
		return spb.NewNumberValue(float64(v)), nil
	case uint:
		return spb.NewNumberValue(float64(v)), nil
	case uint32:
		return spb.NewNumberValue(float64(v)), nil
	case uint64:
		return spb.NewNumberValue(float64(v)), nil
	case float32:
		return spb.NewNumberValue(float64(v)), nil
	case float64:
		return spb.NewNumberValue(v), nil
	case string:
		if !utf8.ValidString(v) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", v)
		}
		return spb.NewStringValue(v), nil
	case []byte:
		s := base64.StdEncoding.EncodeToString(v)
		return spb.NewStringValue(s), nil
	case map[string]interface{}:
		v2, err := Map2pbStruct(v)
		if err != nil {
			return nil, err
		}
		return spb.NewStructValue(v2), nil
	case []interface{}:
		v2, err := spb.NewList(v)
		if err != nil {
			return nil, err
		}
		return spb.NewListValue(v2), nil
	case []map[string]interface{}:
		v2, err := MapSlice2ListValue(v)
		if err != nil {
			return nil, err
		}
		return spb.NewListValue(v2), nil
	case []string:
		v2, err := newStringList(v)
		if err != nil {
			return nil, err
		}
		return spb.NewListValue(v2), nil
	default:
		return nil, protoimpl.X.NewError("invalid type: %T", v)
	}
}

// 参考 structpb.NewList，NewValue(case string) 实现，支持 []string 类型
func newStringList(strList []string) (*spb.ListValue, error) {
	x := &spb.ListValue{Values: make([]*spb.Value, len(strList))}
	for idx, val := range strList {
		if !utf8.ValidString(val) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", val)
		}
		x.Values[idx] = spb.NewStringValue(val)
	}
	return x, nil
}
