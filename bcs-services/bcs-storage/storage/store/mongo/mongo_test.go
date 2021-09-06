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

package mongo

// import (
// 	"context"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/drivers"
// 	mdri "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/drivers/mongo"
// 	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/store"
// 	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
// )

// func createDB() (drivers.DB, error) {
// 	opt := &mdri.Options{
// 		Database:              "test",
// 		Username:              "mongouser",
// 		Password:              "abstractmj",
// 		ConnectTimeoutSeconds: 3,
// 		MaxPoolSize:           10,
// 		MinPoolSize:           2,
// 		Hosts:                 []string{"configure.test.hosts:27018"},
// 	}

// 	db, err := mdri.NewDB(opt)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := db.Ping(); err != nil {
// 		return nil, err
// 	}
// 	return db, nil
// }

// func TestMongoStoreCreate(t *testing.T) {

// 	db, err := createDB()
// 	if err != nil {
// 		t.Errorf("create db failed, err %s", err.Error())
// 	}
// 	storeCli := NewMongoStore(db)
// 	obj := &types.RawObject{
// 		Meta: types.Meta{
// 			Type:      "Dynamic",
// 			ClusterID: "BCS-TEST-00001",
// 			Name:      "test1",
// 			Namespace: "ns1",
// 		},
// 		Data: map[string]interface{}{
// 			"metadata": map[string]interface{}{
// 				"name":              "rso-auth-test6",
// 				"namespace":         "test6",
// 				"creationTimestamp": "0001-01-01T00:00:00Z",
// 				"labels": map[string]interface{}{
// 					"io．tencent．paas．projectid":                 "636db594336748a581391188bae9177b",
// 					"io．tencent．bkdata．container．stdlog．dataid": "9915",
// 					"Owner":                            "accounts-team-rg@riotgames.com",
// 					"io．tencent．bcs．cluster":           "BCS-DEBUGSZLOL000-20001",
// 					"io．tencent．bcs．clusterid":         "BCS-DEBUGSZLOL000-20001",
// 					"io．tencent．bcs．namespace":         "test6",
// 					"io．tencent．paas．versionid":        "1428",
// 					"io．tencent．paas．version":          "rso-auth-wechat",
// 					"io．tencent．paas．instanceid":       "3129",
// 					"io．tencent．bcs．app．appid":         "132",
// 					"BCS-WEIGHT-rso-auth":              "100",
// 					"Env":                              "dev",
// 					"BCSGROUP":                         "test6-loadbalance",
// 					"io．tencent．bkdata．baseall．dataid": "6566",
// 					"io．tencent．paas．templateid":       "50",
// 				},
// 				"annotations": map[string]interface{}{
// 					"io．tencent．paas．webCache": "{\"link_app\": [\"1520222679\"], \"link_app_weight\": [{\"id\": \"1520222679\", \"name\": \"rso-auth\", \"weight\": 100}], \"labelListCache\": [{\"key\": \"Owner\", \"value\": \"accounts-team-rg@riotgames.com\"}, {\"key\": \"Env\", \"value\": \"dev\"}, {\"key\": \"BCSGROUP\", \"value\": \"test6-loadbalance\"}]}",
// 				},
// 			},
// 			"namespace": "test6",
// 			"ports": []map[string]interface{}{
// 				map[string]interface{}{
// 					"name":        "rso-auth",
// 					"BCSVHost":    "",
// 					"path":        "",
// 					"protocol":    "TCP",
// 					"servicePort": 8081,
// 					"backends":    nil,
// 				},
// 			},
// 			"BCSGroup": []string{
// 				"test6-loadbalance",
// 			},
// 			"cluster":     "BCS-DEBUGSZLOL000-20001",
// 			"serviceName": "rso-auth-test6",
// 			"serviceWeight": map[string]interface{}{
// 				"rso-auth": 100,
// 			},
// 			"sslcert": false,
// 			"balance": "",
// 			"maxconn": 0,
// 		},
// 	}

// 	err = storeCli.Create(context.TODO(), obj, &store.CreateOptions{})
// 	if err != nil {
// 		t.Errorf("create obj failed, err %s", err.Error())
// 	}

// }

// func TestMongoStoreGet(t *testing.T) {
// 	db, err := createDB()
// 	if err != nil {
// 		t.Errorf("create db failed, err %s", err.Error())
// 	}
// 	storeCli := NewMongoStore(db)
// 	obj := &types.RawObject{}
// 	obj.SetObjectType("Dynamic")
// 	obj.SetClusterID("BCS-TEST-00001")
// 	obj.SetName("test1")
// 	obj.SetNamespace("ns1")
// 	err = storeCli.Get(context.TODO(), obj)
// 	if err != nil {
// 		t.Errorf("get object failed, err %s", err.Error())
// 	}
// 	t.Logf("%s", obj.ToString())
// 	t.Error()
// }

// func TestMongoStoreWatch(t *testing.T) {
// 	db, err := createDB()
// 	if err != nil {
// 		t.Errorf("create db failed, err %s", err.Error())
// 	}
// 	storeCli := NewMongoStore(db)

// 	mCtx, cancel := context.WithCancel(context.Background())
// 	var wg sync.WaitGroup

// 	go func() {
// 		wg.Add(1)
// 		defer wg.Done()

// 		obj := &types.RawObject{}
// 		obj.SetObjectType("Dynamic")
// 		obj.SetClusterID("BCS-TEST-00001")
// 		ch, err := storeCli.Watch(context.TODO(), obj, &store.WatchOptions{
// 			Selector: &types.ValueSelector{
// 				//Pairs: map[string]interface{}{types.TagClusterID: obj.GetClusterID()},
// 			},
// 			BatchSize:    10,
// 			MaxAwaitTime: time.Second * 10,
// 		})
// 		if err != nil {
// 			t.Errorf("watch failed, err %s", err.Error())
// 		}

// 		if err != nil {
// 			t.Errorf("watch failed, err %s", err.Error())
// 		}

// 		for {
// 			select {
// 			case event, ok := <-ch:
// 				if !ok {
// 					t.Logf("channel broken")
// 					return
// 				}
// 				t.Logf("received %s %s", event.Type, event.Obj.ToString())
// 			case <-mCtx.Done():
// 				t.Logf("context is done")
// 				return
// 			}
// 		}
// 	}()

// 	time.Sleep(2 * time.Second)

// 	obj := &types.RawObject{
// 		Meta: types.Meta{
// 			Type:      "Dynamic",
// 			ClusterID: "BCS-TEST-00001",
// 			Name:      "test1",
// 			Namespace: "ns1",
// 		},
// 		Data: map[string]interface{}{
// 			"metadata": map[string]interface{}{
// 				"name":              "rso-auth-test6",
// 				"namespace":         "test6",
// 				"creationTimestamp": "0001-01-01T00:00:00Z",
// 				"labels": map[string]interface{}{
// 					"io．tencent．paas．projectid":                 "636db594336748a581391188bae9177b",
// 					"io．tencent．bkdata．container．stdlog．dataid": "9915",
// 					"Owner":                            "accounts-team-rg@riotgames.com",
// 					"io．tencent．bcs．cluster":           "BCS-DEBUGSZLOL000-20001",
// 					"io．tencent．bcs．clusterid":         "BCS-DEBUGSZLOL000-20001",
// 					"io．tencent．bcs．namespace":         "test6",
// 					"io．tencent．paas．versionid":        "1428",
// 					"io．tencent．paas．version":          "rso-auth-wechat",
// 					"io．tencent．paas．instanceid":       "3129",
// 					"io．tencent．bcs．app．appid":         "132",
// 					"BCS-WEIGHT-rso-auth":              "100",
// 					"Env":                              "dev",
// 					"BCSGROUP":                         "test6-loadbalance",
// 					"io．tencent．bkdata．baseall．dataid": "6566",
// 					"io．tencent．paas．templateid":       "50",
// 				},
// 				"annotations": map[string]interface{}{
// 					"io．tencent．paas．webCache": "{\"link_app\": [\"1520222679\"], \"link_app_weight\": [{\"id\": \"1520222679\", \"name\": \"rso-auth\", \"weight\": 100}], \"labelListCache\": [{\"key\": \"Owner\", \"value\": \"accounts-team-rg@riotgames.com\"}, {\"key\": \"Env\", \"value\": \"dev\"}, {\"key\": \"BCSGROUP\", \"value\": \"test6-loadbalance\"}]}",
// 				},
// 			},
// 			"namespace": "test6",
// 			"ports": []map[string]interface{}{
// 				map[string]interface{}{
// 					"name":        "rso-auth",
// 					"BCSVHost":    "",
// 					"path":        "",
// 					"protocol":    "TCP",
// 					"servicePort": 8081,
// 					"backends":    nil,
// 				},
// 			},
// 			"BCSGroup": []string{
// 				"test6-loadbalance",
// 			},
// 			"cluster":     "BCS-DEBUGSZLOL000-20001",
// 			"serviceName": "rso-auth-test6",
// 			"serviceWeight": map[string]interface{}{
// 				"rso-auth": 100,
// 			},
// 			"sslcert": false,
// 			"balance": "",
// 			"maxconn": 0,
// 		},
// 	}

// 	err = storeCli.Create(context.TODO(), obj, &store.CreateOptions{UpdateExists: true})
// 	if err != nil {
// 		t.Errorf("create obj failed")
// 	}

// 	obj.SetData(map[string]interface{}{
// 		"metadata": map[string]interface{}{
// 			"name":              "rso-auth-test6",
// 			"namespace":         "test6",
// 			"creationTimestamp": "0001-01-01T00:00:00Z",
// 			"labels": map[string]interface{}{
// 				"io．tencent．paas．projectid":                 "636db594336748a581391188bae9177b",
// 				"io．tencent．bkdata．container．stdlog．dataid": "9915",
// 				"Owner":                            "accounts-team-rg@riotgames.com",
// 				"io．tencent．bcs．cluster":           "BCS-DEBUGSZLOL000-20001",
// 				"io．tencent．bcs．clusterid":         "BCS-DEBUGSZLOL000-20001",
// 				"io．tencent．bcs．namespace":         "test6",
// 				"io．tencent．paas．versionid":        "1428",
// 				"io．tencent．paas．version":          "rso-auth-wechat",
// 				"io．tencent．paas．instanceid":       "3129",
// 				"io．tencent．bcs．app．appid":         "132",
// 				"BCS-WEIGHT-rso-auth":              "100",
// 				"Env":                              "dev",
// 				"BCSGROUP":                         "test6-loadbalance",
// 				"io．tencent．bkdata．baseall．dataid": "6566",
// 				"io．tencent．paas．templateid":       "50",
// 			},
// 			"annotations": map[string]interface{}{
// 				"io．tencent．paas．webCache": "{\"link_app\": [\"1520222679\"], \"link_app_weight\": [{\"id\": \"1520222679\", \"name\": \"rso-auth\", \"weight\": 100}], \"labelListCache\": [{\"key\": \"Owner\", \"value\": \"accounts-team-rg@riotgames.com\"}, {\"key\": \"Env\", \"value\": \"dev\"}, {\"key\": \"BCSGROUP\", \"value\": \"test6-loadbalance\"}]}",
// 			},
// 		},
// 	})

// 	err = storeCli.Update(context.TODO(), obj, &store.UpdateOptions{})
// 	if err != nil {
// 		t.Errorf("create obj failed")
// 	}

// 	time.Sleep(3 * time.Second)

// 	cancel()
// 	wg.Wait()
// 	t.Error()

// }
