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

package util

import (
	"encoding/json"
	"hash/fnv"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

func StructToMap(v interface{}) operator.M {
	data := make(operator.M)
	bytes, _ := json.Marshal(v)
	_ = json.Unmarshal(bytes, &data)
	return data
}

// MapToStruct 尝试从map转换为struct
// debugMsg 仅用于调试，可以为空
func MapToStruct(rawData operator.M, target interface{}, debugMsg string) (err error) {
	bytes, err := json.Marshal(rawData)
	if err != nil {
		return errors.Wrapf(err, "map to json failed. %s", debugMsg)
	}

	if err = json.Unmarshal(bytes, target); err != nil {
		return errors.Wrapf(err, "json to struct failed. %s", debugMsg)
	}

	return nil
}

// ListMapToListStruct 尝试从list map转换为list struct
// debugMsg 仅用于调试，可以为空
func ListMapToListStruct(rawData []operator.M, target interface{}, debugMsg string) (err error) {
	bytes, err := json.Marshal(rawData)
	if err != nil {
		return errors.Wrapf(err, "list map to json list failed. %s", debugMsg)
	}

	if err = json.Unmarshal(bytes, target); err != nil {
		return errors.Wrapf(err, "json list to struct failed. %s", debugMsg)
	}
	return nil
}

// PrettyStruct 美化struct并输出
func PrettyStruct(obj interface{}) string {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return ""
	}
	return string(bytes)
}

// RequestToMap request对象转map
func RequestToMap(tags map[string]string, request interface{}) operator.M {
	ans := make(operator.M)
	elem := reflect.ValueOf(request).Elem()
	for key, value := range tags {
		field := elem.FieldByName(key)
		switch field.Kind() {
		case reflect.String:
			ans[value] = field.String()
		case reflect.Ptr:
			ans[value] = (*structpb.Struct)(unsafe.Pointer(field.Pointer()))
		}
	}
	return ans
}

// GetStructTags 获取struct filed name 与 filed tag json 一一对饮map
func GetStructTags(obj interface{}) map[string]string {
	tags := make(map[string]string)
	elem := reflect.TypeOf(obj).Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			continue
		}
		tags[field.Name] = strings.Split(tag, ",")[0]
	}
	return tags
}

// StructToStruct 异构转换
func StructToStruct(obj interface{}, target interface{}, debugMsg string) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrapf(err, "map to json failed. %s", debugMsg)
	}

	if err = json.Unmarshal(bytes, target); err != nil {
		return errors.Wrapf(err, "json to struct failed. %s", debugMsg)
	}
	return nil
}

// HashString2Time 把字符串哈希到给定时间段之间的某个时间段
func HashString2Time(str string, maxTime time.Duration) time.Duration {
	if maxTime <= time.Duration(0) {
		return time.Duration(0)
	}
	// 计算字符串的哈希值
	h := fnv.New64a()
	h.Write([]byte(str))
	hash := h.Sum64()

	premill := int64(hash % 1000)
	randomDuration := maxTime / 1000 * time.Duration(premill)

	return randomDuration
}
