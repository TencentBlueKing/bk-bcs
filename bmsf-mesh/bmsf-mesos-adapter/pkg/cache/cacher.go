/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package cache

import (
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/pkg/meta"
	"bk-bcs/bcs-common/pkg/storage"
	"bk-bcs/bcs-common/pkg/watch"
	"fmt"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/context"
)

//CacherConfig all config items for Cacher
type CacherConfig struct {
	Name          string          //uniq Cacher name for metric
	Storage       storage.Storage //event-storage implementation for read/write/watch operation
	KeyFn         ObjectKeyFunc   //ObjectKeyFunc for local memory data indexer
	ResyncPeriod  time.Duration   //resync time period for Lister
	ListerWatcher ListerWatcher   //custom Lister and Watcher, optional
}

//VerifyCacherConfig check cacherconfig accuracy
func VerifyCacherConfig(config *CacherConfig) error {
	if len(config.Name) == 0 {
		return fmt.Errorf("lost cacher name")
	}
	if config.Storage == nil {
		return fmt.Errorf("lost storage implementation")
	}
	if config.KeyFn == nil {
		return fmt.Errorf("lost Object Key Function")
	}
	return nil
}

//NewCacher create cacher for resource with config
func NewCacher(config *CacherConfig) (*Cacher, error) {
	if err := VerifyCacherConfig(config); err != nil {
		return nil, err
	}
	//create data object store
	if config.ListerWatcher == nil {
		config.ListerWatcher = &ListWatch{
			ListFn: func() ([]meta.Object, error) {
				return config.Storage.List(context.TODO(), "", &storage.Everything{})
			},
			WatchFn: func() (watch.Interface, error) {
				return config.Storage.Watch(context.TODO(), "", "", &storage.Everything{})
			},
		}
	}
	store := NewCache(config.KeyFn)
	handler := &eventProxy{}
	ref := NewReflector(fmt.Sprintf("Reflector-%s", config.Name), store, config.ListerWatcher, config.ResyncPeriod, handler)
	cxt, canceler := context.WithCancel(context.Background())
	c := &Cacher{
		name:      fmt.Sprintf("Cacher-%s", config.Name),
		cxt:       cxt,
		stopFn:    canceler,
		stopped:   false,
		inComing:  make(chan watch.Event, watch.DefaultChannelBuffer),
		store:     store,
		storage:   config.Storage,
		reflector: ref,
		indexerWatch: &indexerMaps{
			watchMaps: make(map[uint64]*dispatchWatch),
			indexer:   0,
		},
	}
	handler.cache = c
	//running reflector, begin to List & Watch data
	go c.reflector.Run()
	//running local goroutine for event dispathing
	go c.dispatchWatchEvent()
	return c, nil
}

//Cacher implements Storage interface, holds all object data in memory Store.
//* delegates all write operations to inner storage instance
//* delegates all read operations to local inner Store;
//* broadcasts same watch event to cacherwatch
//all datas in Store are pointer, **No Modify** to datas when dispatching to
//different watch event channel
type Cacher struct {
	name         string             //cacher name for metric next
	cxt          context.Context    //context controll exit
	stopFn       context.CancelFunc //stop func relate to cxt
	stopped      bool               //flag for stop
	inComing     chan watch.Event   //event channel for receiving data from reflector
	store        Store              //store for memory cache
	storage      storage.Storage    //delegate Storage implementation for data access
	reflector    *Reflector         //reflector for Lister & Watcher
	indexerWatch *indexerMaps       //watcherMaps for cacherwatch
}

//GetName return data type name for this cacher
func (c *Cacher) GetName() string {
	return c.name
}

//Create implements Storage interface
func (c *Cacher) Create(ctx context.Context, key string, obj meta.Object, ttl int) (out meta.Object, err error) {
	return c.storage.Create(ctx, key, obj, ttl)
}

//Delete implements Storage interface
func (c *Cacher) Delete(ctx context.Context, key string) (obj meta.Object, err error) {
	return c.storage.Delete(ctx, key)
}

//Watch implements Storage interface
//create cacher watcher to dispath data object
func (c *Cacher) Watch(ctx context.Context, key, version string, selector storage.Selector) (watch.Interface, error) {
	if c.stopped {
		return nil, fmt.Errorf("Cacher is stopped")
	}
	//create dispathWatch
	watcher := newDispatchWatch(c.indexerWatch, selector)
	indexer := c.indexerWatch.addWatch(watcher)
	watcher.setIndexer(indexer)
	go watcher.selectWatchEvent()
	return watcher, nil
}

//WatchList implements Storage interface
func (c *Cacher) WatchList(cxt context.Context, key, version string, selector storage.Selector) (watch.Interface, error) {
	return c.Watch(cxt, key, version, selector)
}

//Get implements Storage interface
//param key: string key for target object in Store, it's different with storage like etcd/zookeeper
func (c *Cacher) Get(ctx context.Context, key, version string, ignoreNotFound bool) (meta.Object, error) {
	//get local info
	item, exist, err := c.store.GetByKey(key)
	if err != nil {
		return nil, fmt.Errorf("get data err: %s", err)
	}
	if !exist {
		if ignoreNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("Data Not Found")
	}
	obj, ok := item.(meta.Object)
	if !ok {
		return nil, fmt.Errorf("data structure error")
	}
	return obj, nil
}

//List implements Storage interface
func (c *Cacher) List(ctx context.Context, key string, selector storage.Selector) ([]meta.Object, error) {
	//list all data from local cache, filtering expected data with selector
	items := c.store.List()
	var objs []meta.Object
	if len(items) == 0 {
		return objs, nil
	}
	//modify data type, filter all items
	for _, item := range items {
		obj, ok := item.(meta.Object)
		if !ok {
			return nil, fmt.Errorf("cacher data is not runtime.Object")
		}
		if selector == nil {
			objs = append(objs, obj)
			continue
		}
		need, err := selector.Matchs(obj)
		if err != nil {
			return nil, fmt.Errorf("selector implementation matchs err, %s", err)
		}
		//select target objects
		if need {
			objs = append(objs, obj)
		}
	}
	return objs, nil
}

//Close stop cacher backgroup goroutines
func (c *Cacher) Close() {
	c.stopped = true
	c.reflector.Stop()
	c.stopFn()
	c.indexerWatch.terminatAll()
	close(c.inComing)
}

func (c *Cacher) handleReflectorNotify(e watch.Event) {
	if c.stopped {
		blog.Warnf("cacher %s is stopped, discard all data", c.name)
		return
	}
	c.inComing <- e
	blog.V(5).Infof("%s push watch[%s-%s/%s] inComming channel", c.name, e.Type, e.Data.GetNamespace(), e.Data.GetName())
}

//dispatchWatchEvent dispath all coming events to different watch end user
func (c *Cacher) dispatchWatchEvent() {
	blog.V(3).Infof("%s is running event dispath to all local cache watch", c.name)
	for {
		if c.stopped {
			blog.Infof("%s is stopped, discard all data in inComing channel", c.name)
			return
		}
		select {
		case <-c.cxt.Done():
			blog.Infof("%s is asked stopped, ready to exit", c.name)
			return
		case event, ok := <-c.inComing:
			if !ok || c.stopped {
				//channel closed
				blog.Warnf("cacher %s is disaptch channel is closed, ready to exit", c.name)
				return
			}
			blog.V(3).Infof("%s dispath event %s to all client watch", c.name, event.Type)
			c.indexerWatch.dispatchAll(event)
		}
	}
}

//eventProxy offer event proxy for Cacher, handle Add/Update/Delete
//event when data change
type eventProxy struct {
	cache *Cacher
}

//OnAdd implements EventInterface
func (proxy *eventProxy) OnAdd(obj interface{}) {
	if obj == nil {
		blog.V(5).Info("eventProxy got empty object when Add")
		return
	}
	data, ok := obj.(meta.Object)
	if !ok {
		blog.V(5).Info("event proxy got non-object data in add event.")
		return
	}
	e := watch.Event{Type: watch.EventAdded, Data: data}
	proxy.cache.handleReflectorNotify(e)
	blog.V(5).Infof("%s channel is receiving ADD Event", proxy.cache.name)
}

//OnUpdate implements EventInterface
func (proxy *eventProxy) OnUpdate(old, cur interface{}) {
	if old == nil || cur == nil {
		blog.V(5).Info("eventProxy got empty object when Update")
		return
	}
	if reflect.DeepEqual(old, cur) {
		blog.V(5).Info("eventProxy got equal object when Update")
		return
	}
	data, ok := cur.(meta.Object)
	if !ok {
		blog.V(5).Info("eventProxy got not Object data")
		return
	}
	e := watch.Event{Type: watch.EventUpdated, Data: data}
	proxy.cache.handleReflectorNotify(e)
	blog.V(5).Infof("%s channel is receiving Update Event", proxy.cache.name)
}

//OnDelete implements EventInterface
func (proxy *eventProxy) OnDelete(obj interface{}) {
	if obj == nil {
		blog.V(5).Info("eventProxy got empty objects when Delete")
		return
	}
	data, ok := obj.(meta.Object)
	if !ok {
		blog.V(5).Info("eventProxy got not Object data when Delete")
		return
	}
	e := watch.Event{Type: watch.EventDeleted, Data: data}
	proxy.cache.handleReflectorNotify(e)
	blog.V(5).Infof("%s channel is receiving Delete Event", proxy.cache.name)
}

//newDispatchWatch create dispatch event watch from one Watch request in Cacher.
//selector comes from watch request in Cacher.Watch
func newDispatchWatch(c *indexerMaps, s storage.Selector) *dispatchWatch {
	cxt, canceler := context.WithCancel(context.Background())
	w := &dispatchWatch{
		indexer:       0,
		stopped:       false,
		cxt:           cxt,
		stopFn:        canceler,
		cacher:        c,
		selector:      s,
		filterChannel: make(chan watch.Event, watch.DefaultChannelBuffer),
		eventChannel:  make(chan watch.Event, watch.DefaultChannelBuffer),
	}
	return w
}

//dispatchWatch watch delegate for exact client, dispatch event without block.
//dispatchWatch holds selector from client, when receive event from Cacher,
//it will filter data before push to client event channel
type dispatchWatch struct {
	indexer       uint64             //uniq watch indexer
	stopped       bool               //flag for stopped
	cxt           context.Context    //context for exits
	stopFn        context.CancelFunc //func for stop
	cacher        *indexerMaps       //cacher pointer refference
	selector      storage.Selector   //selector for data filter
	filterChannel chan watch.Event   //original event channel for filter
	eventChannel  chan watch.Event   //channel for watch interface
}

//Stop implements Watch interface
func (cw *dispatchWatch) Stop() {
	cw.cacher.deleteWatch(cw.indexer)
	cw.stop()
}

//inner stop function
func (cw *dispatchWatch) stop() {
	cw.stopped = true
	cw.stopFn()
}

//WatchEvent implements Watch interface
func (cw *dispatchWatch) WatchEvent() <-chan watch.Event {
	return cw.eventChannel
}

func (cw *dispatchWatch) addWatchEvent(e watch.Event) {
	if cw.stopped {
		blog.Warnf("dispathWatch %d is stopped, discard all data", cw.indexer)
		return
	}
	cw.filterChannel <- e
	//data discards
}

//selectWatchEvent filter event by Selector
func (cw *dispatchWatch) selectWatchEvent() {
	blog.Infof("dispathWatch %d is ready to work, selector: %s", cw.indexer, cw.selector.String())
	defer func() {
		blog.Infof("dispathWatch %d ready to exit, close all channel", cw.indexer)
		close(cw.filterChannel)
		close(cw.eventChannel)
	}()
	for {
		select {
		case event, ok := <-cw.filterChannel:
			if !ok {
				blog.Warnf("dispathWatch %d filter channel closed, err exit", cw.indexer)
				return
			}
			if cw.selector == nil {
				cw.eventChannel <- event
				continue
			}
			matched, err := cw.selector.Matchs(event.Data)
			if err != nil {
				blog.Warnf("diapatchWatch %d for Cacher occure error when select data, %s", cw.indexer, err)
				continue
			}
			if matched {
				cw.eventChannel <- event
			}
		case <-cw.cxt.Done():
			blog.Infof("dispathWatch %d is asked to stop", cw.indexer)
			return
		}
	}
}

func (cw *dispatchWatch) setIndexer(indexer uint64) {
	cw.indexer = indexer
}

//indexer storage for indexing cacherwatch
type indexerMaps struct {
	watchMaps map[uint64]*dispatchWatch //maps for cacherwatch storage
	indexer   uint64                    //indexer for cachewatch
	lock      sync.RWMutex              //lock for watchmaps & indexer
}

//addWatch store cacherwatch & return indexer for this watch
func (im *indexerMaps) addWatch(w *dispatchWatch) uint64 {
	im.lock.Lock()
	defer im.lock.Unlock()
	im.indexer++
	im.watchMaps[im.indexer] = w
	return im.indexer
}

//deleteWatch delete watch in store with it's uniq indexer
func (im *indexerMaps) deleteWatch(indexer uint64) {
	im.lock.Lock()
	defer im.lock.Unlock()
	delete(im.watchMaps, indexer)
}

func (im *indexerMaps) dispatchAll(event watch.Event) {
	im.lock.RLock()
	defer im.lock.RUnlock()
	for _, w := range im.watchMaps {
		w.addWatchEvent(event)
	}
}

func (im *indexerMaps) terminatAll() {
	im.lock.Lock()
	defer im.lock.Unlock()
	for index, w := range im.watchMaps {
		delete(im.watchMaps, index)
		w.stop()
	}
	im.indexer = 0
}
