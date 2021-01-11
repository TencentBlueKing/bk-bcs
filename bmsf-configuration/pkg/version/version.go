/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package version

import (
	"fmt"
)

const (
	// LOGO is bk bscp inner logo.
	LOGO = `
===================================================================================
oooooooooo   oooo    oooo         oooooooooo     oooooooo     oooooo    oooooooooo
 888     Y8b  888   8P             888     Y8b d8P      Y8  d8P    Y8b   888    Y88
 888     888  888  d8              888     888 Y88bo       888           888    d88
 888oooo888   88888[               888oooo888     Y8888o   888           888ooo88P
 888     88b  888 88b     8888888  888     88b        Y88b 888           888
 888     88P  888   88b            888     88P oo      d8P  88b    ooo   888
o888bood8P   o888o  o888o         o888bood8P   88888888P     Y8bood8P   o888o
===================================================================================`
)

var (
	// VERSION version info.
	VERSION = "0.1.1-alpha"

	// BUILDTIME  build time.
	BUILDTIME = "2019-07-16T10:23:59"

	// GITHASH git hash for release.
	GITHASH = "1234567890"
)

// ShowVersion shows the version info.
func ShowVersion() {
	fmt.Printf("%s", GetVersion())
}

// GetVersion returns compilling version.
func GetVersion() string {
	version := fmt.Sprintf("Version: %s\nBuildTime: %s\nGitHash: %s\n", VERSION, BUILDTIME, GITHASH)
	return version
}

// GetStartInfo returns start info that includes version and logo.
func GetStartInfo() string {
	startInfo := fmt.Sprintf("%s\n\n%s\n", LOGO, GetVersion())
	return startInfo
}
