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
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/samuel/go-zookeeper/zk"
)

type originDriver struct {
	m    sync.Mutex
	pool *zkclient.ZkClient

	// default settings
	root    string
	timeOut time.Duration
}

func (od *originDriver) copy() *zkclient.ZkClient {
	od.m.Lock()
	if od.pool.ZkConn == nil || od.pool.ZkConn.State() < zk.StateConnected {
		od.pool.Close()
		newDriver := zkclient.NewZkClient(od.pool.ZkHost)
		newDriver.ConnectEx(od.timeOut)
		od.pool = newDriver
	}
	od.m.Unlock()
	return od.pool
}

var driverPool map[string]*originDriver

//RegisterZkTank register zookeeper operation unit
func RegisterZkTank(name string, info *operator.DBInfo) (err error) {
	if driverPool == nil {
		driverPool = make(map[string]*originDriver, 10)
	}
	if _, ok := driverPool[name]; ok {
		err := storageErr.ZookeeperDriverAlreadyInPool
		blog.Errorf("%v: %s", err, name)
	}

	driver := new(originDriver)
	driver.m = sync.Mutex{}
	driver.root = info.Database
	driver.timeOut = info.ConnectTimeout
	client := zkclient.NewZkClient(info.Addr)
	if err = client.ConnectEx(info.ConnectTimeout); err != nil {
		return err
	}
	driver.pool = client
	driverPool[name] = driver
	return nil
}

//NewZkTank create zookeeper operation unit
func NewZkTank(name string) operator.Tank {
	tank := &zkTank{}
	if tank.err = tank.init(name); tank.err != nil {
		blog.Errorf("Init zookeeper tank failed. %v", tank.err)
	}
	return tank
}

type zkTank struct {
	isInit        bool
	hasChild      bool
	isTransaction bool
	name          string

	acl    []zk.ACL
	driver *originDriver
	client *zkclient.ZkClient
	root   string
	node   string
	search *search
	scope  *scope

	data   []operator.M
	tableD interface{}
	err    error
}

func (zt *zkTank) init(name string) error {
	zt.isInit = false
	zt.name = name
	var ok bool
	if zt.driver, ok = driverPool[name]; !ok {
		err := storageErr.ZookeeperDriverNotExist
		blog.Errorf("%v: %s", err, name)
		return err
	}
	zt.acl = zk.DigestACL(zk.PermAll, zkclient.AUTH_USER, zkclient.AUTH_PWD)
	zt.client = zt.driver.copy()
	zt.search = (&search{tank: zt}).clone()
	zt.scope = (&scope{tank: zt}).clone()
	zt.isInit = true
	zt.switchRoot(zt.driver.root)
	return nil
}

func (zt *zkTank) clone() *zkTank {
	if zt.hasChild && zt.isTransaction {
		zt.err = storageErr.TransactionChainBreak
	}
	tank := &zkTank{
		isInit:        zt.isInit,
		name:          zt.name,
		isTransaction: zt.isTransaction,

		acl:    zt.acl,
		driver: zt.driver,
		client: zt.client,
		root:   zt.root,
		node:   zt.node,
	}
	return tank
}

func (zt *zkTank) rootPath() string {
	return "/" + zt.root
}

func (zt *zkTank) nodePath() string {
	return zt.rootPath() + "/" + zt.node
}

func (zt *zkTank) childPath(name string) string {
	return zt.nodePath() + "/" + name
}

func (zt *zkTank) switchRoot(name string) *zkTank {
	if zt.isInit {
		zt.root = name
		zt.client.CheckMulNode("/"+name, nil)
		return zt
	}
	zt.err = storageErr.ZookeeperTankNotInit
	return zt
}

func (zt *zkTank) switchTable(name string) *zkTank {
	if zt.isInit {
		zt.node = name
		zt.client.CheckMulNode(zt.nodePath(), nil)
	}
	zt.err = storageErr.ZookeeperTankNotInit
	return zt
}

func (zt *zkTank) newScope(op operator.OperationType) *scope {
	s := &scope{
		operation: op,
		tank:      zt,
	}
	zt.scope = s
	if !zt.isTransaction {
		zt.scope.do()
	}
	return s
}

func (zt *zkTank) setData(data ...operator.M) *zkTank {
	for _, singleD := range data {
		for k, v := range singleD {
			r, err := getByte(v)
			if err != nil {
				zt.err = err
				return zt
			}
			singleD[k] = r
		}
	}
	zt.data = data
	return zt
}

func (zt *zkTank) setTableD(data interface{}) *zkTank {
	r, err := getByte(data)
	if err != nil {
		zt.err = err
		return zt
	}
	zt.tableD = r
	return zt
}

func (zt *zkTank) Close() {
	// zk use one client no need to close
}

// GetValue Tank implementation
func (zt *zkTank) GetValue() []interface{} {
	if zt.scope.value == nil {
		zt.scope.value = []interface{}{}
	}
	return zt.scope.value
}

// GetLen Tank implementation
func (zt *zkTank) GetLen() int {
	return zt.scope.length
}

// GetError Tank implementation
func (zt *zkTank) GetError() error {
	if zt.err != nil {
		return zt.err
	}
	return zt.scope.err
}

// GetChangeInfo Tank implementation
func (zt *zkTank) GetChangeInfo() *operator.ChangeInfo {
	return zt.scope.changeInfo
}

// Databases Tank implementation
func (zt *zkTank) Databases() operator.Tank {
	return zt.clone().newScope(operator.Databases).tank
}

// Using Tank implementation
func (zt *zkTank) Using(name string) operator.Tank {
	return zt.clone().switchRoot(name)
}

// Tables Tank implementation
func (zt *zkTank) Tables() operator.Tank {
	return zt.clone().newScope(operator.Tables).tank
}

// SetTableV Tank implementation
func (zt *zkTank) SetTableV(data interface{}) operator.Tank {
	return zt.clone().setTableD(data).newScope(operator.SetTableV).tank
}

// GetTableV Tank implementation
func (zt *zkTank) GetTableV() operator.Tank {
	return zt.clone().newScope(operator.GetTableV).tank
}

// From Tank implementation NOT INVOLVED
func (zt *zkTank) From(name string) operator.Tank {
	return zt.clone().switchTable(name)
}

// Distinct Tank implementation
func (zt *zkTank) Distinct(key string) operator.Tank {
	return zt.clone().search.setDistinct(key).tank
}

// OrderBy Tank implementation
func (zt *zkTank) OrderBy(key ...string) operator.Tank {
	return zt.clone().search.setOrder(key...).tank
}

// Select Tank implementation
func (zt *zkTank) Select(key ...string) operator.Tank {
	return zt.clone().search.setSelector(key...).tank
}

// Offset Tank implementation
func (zt *zkTank) Offset(n int) operator.Tank {
	return zt.clone().search.setOffset(n).tank
}

// Limit Tank implementation
func (zt *zkTank) Limit(n int) operator.Tank {
	return zt.clone().search.setLimit(n).tank
}

// Index Tank implementation
func (zt *zkTank) Index(key ...string) operator.Tank {
	return zt.clone()
}

// Filter Tank implementation
func (zt *zkTank) Filter(cond *operator.Condition, args ...interface{}) operator.Tank {
	return zt.clone().search.combineCondition(cond).tank
}

// Count Tank implementation
func (zt *zkTank) Count() operator.Tank {
	return zt.clone().newScope(operator.Count).tank
}

// Query Tank implementation
func (zt *zkTank) Query(args ...interface{}) operator.Tank {
	return zt.clone().newScope(operator.Query).tank
}

// Insert Tank implementation
func (zt *zkTank) Insert(data ...operator.M) operator.Tank {
	return zt.clone().setData(data...).newScope(operator.Insert).tank
}

// Upsert Tank implementation
func (zt *zkTank) Upsert(data operator.M, args ...interface{}) operator.Tank {
	return zt.clone().setData(data).newScope(operator.Upsert).tank
}

// Update Tank implementation
func (zt *zkTank) Update(data operator.M, args ...interface{}) operator.Tank {
	return zt.clone().setData(data).newScope(operator.Update).tank
}

// UpdateAll Tank implementation
func (zt *zkTank) UpdateAll(data operator.M, args ...interface{}) operator.Tank {
	return zt.clone().setData(data).newScope(operator.UpdateAll).tank
}

// Remove Tank implementation
func (zt *zkTank) Remove(args ...interface{}) operator.Tank {
	return zt.clone().newScope(operator.Remove).tank
}

// RemoveAll Tank implementation
func (zt *zkTank) RemoveAll(args ...interface{}) operator.Tank {
	return zt.clone().newScope(operator.RemoveAll).tank
}

// Watch Tank implementation
func (zt *zkTank) Watch(opts *operator.WatchOptions) (chan *operator.Event, context.CancelFunc) {
	return nil, nil
}

func getByte(v interface{}) (r []byte, err error) {
	switch v.(type) {
	case string:
		r = []byte(v.(string))
	case []byte:
		r = v.([]byte)
	case int:
		r = []byte(strconv.Itoa(v.(int)))
	case int64:
		r = []byte(strconv.FormatInt(v.(int64), 10))
	case time.Time:
		r = []byte(v.(time.Time).Format("2006-01-02 15:04:05"))
	case map[string]interface{}:
		err = codec.EncJson(v, &r)
	case []interface{}:
		err = codec.EncJson(v, &r)
	case interface{}:
		err = codec.EncJson(v, &r)
	}
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("%v", err)
		}
	}()
	return
}
