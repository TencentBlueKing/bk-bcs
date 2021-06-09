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

package hpm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	"bk-bscp/pkg/logger"
)

const (
	// ProductKindOA is oa product kind.
	ProductKindOA = "oa"

	// ProductKindEE is ee product kind.
	ProductKindEE = "ee"

	// ProductKindCE is ce product kind.
	ProductKindCE = "ce"
)

// Patch is hot patch struct included a registered name and interface that
// puts the patch content.
type Patch struct {
	// Version is patch version.
	Version string

	// PatchInterface is the Interface which contains the methods that the patch needs to implement.
	PatchInterface PatchInterface
}

// PatchInterface is the Interface which contains the methods that the patch needs to implement.
type PatchInterface interface {
	// GetName return patch version name.
	GetName() string

	// PatchFunc is the func which would puts target patch drived by hpm.
	PatchFunc(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager) error

	// NeedToSkip is the func to decide if the patch should be skipped.
	NeedToSkip(kind string) bool
}

// VersionCmp compare the order of different patch versions
func VersionCmp(patchName1, patchName2 string) int {
	patchName1Arr := strings.Split(patchName1, "-")
	version1 := patchName1Arr[0]
	version1 = version1[1:len(version1)]
	release1Arr := strings.Split(version1, ".")

	patchName2Arr := strings.Split(patchName2, "-")
	version2 := patchName2Arr[0]
	version2 = version2[1:len(version2)]
	release2Arr := strings.Split(version2, ".")

	len1 := len(release1Arr)
	len2 := len(release2Arr)
	var minLen int
	if len1 < len2 {
		minLen = len1
	} else {
		minLen = len2
	}

	cursor := 0
	for {
		if cursor >= minLen {
			break
		}
		result := strings.Compare(release1Arr[cursor], release2Arr[cursor])
		if result != 0 {
			return result
		}
		cursor++
	}
	if len1 != len2 {
		return len1 - len2
	}

	return strings.Compare(patchName1Arr[1], patchName2Arr[1])
}

// PatchManager is hot patch manager that loads and puts all available patchs.
type PatchManager struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	patchs []Patch

	patchMapping map[string]*Patch
}

// NewPatchManager creates a new PatchManager instance.
func NewPatchManager(viper *viper.Viper, smgr *dbsharding.ShardingManager) *PatchManager {
	return &PatchManager{
		viper:        viper,
		smgr:         smgr,
		patchs:       []Patch{},
		patchMapping: make(map[string]*Patch),
	}
}

// Load loads all available Patchs.
func (mgr *PatchManager) Load(patchs []Patch) {
	mgr.patchs = patchs

	mapping := make(map[string]*Patch)
	for _, patch := range mgr.patchs {
		mapping[patch.Version] = &patch
	}
	mgr.patchMapping = mapping
}

// PutPatchs puts all available patchs.
func (mgr *PatchManager) PutPatchs(operator string) (*database.System, error) {
	return mgr.PutPatch(operator, "")
}

// PutPatch puts target patch until limit.
func (mgr *PatchManager) PutPatch(operator, limitVersion string) (*database.System, error) {
	curVersionRecord, err := mgr.GetCurrentVersion()
	if err != nil {
		return nil, err
	}
	curVersion := curVersionRecord.CurrentVersion
	kind := curVersionRecord.Kind

	if len(limitVersion) != 0 {
		patch, isExist := mgr.patchMapping[limitVersion]
		if !isExist {
			return nil, errors.New("target limit version patch not found")
		}
		if patch.PatchInterface.NeedToSkip(kind) {
			return nil, fmt.Errorf("current product kind is %s, can not do this limit version patch", kind)
		}
	} else {
		if len(mgr.patchs) == 0 {
			return curVersionRecord, nil
		}
		limitVersion = mgr.patchs[len(mgr.patchs)-1].Version
	}
	logger.Infof("execute patchs now, current version: %+v, limit version: %+v", curVersionRecord, limitVersion)

	for _, patch := range mgr.patchs {
		if VersionCmp(limitVersion, curVersion) <= 0 {
			curVersionRecord.CurrentVersion = curVersion
			return curVersionRecord, nil
		}

		if patch.PatchInterface.NeedToSkip(kind) {
			logger.Infof("current product kind[%s], skip execute patch[%s]", kind, patch.Version)
			continue
		}

		if VersionCmp(patch.Version, curVersion) <= 0 {
			logger.Infof("current vision[%s], skip execute patch[%s]", curVersion, patch.Version)
			continue
		}

		if err := patch.PatchInterface.PatchFunc(context.Background(), mgr.viper, mgr.smgr); err != nil {
			logger.Errorf("execute patch[%s] failed, %+v", patch.Version, err)
			return nil, err
		}
		curVersionRecord.Operator = operator

		sd, err := mgr.smgr.ShardingDB(dbsharding.BSCPDBKEY)
		if err != nil {
			logger.Errorf("execute patch[%s] done, but can not handle database, %+v", patch.Version, err)
			return nil, err
		}

		ups := map[string]interface{}{
			"CurrentVersion": patch.Version,
			"Operator":       operator,
		}

		exec := sd.DB().
			Model(&database.System{}).
			Where(&database.System{CurrentVersion: curVersion, Kind: kind}).
			Updates(ups)

		if err := exec.Error; err != nil {
			logger.Errorf("execute patch[%s] done, but can not handle database, %+v", patch.Version, err)
			return nil, err
		}
		if exec.RowsAffected == 0 {
			logger.Errorf("execute patch[%s] done, but can not handle database, no affected rows", patch.Version)
			return nil, errors.New("update system version failed, no affected rows")
		}

		// one patch executed.
		curVersion = patch.Version
		logger.Infof("execute and update patch version success, patch[%s]", patch.Version)
	}

	// return newest current version record.
	curVersionRecord.CurrentVersion = curVersion

	return curVersionRecord, nil
}

// GetCurrentVersion returns current system version info.
func (mgr *PatchManager) GetCurrentVersion() (*database.System, error) {
	sd, err := mgr.smgr.ShardingDB(dbsharding.BSCPDBKEY)
	if err != nil {
		logger.Errorf("get current version failed, can not handle database, %+v", err)
		return nil, err
	}

	st := &database.System{}

	err = sd.DB().Last(st).Error
	if err != nil {
		logger.Errorf("get current version failed, %+v", err)
		return nil, err
	}

	return st, nil
}
