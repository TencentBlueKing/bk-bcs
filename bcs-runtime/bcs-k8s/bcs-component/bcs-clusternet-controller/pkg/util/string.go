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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-clusternet-controller/pkg/constant"
)

// MatchAnnotationsKeyPrefix 匹配annotations中带上了 AnnotationSubscriptionKeyPrefix 的key
func MatchAnnotationsKeyPrefix(annotations map[string]string) bool {
	for key := range annotations {
		if strings.HasPrefix(key, constant.AnnotationSubscriptionKeyPrefix) {
			return true
		}
	}
	return false
}

// FindAnnotationsMathKeyPrefix 返回匹配的annotations
func FindAnnotationsMathKeyPrefix(annotations map[string]string) map[string]string {
	ret := make(map[string]string)
	for key, val := range annotations {
		if strings.HasPrefix(key, constant.AnnotationSubscriptionKeyPrefix) {
			ret[key] = val
		}
	}
	return ret
}
