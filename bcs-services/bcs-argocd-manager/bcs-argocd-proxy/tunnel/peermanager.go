/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tunnel

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	privateCommon "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/options"

	"go-micro.dev/v4/registry"
)

const (
	wsConnectURI = "/argocdmanager/v1/websocket/connect"
)

// NewPeerManager return a new PeerManager instance
func NewPeerManager(
	opt *options.ProxyOptions,
	cliTLS *tls.Config,
	tunnelServer *websocketDialer.Server,
	disc *discovery.ModuleDiscovery) *PeerManager {

	tunnelServer.PeerID = fmt.Sprintf("%s:%d", opt.Address, opt.HTTPPort)
	tunnelServer.PeerToken = opt.Tunnel.PeerToken
	return &PeerManager{
		peers:        make(map[string]bool),
		tunnelServer: tunnelServer,
		token:        tunnelServer.PeerToken,
		discovery:    disc,
		cliTLS:       cliTLS,
		protocol: func() string {
			if cliTLS == nil {
				return "ws://"
			}
			return "wss://"
		}(),
	}
}

// PeerManager manages the proxy-peer
type PeerManager struct {
	sync.Mutex
	peers map[string]bool

	token        string
	cliTLS       *tls.Config
	tunnelServer *websocketDialer.Server
	discovery    *discovery.ModuleDiscovery

	ready    bool
	protocol string
}

// Start the PeerManager
func (pm *PeerManager) Start() error {
	if pm.discovery == nil {
		return fmt.Errorf("discovery is empty")
	}

	pm.discovery.RegisterEventHandler(pm.discoveryEventHandler)
	return nil
}

// Stop the PeerManager
func (pm *PeerManager) Stop() {
	if pm.discovery != nil {
		pm.discovery.Stop()
	}
}

func (pm *PeerManager) discoveryEventHandler(servers []*registry.Service) {
	nodes := make([]*registry.Node, 0)

	for _, svc := range servers {
		blog.V(3).Infof("merge discovery nodes %v version %s", svc.Nodes, svc.Version)
		nodes = append(nodes, svc.Nodes...)
	}

	peers := make([]string, 0)
	for _, node := range nodes {
		addr, err := getHTTPEndpointFromMeta(node)
		if err != nil {
			blog.Warnf("get http endpoint from micro service node failed, err %s", err.Error())
			continue
		}

		peers = append(peers, addr)
	}

	blog.V(3).Infof("discovery module %s servers %v", pm.discovery.GetModuleName(), peers)
	if err := pm.syncPeers(peers); err != nil {
		blog.Errorf("sync peers failed, err %s", err.Error())
	}
}

// syncPeers sync peers status, add tunnels to new peers, remove tunnels from deleted peers
func (pm *PeerManager) syncPeers(peers []string) error {
	if len(peers) == 0 {
		return fmt.Errorf("syncPeers event can't discovery self")
	}

	pm.addRemovePeers(peers)
	return nil
}

// addRemovePeers add tunnels with new peers each other, remove tunnels from deleted peers
func (pm *PeerManager) addRemovePeers(peers []string) {
	pm.Lock()
	defer pm.Unlock()

	newSet := map[string]bool{}
	ready := false

	for _, peer := range peers {
		if peer == pm.tunnelServer.PeerID {
			ready = true
		} else {
			newSet[peer] = true
		}
	}

	toCreate, toDelete, _ := diff(newSet, pm.peers)

	// add new peers
	for _, peerServ := range toCreate {
		pm.tunnelServer.AddPeer(pm.getPeerUrl(peerServ), peerServ, pm.token, pm.cliTLS)
	}
	// remove deleted peers
	for _, ip := range toDelete {
		pm.tunnelServer.RemovePeer(ip)
	}

	pm.peers = newSet
	pm.ready = ready
}

func (pm *PeerManager) getPeerUrl(server string) string {
	return pm.protocol + server + wsConnectURI
}

// diff just compare and diff two map
func diff(desired, actual map[string]bool) ([]string, []string, []string) {
	var same, toCreate, toDelete []string
	for key := range desired {
		if actual[key] {
			same = append(same, key)
		} else {
			toCreate = append(toCreate, key)
		}
	}
	for key := range actual {
		if !desired[key] {
			toDelete = append(toDelete, key)
		}
	}
	return toCreate, toDelete, same
}

func getHTTPEndpointFromMeta(node *registry.Node) (string, error) {
	address := node.Address
	strs := strings.Split(address, ":")
	if len(strs) != 2 {
		return "", fmt.Errorf("invalid server address %s", address)
	}
	httpPortStr, ok := node.Metadata[privateCommon.MetaKeyHTTPPort]
	if !ok {
		httpPortStr = strs[1]
	}
	_, err := strconv.Atoi(httpPortStr)
	if err != nil {
		return "", fmt.Errorf("convert port %s to int failed, err %s", httpPortStr, err.Error())
	}
	return strs[0] + ":" + httpPortStr, nil
}
