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
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type scope struct {
	err       error
	tank      *mongoTank
	operation operator.OperationType

	changeInfo  *operator.ChangeInfo
	value       []interface{}
	isRecovered bool
	length      int
	query       *mgo.Query
	iter        *mgo.Iter
}

func (s *scope) clone() *scope {
	ns := &scope{
		tank:       s.tank,
		operation:  s.operation,
		changeInfo: s.changeInfo,
		value:      s.value,
	}
	if ns.operation == "" {
		ns.operation = operator.None
	}
	return ns
}

// Do the actual operation to mongodb
func (s *scope) do() *scope {
	started := time.Now()
	defer func() {
		if s.err != nil {
			reportMongdbMetrics(string(s.operation), "FAILURE", started)
		} else {
			reportMongdbMetrics(string(s.operation), "SUCCESS", started)
		}
	}()
	switch s.operation {
	case operator.None:
	case operator.Query:
		s.doQuery()
	case operator.Insert:
		s.doInsert()
	case operator.Upsert:
		s.doUpsert()
	case operator.Update:
		s.doUpdate(false)
	case operator.UpdateAll:
		s.doUpdate(true)
	case operator.Remove:
		s.doRemove(false)
	case operator.RemoveAll:
		s.doRemove(true)
	case operator.Count:
		s.doCount()
	case operator.Tables:
		s.doTables()
	case operator.Databases:
		s.doDatabases()
	case operator.Tail:
		s.doTail()
	case operator.GetTableV:
		s.err = storageErr.GetTableVNotSupported
	case operator.SetTableV:
		s.err = storageErr.SetTableVNotSupported
	default:
		s.err = storageErr.UnknownOperationType
	}
	return s
}

// Do the count action, save result to scope.length and scope.err
func (s *scope) doCount() {
	if s.doFilter(); s.err != nil {
		return
	}
	s.length, s.err = s.query.Count()
}

// Do the query action, save result to scope.value and scope.err
func (s *scope) doQuery() {
	if s.doFilter(); s.err != nil {
		return
	}
	if s.tank.search.distinct == "" {
		s.err = s.query.All(&(s.value))
	} else {
		s.err = s.query.Distinct(s.tank.search.distinct, &(s.value))
	}
	s.length = len(s.value)
}

// Do the tail action, save mgo.Iter to scope.iter for return
func (s *scope) doTail() {
	if s.doFilter(); s.err != nil {
		return
	}
	s.iter = s.query.Sort("$natural").Tail(-1)
}

// Do the filter action, save mgo.Query to scope.query for doCount and doQuery
func (s *scope) doFilter() {
	if s.tank.collection == nil {
		s.err = storageErr.MongodbCollectionNoFound
		return
	}
	rawCond := s.tank.search.getRawCond()

	query := s.tank.collection.Find(rawCond)
	if order := s.tank.search.orders; order != nil {
		order = append(order, "_id")
		query.Sort(order...)
	}
	query.Skip(s.tank.search.offset)
	if s.tank.search.limit > 0 {
		query.Limit(s.tank.search.limit)
	}
	if s.tank.search.selector != nil {
		query.Select(s.tank.search.selector)
	}
	s.query = query
}

// Do the insert action
func (s *scope) doInsert() {
	if err := s.ensureIndex(); err != nil {
		s.err = err
		return
	}
	s.err = s.tank.collection.Insert(s.tank.data...)
}

// Do the upsert action
func (s *scope) doUpsert() {
	if err := s.ensureIndex(); err != nil {
		s.err = err
		return
	}

	rawCond := s.tank.search.getRawCond()
	s.changeInfo = &operator.ChangeInfo{}
	data := bson.M{"$set": s.tank.data[0]}
	var info *mgo.ChangeInfo
	info, s.err = s.tank.collection.Upsert(rawCond, data)
	if s.err == nil {
		s.changeInfo = &operator.ChangeInfo{Updated: info.Updated, Matched: info.Matched}
	}
}

// Do the update action
func (s *scope) doUpdate(all bool) {
	if err := s.ensureIndex(); err != nil {
		s.err = err
		return
	}
	rawCond := s.tank.search.getRawCond()
	s.changeInfo = &operator.ChangeInfo{}
	data := bson.M{"$set": s.tank.data[0]}

	if all {
		var info *mgo.ChangeInfo
		info, s.err = s.tank.collection.UpdateAll(rawCond, data)
		if s.err == nil {
			s.changeInfo = &operator.ChangeInfo{Updated: info.Updated, Matched: info.Matched}
		}
		return
	}
	s.err = s.tank.collection.Update(rawCond, data)
	if s.err == nil {
		s.changeInfo = &operator.ChangeInfo{Updated: 1, Matched: 1}

		// No found can be known by changeInfo, make it no-error
	} else if s.err.Error() == "not found" {
		s.err = nil
	}
}

// Do the remove action
func (s *scope) doRemove(all bool) {
	if s.tank.collection == nil {
		s.err = storageErr.MongodbCollectionNoFound
		return
	}
	rawCond := s.tank.search.getRawCond()
	s.changeInfo = &operator.ChangeInfo{}

	if all {
		var info *mgo.ChangeInfo
		info, s.err = s.tank.collection.RemoveAll(rawCond)
		if s.err == nil {
			s.changeInfo = &operator.ChangeInfo{Removed: info.Removed, Matched: info.Matched}
		}
		return
	}
	s.err = s.tank.collection.Remove(rawCond)
	if s.err == nil {
		s.changeInfo = &operator.ChangeInfo{Removed: 1, Matched: 1}
	}
}

// Do the tables action, list all tables
func (s *scope) doTables() {
	if s.tank.database == nil {
		s.err = storageErr.MongodbDatabasesNoFound
		return
	}
	value, err := s.tank.database.CollectionNames()
	iValue := make([]interface{}, 0, len(value))
	for _, v := range value {
		iValue = append(iValue, v)
	}

	// each db contains "system.indexes", and this will be screened out
	sand := "system.indexes"
	for i, key := range iValue {
		if key == sand {
			iValue = append(iValue[:i], iValue[i+1:]...)
			break
		}
	}
	s.value = iValue
	s.length = len(s.value)
	s.err = err
}

// Do the databases action, list all dbs
func (s *scope) doDatabases() {
	value, err := s.tank.session.DatabaseNames()
	iValue := make([]interface{}, 0, len(value))
	for _, v := range value {
		iValue = append(iValue, v)
	}
	s.value = iValue
	s.length = len(s.value)
	s.err = err
}

func (s *scope) ensureIndex() error {
	if s.tank.collection == nil {
		return storageErr.MongodbCollectionNoFound
	}
	if s.tank.index != nil && len(s.tank.index) != 0 {
		return s.tank.collection.EnsureIndex(mgo.Index{
			Key:        s.tank.index,
			Unique:     true,
			DropDups:   true,
			Background: false,
			Sparse:     true,
		})
	}
	return nil
}
