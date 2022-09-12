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

// Package version xxx
package version

import "fmt"

var (
	version   = "unknown"
	gitCommit = "$Format:%H$"          // sha1 from git, output of $(git rev-parse HEAD)
	buildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

// SimpleVersion xxx
type SimpleVersion struct {
	version   string
	gitCommit string
	buildDate string
}

// ForDisplay xxx
func (v *SimpleVersion) ForDisplay() string {
	return fmt.Sprintf(
		`Version: "%s", GitCommit: "%s", BuildDate: "%s"`,
		v.version,
		v.gitCommit,
		v.buildDate)
}

// Get xxx
func Get() *SimpleVersion {
	return &SimpleVersion{
		version:   version,
		gitCommit: gitCommit,
		buildDate: buildDate,
	}
}
