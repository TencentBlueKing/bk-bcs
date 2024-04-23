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
 */

// Package reflector xxx
package reflector

import (
	"context"
	"time"

	k8scache "k8s.io/client-go/tools/cache"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

// NewReflector create new reflector
func NewReflector(name string, store k8scache.Indexer,
	lw ListerWatcher, fullSyncPeriod time.Duration, handler EventInterface) *Reflector {

	cxt, stopfn := context.WithCancel(context.Background())
	return &Reflector{
		name:       name,
		cxt:        cxt,
		stopFn:     stopfn,
		listWatch:  lw,
		syncPeriod: fullSyncPeriod,
		store:      store,
		handler:    handler,
		underWatch: false,
	}
}

// Reflector offers lister & watcher mechanism to sync event-storage data
// to local memory store, and meanwhile push data event to predifine event handler
type Reflector struct {
	name       string // reflector name
	cxt        context.Context
	stopFn     context.CancelFunc
	listWatch  ListerWatcher    // lister & watcher for object data
	syncPeriod time.Duration    // period for resync all data
	store      k8scache.Indexer // memory store for all data object
	handler    EventInterface   // event callback when processing store data
	underWatch bool             // flag for watch handler

	// used for delete object in ListAllData
	keyFunc k8scache.KeyFunc
}

// SetKeyFunc set key function used for delete object in ListAllData
func (r *Reflector) SetKeyFunc(keyFunc k8scache.KeyFunc) {
	r.keyFunc = keyFunc
}

// Run running reflector, list all data in period and create stable watcher for
// all data events
func (r *Reflector) Run() {
	blog.V(3).Infof("%s ready to start, begin to cache data", r.name)
	// sync all data object from remote event storage
	// r.listAllData()
	watchCxt, _ := context.WithCancel(r.cxt)
	go r.handleWatch(watchCxt)
	blog.V(3).Infof("%s first resynchronization & watch success, register all ticker", r.name)
	// create ticker for data object resync
	syncTick := time.NewTicker(r.syncPeriod)
	defer syncTick.Stop()
	// create ticker check stable watcher
	watchTick := time.NewTicker(time.Second * 2)
	defer watchTick.Stop()
	for {
		select {
		case <-r.cxt.Done():
			blog.Warnf("%s running exit, lister & watcher stopped", r.name)
			return
		case <-syncTick.C:
			// fully resync all datas in period
			blog.Infof("%s trigger all data synchronization", r.name)
			_ = r.ListAllData()
		case <-watchTick.C:
			// check watch is running
			if r.underWatch {
				continue
			}
			// watch is out, recovery watch loop
			blog.Warnf("%s long watch is out, start watch loop to recovery.", r.name)
			go r.handleWatch(watchCxt)
		}
	}
}

// Stop stop reflector
func (r *Reflector) Stop() {
	blog.V(3).Infof("%s is asked to stop", r.name)
	r.stopFn()
}

// ListAllData list all data from listwatcher
func (r *Reflector) ListAllData() error {
	blog.V(3).Infof("%s begins to list all data...", r.name)
	objMap := make(map[string]meta.Object)
	objs, err := r.listWatch.List()
	if err != nil {
		// some err response, wait for next resync ticker
		blog.Errorf("%s List all data failed, %s", r.name, err)
		return err
	}
	blog.V(3).Infof("%s list all data success, objects number %d", r.name, len(objs))
	for _, obj := range objs {
		if r.keyFunc != nil {
			key, err := r.keyFunc(obj)
			if err != nil {
				blog.Errorf("%s gets obj key failed, err %s", r.name, err)
				continue
			}
			objMap[key] = obj
		}

		oldObj, exist, err := r.store.Get(obj)
		if err != nil {
			blog.Errorf("%s gets local store err under List, %s, discard data", r.name, err)
			continue
		}
		if exist {
			_ = r.store.Update(obj)
			if r.handler != nil {
				r.handler.OnUpdate(oldObj, obj)
			}
			blog.V(5).Infof("%s update %s/%s notify succes in Lister.", r.name, obj.GetNamespace(), obj.GetName())
		} else {
			_ = r.store.Add(obj)
			if r.handler != nil {
				r.handler.OnAdd(obj)
			}
			blog.V(5).Infof("%s add %s/%s notify succes in Lister.", r.name, obj.GetNamespace(), obj.GetName())
		}
	}

	if r.keyFunc != nil {
		cacheObjKeys := r.store.ListKeys()
		for _, key := range cacheObjKeys {
			if _, ok := objMap[key]; !ok {
				obj, _, _ := r.store.GetByKey(key)
				_ = r.store.Delete(obj)
				r.handler.OnDelete(obj)
			}
		}
	}

	return nil
}

func (r *Reflector) handleWatch(cxt context.Context) {
	if r.underWatch {
		return
	}
	r.underWatch = true
	watcher, err := r.listWatch.Watch()
	if err != nil {
		blog.Errorf("Reflector %s create watch by ListerWatcher failed, %s", r.name, err)
		r.underWatch = false
		return
	}
	defer func() {
		r.underWatch = false
		watcher.Stop()
		blog.Infof("Reflector %s watch loop exit", r.name)
	}()
	blog.V(3).Infof("%s enter storage watch loop, waiting for event trigger", r.name)
	channel := watcher.WatchEvent()
	for {
		select {
		case <-cxt.Done():
			blog.Infof("reflector %s is asked to exit.", r.name)
			return
		case event, ok := <-channel:
			if !ok {
				blog.Errorf("%s reads watch.Event from channel failed. channel closed", r.name)
				return
			}
			switch event.Type {
			case watch.EventSync, watch.EventAdded, watch.EventUpdated:
				r.processAddUpdate(&event)
			case watch.EventDeleted:
				r.processDeletion(&event)
			case watch.EventErr:
				// some unexpected err occurred, but channel & watach is still work
				blog.V(3).Infof("Reflector %s catch some data err in watch.Event channel, keep watch running", r.name)
			}
		}
	}
}

func (r *Reflector) processAddUpdate(event *watch.Event) {
	oldObj, exist, err := r.store.Get(event.Data)
	if err != nil {
		blog.V(3).Infof("Reflector %s gets local store err, %s", r.name, err)
		return
	}
	if exist {
		_ = r.store.Update(event.Data)
		if r.handler != nil {
			r.handler.OnUpdate(oldObj, event.Data)
		}
	} else {
		_ = r.store.Add(event.Data)
		if r.handler != nil {
			r.handler.OnAdd(event.Data)
		}
	}
}

func (r *Reflector) processDeletion(event *watch.Event) {
	// fix(DeveloperJim): 2018-06-26 16:42:10
	// when deletion happens in zookeeper, no Object dispatchs, so we
	// need to get object from local cache
	oldObj, exist, err := r.store.Get(event.Data)
	if err != nil {
		blog.V(3).Infof("Reflector %s gets local store err in DeleteEvent, %s", r.name, err)
		return
	}
	if exist {
		_ = r.store.Delete(event.Data)
		if event.Data.GetAnnotations() != nil &&
			event.Data.GetAnnotations()["bk-bcs-inner-storage"] == "bkbcs-zookeeper" {

			// tricky here, zookeeper can't get creation time when deletion
			if r.handler != nil {
				r.handler.OnDelete(oldObj)
			}
			blog.V(5).Infof("reflector %s invoke Delete tricky callback func for %s/%s.",
				r.name, event.Data.GetNamespace(), event.Data.GetName())
		} else {
			if r.handler != nil {
				r.handler.OnDelete(event.Data)
			}
			blog.V(5).Infof("reflector %s invoke Delete callback for %s/%s.",
				r.name, event.Data.GetNamespace(), event.Data.GetName())
		}
		return
	}
	// local cache do not exist, nothing happens
	blog.V(3).Infof("reflector %s lost local cache for %s/%s", r.name, event.Data.GetNamespace(), event.Data.GetName())
}

// EventInterface register interface for event notification
type EventInterface interface {
	OnAdd(obj interface{})
	OnUpdate(old, cur interface{})
	OnDelete(obj interface{})
}
