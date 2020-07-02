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

package zk

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"strings"
	"time"
)

const BranchSymbol = "*"

type Zk interface {
	Watch(paths []string) chan *Event
	List(paths []string) []*Endpoint
	Close()
}

type zk struct {
	zkClient *zkclient.ZkClient
	event    chan *Event

	ctx    context.Context
	cancel context.CancelFunc
}

func New(cfg *config.Config) (z Zk, err error) {
	zkClient := zkclient.NewZkClient(strings.Split(cfg.BCSZk, ","))
	if err = zkClient.Connect(); err != nil {
		return
	}
	z = &zk{
		zkClient: zkClient,
	}
	return
}

func (z *zk) List(paths []string) (endpoint []*Endpoint) {
	endpoint = make([]*Endpoint, 0)
	for _, pathRaw := range paths {
		pair := strings.Split(pathRaw, ":")
		if len(pair) < 2 {
			continue
		}
		endpoint = append(endpoint, z.list(completePath(pair[1]), "", pair[0])...)
	}
	return
}

func (z *zk) list(subPath, head, name string) (endpoint []*Endpoint) {
	endpoint = make([]*Endpoint, 0)
	index := strings.Index(subPath, BranchSymbol)
	if index == -1 {
		path := joinPath(head, subPath)
		nodes, err := z.zkClient.GetChildren(strings.TrimRight(path, "/"))
		if err != nil && err != zkclient.ErrNoNode {
			blog.Errorf("list zk endpoint failed %s: %v", path, err)
			return
		}

		for _, node := range nodes {
			info, err := z.zkClient.Get(joinPath(path, node))
			if err != nil {
				blog.Errorf("get zk endpoint failed: %v", err)
				continue
			}

			var serverInfo types.ServerInfo
			if err = codec.DecJson([]byte(info), &serverInfo); err != nil {
				blog.Errorf("decode zk endpoint info failed: %v", err)
				continue
			}
			endpoint = append(endpoint, &Endpoint{
				IP:     serverInfo.IP,
				Port:   serverInfo.MetricPort,
				Scheme: serverInfo.Scheme,
				Name:   name,
				Path:   path,
			})
		}
		return
	}

	head = joinPath(head, subPath[:index])
	nodes, err := z.zkClient.GetChildren(strings.TrimRight(head, "/"))
	if err != nil && err != zkclient.ErrNoNode {
		blog.Errorf("list zk endpoint failed %s: %v", head, err)
		return
	}
	for _, node := range nodes {
		endpoint = append(endpoint, z.list(subPath[index+1:], joinPath(head, node), name)...)
	}
	return
}

func (z *zk) Watch(paths []string) chan *Event {
	z.ctx, z.cancel = context.WithCancel(context.Background())
	z.event = make(chan *Event, 100)
	for _, pathRaw := range paths {
		pair := strings.Split(pathRaw, ":")
		if len(pair) < 2 {
			continue
		}
		go z.watchBranch(z.ctx, completePath(pair[1]), "", pair[0])
	}
	return z.event
}

func (z *zk) watchBranch(ctx context.Context, subPath, head, name string) {
	index := strings.Index(subPath, BranchSymbol)
	subCtx, subCancel := context.WithCancel(ctx) //nolint

	if index == -1 {
		go z.watchLeaf(subCtx, joinPath(head, subPath), name)
		return //nolint
	}

	head = joinPath(head, subPath[:index])
	blog.Infof("begin watch zk children of %s", head)

	nodes := make(map[string]bool)
	nodesCancel := make(map[string]context.CancelFunc)

	for {
		newList, event, err := z.zkClient.WatchChildren(strings.TrimRight(head, "/"))
		lose := freshNodes(nodes, newList)
		for _, loseNode := range lose {
			if nodeCancel := nodesCancel[loseNode]; nodeCancel != nil {
				nodeCancel()
			}
			delete(nodesCancel, loseNode)
		}

		if err != nil {
			if err == zkclient.ErrNoNode {
				time.Sleep(5 * time.Second)
				continue
			}
			blog.Errorf("watch zk children of %s failed", head)
			subCancel()
			return
		}

		for node := range nodes {
			if nodeCancel := nodesCancel[node]; nodeCancel == nil {
				nodeCtx, nodeCancel := context.WithCancel(subCtx)
				nodesCancel[node] = nodeCancel
				go z.watchBranch(nodeCtx, subPath[index+1:], joinPath(head, node), name)
			}
		}

		select {
		case <-event:
		case <-ctx.Done():
			blog.Infof("end watch zk children of %s", head)
			subCancel()
			return
		}
	}
}

func (z *zk) watchLeaf(ctx context.Context, path, name string) {
	blog.Infof("begin watch zk children of %s", path)

	endpointIds := make(map[string]bool)
	endpoints := make(map[string]*Endpoint)
	for {
		nodes, event, err := z.zkClient.WatchChildren(strings.TrimRight(path, "/"))
		if err != nil {
			if err == zkclient.ErrNoNode {
				time.Sleep(5 * time.Second)
				continue
			}
			blog.Errorf("watch zk endpoint failed: %s", path)
			return
		}

		lose := freshNodes(endpointIds, nodes)
		if len(lose) > 0 {
			endpointLoseEvent := &Event{Type: EventNodeDown, Endpoints: make([]*Endpoint, 0)}
			for _, node := range lose {
				endpointLoseEvent.Endpoints = append(endpointLoseEvent.Endpoints, endpoints[node])
				delete(endpoints, node)
			}
			z.event <- endpointLoseEvent
		}

		endpointUpdateEvent := &Event{Type: EventNodeUp, Endpoints: make([]*Endpoint, 0)}
		for _, node := range nodes {
			info, err := z.zkClient.Get(joinPath(path, node))
			if err != nil {
				blog.Errorf("get zk endpoint failed: %v", err)
				continue
			}

			var serverInfo types.ServerInfo
			if err = codec.DecJson([]byte(info), &serverInfo); err != nil {
				blog.Errorf("decode zk endpoint info failed: %v", err)
				continue
			}

			endpoints[node] = &Endpoint{
				IP:     serverInfo.IP,
				Port:   serverInfo.MetricPort,
				Scheme: serverInfo.Scheme,
				Name:   name,
				Path:   path,
			}
			endpointUpdateEvent.Endpoints = append(endpointUpdateEvent.Endpoints, endpoints[node])
		}
		z.event <- endpointUpdateEvent

		select {
		case <-event:
		case <-ctx.Done():
			blog.Infof("end watch zk children of %s", path)
			return
		}
	}
}

func (z *zk) Close() {
	if z.cancel != nil {
		z.cancel()
	}
}

func completePath(path string) string {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func joinPath(leftPath, rightPath string) string {
	return strings.TrimRight(leftPath, "/") + "/" + strings.TrimLeft(rightPath, "/")
}

func freshNodes(nodes map[string]bool, newList []string) (lose []string) {
	lose = make([]string, 0)

	for node := range nodes {
		nodes[node] = false
	}
	for _, node := range newList {
		nodes[node] = true
	}
	for node := range nodes {
		if !nodes[node] {
			lose = append(lose, node)
		}
	}
	for _, node := range lose {
		delete(nodes, node)
	}
	return
}

type Event struct {
	Type      EventType
	Endpoints []*Endpoint
}

type Endpoint struct {
	IP     string
	Port   uint
	Scheme string
	Name   string
	Path   string
}

type EventType int

const (
	EventNodeUp EventType = iota
	EventNodeDown
)

func (et EventType) String() string {
	return eventTypeNames[et]
}

var (
	eventTypeNames = map[EventType]string{
		EventNodeUp:   "ZkNodeUp",
		EventNodeDown: "ZkNodeDown",
	}
)
