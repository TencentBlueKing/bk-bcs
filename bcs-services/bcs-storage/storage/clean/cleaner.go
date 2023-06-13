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
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/util"
)

const (
	// softdelete field name
	databaseFieldNameForDeletionFlag = "_isBcsObjectDeleted"
	databaseEvent                    = "event"
	databaseDynamic                  = "dynamic"
	databaseAlarm                    = "alarm"

	tableEvent       = "Event"
	tableEventPrefix = "event_"
)

// DBCleaner db cleaner
type DBCleaner struct {
	db            drivers.DB
	checkInterval time.Duration
	tableName     string
	maxEntryNum   int64
	maxDuration   time.Duration
	sleepDuration time.Duration
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
func (dbc *DBCleaner) WithMaxDuration(maxDuration time.Duration, maxRandomDuration time.Duration, timeTagName string) {
	dbc.maxDuration = maxDuration
	dbc.timeTagName = timeTagName

	// 到达ticker触发时，延迟时间启动删除程序，避免多个cleaner同时启动删除造成高负载
	if maxRandomDuration != time.Duration(0) {
		maxRandomDuration = util.HashString2Time(dbc.tableName, maxRandomDuration)
	}
	dbc.sleepDuration = maxRandomDuration
	blog.Infof("[todelete] set maxRandomDuration to %s for db [%s] table [%s]", maxRandomDuration.String(), dbc.db.DataBase(), dbc.tableName)
}

func (dbc *DBCleaner) doNumClean() error {
	blog.Infof("table(%s) max entry num: %d", dbc.tableName, dbc.maxEntryNum)
	if dbc.maxEntryNum != 0 {
		total, err := dbc.db.Table(dbc.tableName).Find(operator.EmptyCondition).Count(context.TODO())
		if err != nil {
			return fmt.Errorf("count table %s failed, err %s", dbc.tableName, err.Error())
		}
		blog.Infof("table(%s) total entry num: %d", dbc.tableName, total)
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
			blog.Infof("timeTag %s type: %s", dbc.timeTagName, reflect.TypeOf(timeObj))
			timeEdge := time.Time{}

			if timeObjDT, asok := timeObj.(primitive.DateTime); asok {
				timeEdge = timeObjDT.Time()
			} else {
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
	// avoid high concurrency
	time.Sleep(dbc.sleepDuration)

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

func (dbc *DBCleaner) doSoftDeleteClean() error {
	deleteCounter, err := dbc.db.Table(dbc.tableName).Delete(context.TODO(),
		operator.NewLeafCondition(operator.Eq, operator.M{
			databaseFieldNameForDeletionFlag: true,
		}))
	if err != nil {
		return fmt.Errorf("delete entry with deletion flag true")
	}
	blog.Infof("cleaned %d entry of table %s", deleteCounter, dbc.tableName)
	return nil
}

// Run run cleaner
func (dbc *DBCleaner) Run(ctx context.Context) {
	ticker := time.NewTicker(dbc.checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			switch dbc.db.DataBase() {
			case databaseDynamic:
				if err := dbc.doSoftDeleteClean(); err != nil {
					blog.Errorf("do soft delete clean failed, err %s", err.Error())
				}
				if dbc.tableName == tableEvent {
					if err := dbc.doTimeClean(); err != nil {
						blog.Errorf("do time clean failed, err %s", err.Error())
					}
				}
			case databaseAlarm:
				if err := dbc.doNumClean(); err != nil {
					blog.Errorf("do num clean failed, err %s", err.Error())
				}
				if err := dbc.doTimeClean(); err != nil {
					blog.Errorf("do time clean failed, err %s", err.Error())
				}
			case databaseEvent:
				if err := dbc.doTimeClean(); err != nil {
					blog.Errorf("do time clean failed, err %s", err.Error())
				}
			}
		case <-ctx.Done():

			return
		}
	}
}
