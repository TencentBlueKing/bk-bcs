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

package zookeeper

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

func convertToAppData(pushFunc PushWatchEventFn) map[int]*Layer {
	layerConfig := make(map[int]*Layer)
	dataType := &Layer{
		Index:           0,
		Name:            "DataType",
		IsData:          false,
		PushEventFunc:   nil,
		IsWatchChildren: true,
	}
	layerConfig[dataType.Index] = dataType
	ns := &Layer{
		Index:           1,
		Name:            "namespace",
		IsData:          false,
		PushEventFunc:   nil,
		IsWatchChildren: true,
	}
	layerConfig[ns.Index] = ns
	app := &Layer{
		Index:           2,
		Name:            "Application",
		IsData:          false,
		PushEventFunc:   nil,
		IsWatchChildren: true,
	}
	layerConfig[app.Index] = app
	data := &Layer{
		Index:           3,
		Name:            "TaskGroup",
		IsData:          true,
		PushEventFunc:   pushFunc,
		IsWatchChildren: false,
	}
	layerConfig[data.Index] = data
	return layerConfig
}

func convertToAppNS(pushFunc PushWatchEventFn) map[int]*Layer {
	layerConfig := make(map[int]*Layer)
	ns := &Layer{
		Index:           0,
		Name:            "namespace",
		IsData:          false,
		PushEventFunc:   nil,
		IsWatchChildren: true,
	}
	layerConfig[ns.Index] = ns
	app := &Layer{
		Index:           1,
		Name:            "Application",
		IsData:          false,
		PushEventFunc:   nil,
		IsWatchChildren: true,
	}
	layerConfig[app.Index] = app
	data := &Layer{
		Index:           2,
		Name:            "TaskGroup",
		IsData:          true,
		PushEventFunc:   pushFunc,
		IsWatchChildren: false,
	}
	layerConfig[data.Index] = data
	return layerConfig
}

func convertToAppNode(pushFunc PushWatchEventFn) map[int]*Layer {
	layerConfig := make(map[int]*Layer)
	app := &Layer{
		Index:           0,
		Name:            "Application",
		IsData:          false,
		PushEventFunc:   nil,
		IsWatchChildren: true,
	}
	layerConfig[app.Index] = app
	data := &Layer{
		Index:           1,
		Name:            "TaskGroup",
		IsData:          true,
		PushEventFunc:   pushFunc,
		IsWatchChildren: false,
	}
	layerConfig[data.Index] = data
	return layerConfig
}

// NewPodClient create pod client for bcs-scheduler in
func NewPodClient(config *ZkConfig) (*PodClient, error) {
	if len(config.Hosts) == 0 {
		return nil, fmt.Errorf("Lost zookeeper server info")
	}
	if len(config.PrefixPath) == 0 {
		return nil, fmt.Errorf("Lost zookeeper prefix path")
	}
	if config.Codec == nil {
		return nil, fmt.Errorf("Lost Codec in config")
	}
	if config.ObjectNewFunc == nil {
		return nil, fmt.Errorf("Lost Object New function")
	}
	// create client for zookeeper
	c := zkclient.NewZkClient(config.Hosts)
	if err := c.ConnectEx(time.Second * 3); err != nil {
		return nil, fmt.Errorf("podclient zookeeper connect: %s", err)
	}
	s := &PodClient{
		config:      config,
		client:      c,
		prefixPath:  config.PrefixPath,
		codec:       config.Codec,
		objectNewFn: config.ObjectNewFunc,
	}
	return s, nil
}

// PodClient implements storage interface with zookeeper client, all operations
// are based on one object data types, but data may be stored at different levels of nodes.
type PodClient struct {
	config      *ZkConfig          // configuration for nsclient
	prefixPath  string             // zookeeper storage prefix, like bcs/mesh/v1/datatype
	client      *zkclient.ZkClient // http client
	codec       meta.Codec         // json encoder/decoder
	objectNewFn meta.ObjectNewFn   // injection for new object
}

// Create implements storage interface
func (s *PodClient) Create(_ context.Context, key string, obj meta.Object, _ int) (out meta.Object, err error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("zk podclient lost object key")
	}
	if obj == nil {
		blog.V(3).Infof("zk podclient lost object data for %s", key)
		return nil, fmt.Errorf("lost object data")
	}
	fullPath := path.Join(s.prefixPath, key)
	// check history node's data
	exist, eerr := s.client.Exist(fullPath)
	if eerr != nil {
		blog.V(3).Infof("zk podclient check %s old Object failed, %s", fullPath, eerr)
		return nil, fmt.Errorf("check old Object: %s", eerr)
	}
	if exist {
		out = s.objectNewFn()
		oldRawBytes, _, gerr := s.client.GetEx(fullPath)
		if gerr != nil {
			blog.V(3).Infof("zk podclient ready to get exist target %s data failed, %s", fullPath, gerr)
		} else {
			if derr := s.codec.Decode(oldRawBytes, out); derr != nil {
				blog.V(3).Infof("zk client decode %s json failed, %s", fullPath, derr)
			}
		}
	}
	// create new data
	targetBytes, err := s.codec.Encode(obj)
	if err != nil {
		blog.V(3).Infof("zk podclient encode %s object to json failed, %s", fullPath, err)
		return nil, err
	}
	if err := s.client.Update(fullPath, string(targetBytes)); err != nil {
		blog.V(3).Infof("zk podclient created %s failed, %s", fullPath, err)
		return out, err
	}
	blog.V(3).Infof("zk podclient create %s success, prev exist: %v", fullPath, exist)
	return out, nil
}

// Delete implements storage interface
// for http api operation, there are three situations for key
// * if key likes apis/v1/dns, clean all dns data under version v1
// * if key likes apis/v1/dns/namespace/bmsf-system, delete all data under namespace
// * if key likes apis/v1/dns/namespace/bmsf-system/data, delete detail data
// in this version, no delete objects reply
func (s *PodClient) Delete(_ context.Context, key string) (obj meta.Object, err error) {
	if len(key) != 0 {
		return nil, fmt.Errorf("empty key")
	}
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("error format key, cannot end with /")
	}
	fullPath := path.Join(s.prefixPath, key)
	// delete all children under this node
	if err := s.recursiveDelete(fullPath); err != nil {
		blog.Errorf("zk podclient delete fullpath resource %s failed, %s", fullPath, err)
		return nil, err
	}
	blog.V(3).Infof("zk podclient delete %s success", fullPath)
	return nil, nil
}

func (s *PodClient) recursiveDelete(p string) error {
	children, err := s.client.GetChildren(p)
	if err != nil {
		blog.Errorf("zk podclient check %s in recursive deletion failed, %s", p, err)
		return err
	}
	if len(children) == 0 {
		if err := s.client.Del(p, 0); err != nil {
			blog.Errorf("zk podclient recursively delete %s failed, %s", p, err)
			return err
		}
		blog.V(3).Infof("zk podclient delete leaf node %s successfully", p)
		return nil
	}
	// we got children here
	for _, child := range children {
		childpath := path.Join(p, child)
		if err := s.recursiveDelete(childpath); err != nil {
			return err
		}
	}
	blog.V(5).Infof("zk podclient recursive delete %s success", p)
	return nil
}

// Watch implements storage interface
// * if key empty, watch all data
// * if key is namespace, watch all data under namespace
// * if key is namespace/name, watch detail data based-on application
// watch is Stopped when any error occure, close event channel immediately
// param cxt: context for background running, not used, only reserved now
// param version: data version, not used, reserved
// param selector: selector for target object data
// return:
//
//	watch: watch implementation for changing event, need to Stop manually
func (s *PodClient) Watch(_ context.Context, key, _ string, selector storage.Selector) (watch.Interface,
	error) {
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("error key formate")
	}
	fullpath := s.prefixPath
	if len(key) != 0 {
		fullpath = path.Join(s.prefixPath, key)
	}
	podwatch := newPodWatch(fullpath, s.config, s.client, selector)
	level := strings.Count(key, "/")
	var layerConfig map[int]*Layer
	if len(key) == 0 {
		// data type path watch, watch all pod data
		layerConfig = convertToAppData(podwatch.pushWatchEventFn)
	} else if level == 0 {
		layerConfig = convertToAppNS(podwatch.pushWatchEventFn)
	} else if level == 1 {
		layerConfig = convertToAppNode(podwatch.pushWatchEventFn)
	}
	nodewatch, err := NewNodeWatch(0, fullpath, nil, s.client, layerConfig)
	if err != nil {
		blog.V(3).Infof("zk podclient create watch for %s failed, %s", fullpath, err)
		return nil, err
	}
	podwatch.nodeWatch = nodewatch
	// running watch
	go podwatch.run()
	return podwatch, nil
}

// WatchList implements storage interface
// Watch & WatchList are the same for http api
func (s *PodClient) WatchList(ctx context.Context, key, version string, selector storage.Selector) (watch.Interface,
	error) {
	return s.Watch(ctx, key, version, selector)
}

// Get implements storage interface
// get exactly data object from http event storage. so key must be resource fullpath
// param cxt: not used
// param version: reserved for future
func (s *PodClient) Get(_ context.Context, key, _ string, ignoreNotFound bool) (meta.Object, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("lost object key")
	}
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("err key format, no / ends")
	}
	fullPath := path.Join(s.prefixPath, key)
	rawBytes, _, err := s.client.GetEx(fullPath)
	if err != nil {
		blog.V(3).Infof("zk podclient got %s data failed, %s", fullPath, err)
		if err == zkclient.ErrNoNode && ignoreNotFound {
			return nil, nil
		}
		return nil, err
	}
	if len(rawBytes) <= 23 {
		blog.V(3).Infof("zk podclient got not enough data at %s, rawBytes: %s", fullPath, string(rawBytes))
		return nil, fmt.Errorf("not enough data for decoder")
	}
	target := s.objectNewFn()
	if err := s.codec.Decode(rawBytes, target); err != nil {
		blog.V(3).Infof("zk podclient decode %s data object failed, %s", fullPath, err)
		return nil, fmt.Errorf("json decode: %s", err)
	}
	blog.V(3).Infof("zk podclient got %s success", fullPath)
	return target, nil
}

// List implements storage interface
// list all data under prefixPath/key, there may be several data type under children
// nodes of prefixPath, we use selector to filt all types of data.
// param cxt: context for cancel, not used now
// param key: key only can be empty, namespace or namespace/{name}
// param selector: data filter
func (s *PodClient) List(_ context.Context, key string, _ storage.Selector) ([]meta.Object, error) {
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("error key format")
	}
	fullPath := s.prefixPath
	keyLvl := 0
	if len(key) != 0 {
		fullPath = path.Join(s.prefixPath, key)
		num := strings.Count(key, "/")
		if num == 0 {
			keyLvl = 1
		} else if num == 1 {
			keyLvl = 2
		} else {
			return nil, fmt.Errorf("key formation error")
		}
	}
	blog.V(3).Infof("zk podclient begins to list %s", fullPath)
	// detecte all data nodes for reading detail contents
	nodes, err := s.getDataNode(fullPath, keyLvl, 3)
	if err != nil {
		blog.V(3).Infof("zk podclient got all childrens under %s failed, %s", fullPath, err)
		return nil, err
	}
	if len(nodes) == 0 {
		blog.V(3).Infof("zk podclient list %s success, got nothing.", fullPath)
		return nil, nil
	}
	var outs []meta.Object
	for _, node := range nodes {
		// get node content from zookeeper
		rawBytes, _, err := s.client.GetEx(node)
		if err != nil {
			if err == zkclient.ErrNoNode {
				// data consistency problem here, maybe node is deleted after we
				// got it, just skip disappearing nodes
				blog.V(3).Infof("zk podclient find %s lost when getting contents", node)
				continue
			}
			blog.V(3).Infof("zk podclient gets %s content failed, %s", node, err)
			return nil, fmt.Errorf("list %s err: %s", node, err)
		}
		if len(rawBytes) < 23 {
			blog.V(3).Infof("zk podclient gets %s not enough data for decode", node)
			continue
		}
		target := s.objectNewFn()
		if err := s.codec.Decode(rawBytes, target); err != nil {
			blog.V(3).Infof("zk podclient decode %s data failed, %s", node, err)
			return nil, fmt.Errorf("decode %s failed, %s", node, err)
		}
		outs = append(outs, target)
	}
	blog.V(3).Infof("zk podclient list %s success, got %d objects", fullPath, len(outs))
	return outs, nil
}

// Close storage connection, clean resource
// warnning: if you want to Close client, you have better close Watch first
func (s *PodClient) Close() {
	blog.V(3).Infof("podclient zookeeper event storage %s exit.", s.prefixPath)
	s.client.Close()
}

// getDataNode recursive get all leaf nodes
func (s *PodClient) getDataNode(node string, now, target int) ([]string, error) {
	if len(node) == 0 {
		return nil, fmt.Errorf("empty node")
	}
	childrens, err := s.client.GetChildren(node)
	if err != nil {
		blog.V(3).Infof("zk podclient check %s children nodes failed, %s", node, err)
		return nil, err
	}
	if len(childrens) == 0 {
		return nil, nil
	}
	var dataNodes []string
	done := false
	if now+1 == target {
		done = true
	}
	for _, child := range childrens {
		subNode := path.Join(node, child)
		if done {
			dataNodes = append(dataNodes, subNode)
			continue
		}
		subNodes, err := s.getDataNode(subNode, now+1, target)
		if err != nil {
			return nil, err
		}
		if len(subNodes) != 0 {
			dataNodes = append(dataNodes, subNodes...)
		}
	}
	return dataNodes, nil
}

// newPodWatch create zookeeper watch
func newPodWatch(basic string, config *ZkConfig, _ *zkclient.ZkClient, se storage.Selector) *podWatch {
	w := &podWatch{
		selfpath:    basic,
		config:      config,
		selector:    se,
		dataChannel: make(chan watch.Event, watch.DefaultChannelBuffer),
		isStop:      false,
	}
	return w
}

// podWatch wrapper for zookeeper all pod data
type podWatch struct {
	selfpath    string // zookeeper path
	config      *ZkConfig
	selector    storage.Selector
	nodeWatch   NodeWatch
	dataChannel chan watch.Event
	isStop      bool
}

// Stop watch channel
func (e *podWatch) Stop() {
	e.isStop = true
	e.nodeWatch.Stop()
	close(e.dataChannel)
}

// WatchEvent get watch events, if watch stopped/error, watch must close
// channel and exit, watch user must read channel like
// e, ok := <-channel
func (e *podWatch) WatchEvent() <-chan watch.Event {
	return e.dataChannel
}

func (e *podWatch) run() {
	// just running node watch
	e.nodeWatch.Run()
}

// pushWatchEventFn xxx
// addEvent dispatch event to client
// param eventType: zookeeper event type
// param nodepath: pod data path
// param rawBytes: raw json bytes
func (e *podWatch) pushWatchEventFn(eventType watch.EventType, nodepath string, rawBytes []byte) {
	if e.isStop {
		return
	}
	event := new(watch.Event)
	event.Type = eventType
	if len(rawBytes) != 0 {
		target := e.config.ObjectNewFunc()
		if err := e.config.Codec.Decode(rawBytes, target); err != nil {
			blog.Errorf("zk podwatch decode path %s failed: %s", nodepath, err)
			return
		}
		event.Data = target
	}
	// check deletion eventType
	if eventType == watch.EventDeleted {
		// deletion, this event is especial because
		// no data can obtain from zookeeper, we only know that node is deleted.
		// so we construct empty data for this object from nodepath
		nodes := strings.Split(nodepath, "/")
		if len(nodes) < 3 {
			blog.Errorf("zk podwatch match error path in zookeeper, %s", nodepath)
			return
		}
		target := e.config.ObjectNewFunc()
		// fix(DeveloperJim): construting name from nodepath
		target.SetNamespace(nodes[len(nodes)-3])
		parts := strings.Split(nodes[len(nodes)-1], ".")
		if len(parts) < 4 {
			blog.Errorf("zk podwatch got error formation podname: %s", nodes[len(nodes)-1])
			return
		}
		name := fmt.Sprintf("%s-%s", parts[1], parts[0])
		target.SetName(name)
		zkFlag := make(map[string]string)
		zkFlag["bk-bcs-inner-storage"] = "bkbcs-zookeeper"
		target.SetAnnotations(zkFlag)
		event.Data = target
		// only for debug
		blog.V(3).Infof("zk client podwatch %s/%s is delete, path %s", target.GetNamespace(), target.GetName(), nodepath)
	}
	if e.selector != nil {
		ok, _ := e.selector.Matchs(event.Data)
		if !ok {
			blog.V(5).Infof("zk podwatch %s discard %s by filter", e.selfpath, nodepath)
			return
		}
	}
	e.dataChannel <- *event
}
