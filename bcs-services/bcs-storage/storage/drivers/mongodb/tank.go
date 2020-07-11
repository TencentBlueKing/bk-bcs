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

package mongodb

import (
	"context"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	mgo "gopkg.in/mgo.v2"
)

type originDriver struct {
	pool *mgo.Session

	// default settings
	database     string
	listenerName string
}

var driverPool map[string]*originDriver

// RegisterMongodbTank register mongodb operation unit
func RegisterMongodbTank(name string, info *operator.DBInfo) (err error) {
	if driverPool == nil {
		driverPool = make(map[string]*originDriver, 10)
	}
	if _, ok := driverPool[name]; ok {
		err := storageErr.MongodbDriverAlreadyInPool
		blog.Errorf("%v: %s", err, name)
		return err
	}

	var mgoInfo *mgo.DialInfo
	if mgoInfo, err = mgo.ParseURL(strings.Join(info.Addr, ",")); err != nil {
		return
	}
	mgoInfo.Timeout = info.ConnectTimeout
	mgoInfo.Database = info.Database
	mgoInfo.Username = info.Username
	mgoInfo.Password = info.Password
	mgoInfo.Source = "admin"

	var session *mgo.Session
	if session, err = mgo.DialWithInfo(mgoInfo); err != nil {
		return
	}
	session.SetMode(mgo.Mode(info.Mode), true)

	driver := new(originDriver)
	driver.database = info.Database
	driver.listenerName = info.ListenerName
	driver.pool = session
	driverPool[name] = driver
	return
}

// NewMongodbTank create mongodb operation unit
func NewMongodbTank(name string) operator.Tank {
	tank := &mongoTank{}
	if tank.err = tank.init(name); tank.err != nil {
		blog.Errorf("Init mongodb tank failed. %v", tank.err)
	}
	return tank
}

type mongoTank struct {
	isInit        bool
	hasChild      bool
	isTransaction bool
	name          string
	listenerName  string

	driver     *originDriver
	session    *mgo.Session
	dbName     string
	cName      string
	database   *mgo.Database
	collection *mgo.Collection
	search     *search
	scope      *scope
	index      []string

	data []interface{}
	err  error
}

func (mt *mongoTank) init(name string) error {
	mt.isInit = false
	mt.name = name
	var ok bool
	if mt.driver, ok = driverPool[name]; !ok {
		err := storageErr.MongodbDriverNotExist
		blog.Errorf("%v: %s", err, name)
		return err
	}
	mt.listenerName = mt.driver.listenerName
	mt.session = mt.driver.pool.Copy()
	mt.search = (&search{tank: mt}).clone()
	mt.scope = (&scope{tank: mt}).clone()
	mt.isInit = true
	mt.dbName = mt.driver.database
	mt.switchDB(mt.dbName)
	return nil
}

func (mt *mongoTank) clone() *mongoTank {
	if mt.hasChild && mt.isTransaction {
		mt.err = storageErr.TransactionChainBreak
	}
	tank := &mongoTank{
		isInit:        mt.isInit,
		name:          mt.name,
		listenerName:  mt.listenerName,
		isTransaction: mt.isTransaction,
		index:         mt.index,

		driver:     mt.driver,
		session:    mt.session,
		database:   mt.database,
		collection: mt.collection,
		dbName:     mt.dbName,
		cName:      mt.cName,
	}
	tank.scope = (&scope{tank: tank}).clone()
	if mt.search == nil {
		tank.search = &search{limit: -1, offset: 0}
	} else {
		tank.search = mt.search.clone()
	}
	tank.search.tank = tank
	return tank
}

func (mt *mongoTank) switchDB(name string) *mongoTank {
	if mt.isInit {
		mt.dbName = name
		mt.database = mt.session.DB(name)
		return mt
	}
	mt.err = storageErr.MongodbTankNotInit
	return mt
}

func (mt *mongoTank) switchCollection(name string) *mongoTank {
	if mt.isInit {
		mt.cName = name
		mt.collection = mt.database.C(name)
		return mt
	}
	mt.err = storageErr.MongodbTankNotInit
	return mt
}

func (mt *mongoTank) newScope(op operator.OperationType) *scope {
	s := &scope{
		operation: op,
		tank:      mt,
	}
	mt.scope = s
	if !mt.isTransaction {
		mt.scope.do()
	}
	return s
}

func (mt *mongoTank) setIndex(key ...string) *mongoTank {
	mt.index = key
	return mt
}

func (mt *mongoTank) setData(data ...operator.M) *mongoTank {
	dataList := make([]interface{}, 0, len(data))
	for _, d := range data {
		tmp := dotHandler(map[string]interface{}(d)).(map[string]interface{})
		dataList = append(dataList, tmp)
	}
	mt.data = dataList
	return mt
}

func (mt *mongoTank) Tail() *mgo.Iter {
	return mt.clone().newScope(operator.Tail).iter
}

// Copy a tank with a new session, a new database and a new collection.
func (mt *mongoTank) Copy() *mongoTank {
	m := &mongoTank{}
	m.init(mt.name)
	return m.switchDB(mt.dbName).switchCollection(mt.cName)
}

// close session after use
func (mt *mongoTank) Close() {
	if mt.session != nil {
		mt.session.Close()
	}
}

// get value from scope, so it must be called after options,
// or will return []interface{}{}
func (mt *mongoTank) GetValue() []interface{} {
	if !mt.scope.isRecovered {
		value := dotRecover(mt.scope.value)
		if value == nil {
			value = []interface{}{}
		}
		mt.scope.value = value
		mt.scope.isRecovered = true
	}
	return mt.scope.value
}

// get the value length, or the Count() value
func (mt *mongoTank) GetLen() int {
	return mt.scope.length
}

// get the changeInfo after update or remove
func (mt *mongoTank) GetChangeInfo() *operator.ChangeInfo {
	return mt.scope.changeInfo
}

// get the last error during the operations
func (mt *mongoTank) GetError() error {
	if mt.err != nil {
		return mt.err
	}
	return mt.scope.err
}

// list databases, like "show dbs"
func (mt *mongoTank) Databases() operator.Tank {
	return mt.clone().newScope(operator.Databases).tank
}

// switch database, like "use db"
func (mt *mongoTank) Using(name string) operator.Tank {
	return mt.clone().switchDB(name)
}

// list collections, should be called after Using()
func (mt *mongoTank) Tables() operator.Tank {
	return mt.clone().newScope(operator.Tables).tank
}

// NOT INVOLVED
func (mt *mongoTank) SetTableV(data interface{}) operator.Tank {
	return mt.clone().newScope(operator.SetTableV).tank
}

// NOT INVOLVED
func (mt *mongoTank) GetTableV() operator.Tank {
	return mt.clone().newScope(operator.GetTableV).tank
}

// switch collection
func (mt *mongoTank) From(name string) operator.Tank {
	return mt.clone().switchCollection(name)
}

// set distinct key, will no reach db until Query() called
func (mt *mongoTank) Distinct(key string) operator.Tank {
	return mt.clone().search.setDistinct(key).tank
}

// OrderBy set order key, will no reach db until Query() called
func (mt *mongoTank) OrderBy(key ...string) operator.Tank {
	return mt.clone().search.setOrder(key...).tank
}

// Select set select key, will no reach db until Query() called
func (mt *mongoTank) Select(key ...string) operator.Tank {
	return mt.clone().search.setSelector(key...).tank
}

// Offset set offset value, will no reach db until Query() called
func (mt *mongoTank) Offset(n int) operator.Tank {
	return mt.clone().search.setOffset(n).tank
}

// Limit set limit value, will no reach db until Query() called
func (mt *mongoTank) Limit(n int) operator.Tank {
	return mt.clone().search.setLimit(n).tank
}

// Index set unique index
func (mt *mongoTank) Index(key ...string) operator.Tank {
	return mt.clone().setIndex(key...)
}

// Filter add condition for filter, multi filter will be combine with AND
func (mt *mongoTank) Filter(cond *operator.Condition, args ...interface{}) operator.Tank {
	return mt.clone().search.combineCondition(cond).tank
}

// Count the data length according to filters before
func (mt *mongoTank) Count() operator.Tank {
	return mt.clone().newScope(operator.Count).tank
}

// Query the value according to filters before
func (mt *mongoTank) Query(args ...interface{}) operator.Tank {
	return mt.clone().newScope(operator.Query).tank
}

// Insert multi value
func (mt *mongoTank) Insert(data ...operator.M) operator.Tank {
	return mt.clone().setData(data...).newScope(operator.Insert).tank
}

// Upsert value according to filters before
func (mt *mongoTank) Upsert(data operator.M, args ...interface{}) operator.Tank {
	return mt.clone().setData(data).newScope(operator.Upsert).tank
}

// Update value to first match according to filters before
func (mt *mongoTank) Update(data operator.M, args ...interface{}) operator.Tank {
	return mt.clone().setData(data).newScope(operator.Update).tank
}

// UpdateAll value to all matches according to filters before
func (mt *mongoTank) UpdateAll(data operator.M, args ...interface{}) operator.Tank {
	return mt.clone().setData(data).newScope(operator.UpdateAll).tank
}

// Remove first match according to filters before
func (mt *mongoTank) Remove(args ...interface{}) operator.Tank {
	return mt.clone().newScope(operator.Remove).tank
}

// Removeall matches according to filters before
func (mt *mongoTank) RemoveAll(args ...interface{}) operator.Tank {
	return mt.clone().newScope(operator.RemoveAll).tank
}

// Watch make a watch to collections and its documents, then return a chan Event.
func (mt *mongoTank) Watch(opts *operator.WatchOptions) (chan *operator.Event, context.CancelFunc) {
	return newWatchHandler(opts, mt).watch()
}
