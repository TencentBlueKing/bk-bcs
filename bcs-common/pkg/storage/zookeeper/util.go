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
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

// const (
// 	//mesos scheduler data prefix, followed by dataType, such as configmaps, applications
// 	defaultPrefixPath = "/blueking"
// 	//DefaultCRDPrefixPath default prefix for crd, followed by dataType/namespace/name
// 	defaultCRDPrefixPath = "/blueking/crd"
// 	//second for watch
// 	defaultWatchCheckPeriod = 60
// )

// PushWatchEventFn function to dispatch event
// param EventType: zk event type
// param string: event path, especially for delete event
// param []byte: detail data for object, nil when it's deletion
type PushWatchEventFn func(watch.EventType, string, []byte)

// NodeWatch interface for watch definition
type NodeWatch interface {
	GetSelfPath() string            // get self node path
	DeleteNextWatch(next NodeWatch) // delete children watch
	Run()                           // ready to start up watch
	Stop()                          // stop watch, only parent watch to stop
}

// NewNodeWatch create one nodewatch from configuration
func NewNodeWatch(index int, selfpath string, parent NodeWatch, c *zkclient.ZkClient,
	configs map[int]*Layer) (NodeWatch, error) {
	config, ok := configs[index]
	if !ok {
		return nil, fmt.Errorf("Lost layer config for node %s", selfpath)
	}
	cxt, stopFn := context.WithCancel(context.Background())
	n := &Node{
		selfpath:          selfpath,
		config:            config,
		allConfig:         configs,
		parent:            parent,
		client:            c,
		children:          make(map[string]NodeWatch),
		watchCxt:          cxt,
		stopFn:            stopFn,
		isStopped:         false,
		underSelfloop:     false,
		underChildrenloop: false,
	}
	return n, nil
}

// Layer info
type Layer struct {
	Index           int              // layer index from path
	IsData          bool             // flag to present this layer store data
	IsWatchChildren bool             // flag for watch children if not
	Name            string           // data type for this layer
	PushEventFunc   PushWatchEventFn // event dispatch function
}

// Node for zookeeper every node
type Node struct {
	selfpath          string               // node absolute path
	config            *Layer               // node config
	allConfig         map[int]*Layer       // next layer configuration
	parent            NodeWatch            // parent node for reference
	client            *zkclient.ZkClient   // zookeeper client
	childrenLock      sync.Mutex           // lock for map
	children          map[string]NodeWatch // all children's watch
	watchCxt          context.Context      // root context for self
	stopFn            context.CancelFunc   // stop func to stop all backgroup context
	isStopped         bool                 // flag for stop
	underSelfloop     bool                 // flag for selfLoop
	underChildrenloop bool                 // flag for childrenLoop
}

// GetSelfPath get self node path
func (n *Node) GetSelfPath() string {
	return n.selfpath
}

// DeleteNextWatch clen next watch when child node deletion
func (n *Node) DeleteNextWatch(next NodeWatch) {
	n.childrenLock.Lock()
	defer n.childrenLock.Unlock()
	delete(n.children, next.GetSelfPath())
}

// Run start to run all inner event loop
func (n *Node) Run() {
	go n.selfLoop()
	if n.config.IsWatchChildren {
		go n.childrenLoop()
	}
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()
	for {
		if n.isStopped {
			return
		}
		select {
		case <-n.watchCxt.Done():
			return
		case <-tick.C:
			if n.isStopped {
				return
			}
			if !n.underSelfloop {
				go n.selfLoop()
			}
			if !n.underChildrenloop && n.config.IsWatchChildren {
				go n.childrenLoop()
			}
		}
	}
}

// Stop all events & clean sub node events
func (n *Node) Stop() {
	n.childrenLock.Lock()
	defer n.childrenLock.Unlock()
	if n.isStopped {
		return
	}
	n.isStopped = true
	n.stopFn()
	blog.V(3).Infof("zk node %s ready to stop, clean all NextWatch.", n.selfpath)
	for node, next := range n.children {
		blog.V(3).Infof("zk stop Children watch %s", next.GetSelfPath())
		delete(n.children, node)
		next.Stop()
	}
}

// selfLoop check self node & ends
// for zookeeper, it's not easy to iterate all data when Synchronization,
// so after watch data nodes, we decide to force sync datas every 45 seconds
// nolint
func (n *Node) selfLoop() {
	if n.isStopped {
		return
	}
	blog.V(5).Infof("node %s is under watch", n.selfpath)
	n.underSelfloop = true
	// check node existence
	exist, err := n.client.Exist(n.selfpath)
	if err != nil {
		blog.Errorf("zk node %s Exist failed, %s", n.selfpath, err)
		if n.parent != nil {
			// this Node is parent, we can not Stop
			// and must recovery from next tick
			n.Stop()
			blog.V(3).Infof("zk node %s notify parent node clean reference", n.selfpath)
			n.parent.DeleteNextWatch(n)
		}
		n.underSelfloop = false
		return
	}
	if !exist {
		blog.V(3).Infof("zk node %s do not exist", n.selfpath)
		if n.parent != nil {
			// this Node is parent, we can not Stop
			// and must recovery from next tick
			n.Stop()
			blog.V(3).Infof("zk node %s notify parent node clean reference", n.selfpath)
			n.parent.DeleteNextWatch(n)
		}
		n.underSelfloop = false
		return
	}
	// get watch
	rawBytes, _, eventCh, err := n.client.GetW(n.selfpath)
	if err != nil {
		blog.V(3).Infof("zk client node watch %s failed, %s.", n.selfpath, err)
		if n.parent != nil {
			// this Node is parent, we can not Stop
			// and must recovery from next tick
			n.Stop()
			blog.V(3).Infof("zk node %s notify parent node clean reference", n.selfpath)
			n.parent.DeleteNextWatch(n)
		}
		n.underSelfloop = false
		return
	}
	// format to object datas
	if len(rawBytes) > 23 && n.config.IsData {
		n.config.PushEventFunc(watch.EventUpdated, n.selfpath, rawBytes)
	}
	// wait for next event
	forceTick := time.NewTicker(time.Second * 300)
	defer forceTick.Stop()
	for {
		select {
		case <-n.watchCxt.Done():
			blog.V(3).Infof("zk client node %s watch is asked to exit", n.selfpath)
			// n.underSelfloop = false
			return
		case event := <-eventCh:
			// here we need to focus on deletion & changed
			if event.Type == zkclient.EventNodeDeleted {
				blog.V(3).Infof("zk client got node %s deletion, clean watch", n.selfpath)
				if n.config.IsData {
					n.config.PushEventFunc(watch.EventDeleted, n.selfpath, nil)
				}
				// self node deletion, clean all sub watch
				n.Stop()
				if n.parent != nil {
					n.parent.DeleteNextWatch(n)
				}
				n.underSelfloop = false
				return
			}
			if event.Type == zkclient.EventNodeDataChanged {
				blog.V(3).Infof("zk client node %s data changed. refresh watch again", n.selfpath)
				go n.selfLoop()
				return
			}
			// Creation event can not happen here
		case <-forceTick.C:
			if !n.config.IsData {
				continue
			}
			blog.V(5).Infof("zk client watch %s force synchronization", n.selfpath)
			exist, err := n.client.Exist(n.selfpath)
			if err != nil {
				blog.Errorf("zk client %s force synchronization exist failed, %s, wait next force tick.", n.selfpath, err)
				continue
			}
			if !exist {
				blog.V(3).Infof("zk node %s force synchronization found no data, clean watch", n.selfpath)
				if n.parent != nil {
					// this Node is parent, we can not Stop
					// and must recovery from next tick
					n.Stop()
					blog.V(3).Infof("zk node %s notify parent node clean reference", n.selfpath)
					n.parent.DeleteNextWatch(n)
				}
				n.underSelfloop = false
				return
			}
			// get watch
			rawBytes, _, err := n.client.GetEx(n.selfpath)
			if err != nil {
				blog.Errorf("zk client node watch %s forceSync get failed, %s.", n.selfpath, err)
				continue
			}
			n.config.PushEventFunc(watch.EventSync, n.selfpath, rawBytes)
		}
	}
}

// childrenLoop check self node & ends
func (n *Node) childrenLoop() {
	if n.isStopped {
		return
	}
	n.underChildrenloop = true
	// check node existence
	children, evCh, err := n.client.WatchChildren(n.selfpath)
	if err != nil {
		blog.Errorf("zk node %s childrenLoop failed, %s", n.selfpath, err)
		if n.parent != nil {
			n.Stop()
			n.parent.DeleteNextWatch(n)
		}
		n.underChildrenloop = false
		return
	}
	n.childrenLock.Lock()
	// get all node from local children map
	localChildren := make(map[string]bool)
	for key := range n.children {
		localChildren[key] = true
	}
	// find out new children, create watch for it
	for _, child := range children {
		node := path.Join(n.selfpath, child)
		if _, ok := n.children[node]; !ok {
			nodeWatch, err := NewNodeWatch(n.config.Index+1, node, n, n.client, n.allConfig)
			if err != nil {
				blog.Error("zk create node watch for %s failed, %s", node, err)
				continue
			}
			n.children[node] = nodeWatch
			// starting children node watch
			go nodeWatch.Run()
		} else {
			// clean exist key for searching lost key in zookeeper
			delete(localChildren, node)
		}
	}
	// clean keys that lost in zookeeper but exist in local children map
	for key := range localChildren {
		nodeWatch := n.children[key]
		delete(n.children, key)
		nodeWatch.Stop()
		blog.V(3).Infof("zk NodeWatch %s lost in zookeeper, but NodeWatch got no event. ready to clean it.", key)
	}
	n.childrenLock.Unlock()
	select {
	case <-n.watchCxt.Done():
		blog.V(3).Infof("zk NodeWatch %s is asked to exit", n.selfpath)
		return
	case event := <-evCh:
		if event.Type == zkclient.EventNodeChildrenChanged {
			go n.childrenLoop()
			return
		}
		// we do not sure that zookeeper client will report other events.
		// so we consider children loop here will exit. we will recovery
		// children loop by time ticker under Run()
		blog.V(3).Infof("zk %s NodeWatch detects children loop exit, eventType: %d, state: %d, "+
			"Err: %s, server: %s, path: %s", n.selfpath, event.Type, event.State, event.Err, event.Server, event.Path)
		n.underChildrenloop = false
		return
	}
}
