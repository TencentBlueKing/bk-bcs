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

import (
	"fmt"
	"runtime"
)

// BcsVersion describes version
// BcsTag show the git tag for this version
// BcsBuildTime show the compile time
var (
	BcsVersion   = "17.03.28"
	BcsTag       = "2017-03-28 Release"
	BcsBuildTime = "2017-03-28 19:50:00"
	BcsGitHash   = "unknown"
	GoVersion    = runtime.Version()
)

// ShowVersion is the default handler which match the --version flag
func ShowVersion() {
	fmt.Printf("%s", GetVersion())
}

// GetVersion xxx
func GetVersion() string {
	version := fmt.Sprintf("Version  :  %s\nTag      :  %s\nBuildTime:  %s\nGitHash  :  %s\nGoVersion:  %s\n",
		BcsVersion, BcsTag, BcsBuildTime, BcsGitHash, GoVersion)
	return version
}
