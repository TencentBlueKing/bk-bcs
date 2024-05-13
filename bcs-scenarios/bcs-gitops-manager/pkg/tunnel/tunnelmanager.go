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
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"go-micro.dev/v4/registry"
	"k8s.io/apimachinery/pkg/util/proxy"
)

// TunnelOptions for manage tunnel server
// nolint
type TunnelOptions struct {
	Context context.Context
	// ID to identify tunnel instance
	TunnelID    string
	TunnelToken string
	// Key for collecting cluster info
	ClusterAddressKey string
	// ConnectURL for tunnel cluster inter-connection
	ConnectURL string
	// ClientTLS for tunnel cluster inter-connection
	ClientTLS *tls.Config
	// Registry implementation depend on go-micro
	PeerServiceName string
	Registry        registry.Registry
	// Indexer for backend tunnel selection
	Indexer ClusterIndexer
}

// Validate options
func (topt *TunnelOptions) Validate() error {
	if len(topt.TunnelID) == 0 || len(topt.TunnelToken) == 0 {
		return fmt.Errorf("lost Tunnel identification")
	}
	if len(topt.ClusterAddressKey) == 0 {
		return fmt.Errorf("lost managed cluster indexer")
	}
	if len(topt.ConnectURL) == 0 || topt.ClientTLS == nil {
		return fmt.Errorf("lost tunnel cluster management information")
	}
	if len(topt.PeerServiceName) == 0 || topt.Registry == nil {
		return fmt.Errorf("lost tunnel cluster discovery")
	}
	if topt.Indexer == nil {
		return fmt.Errorf("lost backend cluster indexer")
	}
	return nil
}

// ClusterIndexer search specific clusterID for backend tunnel
type ClusterIndexer func(req *http.Request) (string, error)

// NewTunnelManager return a new TunnelManager instance
func NewTunnelManager(opt *TunnelOptions) *TunnelManager {
	tm := &TunnelManager{
		option:   opt,
		clusters: make(map[string]*ClusterInfo),
	}
	tm.tunnelSvr = websocketDialer.New(
		tm.authorizer, // tunnelManager get cluster info from authorizer
		websocketDialer.DefaultErrorWriter,
		tm.cleanCredential,
	)
	return tm
}

// TunnelManager holds tunnel and manage tunnel entry points(tranport).
// nolint
type TunnelManager struct {
	option *TunnelOptions
	// peer manager handle proxy mutual discovery
	peerMgr     *PeerManager
	peerMgrStop context.CancelFunc

	clsMutex sync.RWMutex
	// clusters mean different lower clusters/groups connecting to tunnel cluster
	clusters map[string]*ClusterInfo
	// tunnelServer holding websocket tunnel
	tunnelSvr *websocketDialer.Server
}

// Init tunnel manager
func (tm *TunnelManager) Init() error {
	if tm.option == nil {
		return fmt.Errorf("lost Tunnel Options")
	}
	if err := tm.option.Validate(); err != nil {
		return fmt.Errorf("option is invalid, %s", err.Error())
	}
	peerCxt, peerCancel := context.WithCancel(tm.option.Context)
	// init peer manager
	mgrOption := &PeerManagerOptions{
		Context:         peerCxt,
		PeerID:          tm.option.TunnelID,
		PeerToken:       tm.option.TunnelToken,
		PeerConnectURL:  tm.option.ConnectURL,
		PeerTLS:         tm.option.ClientTLS,
		PeerServiceName: tm.option.PeerServiceName,
		Discovery:       tm.option.Registry,
		Tunnel:          tm.tunnelSvr,
	}
	tm.peerMgr = NewPeerManager(mgrOption)
	tm.peerMgrStop = peerCancel
	return nil
}

// Start tunnel manager, non-blocking
func (tm *TunnelManager) Start() error {
	if err := tm.peerMgr.Start(); err != nil {
		return fmt.Errorf("tunnel peer handle err, %s", err.Error())
	}
	return nil
}

// GetTunnelServer return the tunnel server inside struct.
func (tm *TunnelManager) GetTunnelServer() *websocketDialer.Server {
	return tm.tunnelSvr
}

// ServeHTTP proxy specific flow to tunnel
func (tm *TunnelManager) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// feature(DeveloperJim): ClusterID indexer for multiple cluster selection
	clusterID, err := tm.option.Indexer(req)
	if err != nil {
		blog.Errorf("TunnelManager indexer clusterID failed, %s", err.Error())
		resp := fmt.Sprintf("Internal ClusterID error: %s", err.Error())
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(resp)) // nolint
		return
	}
	backendEntrypoint, err := tm.lookupBackendEntryPoint(clusterID)
	if err != nil {
		blog.Errorf("TunnelManager handle clusterID %s backend transport failed, %s", clusterID, err.Error())
		resp := fmt.Sprintf("Internal ClusterID %s transport err: %s", clusterID, err.Error())
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(resp)) // nolint
		return
	}
	// Serve to Backend
	blog.Infof("cluster %s serving RequestURI %s", clusterID, req.URL.RequestURI())
	backendEntrypoint.ServeHTTP(rw, req)
}

// lookupBackendTransport according clusterID, transport is entry point
// for backend cluster. Transport will be cache for reuse.
// NOTE: health check for transport?
func (tm *TunnelManager) lookupBackendEntryPoint(clusterID string) (*proxy.UpgradeAwareHandler, error) {
	if !tm.tunnelSvr.HasSession(clusterID) {
		return nil, fmt.Errorf("no session in tunnel")
	}
	tm.clsMutex.Lock()
	defer tm.clsMutex.Unlock()
	cluster, ok := tm.clusters[clusterID]
	if !ok {
		return nil, fmt.Errorf("no session proxy in tunnel")
	}
	// check transport in cache, build it if transport lost
	if cluster.MiddleTransport == nil {
		dialer := tm.tunnelSvr.Dialer(clusterID, time.Second*15)
		transport := &http.Transport{
			MaxIdleConnsPerHost: 10,
			DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
				return dialer(network, addr)
			},
			// feature: tls verify as client
			TLSClientConfig: tm.option.ClientTLS,
		}
		cluster.MiddleTransport = transport
		blog.Infof("tunnel manager build backend %s middle transport", clusterID)
	}
	// wrap transport with UpgradeHandler
	reqLocation := cluster.ServerAddress
	if !strings.HasSuffix(reqLocation, "/") {
		reqLocation += "/"
	}
	reqURL, err := url.Parse(reqLocation)
	if err != nil {
		blog.Errorf("TunnelManager lookup backend %s transport met mis-formate: %s. addr[%s]",
			clusterID, err.Error(), reqLocation)
		return nil, fmt.Errorf("backend location mis-format")
	}
	response := &responder{}
	upgrader := proxy.NewUpgradeAwareHandler(
		reqURL, cluster.MiddleTransport, true, false, response,
	)
	upgrader.UseRequestLocation = true
	blog.Infof("tunnel manager init cluster %s upgrade handler, %s", clusterID, reqURL.String())
	return upgrader, nil
}

// authorizer should implement the websocketDialer.Authorizer, return clientKey, isAuthed and error
// the request may come from proxy-peer or a managed-cluster for argocd-server
func (tm *TunnelManager) authorizer(req *http.Request) (string, bool, error) {
	clusterID := req.Header.Get(websocketDialer.ID)
	token := req.Header.Get(websocketDialer.Token)
	serverAddr := req.Header.Get(tm.option.ClusterAddressKey)

	// first check token
	if token != tm.option.TunnelToken {
		blog.Errorf("cluster %s token %s in Unauthorized", clusterID, token)
		return "", false, nil
	}

	tm.clsMutex.Lock()
	defer tm.clsMutex.Unlock()
	oldCls, ok := tm.clusters[clusterID]
	if !ok {
		tm.clusters[clusterID] = &ClusterInfo{
			ClusterID:     clusterID,
			ServerAddress: serverAddr,
		}
		blog.Infof("tunnel manager construct new cluster %s backend info", clusterID)
		return clusterID, true, nil
	}
	if oldCls.ServerAddress != serverAddr {
		// backend cluster instance switches,
		// reset tranport for reconnection
		oldCls.ServerAddress = serverAddr
		if oldCls.MiddleTransport != nil {
			oldCls.MiddleTransport.CloseIdleConnections()
			oldCls.MiddleTransport = nil
		}
		blog.Infof("tunnel manager found cluster %s backend server changed, clean middle transport", clusterID)
	}
	// nothing changed
	return clusterID, true, nil
}

// cleanCredential receive serverKey and do the clean work.
func (tm *TunnelManager) cleanCredential(_ string) {
}

// responder implements k8s.io/apimachinery/pkg/util/proxy.ErrorResponder
type responder struct{}

// Error implements ErrorResponder Error function
func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	blog.Errorf("serving %s failed, %s", req.URL.String(), err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// ClusterInfo describe the info of managed-cluster connecting from tunnel
type ClusterInfo struct {
	// identity for remote services
	ClusterID string
	// address for remote services
	ServerAddress string
	// transport for tunnel connecting
	MiddleTransport *http.Transport
}
