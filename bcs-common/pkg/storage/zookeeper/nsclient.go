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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
	"path"
	"strings"
	"time"
)

//ZkConfig only data type node config
type ZkConfig struct {
	Hosts         []string         //http api host link, ip:host
	PrefixPath    string           //basic path for namespace data
	Name          string           //data type name
	Codec         meta.Codec       //Codec for encoder & decoder
	ObjectNewFunc meta.ObjectNewFn //object pointer for serialization
}

func convertToNodeConfig(config *ZkConfig, pushEventfn PushWatchEventFn) map[int]*Layer {
	nodeConfig := make(map[int]*Layer)
	datalayer := &Layer{
		Index:           0,
		Name:            config.Name,
		IsData:          true,
		IsWatchChildren: false,
		PushEventFunc:   pushEventfn,
	}
	nodeConfig[datalayer.Index] = datalayer
	return nodeConfig
}

func convertToNamespaceConfig(config *ZkConfig, pushEventfn PushWatchEventFn) map[int]*Layer {
	namespaceConfig := make(map[int]*Layer)
	nslayer := &Layer{
		Index:           0,
		Name:            "namespace",
		IsData:          false,
		PushEventFunc:   nil,
		IsWatchChildren: true,
	}
	namespaceConfig[nslayer.Index] = nslayer
	//detail data node
	datalayer := &Layer{
		Index:           1,
		Name:            config.Name,
		IsData:          true,
		IsWatchChildren: false,
		PushEventFunc:   pushEventfn,
	}
	namespaceConfig[datalayer.Index] = datalayer
	return namespaceConfig
}

func convertToTypeConfig(config *ZkConfig, pushEventfn PushWatchEventFn) map[int]*Layer {
	layerConfig := make(map[int]*Layer)
	//create first layer for namespace node
	typelayer := &Layer{
		Index:           0,
		Name:            config.Name,
		IsData:          false,
		PushEventFunc:   nil,
		IsWatchChildren: true,
	}
	layerConfig[typelayer.Index] = typelayer
	nslayer := &Layer{
		Index:           1,
		Name:            "namespace",
		IsData:          false,
		PushEventFunc:   nil,
		IsWatchChildren: true,
	}
	layerConfig[nslayer.Index] = nslayer
	//detail data node
	datalayer := &Layer{
		Index:           2,
		Name:            config.Name,
		IsData:          true,
		IsWatchChildren: false,
		PushEventFunc:   pushEventfn,
	}
	layerConfig[datalayer.Index] = datalayer
	return layerConfig
}

//NewStorage create etcd accessor implemented storage interface
func NewStorage(config *ZkConfig) (storage.Storage, error) {
	return NewNSClient(config)
}

//NewNSClient create new client for namespace-based datas that store in zookeeper.
//namespace-based datas store in path {namespace}/{name}, namespace nodes store
//nothing, only name nodes store json data
func NewNSClient(config *ZkConfig) (*NSClient, error) {
	if len(config.Hosts) == 0 {
		return nil, fmt.Errorf("Lost http api hosts info")
	}
	//create client for zookeeper
	c := zkclient.NewZkClient(config.Hosts)
	if err := c.ConnectEx(time.Second * 3); err != nil {
		return nil, fmt.Errorf("zookeeper connect: %s", err)
	}
	s := &NSClient{
		config:      config,
		client:      c,
		codec:       config.Codec,
		objectNewFn: config.ObjectNewFunc,
		prefixPath:  config.PrefixPath,
	}
	return s, nil
}

//NSClient implementation storage interface with zookeeper client, all operations
//are based on namespace-like data types. All data store with namespace feature.
//so full paths of objects are like /prefix/{dataType}/{namespace}/{name},
//all object json contents are hold by name node
type NSClient struct {
	config      *ZkConfig          //configuration for nsclient
	client      *zkclient.ZkClient //http client
	codec       meta.Codec         //json Codec for object
	objectNewFn meta.ObjectNewFn   //create new object for json Decode
	prefixPath  string             //zookeeper storage prefix, like /bcs/mesh/v1/discovery
}

//Create implements storage interface
//param cxt: context for use decline Creation, not used
//param key: http full api path
//param obj: object for creation
//param ttl: second for time-to-live, not used
//return out: exist object data
func (s *NSClient) Create(cxt context.Context, key string, obj meta.Object, ttl int) (out meta.Object, err error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("zk client lost object key")
	}
	if obj == nil {
		blog.V(3).Infof("zk client lost object data for %s", key)
		return nil, fmt.Errorf("lost object data")
	}
	fullPath := path.Join(s.prefixPath, key)
	//check history node's data
	exist, err := s.client.Exist(fullPath)
	if err != nil {
		blog.V(3).Infof("zk client check %s old Object failed, %s", fullPath, err)
		return nil, fmt.Errorf("check old Object: %s", err)
	}
	if exist {
		out = s.objectNewFn()
		oldRawBytes, _, gerr := s.client.GetEx(fullPath)
		if gerr != nil {
			blog.V(3).Infof("zk client ready to get exist target %s data failed, %s", fullPath, gerr)
		} else {
			if derr := s.codec.Decode(oldRawBytes, out); derr != nil {
				blog.V(3).Infof("zk client decode %s json failed, %s", fullPath, derr)
			}
		}
	}
	//create new data
	targetBytes, err := s.codec.Encode(obj)
	if err != nil {
		blog.V(3).Infof("zk client encode %s object to json failed, %s", fullPath, err)
		return nil, err
	}
	if err := s.client.Update(fullPath, string(targetBytes)); err != nil {
		blog.V(3).Infof("zk client created %s failed, %s", fullPath, err)
		return out, err
	}
	blog.V(3).Infof("zk client create %s success, prev exist: %v", fullPath, exist)
	return out, nil
}

//Delete implements storage interface
//for http api operation, there are three situations for key
//* if key likes apis/v1/dns, clean all dns data under version v1
//* if key likes apis/v1/dns/namespace/bmsf-system, delete all data under namespace
//* if key likes apis/v1/dns/namespace/bmsf-system/data, delete detail data
// in this version, no delete objects reply
func (s *NSClient) Delete(ctx context.Context, key string) (obj meta.Object, err error) {
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("error format key, cannot end with /")
	}
	fullPath := s.prefixPath
	if len(key) == 0 {
		blog.Warnf("zk namespace client ready to clean all data under %s", s.prefixPath)
	} else {
		fullPath = path.Join(s.prefixPath, key)
	}
	if err := s.client.Del(fullPath, 0); err != nil {
		blog.V(3).Infof("zk client delete %s failed, %s", fullPath, err)
		return nil, err
	}
	blog.V(3).Infof("zk delete %s success", fullPath)
	return nil, nil
}

//Watch implements storage interface
//* if key empty, watch all data
//* if key is namespace, watch all data under namespace
//* if key is namespace/name, watch detail data
//watch is Stopped when any error occure, close event channel immediatly
//param cxt: context for background running, not used, only reserved now
//param version: data version, not used, reserved
//param selector: labels selector
//return:
//  watch: watch implementation for changing event, need to Stop manually
func (s *NSClient) Watch(cxt context.Context, key, version string, selector storage.Selector) (watch.Interface, error) {
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("error key formate")
	}
	fullpath := s.prefixPath
	if len(key) != 0 {
		fullpath = path.Join(s.prefixPath, key)
	}
	nswatch := newNSWatch(fullpath, s.config, s.client, selector)
	var layer map[int]*Layer
	level := strings.Count(key, "/")
	if len(key) == 0 {
		layer = convertToTypeConfig(s.config, nswatch.pushEventFunc)
	} else if level == 0 {
		layer = convertToNamespaceConfig(s.config, nswatch.pushEventFunc)
	} else if level == 1 {
		layer = convertToNodeConfig(s.config, nswatch.pushEventFunc)
	}
	nswatch.setLayerConfig(layer)
	//running watch
	go nswatch.run()
	return nswatch, nil
}

//WatchList implements storage interface
//Watch & WatchList are the same for http api
func (s *NSClient) WatchList(ctx context.Context, key, version string, selector storage.Selector) (watch.Interface, error) {
	return s.Watch(ctx, key, version, selector)
}

//Get implements storage interface
//get exactly data object from http event storage. so key must be resource fullpath
//param cxt: not used
//param version: reserved for future
func (s *NSClient) Get(cxt context.Context, key, version string, ignoreNotFound bool) (meta.Object, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("lost object key")
	}
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("err key format, no / ends")
	}
	fullPath := path.Join(s.prefixPath, key)
	rawBytes, _, err := s.client.GetEx(fullPath)
	if err != nil {
		blog.V(3).Infof("zk client got %s data failed, %s", fullPath, err)
		if err == zkclient.ErrNoNode && ignoreNotFound {
			return nil, nil
		}
		return nil, err
	}
	if len(rawBytes) < 4 {
		blog.V(3).Infof("zk client got nothing in %s.", fullPath)
		return nil, fmt.Errorf("data content empty")
	}
	target := s.objectNewFn()
	if err := s.codec.Decode(rawBytes, target); err != nil {
		blog.V(3).Infof("zk client decode %s data object failed, %s", fullPath, err)
		return nil, fmt.Errorf("json decode: %s", err)
	}
	blog.V(3).Infof("zk client got %s success", fullPath)
	return target, nil
}

//List implements storage interface
//list namespace-based data or all data, detail node should use Get
//* if key is empty, list all data under prefix, data must be target object
//* if key is not empty, list all data under prefix/key, data must be target object.
//if one node errors, we consider all errors. List Function only list leaf data nodes
func (s *NSClient) List(cxt context.Context, key string, selector storage.Selector) ([]meta.Object, error) {
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("error key format")
	}
	fullPath := s.prefixPath
	if len(key) != 0 {
		fullPath = path.Join(s.prefixPath, key)
	}
	blog.V(3).Infof("zk ready to list all objects under %s...", fullPath)
	//detecte all leaf nodes for reading detail contents
	nodes, err := s.getLeafNode(fullPath)
	if err != nil {
		blog.V(3).Infof("zk client got all childrens under %s failed, %s", fullPath, err)
		return nil, err
	}
	if len(nodes) == 0 {
		blog.V(3).Infof("zk client list %s success, got nothing.", fullPath)
		return nil, nil
	}
	var outs []meta.Object
	for _, node := range nodes {
		//get node content from zookeeper
		rawBytes, _, err := s.client.GetEx(node)
		if err != nil {
			if err == zkclient.ErrNoNode {
				//data consistency problem here, maybe node is deleted after we
				//got it, just skip disappearing nodes
				blog.V(3).Infof("zk namespace client find %s lost when getting contents", node)
				continue
			}
			blog.V(3).Infof("zk client gets %s content failed, %s", node, err)
			return nil, fmt.Errorf("list %s err: %s", node, err)
		}
		target := s.objectNewFn()
		if err := s.codec.Decode(rawBytes, target); err != nil {
			blog.V(3).Infof("zk client decode %s data failed, %s", node, err)
			return nil, fmt.Errorf("decode %s failed, %s", node, err)
		}
		if selector != nil {
			ok, _ := selector.Matchs(target)
			if !ok {
				continue
			}
		}
		outs = append(outs, target)
	}
	blog.V(3).Infof("zk list %s success, got %d objects", fullPath, len(outs))
	return outs, nil
}

//Close storage conenction, clean resource
func (s *NSClient) Close() {
	blog.V(3).Infof("zookeeper event storage %s exit.", s.prefixPath)
	s.client.Close()
}

//getLeafNode recursive get all leaf nodes, pay more attension
func (s *NSClient) getLeafNode(node string) ([]string, error) {
	if len(node) == 0 {
		return nil, fmt.Errorf("empty node")
	}
	childrens, err := s.client.GetChildren(node)
	if err != nil {
		blog.V(3).Infof("zk client check %s children nodes failed, %s", node, err)
		return nil, err
	}
	if len(childrens) == 0 {
		return nil, nil
	}
	var leafNodes []string
	for _, child := range childrens {
		subNode := path.Join(node, child)
		subNodes, err := s.getLeafNode(subNode)
		if err != nil {
			return nil, err
		}
		if len(subNodes) == 0 {
			//based-on namespace data path rule,
			//check if leafNode is a really data node
			//for example:
			//  prefix is /data/application
			//  data node is /data/application/namespace/mydata
			//  sub is namespace/mydata
			sub := subNode[len(s.prefixPath)+1:]
			if len(strings.Split(sub, "/")) == 2 {
				leafNodes = append(leafNodes, subNode)
			}
		} else {
			leafNodes = append(leafNodes, subNodes...)
		}
	}
	return leafNodes, nil
}

//newZookeeperWatch create zookeeper watch
func newNSWatch(basic string, config *ZkConfig, c *zkclient.ZkClient, s storage.Selector) *nsWatch {
	w := &nsWatch{
		selfpath:    basic,
		config:      config,
		client:      c,
		dataChannel: make(chan watch.Event, watch.DefaultChannelBuffer),
		isStop:      false,
		selector:    s,
	}
	return w
}

//NSWatch implements watch interface, and wrap for zookeeper layer watch
type nsWatch struct {
	selfpath    string
	config      *ZkConfig
	client      *zkclient.ZkClient
	nodeWatch   NodeWatch
	layers      map[int]*Layer
	dataChannel chan watch.Event
	isStop      bool
	selector    storage.Selector
}

func (e *nsWatch) setLayerConfig(c map[int]*Layer) {
	e.layers = c
}

//Stop watch channel
func (e *nsWatch) Stop() {
	e.isStop = true
	e.nodeWatch.Stop()
	close(e.dataChannel)
}

func (e *nsWatch) run() error {
	nodeWatch, err := NewNodeWatch(0, e.selfpath, nil, e.client, e.layers)
	if err != nil {
		blog.Errorf("zk namespace client running watch for %s failed, %s", e.selfpath, err)
		return err
	}
	e.nodeWatch = nodeWatch
	nodeWatch.Run()
	return nil
}

//WatchEvent get watch events, if watch stopped/error, watch must close
// channel and exit, watch user must read channel like
// e, ok := <-channel
func (e *nsWatch) WatchEvent() <-chan watch.Event {
	return e.dataChannel
}

//pushEventFunc callback event when object data trigger
//warnning(DeveloperJim): nsWatch do not support detail data node watch
func (e *nsWatch) pushEventFunc(eventType watch.EventType, nodepath string, rawBytes []byte) {
	if e.isStop {
		return
	}
	event := new(watch.Event)
	event.Type = eventType
	if len(rawBytes) != 0 {
		target := e.config.ObjectNewFunc()
		if err := e.config.Codec.Decode(rawBytes, target); err != nil {
			blog.Errorf("zk namespace client decode path %s failed: %s", nodepath, err)
			return
		}
		event.Data = target
	}
	//check delete eventType
	if eventType == watch.EventDeleted {
		//deletion, this event is especial because
		//no data can obtain from zookeeper, we only know node is deleted.
		//so we construct empty data for this object from nodepath
		nodes := strings.Split(nodepath, "/")
		if len(nodes) < 2 {
			blog.Errorf("zk namespace watch match error path in zookeeper, %s", nodepath)
			return
		}
		target := e.config.ObjectNewFunc()
		target.SetNamespace(nodes[len(nodes)-2])
		target.SetName(nodes[len(nodes)-1])
		zkFlag := make(map[string]string)
		zkFlag["bk-bcs-inner-storage"] = "bkbcs-zookeeper"
		target.SetAnnotations(zkFlag)
		event.Data = target
	}
	if e.selector != nil {
		ok, _ := e.selector.Matchs(event.Data)
		if !ok {
			blog.V(5).Infof("zk namespace %s watch discard %s by filter", e.selfpath, nodepath)
			return
		}
	}
	e.dataChannel <- *event
}
