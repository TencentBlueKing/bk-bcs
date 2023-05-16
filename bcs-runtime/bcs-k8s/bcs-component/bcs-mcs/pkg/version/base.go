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

package version

import "github.com/Tencent/bk-bcs/bcs-common/common/version"

// Base version information.
//
// This is the fallback data used when version information from git is not
// provided via go ldflags. It provides an approximation of the Karmada
// version for ad-hoc builds (e.g. `go build`) that cannot get the version
// information from git.
var (
	gitVersion   = version.BcsVersion
	gitCommit    = version.BcsGitHash   // sha1 from git, output of $(git rev-parse HEAD)
	gitTreeState = "unknown"            // state of git tree, either "clean" or "dirty"
	buildDate    = version.BcsBuildTime // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)
