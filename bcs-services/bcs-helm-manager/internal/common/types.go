/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"encoding/json"
	"fmt"

	_struct "github.com/golang/protobuf/ptypes/struct"
)

// MarshalInterfacesToListValue trans interface slice to ListValue
func MarshalInterfacesToListValue(objs interface{}) (*_struct.ListValue, error) {
	b, err := json.Marshal(objs)
	if err != nil {
		return nil, err
	}
	l := &_struct.ListValue{}
	err = l.UnmarshalJSON(b)
	if err != nil {
		return nil, err
	}
	return l, nil
}

// MarshalInterfaceToValue trans interface to Struct
func MarshalInterfaceToValue(data interface{}) (*_struct.Struct, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	l := &_struct.Struct{}
	err = l.UnmarshalJSON(b)
	if err != nil {
		return nil, err
	}
	return l, nil
}

// GetPublicRepoURL get public repo url
func GetPublicRepoURL(url, project, name string) string {
	return fmt.Sprintf("%s/helm/%s/%s", url, project, name)
}
