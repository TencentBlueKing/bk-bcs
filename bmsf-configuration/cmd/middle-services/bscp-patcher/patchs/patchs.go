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

package patchs

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"bk-bscp/cmd/middle-services/bscp-patcher/modules/hpm"
	"bk-bscp/cmd/middle-services/bscp-patcher/patchs/v0.0.0-202011201517"
	"bk-bscp/pkg/logger"
)

var (
	// all registered patchs.
	patchs = []hpm.Patch{}
)

// register one patch.
func register(patchInterface hpm.PatchInterface) {
	patchName := patchInterface.GetName()
	if err := validatePatchName(patchName); err != nil {
		logger.Fatal("patch name format error, val: %s, err: %+v", patchName, err)
		return
	}
	patchs = append(patchs, hpm.Patch{Version: patchName, PatchInterface: patchInterface})
}

// validatePatchName verify the name of the patch
func validatePatchName(patchName string) error {
	// example: v0.0.0-202011201517
	nameArr := strings.Split(patchName, "-")
	if len(nameArr) != 2 {
		return fmt.Errorf("invalid patch name")
	}

	version := nameArr[0]
	if !strings.HasPrefix(version, "v") {
		return fmt.Errorf("%s is invalid version, dont't have prefix v", version)
	}
	version = version[1:len(version)]
	versionArr := strings.Split(version, ".")
	for _, val := range versionArr {
		_, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("%s is invalid version, val: %s", version, val)
		}
	}

	createTime := nameArr[1]
	patchRegex := regexp.MustCompile(`^\d{12}$`)
	match := patchRegex.MatchString(createTime)
	if !match {
		return fmt.Errorf("%s is invalid createTime", createTime)
	}
	timeFormat := "200601021504"
	maxTime := time.Now().AddDate(0, 0, 1)
	maxVersionCurrently := maxTime.Format(timeFormat)
	if createTime >= maxVersionCurrently {
		return fmt.Errorf("%s is invalid createTime", createTime)
	}

	return nil
}

// Patchs returns all registered patchs.
func Patchs() []hpm.Patch {
	sort.Slice(patchs, func(i, j int) bool {
		return hpm.VersionCmp(patchs[i].Version, patchs[j].Version) < 0
	})
	return patchs
}

// register patchs here.
func init() {
	// register patchs.
	register(&v0v0v0v202011201517.Patch{Name: "v0.0.0-202011201517"})

	// TODO add your patch pkg here.
}
