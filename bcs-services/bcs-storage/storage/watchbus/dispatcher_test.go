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

package watchbus

import (
	"context"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
)

// TestWatchBus test watch bus
func TestWatchBus(t *testing.T) {
	opt := &mongo.Options{
		Database:              "test",
		Username:              "mongouser",
		Password:              "abstractmj",
		ConnectTimeoutSeconds: 3,
		MaxPoolSize:           10,
		MinPoolSize:           2,
		Hosts:                 []string{"configure.test.hosts:27018"},
	}

	db, err := mongo.NewDB(opt)
	if err != nil {
		t.Errorf("create db client err %s", err.Error())
	}

	if err := db.Ping(); err != nil {
		t.Errorf("ping db err %s", err.Error())
	}

	eb := NewEventBus(db)

	go func() {
		ch := make(chan *drivers.WatchEvent, 100)
		err := eb.Subscribe("test", "1", ch)
		if err != nil {
			t.Errorf("%+v", err)
		}
		for {
			select {
			case data := <-ch:
				t.Logf("1: recv %+v", data)
			}
		}
	}()

	go func() {
		ch := make(chan *drivers.WatchEvent, 100)
		err := eb.Subscribe("test", "2", ch)
		if err != nil {
			t.Errorf("%+v", err)
		}
		for {
			select {
			case data := <-ch:
				t.Logf("2: recv %+v", data)
			}
		}
	}()

	time.Sleep(1 * time.Second)

	insertNum, err := db.Table("test").Insert(context.TODO(), []interface{}{map[string]interface{}{
		"name":      "test3",
		"namespace": "ns1",
		"content": map[string]interface{}{
			"haahah": "hahaha",
		},
	}})
	if err != nil {
		t.Errorf("%+v", err)
	}
	t.Logf("insert %d", insertNum)

	time.Sleep(1 * time.Millisecond)

	insertNum, err = db.Table("test").Insert(context.TODO(), []interface{}{map[string]interface{}{
		"name":      "test5",
		"namespace": "ns1",
		"content": map[string]interface{}{
			"haahah": "hahaha",
		},
	}})
	if err != nil {
		t.Errorf("%+v", err)
	}
	t.Logf("insert %d", insertNum)

	time.Sleep(1 * time.Second)
	t.Error()
}
