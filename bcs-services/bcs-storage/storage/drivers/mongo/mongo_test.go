// /*
//  * Tencent is pleased to support the open source community by making Blueking Container Service available.,
//  * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
//  * Licensed under the MIT License (the "License"); you may not use this file except
//  * in compliance with the License. You may obtain a copy of the License at
//  * http://opensource.org/licenses/MIT
//  * Unless required by applicable law or agreed to in writing, software distributed under,
//  * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  * either express or implied. See the License for the specific language governing permissions and
//  * limitations under the License.
//  */

package mongo

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
)

func TestMongoGet(t *testing.T) {
	opt := &Options{
		Database:              "test",
		Username:              "mongouser",
		Password:              "abstractmj",
		ConnectTimeoutSeconds: 3,
		MaxPoolSize:           10,
		MinPoolSize:           2,
		Hosts:                 []string{"9.134.65.243:27018"},
	}

	db, err := NewDB(opt)
	if err != nil {
		t.Errorf("create db client err %s", err.Error())
	}

	if err := db.Ping(); err != nil {
		t.Errorf("ping db err %s", err.Error())
	}

	ret := make(map[string]interface{})
	if err := db.Table("Dynamic").Find(operator.NewLeafCondition(operator.Eq, operator.M{
		types.TagResourceName: "test1",
		types.TagNamespace:    "ns1",
		types.TagClusterID:    "BCS-TEST-00001",
	})).One(context.TODO(), &ret); err != nil {
		t.Errorf("find one data failed, err %s", err.Error())
	}
	bytes, _ := json.Marshal(ret)
	obj := &types.RawObject{}
	json.Unmarshal(bytes, obj)
	nbytes, _ := json.Marshal(obj)
	t.Logf("%s", string(nbytes))
	t.Error()
}

func TestMongo(t *testing.T) {
	opt := &Options{
		Database:              "test",
		Username:              "mongouser",
		Password:              "abstractmj",
		ConnectTimeoutSeconds: 3,
		MaxPoolSize:           10,
		MinPoolSize:           2,
		Hosts:                 []string{"9.134.65.243:27018"},
	}

	db, err := NewDB(opt)
	if err != nil {
		t.Errorf("create db client err %s", err.Error())
	}

	if err := db.Ping(); err != nil {
		t.Errorf("ping db err %s", err.Error())
	}

	if err := db.CreateTable(context.TODO(), "collection1"); err != nil {
		t.Errorf("create collection failed, err %s", err.Error())
	}

	hasTable, err := db.HasTable(context.TODO(), "collection1")
	if err != nil {
		t.Errorf("judge collection failed, err %s", err.Error())

	}
	if !hasTable {
		t.Errorf("created table not found")
	}

	tableNames, err := db.ListTableNames(context.TODO())
	if err != nil {
		t.Errorf("list table names failed, err %s", err.Error())
	}
	isFound := false
	for _, tableName := range tableNames {
		if tableName == "collection1" {
			isFound = true
		}
	}
	if !isFound {
		t.Errorf("table name not found")
	}

	newIndex := drivers.Index{
		Name: "index-1",
		Key: map[string]int32{
			"name":      1,
			"namespace": 1,
		},
		Unique: true,
	}
	if err := db.Table("collection1").CreateIndex(context.TODO(), newIndex); err != nil {
		t.Errorf("create index %+v failed, err %s", newIndex, err.Error())
	}

	hasIndex, err := db.Table("collection1").HasIndex(context.TODO(), "index-1")
	if err != nil {
		t.Errorf("judge index failed, err %s", err.Error())
	}
	if !hasIndex {
		t.Errorf("create index not found")
	}

	indexes, err := db.Table("collection1").Indexes(context.TODO())
	if err != nil {
		t.Errorf("list indexes failed, err %s", err.Error())
	}
	if len(indexes) != 2 {
		t.Errorf("invalid indexes %+v", indexes)
	}
	t.Logf("indexes %+v", indexes)

	if err := db.Table("collection1").Insert(context.TODO(), []interface{}{map[string]interface{}{
		"name":      "test1",
		"namespace": "ns1",
		"content": map[string]interface{}{
			"haahah": "hahaha",
		},
	},
	}); err != nil {
		t.Errorf("create doc failed, err %s", err.Error())
	}

	ret := make(map[string]interface{})
	if err := db.Table("collection1").Find(operator.EmptyCondition).One(context.TODO(), &ret); err != nil {
		t.Errorf("find one data failed, err %s", err.Error())
	}
	t.Logf("%+v", ret)

	retArr := make([]map[string]interface{}, 0)
	if err := db.Table("collection1").Find(operator.EmptyCondition).All(context.TODO(), &retArr); err != nil {
		t.Errorf("find many data failed, err %s", err.Error())
	}
	t.Logf("%+v", retArr)

	if err := db.Table("collection1").Update(
		context.TODO(),
		operator.NewLeafCondition(operator.Eq, operator.M{"name": "test1"}),
		map[string]interface{}{"$set": map[string]interface{}{
			"name":      "test1",
			"namespace": "ns1",
			"content": map[string]interface{}{
				"aaaaa": "aaaaaa",
			}}}); err != nil {
		t.Errorf("update failed, err %s", err.Error())
	}

	if err := db.Table("collection1").Find(operator.EmptyCondition).One(context.TODO(), &ret); err != nil {
		t.Errorf("find one data failed, err %s", err.Error())
	}
	t.Logf("%+v", ret)

	if err := db.Table("collection1").DropIndex(context.TODO(), "index-1"); err != nil {
		t.Errorf("drop index index-1 failed, err %s", err.Error())
	}

	hasIndex, err = db.Table("collection1").HasIndex(context.TODO(), "index-1")
	if err != nil {
		t.Errorf("judge index failed, err %s", err.Error())
	}
	if hasIndex {
		t.Errorf("drop index still found")
	}

	if err := db.DropTable(context.TODO(), "collection1"); err != nil {
		t.Errorf("drop collection failed, err %s", err.Error())
	}
	t.Error()
}

func TestMongoWatch(t *testing.T) {
	opt := &Options{
		Database:              "test",
		Username:              "mongouser",
		Password:              "abstractmj",
		ConnectTimeoutSeconds: 3,
		MaxPoolSize:           10,
		MinPoolSize:           2,
		Hosts:                 []string{"9.134.65.243:27018"},
	}

	db, err := NewDB(opt)
	if err != nil {
		t.Errorf("create db client err %s", err.Error())
	}

	if err := db.Ping(); err != nil {
		t.Errorf("ping db err %s", err.Error())
	}

	// newIndex := drivers.Index{
	// 	Name: "index-1",
	// 	Key: map[string]int32{
	// 		"name":      1,
	// 		"namespace": 1,
	// 	},
	// 	Unique: true,
	// }
	// if err := db.Table("collection1").CreateIndex(context.TODO(), newIndex); err != nil {
	// 	t.Errorf("create index %+v failed, err %s", newIndex, err.Error())
	// }

	mCtx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		// eqCondition := drivers.NewLeafCondition(
		// 	drivers.Eq,
		// 	drivers.M{
		// 		"fullDocument.name": "test1",
		// 	})
		// condition := drivers.NewBranchCondition(drivers.Mat,
		// eqCondition)
		ch, err := db.Table("collection1").Watch(
			[]*operator.Condition{
				//condition,
			}).
			WithBatchSize(10).
			WithFullContent(false).
			WithMaxAwaitTime(10 * time.Second).
			DoWatch(mCtx)
		if err != nil {
			t.Errorf("watch failed, err %s", err.Error())
		}
		for {
			select {
			case event, ok := <-ch:
				if !ok {
					t.Logf("channel broken")
					return
				}
				t.Logf("received %+v", event)
			case <-mCtx.Done():
				t.Logf("context is done")
				return
			}
		}
	}()

	// if err := db.Table("collection1").Insert(context.TODO(), []interface{}{map[string]interface{}{
	// 	"name":      "test3",
	// 	"namespace": "ns1",
	// 	"content": map[string]interface{}{
	// 		"haahah": "hahaha",
	// 	},
	// },
	// }); err != nil {
	// 	t.Errorf("create doc failed, err %s", err.Error())
	// }

	time.Sleep(2 * time.Second)

	if err := db.Table("collection1").Update(
		context.TODO(),
		operator.NewLeafCondition(operator.Eq, operator.M{"name": "test1"}),
		map[string]interface{}{"$set": map[string]interface{}{
			"name":      "test1",
			"namespace": "ns1",
			"content": map[string]interface{}{
				"cccccc": time.Now().String(),
			}}}); err != nil {
		t.Errorf("update failed, err %s", err.Error())
	}

	time.Sleep(2 * time.Second)
	cancel()
	wg.Wait()
	t.Error()
}
