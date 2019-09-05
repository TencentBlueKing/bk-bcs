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

package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/pkg/meta"
	"bk-bcs/bcs-common/pkg/storage"
	"bk-bcs/bcs-common/pkg/storage/http"
	"bk-bcs/bcs-common/pkg/watch"
)

var _ ListWatchStore = Store(&Cache{})

type ListWatchStore interface {
	List() []interface{}
	ListKeys() []string
	Get(obj interface{}) (item interface{}, exists bool, err error)
	GetByKey(key string) (item interface{}, exists bool, err error)
	//Num will return data counts in Store
	Num() int
}

type OnEventFunc struct {
	// called when an add object event occurred.
	OnAddFunc func(object meta.Object)

	// called when a update object event occurred.
	OnUpdateFunc func(object meta.Object)

	// called when a delete object event occurred.
	OnDeleteFunc func(object meta.Object)
}

type LWConfig struct {
	TLS *tls.Config

	// the duration to sync all from mesh api server.
	SyncSeconds int

	// mesh api server address
	APIAddress []string

	// the api version to be used, default is v1.
	Version string

	// the path that is watched to mesh api
	WatchPath string

	// func(s) that will be used in the mesh api watch api
	ObjectFunc ObjectFunc

	// func(s) that will be called when add/update/delete event occurred.
	OnEventFunc OnEventFunc
}

type ObjectFunc struct {
	KeyFunc  ObjectKeyFunc
	NewFunc  meta.ObjectNewFn
	ListFunc meta.ObjectListNewFn
}

func NewListWatchStore(lwc LWConfig) (ListWatchStore, error) {
	hc := http.Config{
		Hosts:          lwc.APIAddress,
		Codec:          &meta.JsonCodec{},
		ObjectNewFunc:  lwc.ObjectFunc.NewFunc,
		ObjectListFunc: lwc.ObjectFunc.ListFunc,
		TLS:            lwc.TLS,
	}
	blog.Infof("list watch store, mesh api server address: %v", lwc.APIAddress)

	stg, err := http.NewStorage(&hc)
	if err != nil {
		return nil, fmt.Errorf("new cache storage failed, err: %v", err)
	}

	listWatchCache := &ListWatchCache{
		syncSeconds: lwc.SyncSeconds,
		watchPath:   lwc.WatchPath,
		version:     "v1",
		client:      stg,
		retryChan:   make(chan struct{}, 2),
		eventFunc:   lwc.OnEventFunc,
		keyFunc:     lwc.ObjectFunc.KeyFunc,
		Store:       NewCache(lwc.ObjectFunc.KeyFunc),
	}
	if err := listWatchCache.watch(); err != nil {
		return nil, fmt.Errorf("initial list watch store, start watch failed, err: %v", err)
	}

	listWatchCache.reList()
	listWatchCache.reWatch()
	return listWatchCache, nil
}

type ListWatchCache struct {
	Store
	eventFunc OnEventFunc
	keyFunc   ObjectKeyFunc
	// default 2s
	syncSeconds int
	watchPath   string
	version     string
	client      storage.Storage
	retryChan   chan struct{}
}

func (lw *ListWatchCache) reList() {
	tick := 2
	if lw.syncSeconds != 0 {
		tick = lw.syncSeconds
	}
	blog.Infof("start relist path[%s] every %d seconds", lw.watchPath, tick)
	go func() {
		duration := time.Duration(tick) * time.Second
		for {
			if err := lw.list(); err != nil {
				blog.Errorf("relist from api server with path[%s] failed, err: %v", lw.watchPath, err)
				time.Sleep(1 * time.Second)
				continue
			}
			time.Sleep(duration)
			blog.V(2).Infof("list watch store, relist path[%s] success.", lw.watchPath)
		}
	}()
}

func (lw *ListWatchCache) reWatch() {
	blog.Infof("rewatch go routine started.")
	go func() {
		for range lw.retryChan {
			blog.Infof("start retry watch, path[%s]", lw.watchPath)
			time.Sleep(1 * time.Second)

			if err := lw.list(); err != nil {
				blog.Errorf("retry watch path[%s] with list failed, err: %v", lw.watchPath, err)
				lw.retryChan <- struct{}{}
				continue
			}

			if err := lw.watch(); err != nil {
				blog.Errorf("retry watch path[%s] failed, err: %v", lw.watchPath, err)
				lw.retryChan <- struct{}{}
				continue
			}

			blog.Infof("retry watch path[%s] success.", lw.watchPath)
		}
	}()
}

func (lw *ListWatchCache) watch() error {
	blog.Infof("list watch store, start watch path[%s].", lw.watchPath)
	w, err := lw.client.Watch(context.Background(), lw.watchPath, lw.version, nil)
	if err != nil {
		return fmt.Errorf("watch path[%s] failed, err: %v", lw.watchPath, err)
	}
	blog.Infof("list watch store, watch path[%s] success.", lw.watchPath)

	go func() {
		for {
			event, exit := <-w.WatchEvent()
			if !exit {
				blog.Errorf("watch path[%s], but watch connection closed, need to retry again.", lw.watchPath)
				lw.retryChan <- struct{}{}
				break
			}

			e := event.Data
			js, _ := json.Marshal(e)
			blog.Warnf("-> list watch store, received event[%s] with details: %s", event.Type, string(js))

			switch event.Type {
			case watch.EventAdded:
				if lw.eventFunc.OnAddFunc != nil {
					lw.eventFunc.OnAddFunc(e)
				}

				if err := lw.Store.Add(e); err != nil {
					blog.Errorf("add object[%s/%s] failed, err: %v", e.GetName(), e.GetNamespace(), err)
					continue
				}
			case watch.EventUpdated:
				if lw.eventFunc.OnUpdateFunc != nil {
					lw.eventFunc.OnUpdateFunc(e)
				}

				if err := lw.Store.Update(e); err != nil {
					blog.Errorf("update object[%s/%s] failed, err: %v", e.GetName(), e.GetNamespace(), err)
					continue
				}
			case watch.EventDeleted:
				if lw.eventFunc.OnDeleteFunc != nil {
					lw.eventFunc.OnDeleteFunc(e)
				}

				if err := lw.Store.Delete(e); err != nil {
					blog.Errorf("update object[%s/%s] failed, err: %v", e.GetName(), e.GetNamespace(), err)
					continue
				}
			case watch.EventSync:
				if lw.eventFunc.OnUpdateFunc != nil {
					lw.eventFunc.OnUpdateFunc(e)
				}

				if err := lw.Store.Update(e); err != nil {
					blog.Errorf("update object[%s/%s] failed, err: %v", e.GetName(), e.GetNamespace(), err)
					continue
				}
			case watch.EventErr:
				blog.Errorf("event got err with data: %+#v", e)
				continue
			default:
				blog.Errorf("unknown event type[%s] with data: %+#v", event.Type, e)
				continue
			}
		}
	}()

	return nil
}

func (lw *ListWatchCache) list() error {
	list, err := lw.client.List(context.Background(), lw.watchPath, nil)
	if err != nil {
		return fmt.Errorf("list path[%s] failed, err: %v", lw.watchPath, err)
	}
	blog.V(2).Infof("list watch store, list[%s] got %d objects.", lw.watchPath, len(list))
	listMapper := make(map[string]interface{})
	for _, v := range list {
		key, err := lw.keyFunc(v)
		if err != nil {
			blog.Errorf("list watch store, list and get object key failed, err: %v, obj: %+v", err, v)
		} else {
			listMapper[key] = v
		}

		obj, exist, err := lw.Store.Get(v)
		if err != nil {
			blog.Errorf("list watch store, get obj failed, err: %v, obj: %+v", err, v)
			continue
		}

		// new object
		if !exist {
			lw.Store.Add(v)
			lw.eventFunc.OnAddFunc(v)
			continue
		}

		if reflect.DeepEqual(v, obj) {
			blog.V(3).Infof("list watch store, list[%s] compare is equal, obj: %+v", lw.watchPath, v)
			continue
		}

		// need update object
		lw.Store.Update(v)
		lw.eventFunc.OnUpdateFunc(v)
	}

	// find redundant object in store and delete it.
	// call delete event at the same time.
	cachedObj := lw.Store.List()
	for _, obj := range cachedObj {
		key, err := lw.keyFunc(obj)
		if err != nil {
			blog.Errorf("list watch store, sync cached list, get key failed, err: %v, obj: %+v", err, obj)
			continue
		}
		_, exist := listMapper[key]
		if exist {
			// this obj already exist, so do not need compare again, because
			// it is already updated before.
			continue

		}
		// this is a redundant object, delete now.
		blog.Warnf("find a redundant object in list store, delete now, object: %+v", obj)
		if mObj, ok := obj.(meta.Object); ok {
			// send the event now.
			lw.eventFunc.OnDeleteFunc(mObj)
		} else {
			blog.Error("assert object to meta object failed.")
		}

		if err := lw.Store.Delete(obj); err != nil {
			blog.Errorf("delete redundant object in list store failed, err: %v, obj: %+v", err, obj)
			continue
		}
	}
	return nil
}
