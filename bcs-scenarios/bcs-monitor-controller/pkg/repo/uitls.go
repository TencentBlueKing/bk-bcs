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

package repo

import (
	"fmt"

	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
)

func genRepoKey(repoURL, targetRevision string) string {
	key := fmt.Sprintf("%s@%s", repoURL, targetRevision)
	// // k8s 名称段是必需的，必须小于等于 63 个字符，以字母数字字符（[a-z0-9A-Z]）开头和结尾， 带有破折号（-），下划线（_），点（ .）和之间的字母数字。
	// encoded := base64.URLEncoding.EncodeToString([]byte(key))
	// encoded = strings.TrimRight(encoded, "=")
	// if len(encoded) > 60 {
	// 	encoded = encoded[:60]
	// }
	// return encoded
	return key
}

// GenRepoKeyFromAppMonitor get repoKey from AppMonitor
func GenRepoKeyFromAppMonitor(appMonitor *monitorextensionv1.AppMonitor) string {
	if appMonitor == nil {
		return ""
	}

	repoRef := appMonitor.Spec.RepoRef
	if repoRef == nil {
		return RepoKeyDefault
	}

	return genRepoKey(repoRef.URL, repoRef.TargetRevision)
}
