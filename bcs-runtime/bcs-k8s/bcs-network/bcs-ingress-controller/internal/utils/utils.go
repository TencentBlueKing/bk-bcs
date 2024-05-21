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

// Package utils
package utils

import (
	"crypto/md5"
	"fmt"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// GenPortBindingLabel 生成portBinding label, 当长度超过63时（k8s限制）， 将截取前50位+md5值的前13位作为label key
func GenPortBindingLabel(namespace string, name string) string {
	result := fmt.Sprintf(networkextensionv1.PortPoolBindingLabelKeyFormat, name, namespace)
	if len(result) > 63 {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(result)))
		return result[:50] + hash[:13]
	}
	return result
}
