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

const (
	// MaxK8sLabelLen is the max length of a kubernetes label key or value.
	MaxK8sLabelLen = 63
	// labelKeyPrefixLen is the prefix length kept when truncating overlong label keys.
	labelKeyPrefixLen = 50
	// labelKeyHashLen is the md5 hex prefix length appended after truncation.
	labelKeyHashLen = 13
)

// GenPortBindingLabel 生成portBinding label, 当长度超过63时（k8s限制）， 将截取前50位+md5值的前13位作为label key
func GenPortBindingLabel(name string, namespace string) string {
	result := fmt.Sprintf(networkextensionv1.PortPoolBindingLabelKeyFormat, name, namespace)
	return truncateLabelKey(result)
}

// GenIngressLabelKey generates a k8s-safe label key/value from an Ingress name.
// Listener CRs use the Ingress name as a label key (and as owner-name value);
// both are limited to 63 characters. Longer names are truncated with an md5
// suffix (prefix 50 + hash 13) to stay unique and valid.
func GenIngressLabelKey(ingressName string) string {
	return truncateLabelKey(ingressName)
}

// truncateLabelKey returns key unchanged if len<=63; otherwise prefix50+md5[:13].
func truncateLabelKey(key string) string {
	if len(key) <= MaxK8sLabelLen {
		return key
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	return key[:labelKeyPrefixLen] + hash[:labelKeyHashLen]
}
