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

// Package util xxx
package util

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// CreateMergePatch return kube generated from original and new interfaces
func CreateMergePatch(original, new interface{}) ([]byte, error) {
	pvByte, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}
	cloneByte, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}
	patch, err := strategicpatch.CreateTwoWayMergePatch(pvByte, cloneByte, original)
	if err != nil {
		return nil, err
	}
	return patch, nil
}
