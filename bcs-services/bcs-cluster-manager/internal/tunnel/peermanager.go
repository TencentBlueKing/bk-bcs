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
	cmcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"

	"github.com/asim/go-micro/v3/registry"
)

// PeerManager tunnel peer manager
type PeerManager struct {
	sync.Mutex
	ready           bool
	token           string
	urlFormat       string
	wssTunnelServer *websocketDialer.Server
	peers           map[string]bool
	discovery       *discovery.ModuleDiscovery
	cliTLS          *tls.Config
}

// NewPeerManager create peer manager
func NewPeerManager(
	opt *options.ClusterManagerOptions,
	cliTLS *tls.Config,
	dialerServer *websocketDialer.Server,
	disc *discovery.ModuleDiscovery) *PeerManager {
	// self peerID is ip:port
	dialerServer.PeerID = fmt.Sprintf("%s:%d", opt.Address, opt.HTTPPort)
	dialerServer.PeerToken = opt.Tunnel.PeerToken

	var urlPrefix string
	if cliTLS == nil {
		urlPrefix = "ws://"
	} else {
		urlPrefix = "wss://"
	}
	pm := &PeerManager{
		token:           dialerServer.PeerToken,
		urlFormat:       urlPrefix + "%s/clustermanager/v1/websocket/connect",
		wssTunnelServer: dialerServer,
		peers:           map[string]bool{},
		discovery:       disc,
		cliTLS:          cliTLS,
	}

	return pm
}

// Start start peer manager
func (pm *PeerManager) Start() error {
	if pm.discovery == nil {
		return fmt.Errorf("discovery is empty")
	}
	pm.discovery.RegisterEventHandler(pm.discoveryEventHandler)
	return nil
}

// discoveryEventHandler
func (pm *PeerManager) discoveryEventHandler(svcs []*registry.Service) {
	nodes := make([]*registry.Node, 0)
	for _, svc := range svcs {
		blog.V(3).Infof("merge discovery nodes %v version %s", svc.Nodes, svc.Version)
		nodes = append(nodes, svc.Nodes...)
	}
	servs := make([]string, 0)
	for _, node := range nodes {
		httpAddr, err := getHTTPEndpointFromMeta(node)
		if err != nil {
			blog.Warnf("get http endpoint from micro service node failed, err %s", err.Error())
		}
		servs = append(servs, httpAddr)
	}
	blog.V(3).Infof("discovery module %s servers %v", pm.discovery.GetModuleName(), servs)
	if err := pm.syncPeers(servs); err != nil {
		blog.Errorf("sync peers failed, err %s", err.Error())
	}
}

func getHTTPEndpointFromMeta(node *registry.Node) (string, error) {
	address := node.Address
	strs := strings.Split(address, ":")
	if len(strs) != 2 {
		return "", fmt.Errorf("invalid server address %s", address)
	}
	httpPortStr, ok := node.Metadata[cmcommon.MicroMetaKeyHTTPPort]
	if !ok {
		httpPortStr = strs[1]
	}
	_, err := strconv.Atoi(httpPortStr)
	if err != nil {
		return "", fmt.Errorf("convert port %s to int failed, err %s", httpPortStr, err.Error())
	}
	return strs[0] + ":" + httpPortStr, nil
}

// syncPeers sync peers status, add tunnels to new peers, remove tunnels from deleted peers
func (pm *PeerManager) syncPeers(servs []string) error {
	if len(servs) == 0 {
		return fmt.Errorf("syncPeers event can't discovery self")
	}
	pm.addRemovePeers(servs)
	return nil
}

// addRemovePeers add tunnels with new peers each other, remove tunnels from deleted peers
func (pm *PeerManager) addRemovePeers(servs []string) {
	pm.Lock()
	defer pm.Unlock()

	newSet := map[string]bool{}
	ready := false

	for _, serv := range servs {
		if serv == pm.wssTunnelServer.PeerID {
			ready = true
		} else {
			newSet[serv] = true
		}
	}

	toCreate, toDelete, _ := diff(newSet, pm.peers)

	// add new peers
	for _, peerServ := range toCreate {
		pm.wssTunnelServer.AddPeer(fmt.Sprintf(pm.urlFormat, peerServ), peerServ, pm.token, pm.cliTLS)
	}
	// remove deleted peers
	for _, ip := range toDelete {
		pm.wssTunnelServer.RemovePeer(ip)
	}

	pm.peers = newSet
	pm.ready = ready
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

// Stop stop peer manager
func (pm *PeerManager) Stop() {
	if pm.discovery != nil {
		pm.discovery.Stop()
	}
}
