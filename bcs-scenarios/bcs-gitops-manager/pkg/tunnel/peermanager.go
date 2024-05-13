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

package tunnel

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"go-micro.dev/v4/registry"

	localcommon "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

// PeerManagerOptions config options
type PeerManagerOptions struct {
	// cxt for graceful exit
	Context context.Context
	// PeerID uniq ID
	PeerID string
	// peer token for validation
	PeerToken string
	// connect URL for peer inter-connection
	PeerConnectURL string
	// ClientTLS for peer inter-connection
	PeerTLS *tls.Config
	// Peer service name in registry
	PeerServiceName string
	// discovery all peer info from go-micro registry
	Discovery registry.Registry
	// basic tunnel for peer inter-connection
	Tunnel *websocketDialer.Server
}

// NewPeerManager return a new PeerManager instance
func NewPeerManager(opt *PeerManagerOptions) *PeerManager {
	return &PeerManager{
		peers:   make(map[string]bool),
		ready:   false,
		options: opt,
		protocol: func() string {
			if opt.PeerTLS == nil {
				return "ws://"
			}
			return "wss://"
		}(),
	}
}

// PeerManager discovery all peers from registry,
// then dynamically update peer information to tunnel server.
// PeerManager can use alone.
type PeerManager struct {
	sync.Mutex
	peers    map[string]bool
	ready    bool
	protocol string
	options  *PeerManagerOptions
}

// Start the PeerManager, non-blocking start
func (pm *PeerManager) Start() error {
	if len(pm.options.PeerServiceName) == 0 {
		return fmt.Errorf("lost peer service name")
	}
	if pm.options.Discovery == nil {
		return fmt.Errorf("lost micro service registry")
	}
	go pm.peerSyncLoop()
	return nil
}

func (pm *PeerManager) peerSyncLoop() {
	// get all peers first
	services, err := pm.options.Discovery.GetService(pm.options.PeerServiceName)
	if err != nil {
		blog.Errorf("PeerManager get service %s failed in syncLoop, %s. try after back-off",
			pm.options.PeerServiceName, err.Error())
		time.Sleep(time.Second * 3)
		go pm.peerSyncLoop()
		return
	}
	if services != nil {
		blog.Infof("PeerManager get service %s with %d services",
			pm.options.PeerServiceName, len(services))
		pm.updatePeerByServices(services)
	}

	// watch all changes
	watcher, err := pm.options.Discovery.Watch(registry.WatchService(pm.options.PeerServiceName))
	if err != nil {
		blog.Errorf("PeerManager watch service %s failed in syncLoop, %s. try after back-off",
			pm.options.PeerServiceName, err.Error())
		time.Sleep(time.Second * 3)
		go pm.peerSyncLoop()
		return
	}
	event := make(chan struct{})
	defer func() {
		watcher.Stop()
		close(event)
	}()

	go pm.handleWatchEvent(watcher, event)
	for {
		select {
		case <-pm.options.Context.Done():
			blog.Infof("PeerManager prepare to exit, stop current watcher")
			return
		case <-event:
			// handle read service
			blog.Infof("PeerManager received %s event, ready to update peer information",
				pm.options.PeerServiceName)
			go pm.peerSyncLoop()
			return
		}
	}
}

func (pm *PeerManager) handleWatchEvent(watcher registry.Watcher, ch chan<- struct{}) {
	results, err := watcher.Next()
	if err != nil {
		if err == registry.ErrWatcherStopped {
			blog.Errorf("PeerManager discovery watch is stopped, ready to exit")
			return
		}
		// when watcher was cancel, err is 'could not get next'
		blog.Errorf("PeerManager discovery watch faild, %s. GetService & recover watch", err.Error())
		return
	}
	blog.Infof("PeerManager watch %s event %s, services details: %+v",
		pm.options.PeerServiceName, results.Action, results.Service.Nodes)
	ch <- struct{}{}
}

func (pm *PeerManager) updatePeerByServices(services []*registry.Service) {
	nodes := make([]*registry.Node, 0)
	for _, svc := range services {
		blog.V(3).Infof("merge discovery nodes %v version %s", svc.Nodes, svc.Version)
		nodes = append(nodes, svc.Nodes...)
	}

	peers := make([]string, 0)
	for _, node := range nodes {
		addr, err := getHTTPEndpointFromMeta(node)
		if err != nil {
			blog.Warnf("node %s Endpoint information convert failed, %s", err.Error())
			continue
		}
		peers = append(peers, addr)
	}

	blog.Infof("PeerManager discover services %s with all nodes %v",
		pm.options.PeerServiceName, peers)

	pm.syncPeersToTunnelServer(peers)
}

// addRemovePeers add tunnels with new peers each other, remove tunnels from deleted peers
func (pm *PeerManager) syncPeersToTunnelServer(peers []string) {
	if len(peers) == 0 {
		blog.Errorf("PeerManager discovery self peer failed, wait next event to recovery")
		return
	}

	pm.Lock()
	defer pm.Unlock()

	newSet := map[string]bool{}
	ready := false
	for _, peer := range peers {
		if peer == pm.options.PeerID {
			ready = true
		} else {
			newSet[peer] = true
		}
	}

	newPeers, outDatedPeers, _ := diff(newSet, pm.peers)
	// add new peers
	for _, peer := range newPeers {
		blog.Infof("PeerManager add new peer %s", peer)
		pm.options.Tunnel.AddPeer(
			pm.protocol+peer+pm.options.PeerConnectURL,
			peer, pm.options.PeerToken, pm.options.PeerTLS)
	}
	// remove deleted peers
	for _, peer := range outDatedPeers {
		blog.Infof("PeerManager clean outdated peer %s", peer)
		pm.options.Tunnel.RemovePeer(peer)
	}

	pm.peers = newSet
	pm.ready = ready
}

// diff just compare and diff two map
func diff(desired, actual map[string]bool) ([]string, []string, []string) {
	var same, news, outdated []string
	for key := range desired {
		if actual[key] {
			same = append(same, key)
		} else {
			news = append(news, key)
		}
	}
	for key := range actual {
		if !desired[key] {
			outdated = append(outdated, key)
		}
	}
	return news, outdated, same
}

func getHTTPEndpointFromMeta(node *registry.Node) (string, error) {
	address := node.Address
	strs := strings.Split(address, ":")
	if len(strs) != 2 {
		return "", fmt.Errorf("invalid server address %s", address)
	}
	httpPortStr, ok := node.Metadata[localcommon.MetaHTTPKey]
	if !ok {
		httpPortStr = strs[1]
	}
	_, err := strconv.Atoi(httpPortStr)
	if err != nil {
		return "", fmt.Errorf("convert port %s to int failed, err %s", httpPortStr, err.Error())
	}
	return strs[0] + ":" + httpPortStr, nil
}
