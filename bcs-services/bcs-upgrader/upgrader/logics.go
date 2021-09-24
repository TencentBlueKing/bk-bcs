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
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// RunUpgrade upgrade the db data to newest version
func RunUpgrade(ctx context.Context, helper *Helper) (
	currentVersion string, finishedUpgrades []string, err error) {

	sort.Slice(upgradePool, func(i, j int) bool {
		return CompareVersion(upgradePool[i].version, upgradePool[j].version) < 0
	})

	bcsVersion, err := getVersion(ctx, helper.DB)
	if err != nil {
		return "", nil, fmt.Errorf("getVersion failed, err: %s", err)
	}
	bcsVersion.Edition = version.BcsEdition
	currentVersion = bcsVersion.CurrentVersion
	lastVersion := ""
	finishedUpgrades = make([]string, 0)
	blog.Infof("upgradePool:%#v", upgradePool)
	for _, v := range upgradePool {
		lastVersion = v.version
		if CompareVersion(v.version, currentVersion) <= 0 {
			blog.Infof(`current version is "%s", skip upgrade "%s"`, currentVersion, v.version)
			continue
		}
		blog.Infof(`upgrade version %s`, v.version)
		err = v.do(ctx, helper)
		if err != nil {
			blog.Errorf("upgrade version %s error: %s", v.version, err)
			return currentVersion, finishedUpgrades, fmt.Errorf("upgrade version %s failed, err: %s",
				v.version, err)
		}
		bcsVersion.CurrentVersion = v.version
		err = saveVersion(ctx, helper.DB, bcsVersion)
		if err != nil {
			blog.Errorf("save version %s error: %s", v.version, err)
			return currentVersion, finishedUpgrades, fmt.Errorf("saveVersion failed, err: %s", err)
		}
		finishedUpgrades = append(finishedUpgrades, v.version)
		blog.Infof("upgrade to version %s success", v.version)
	}

	if "" == bcsVersion.PreVersion {
		bcsVersion.PreVersion = lastVersion
		if err := saveVersion(ctx, helper.DB, bcsVersion); err != nil {
			return currentVersion, finishedUpgrades, fmt.Errorf("saveVersion failed, err: %s", err)
		}
	}

	return currentVersion, finishedUpgrades, nil
}

func getVersion(ctx context.Context, db drivers.DB) (*VersionInfo, error) {
	data := new(VersionInfo)
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"type": SystemTypeVersion,
	})
	if err := db.Table(UpgraderTableName).Find(cond).One(ctx, data); err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			info := &VersionInfo{
				Type:           SystemTypeVersion,
				CurrentVersion: InitialVersion,
				LastTime:       time.Now(),
			}
			docs := []interface{}{info}
			if _, err = db.Table(UpgraderTableName).Insert(ctx, docs); err != nil {
				return nil, err
			}
			return info, nil
		}
		blog.Errorf("get system version error, err:%s", err)
		return nil, err
	}

	return data, nil
}

func saveVersion(ctx context.Context, db drivers.DB, info *VersionInfo) error {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"type": SystemTypeVersion,
	})
	info.LastTime = time.Now()
	return db.Table(UpgraderTableName).Update(ctx, cond, operator.M{"$set": info})
}

// VersionInfo is bcs version info
type VersionInfo struct {
	Type           string    `bson:"type"`
	PreVersion     string    `bson:"pre_version"`
	CurrentVersion string    `bson:"current_version"`
	Edition        string    `bson:"edition"`
	LastTime       time.Time `bson:"last_time"`
}
