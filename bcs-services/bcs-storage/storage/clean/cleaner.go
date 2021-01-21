/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clean

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// DBCleaner db cleaner
type DBCleaner struct {
	db            drivers.DB
	checkInterval time.Duration
	tableName     string
	maxEntryNum   int64
	maxDuration   time.Duration
	timeTagName   string
}

// NewDBCleaner create db cleaner
func NewDBCleaner(db drivers.DB, tableName string, checkInterval time.Duration) *DBCleaner {
	return &DBCleaner{
		db:            db,
		tableName:     tableName,
		checkInterval: checkInterval,
	}
}

// WithMaxEntryNum set max entry
func (dbc *DBCleaner) WithMaxEntryNum(num int64) {
	dbc.maxEntryNum = num
}

// WithMaxDuration set max time duration
func (dbc *DBCleaner) WithMaxDuration(maxDuration time.Duration, timeTagName string) {
	dbc.maxDuration = maxDuration
	dbc.timeTagName = timeTagName
}

func (dbc *DBCleaner) doNumClean() error {
	if dbc.maxEntryNum != 0 {
		total, err := dbc.db.Table(dbc.tableName).Find(operator.EmptyCondition).Count(context.TODO())
		if err != nil {
			return fmt.Errorf("count table %s failed, err %s", dbc.tableName, err.Error())
		}
		if total > dbc.maxEntryNum {
			var toDelete operator.M
			if err := dbc.db.Table(dbc.tableName).Find(operator.EmptyCondition).
				WithSort(map[string]interface{}{
					dbc.timeTagName: -1,
				}).WithStart(dbc.maxEntryNum-1).
				WithLimit(1).
				One(context.TODO(), &toDelete); err != nil {
				return fmt.Errorf("find delete edge failed, err %s", err.Error())
			}

			timeObj, ok := toDelete[dbc.timeTagName]
			if !ok {
				return fmt.Errorf("data %+v does not have time tag %s", toDelete, dbc.timeTagName)
			}
			timeEdge, asok := timeObj.(time.Time)
			if !asok {
				return fmt.Errorf("field %+v with time tag %s is not time.Time", timeObj, dbc.timeTagName)
			}
			deleteCounter, err := dbc.db.Table(dbc.tableName).Delete(context.TODO(),
				operator.NewLeafCondition(operator.Lt, operator.M{
					dbc.timeTagName: timeEdge,
				}))
			if err != nil {
				return fmt.Errorf("delete entry with time less than %s", timeEdge.String())
			}
			blog.Infof("cleaned %d entry of table %s", deleteCounter, dbc.tableName)
		}
	}
	return nil
}

func (dbc *DBCleaner) doTimeClean() error {
	if dbc.maxDuration != 0 {
		now := time.Now()
		timeEdge := now.Add(-dbc.maxDuration)
		deleteCounter, err := dbc.db.Table(dbc.tableName).Delete(context.TODO(),
			operator.NewLeafCondition(operator.Lt, operator.M{
				dbc.timeTagName: timeEdge,
			}))
		if err != nil {
			return fmt.Errorf("delete entry with time less than %s", timeEdge.String())
		}
		blog.Infof("cleaned %d entry of table %s", deleteCounter, dbc.tableName)
	}
	return nil
}

// Run run cleaner
func (dbc *DBCleaner) Run(ctx context.Context) {
	ticker := time.NewTicker(dbc.checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := dbc.doNumClean(); err != nil {
				blog.Warnf("do number clean failed, err %s", err.Error())
			}
			if err := dbc.doTimeClean(); err != nil {
				blog.Warnf("do time clean failed, err %s", err.Error())
			}
		case <-ctx.Done():

			return
		}
	}
}
