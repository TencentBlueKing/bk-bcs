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

// IsVersionSupported 判断当前 clusterVersion 是否满足 supportVersion 区间
func IsVersionSupported(clusterVersion, supportVersion string) bool {
	if supportVersion == "" {
		return true
	}
	// 去除 '-' 后的所有内容
	if idx := len(clusterVersion); idx > 0 {
		if dash := stringIndex(clusterVersion, "-"); dash >= 0 {
			clusterVersion = clusterVersion[:dash]
		}
	}
	// 去除前缀 'v'，clusterVersion 和 supportVersion 都去除前缀 'v'
	if len(clusterVersion) > 0 && clusterVersion[0] == 'v' {
		clusterVersion = clusterVersion[1:]
	}
	if len(supportVersion) > 0 && supportVersion[0] == 'v' {
		supportVersion = supportVersion[1:]
	}

	v, err := semver.NewVersion(clusterVersion)
	if err != nil {
		blog.Errorf("failed to parse cluster version %s, err %s", clusterVersion, err.Error())
		return false
	}
	c, err := semver.NewConstraint(supportVersion)
	if err != nil {
		blog.Errorf("failed to parse support version %s, err %s", supportVersion, err.Error())
		return false
	}
	return c.Check(v)
}

// stringIndex returns the index of the first instance of substr in s, or -1 if substr is not present in s.
func stringIndex(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
