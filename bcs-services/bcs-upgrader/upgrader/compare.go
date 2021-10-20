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

package upgrader

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// CompareVersion compares two upgrade program versions
// the compare priority order is Major > Minor > Patch
// eg：u2.22.202109151940 > u1.22.202109151940 > u1.21.202109151940 > u1.21.202108051940
func CompareVersion(version1, version2 string) int {
	return compareVersion(version1, version2)
}

func int64Compare(val1, val2 int64) int {
	if val1 == val2 {
		return 0
	} else if val1 > val2 {
		return 1
	} else {
		return -1
	}
}

func stringCompare(s1, s2 string) int {
	if s1 == s2 {
		return 0
	} else if s1 > s2 {
		return 1
	} else {
		return -1
	}
}

func compareVersion(version1, version2 string) int {
	// version format should be validate before compare
	v1, err := ParseVersion(version1)
	if err != nil {
		blog.Fatalf(err.Error())
	}
	v2, err := ParseVersion(version2)
	if err != nil {
		blog.Fatalf(err.Error())
	}
	result := int64Compare(v1.Major, v2.Major)
	if result != 0 {
		return result
	}
	result = int64Compare(v1.Minor, v2.Minor)
	if result != 0 {
		return result
	}
	return stringCompare(v1.Patch, v2.Patch)
}

// Version define a version object
type Version struct {
	Major int64
	Minor int64
	Patch string
}

var patchRegex = regexp.MustCompile(`^\d{12}$`)

// ParseVersion parse the string type to Version type
func ParseVersion(version string) (Version, error) {
	v := Version{}
	invalidMessage := fmt.Errorf("invalid version [%s]", version)
	version = strings.TrimLeft(version, VersionPrefix)
	fields := strings.Split(version, ".")
	if len(fields) != 3 {
		return v, invalidMessage
	}

	major, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return v, invalidMessage
	}
	v.Major = major

	minor, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return v, invalidMessage
	}
	v.Minor = minor

	patch := fields[2]
	match := patchRegex.MatchString(patch)
	if !match {
		return v, invalidMessage
	}
	v.Patch = patch
	return v, nil
}
