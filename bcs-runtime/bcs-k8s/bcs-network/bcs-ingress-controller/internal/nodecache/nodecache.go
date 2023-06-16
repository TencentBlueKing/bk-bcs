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

package nodecache

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloudnode"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// NodeCache cache node info
type NodeCache struct {
	isInit            bool
	nodeInfoMapByName *sync.Map
	nodeInfoMapByIP   *sync.Map
	k8sClient         client.Client
	nodeClient        cloudnode.NodeClient

	sync.Mutex
}

// NewNodeCache return new NodeCache
func NewNodeCache(k8sClient client.Client, nodeClient cloudnode.NodeClient) *NodeCache {
	cache := &NodeCache{
		nodeInfoMapByName: &sync.Map{},
		nodeInfoMapByIP:   &sync.Map{},
		isInit:            false,
		k8sClient:         k8sClient,
		nodeClient:        nodeClient,
	}
	go func() {
		timeTicker := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-timeTicker.C:
				cache.nodeInfoMapByName.Range(func(key, value interface{}) bool {
					blog.V(4).Infof("node cache info: %v %v", key, value)
					return true
				})
			}
		}
	}()

	return cache
}

// SetNodeIps set node ip to cache
func (n *NodeCache) SetNodeIps(node corev1.Node, nodeIPs []string) {
	// init失败时仍尝试保存数据
	if err := n.checkInit(); err != nil {
		err = errors.Wrapf(err, "init node cache failed")
		blog.Errorf("%s", err.Error())
	}

	n.nodeInfoMapByName.Store(node.GetName(), nodeIPs)
	n.nodeInfoMapByIP.Store(getNodeInternalIP(node), nodeIPs)
}

// GetNodeExternalIPsByName get node ip from cache
func (n *NodeCache) GetNodeExternalIPsByName(nodeName string) ([]string, error) {
	if n.isInit == false {
		blog.Errorf("try to get node ip without init")
		return nil, errors.New("try to get node ip without init")
	}
	val, ok := n.nodeInfoMapByName.Load(nodeName)
	if !ok {
		metrics.IncreaseNodeNotFoundCounter(nodeName)
		err := errors.Errorf("node[%s] external ips is empty", nodeName)
		blog.Errorf("%s", err.Error())
		return nil, err
	}

	nodeIPs, ok := val.([]string)
	if !ok {
		err := errors.Errorf("unknown type in node cache, value: %+v", val)
		blog.Errorf("%s", err.Error())
		return nil, err
	}
	return nodeIPs, nil
}

// GetNodeExternalIPsByIP get node ip from cache
func (n *NodeCache) GetNodeExternalIPsByIP(nodeInternalIP string) ([]string, error) {
	if n.isInit == false {
		blog.Errorf("try to get node ip without init")
		return nil, errors.New("try to get node ip without init")
	}
	val, ok := n.nodeInfoMapByIP.Load(nodeInternalIP)
	if !ok {
		metrics.IncreaseNodeNotFoundCounter(nodeInternalIP)
		err := errors.Errorf("node[%s] external ips is empty", nodeInternalIP)
		blog.Errorf("%s", err.Error())
		return nil, err
	}

	nodeIPs, ok := val.([]string)
	if !ok {
		err := errors.Errorf("unknown type in node cache, value: %+v", val)
		blog.Errorf("%s", err.Error())
		return nil, err
	}
	return nodeIPs, nil
}

// initCache use node list to build cache
func (n *NodeCache) initCache() error {
	nodeList := &corev1.NodeList{}
	if err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		return n.k8sClient.List(context.TODO(), nodeList)
	}); err != nil {
		blog.Errorf("get node list failed, err: %s", err.Error())
		return err
	}

	for _, node := range nodeList.Items {
		externalIPList, err := n.nodeClient.GetNodeExternalIpList(&node)
		if err != nil {
			blog.Errorf("get node[%s] external ip list failed, err: %s", node.GetName(), err.Error())
			continue
		}
		n.SetNodeIps(node, externalIPList)
	}

	return nil
}

func (n *NodeCache) checkInit() error {
	if !n.isInit {
		n.Lock()
		defer n.Unlock()
		if !n.isInit {
			// 避免无限递归
			n.isInit = true
			if err := n.initCache(); err != nil {
				n.isInit = false
				return err
			}
		}
	}
	return nil
}

func getNodeInternalIP(node corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}

	return ""
}
