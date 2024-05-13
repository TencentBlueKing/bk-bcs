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

package ingresscache

import (
	"fmt"
	"strings"
)

const (
	// ingressKeyFmt namespace/name
	ingressKeyFmt = "%s/%s"
	// serviceKeyFmt namespace/name
	serviceKeyFmt = "%s/%s/%s"
	// workloadKeyFmt kind/namespace/name
	// kind is StatefulSet or GameStatefulSet
	workloadKeyFmt = "%s/%s/%s"
)

// buildIngressKey return cache key of ingress
func buildIngressKey(namespace, name string) string {
	return fmt.Sprintf(ingressKeyFmt, namespace, name)
}

// buildServiceKey return cache key of service
func buildServiceKey(kind, namespace, name string) string {
	return fmt.Sprintf(serviceKeyFmt, strings.ToLower(kind), namespace, name)
}

// buildWorkloadKey return cache key of workload
func buildWorkloadKey(kind, namespace, name string) string {
	return fmt.Sprintf(workloadKeyFmt, kind, namespace, name)
}
