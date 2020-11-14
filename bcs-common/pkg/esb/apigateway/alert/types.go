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

package alert

import "time"

const (
	startKey = "startsAt"
	endKey   = "endsAt"

	annotationKey        = "annotations"
	annotationMessageKey = "message"

	labelKey        = "labels"
	labelAlertKey   = "alert_type"
	labelClusterKey = "cluster_id"
	labelNSKey      = "namespace"
	labelIPKey      = "ip"
	labelModuleKey  = "module_name"
)

func newServiceAlert(module, message, ip string) map[string]interface{} {
	return map[string]interface{}{
		startKey: time.Now(),
		endKey:   time.Now(),
		annotationKey: map[string]string{
			annotationMessageKey: message,
		},
		labelKey: map[string]string{
			labelAlertKey:   "Error",
			labelClusterKey: "bcs-service",
			labelIPKey:      ip,
			labelNSKey:      "bcs-service",
			labelModuleKey:  module,
		},
	}
}

func newClusterAlert(cluster, module, message, ip string) map[string]interface{} {
	return map[string]interface{}{
		startKey: time.Now(),
		endKey:   time.Now(),
		annotationKey: map[string]string{
			annotationMessageKey: message,
		},
		labelKey: map[string]string{
			labelAlertKey:   "Error",
			labelClusterKey: cluster,
			labelIPKey:      ip,
			labelNSKey:      "bcs-cluster",
			labelModuleKey:  module,
		},
	}
}
