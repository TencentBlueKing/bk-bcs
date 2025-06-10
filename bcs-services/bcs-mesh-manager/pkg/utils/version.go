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

package utils

import (
	"github.com/Masterminds/semver/v3"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// IsVersionSupported 判断当前 istioVersion 是否满足 supportVersion 区间
func IsVersionSupported(istioVersion, supportVersion string) bool {
	if supportVersion == "" {
		return true
	}
	v, err := semver.NewVersion(istioVersion)
	if err != nil {
		blog.Errorf("failed to parse istio version %s, err %s", istioVersion, err.Error())
		return false
	}
	c, err := semver.NewConstraint(supportVersion)
	if err != nil {
		blog.Errorf("failed to parse support version %s, err %s", supportVersion, err.Error())
		return false
	}
	return c.Check(v)
}
