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

package discovery

import (
	"context"
	"fmt"
	"time"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	k8scache "k8s.io/client-go/tools/cache"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/reflector"
	schedypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage/zookeeper"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

// NodeController controller for Node
type NodeController interface {
	// List list all node datas
	List(selector labels.Selector) (ret []*schedypes.Agent, err error)
	// GetByHostname xxx
	// Get node by the specified hostname
	GetByHostname(hostname string) (*schedypes.Agent, error)
}

// nodeController for dataType resource
type nodeController struct {
	cxt          context.Context
	stopFn       context.CancelFunc
	eventStorage storage.Storage      // remote event storage
	indexer      k8scache.Indexer     // indexer
	reflector    *reflector.Reflector // reflector list/watch all datas to local memory cache
}

// List list all node datas
func (s *nodeController) List(selector labels.Selector) (ret []*schedypes.Agent, err error) {
	err = ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*schedypes.Agent))
	})
	return ret, err
}

// GetByHostname xxx
func (s *nodeController) GetByHostname(hostname string) (*schedypes.Agent, error) {
	obj, exists, err := s.indexer.GetByKey(hostname)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(k8sv1.Resource("Node"), hostname)
	}
	return obj.(*schedypes.Agent), nil
}

// NewNodeController xxx
func NewNodeController(hosts []string, eventHandler reflector.EventInterface) (NodeController, error) {
	indexers := k8scache.Indexers{}

	ts := k8scache.NewIndexer(NodeObjectKeyFn, indexers)
	// create namespace client for zookeeper
	zkConfig := &zookeeper.ZkConfig{
		Hosts:         hosts,
		PrefixPath:    "/blueking",
		Name:          "node",
		Codec:         &meta.JsonCodec{},
		ObjectNewFunc: NodeObjectNewFn,
	}
	nodeclient, err := zookeeper.NewNSClient(zkConfig)
	if err != nil {
		blog.Errorf("bk-bcs mesos discovery create node pod client failed, %s", err)
		return nil, err
	}
	// create listwatcher
	listwatcher := &reflector.ListWatch{
		ListFn: func() ([]meta.Object, error) {
			return nodeclient.List(context.Background(), "agent", nil)
		},
		WatchFn: func() (watch.Interface, error) {
			return nodeclient.Watch(context.Background(), "agent", "", nil)
		},
	}
	cxt, stopfn := context.WithCancel(context.Background())
	ctl := &nodeController{
		cxt:          cxt,
		stopFn:       stopfn,
		eventStorage: nodeclient,
		indexer:      ts,
	}

	// create reflector
	ctl.reflector = reflector.NewReflector(fmt.Sprintf("Reflector-%s", zkConfig.Name), ts, listwatcher, time.Second*600,
		eventHandler)
	// sync all data object from remote event storage
	err = ctl.reflector.ListAllData()
	if err != nil {
		return nil, err
	}
	// run reflector controller
	go ctl.reflector.Run()

	return ctl, nil
}

// NodeObjectKeyFn create key for node
func NodeObjectKeyFn(obj interface{}) (string, error) {
	node, ok := obj.(*schedypes.Agent)
	if !ok {
		return "", fmt.Errorf("error object type")
	}
	return node.Key, nil
}

// NodeObjectNewFn create new Node Object
func NodeObjectNewFn() meta.Object {
	return new(schedypes.Agent)
}
