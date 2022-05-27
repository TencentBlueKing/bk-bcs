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

package etcd

import (
	"crypto/tls"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
)

//Config etcd storage config
type Config struct {
	Host       string           //etcd host info
	PathPrefix string           //operation path prefix to join with key, if needed
	User       string           //user name for authentication
	Passwd     string           //password relative to user
	NewFunc    meta.ObjectNewFn //func for object creation
	Codec      meta.Codec       //Codec for encoder & decoder
	TLS        *tls.Config      //tls config for https
}

//NewStorage create etcd accessor implemented storage interface
func NewStorage(config *Config) (storage.Storage, error) {
	endpoints := strings.Split(config.Host, ",")
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("Lost etcd host with prefix: %s", config.PathPrefix)
	}
	c, err := clientv3.New(clientv3.Config{
		Endpoints:        endpoints,
		DialTimeout:      time.Second * 5,
		AutoSyncInterval: time.Minute * 5,
		TLS:              config.TLS,
		Username:         config.User,
		Password:         config.Passwd,
	})
	if err != nil {
		blog.V(3).Infof("create etcd storage with %s prefix %s failed, %s", config.Host, config.PathPrefix, err)
		return nil, err
	}
	s := &Storage{
		client:      c,
		objectNewFn: config.NewFunc,
		codec:       config.Codec,
		pathPrefix:  config.PathPrefix,
	}
	return s, nil
}

//Storage implementation storage interface with etcd client
type Storage struct {
	client      *clientv3.Client //etcd client
	objectNewFn meta.ObjectNewFn //create new object for codec.decode
	codec       meta.Codec       //json Codec for object
	pathPrefix  string           //etcd prefix
}

//Create implements storage interface
//param key: full path for etcd
func (s *Storage) Create(cxt context.Context, key string, obj meta.Object, ttl int) (out meta.Object, err error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("lost object key")
	}
	var clientCxt context.Context
	if ttl > 0 {
		var cancelFn context.CancelFunc
		clientCxt, cancelFn = context.WithTimeout(cxt, time.Second*time.Duration(ttl))
		defer cancelFn()
	} else {
		clientCxt = cxt
	}
	//serialize object
	data, err := s.codec.Encode(obj)
	if err != nil {
		blog.V(5).Infof("etcd storage %s encode %s/%s failed, %s", s.pathPrefix, obj.GetNamespace(), obj.GetName(), err)
		return nil, fmt.Errorf("encode %s/%s: %s", obj.GetNamespace(), obj.GetName(), err)
	}
	fullPath := path.Join(s.pathPrefix, key)
	if len(fullPath) == 0 {
		blog.V(5).Infof("etcd storage %s construct path for %s/%s failed. discard operation", s.pathPrefix, obj.GetNamespace(), obj.GetName())
		return nil, fmt.Errorf("empty object storage full path")
	}
	response, err := s.client.Put(clientCxt, fullPath, string(data), clientv3.WithPrevKV())
	if err != nil {
		blog.V(3).Infof("etcd storage %s Create %s/%s failed, %s", s.pathPrefix, obj.GetNamespace(), obj.GetName(), err)
		return nil, err
	}
	if response.PrevKv != nil && len(response.PrevKv.Value) > 0 {
		//we got previous key-value from creation
		target := s.objectNewFn()
		if err := s.codec.Decode(response.PrevKv.Value, target); err != nil {
			blog.V(3).Infof("etcd storage %s decode %s/%s Previous value failed, %s", s.pathPrefix, obj.GetNamespace(), obj.GetName(), err)
			//even got previous data failed, we still consider Create successfully
			return nil, nil
		}
		blog.V(3).Infof("etcd storage %s update %s/%s & got previous kv success", s.pathPrefix, obj.GetNamespace(), obj.GetName())
		return target, nil
	}
	blog.V(3).Infof("etcd storage %s create %s/%s success", s.pathPrefix, obj.GetNamespace(), obj.GetName())
	return nil, nil
}

//Delete implements storage interface
//for etcd operation, there are two situations for key
//* if key is empty, delete all data node under storage.pathPrefix
//* if key is not empty, delete all data under pathPrefix/key
//actually, Storage do not return object deleted
func (s *Storage) Delete(ctx context.Context, key string) (obj meta.Object, err error) {
	fullPath := s.pathPrefix
	if len(key) != 0 {
		fullPath = path.Join(s.pathPrefix, key)
	}
	response, err := s.client.Delete(ctx, fullPath, clientv3.WithPrefix())
	if err != nil {
		blog.V(3).Infof("etcd storage delete %s failed, %s", fullPath, err)
		return nil, err
	}
	blog.V(3).Infof("etcd storage clean data node under %s success, node num: %d", fullPath, response.Deleted)
	return nil, nil
}

//Watch implements storage interface
//* if key empty, watch all data
//* if key is namespace, watch all data under namespace
//* if key is namespace/name, watch detail data
//watch is Stopped when any error occure
func (s *Storage) Watch(cxt context.Context, key, version string, selector storage.Selector) (watch.Interface, error) {
	fullPath := s.pathPrefix
	if len(key) != 0 {
		fullPath = path.Join(s.pathPrefix, key)
	}
	proxy := newEtcdProxyWatch(cxt, s.codec, selector)
	//create watchchan
	etcdChan := s.client.Watch(cxt, fullPath, clientv3.WithPrefix(), clientv3.WithPrevKV())
	go proxy.eventProxy(etcdChan, s.objectNewFn)
	blog.V(3).Infof("etcd client is ready to watch %s", fullPath)
	return proxy, nil
}

//WatchList implements storage interface
//Watch & WatchList are the same for etcd storage
func (s *Storage) WatchList(ctx context.Context, key, version string, selector storage.Selector) (watch.Interface, error) {
	return s.Watch(ctx, key, version, selector)
}

//Get implements storage interface
//get exactly data object from etcd client. so key must be resource fullpath
func (s *Storage) Get(cxt context.Context, key, version string, ignoreNotFound bool) (obj meta.Object, err error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("lost object key")
	}
	fullPath := path.Join(s.pathPrefix, key)
	response, err := s.client.Get(cxt, fullPath)
	if err != nil {
		blog.V(3).Infof("etcd storage exact get %s failed, %s", fullPath, err)
		return nil, err
	}
	if len(response.Kvs) == 0 {
		if ignoreNotFound {
			blog.V(5).Infof("etcd storage got nothing under %s", fullPath)
			return nil, nil
		}
		return nil, storage.ErrNotFound
	}
	value := response.Kvs[0].Value
	if len(value) == 0 {
		blog.V(3).Infof("etcd storage got empty data for %s", fullPath)
		return nil, storage.ErrNotFound
	}
	target := s.objectNewFn()
	if err := s.codec.Decode(value, target); err != nil {
		blog.V(3).Infof("etcd storage decode data object %s failed, %s", fullPath, err)
		return nil, fmt.Errorf("%s decode: %s", s.pathPrefix, err)
	}
	blog.V(3).Infof("etcd client got %s success", fullPath)
	return target, nil
}

//List implements storage interface
//list namespace-based data or all data
func (s *Storage) List(cxt context.Context, key string, selector storage.Selector) (objs []meta.Object, err error) {
	fullPath := s.pathPrefix
	if len(key) != 0 {
		fullPath = path.Join(s.pathPrefix, key)
	}
	response, err := s.client.Get(cxt, fullPath, clientv3.WithPrefix())
	if err != nil {
		blog.V(3).Infof("etcd storage exact get %s failed, %s", fullPath, err)
		return nil, err
	}
	if len(response.Kvs) == 0 {
		blog.V(5).Infof("etcd storage list nothing under %s", fullPath)
		return nil, nil
	}
	for _, node := range response.Kvs {
		target := s.objectNewFn()
		if err := s.codec.Decode(node.Value, target); err != nil {
			blog.V(3).Infof("etcd storage decode data object %s failed, %s", fullPath, err)
			continue
		}
		if selector == nil {
			objs = append(objs, target)
			continue
		}
		if ok, _ := selector.Matchs(target); ok {
			objs = append(objs, target)
		}
	}
	blog.V(3).Infof("etcd storage list %s success, got %d objects", fullPath, len(objs))
	return objs, nil
}

//Close storage conenction, clean resource
func (s *Storage) Close() {
	blog.V(3).Infof("etcd storage %s exit.", s.pathPrefix)
	s.client.Close()
}

//etcdProxyWatch create etcdproxy watch
func newEtcdProxyWatch(cxt context.Context, codec meta.Codec, s storage.Selector) *etcdProxyWatch {
	localCxt, canceler := context.WithCancel(cxt)
	proxy := &etcdProxyWatch{
		selector:      s,
		codec:         codec,
		filterChannel: make(chan watch.Event, watch.DefaultChannelBuffer),
		cxt:           localCxt,
		stopFn:        canceler,
	}
	return proxy
}

//etcdProxyWatch wrapper for etcd watch, filter data by selector if needed.
//* delegates etcd client watch, decode raw bytes to object
//* constructs event and dispaths to user channel
//* maintains watch availability even if network connection is broken until user stop watach
type etcdProxyWatch struct {
	selector      storage.Selector
	codec         meta.Codec
	filterChannel chan watch.Event
	cxt           context.Context //context from storage.Watch
	stopFn        context.CancelFunc
}

//Stop watch channel
func (e *etcdProxyWatch) Stop() {
	e.stopFn()
}

//WatchEvent get watch events, if watch stopped/error, watch must close
// channel and exit, watch user must read channel like
// e, ok := <-channel
func (e *etcdProxyWatch) WatchEvent() <-chan watch.Event {
	return e.filterChannel
}

func (e *etcdProxyWatch) eventProxy(etcdCh clientv3.WatchChan, objectNewFn meta.ObjectNewFn) {
	defer func() {
		close(e.filterChannel)
	}()
	for {
		select {
		case <-e.cxt.Done():
			blog.V(3).Infof("etcdProxyWatch is stopped by user")
			return
		case response, ok := <-etcdCh:
			if !ok || response.Err() != nil {
				//etcd channel error happened
				blog.V(3).Infof("etcd proxy watch got etcd channel err, channel[%v], response[%s]", ok, response.Err())
				return
			}
			for _, event := range response.Events {
				targetEvent := e.eventConstruct(event, objectNewFn)
				if targetEvent == nil {
					blog.V(3).Infof("etcdProxyWatch construct event failed, event LOST")
					continue
				}
				e.filterChannel <- *targetEvent
			}
		}
	}
}

func (e *etcdProxyWatch) eventConstruct(event *clientv3.Event, objectNewFn meta.ObjectNewFn) *watch.Event {
	obj := objectNewFn()
	targetEvent := &watch.Event{}
	if event.IsCreate() {
		targetEvent.Type = watch.EventAdded
		if err := e.codec.Decode(event.Kv.Value, obj); err != nil {
			blog.V(3).Infof("etcdProxyWatch decode create event failed, %s", err)
			return nil
		}
		targetEvent.Data = obj
		if e.selector == nil {
			return targetEvent
		}
		if ok, _ := e.selector.Matchs(obj); !ok {
			blog.V(3).Infof("etcdProxyWatch filter block data, filter: %s", e.selector.String())
			return nil
		}
		return targetEvent
	}
	if event.IsModify() {
		targetEvent.Type = watch.EventUpdated
		if err := e.codec.Decode(event.Kv.Value, obj); err != nil {
			blog.V(3).Infof("etcdProxyWatch decode modifyEvent newData failed, %s", err)
			return nil
		}
		if e.selector == nil {
			targetEvent.Data = obj
			return targetEvent
		}
		cur, _ := e.selector.Matchs(obj)
		oldObj := objectNewFn()
		if err := e.codec.Decode(event.PrevKv.Value, oldObj); err != nil {
			blog.V(3).Infof("etcdProxyWatch decode modifyEvent oldData failed, %s", err)
			return nil
		}
		old, _ := e.selector.Matchs(oldObj)
		switch {
		case cur && old:
			targetEvent.Data = obj
		case cur && !old:
			targetEvent.Type = watch.EventAdded
			targetEvent.Data = obj
		case !cur && old:
			targetEvent.Type = watch.EventDeleted
			targetEvent.Data = oldObj
		case !cur && !old:
			blog.V(3).Infof("etcdProxyWatch filter etcd modify event")
			return nil
		}
		return targetEvent
	}
	if event.Type == clientv3.EventTypeDelete {
		targetEvent.Type = watch.EventDeleted
		if err := e.codec.Decode(event.PrevKv.Value, obj); err != nil {
			blog.V(3).Infof("etcdProxyWatch decode deleteEvent data failed, %s", err)
			return nil
		}
		targetEvent.Data = obj
		if e.selector == nil {
			return targetEvent
		}
		if ok, _ := e.selector.Matchs(obj); !ok {
			return nil
		}
		return targetEvent
	}
	blog.V(3).Infof("etcdProxyWatch got unexpect event: %v", event.Type)
	return nil
}
