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

package zookeeper

import (
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

type node struct {
	children []string
	version  int32
}

func (n *node) hasChild(name string) bool {
	for _, v := range n.children {
		if name == v {
			return true
		}
	}
	return false
}

type scope struct {
	err       error
	tank      *zkTank
	operation operator.OperationType

	changeInfo *operator.ChangeInfo
	value      []interface{}
	length     int
	targetNode *node
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
			reportZKMetrics(string(s.operation), "FAILURE", started)
		} else {
			reportZKMetrics(string(s.operation), "SUCCESS", started)
		}
	}()
	switch s.operation {
	case operator.None:
	case operator.Query:
		s.doQuery()
	case operator.Insert:
		s.doInsert()
	case operator.Upsert:
		s.doUpdate()
	case operator.Update:
		s.doUpdate()
	case operator.UpdateAll:
		s.doUpdate()
	case operator.Remove:
		s.doRemove()
	case operator.RemoveAll:
		s.doRemove()
	case operator.Count:
		s.doCount()
	case operator.Tables:
		s.doTables()
	case operator.Databases:
		s.doDatabases()
	case operator.GetTableV:
		s.doGetTableV()
	case operator.SetTableV:
		s.doSetTableV()
	default:
		s.err = storageErr.UnknownOperationType
	}
	return s
}

func (s *scope) inspectNode(path string) {
	if s.tank.client == nil {
		s.err = storageErr.ZookeeperClientNoFound
		return
	}
	children, stat, err := s.tank.client.GetChildrenEx(path)
	if s.err = err; s.err != nil {
		return
	}
	s.targetNode = &node{
		children: children,
		version:  stat.Version,
	}
	s.err = err
}

func (s *scope) doFilter() {
	s.inspectNode(s.tank.nodePath())
}

func (s *scope) doCount() {
	if s.doFilter(); s.err != nil {
		return
	}
	if len(s.targetNode.children) == 0 {
		s.length = 0
	} else {
		s.length = 1
	}
}

func (s *scope) doQuery() {
	if s.doCount(); s.err != nil {
		return
	}

	var value string
	var err error
	r := make(operator.M)
	for _, child := range s.targetNode.children {
		if value, err = s.tank.client.Get(s.tank.childPath(child)); err != nil {
			s.err = err
			return
		}

		r[child] = value
	}
	s.value = []interface{}{r}
}

func (s *scope) doInsert() {
	if s.doFilter(); s.err != nil {
		return
	}
	length := len(s.tank.data)
	if length == 0 {
		return
	}
	data := s.tank.data[0]
	queue := make([]interface{}, 0, len(data))
	for k, v := range data {
		var request interface{}
		if s.targetNode.hasChild(k) {
			request = &zk.SetDataRequest{Path: s.tank.childPath(k), Data: v.([]byte), Version: -1}
		} else {
			request = &zk.CreateRequest{Path: s.tank.childPath(k), Data: v.([]byte), Acl: s.tank.acl}
		}
		queue = append(queue, request)
	}
	_, s.err = s.tank.client.ZkConn.Multi(queue...)
}

func (s *scope) doUpdate() {
	s.doInsert()

	var updated int
	if s.err == nil {
		updated = 1
	} else {
		updated = 0
	}
	s.changeInfo = &operator.ChangeInfo{
		Matched: 1,
		Updated: updated,
	}
}

func (s *scope) doRemove() {
	if s.doFilter(); s.err != nil {
		return
	}
	s.err = s.tank.client.Del(s.tank.nodePath(), s.targetNode.version)

	var removed int
	if s.err == nil {
		removed = 1
	} else {
		removed = 0
	}
	s.changeInfo = &operator.ChangeInfo{
		Matched: 1,
		Removed: removed,
	}
}

func (s *scope) getChildren(path string) []interface{} {
	if s.inspectNode(path); s.err != nil {
		return nil
	}

	value := make([]interface{}, 0, len(s.targetNode.children))
	for _, v := range s.targetNode.children {
		value = append(value, v)
	}
	return value
}

// list the table node children
func (s *scope) doTables() {
	s.value = s.getChildren(s.tank.nodePath())
}

// list the root children
func (s *scope) doDatabases() {
	s.value = s.getChildren(s.tank.rootPath())
}

func (s *scope) doSetTableV() {
	if s.tank.client == nil {
		s.err = storageErr.ZookeeperClientNoFound
		return
	}
	s.err = s.tank.client.Update(s.tank.nodePath(), string(s.tank.tableD.([]byte)))
}

func (s *scope) doGetTableV() {
	if s.tank.client == nil {
		s.err = storageErr.ZookeeperClientNoFound
		return
	}
	value, err := s.tank.client.Get(s.tank.nodePath())
	s.value = []interface{}{value}
	s.err = err
}
