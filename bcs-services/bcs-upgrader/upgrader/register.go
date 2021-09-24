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
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// Upgrade define a version upgrade
type Upgrade struct {
	version string // eg: u1.21.202109241520
	do      func(context.Context, UpgradeHelper) error
}

var (
	upgradePool = []Upgrade{}

	registLock sync.Mutex

	versionRegexp = regexp.MustCompile(`^u(\d+\.){2}\d{12}$`)
)

// ValidateVersionFormat validate the format of version
func ValidateVersionFormat(version string) error {
	if !versionRegexp.MatchString(version) {
		err := fmt.Errorf("invalid upgrade version: %s,please use a valid format:eg: u1.21.202109241520", version)
		return err
	}

	v, err := ParseVersion(version)
	if err != nil {
		return err
	}

	// third field in version split by `.` shouldn't greater than tomorrow
	timeFormat := "200601021504"
	maxMigrationTime := time.Now().AddDate(0, 0, 1)
	maxVersionCurrently := maxMigrationTime.Format(timeFormat)
	if v.Patch >= maxVersionCurrently {
		err := fmt.Errorf("invalid time field of upgrade version: %s, "+
			"please use current time as part of upgrade version:eg: u1.21.%s", version, time.Now().Format(timeFormat))
		return err
	}
	return nil
}

// RegisterUpgrade register upgrade programe
func RegisterUpgrade(version string, handlerFunc func(context.Context, UpgradeHelper) error) {
	if err := ValidateVersionFormat(version); err != nil {
		blog.Fatalf("ValidateVersionFormat failed, err: %s", err.Error())
	}
	registLock.Lock()
	defer registLock.Unlock()
	u := Upgrade{
		version: version,
		do: func(ctx context.Context, helper UpgradeHelper) error {
			return handlerFunc(ctx, helper)
		},
	}
	upgradePool = append(upgradePool, u)
}
