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
	"fmt"
	"net/http"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
)

const (
	BcsArgocdManagedClusterID     = "BCS-ARGOCD-ClusterId"
	BcsArgocdManagedServerAddress = "BCS-ARGOCD-ServerAddress"
)

// NewWsTunnelServerCallback return a new WsTunnelServerCallback instance
func NewWsTunnelServerCallback() *WsTunnelServerCallback {
	wts := &WsTunnelServerCallback{
		cls: make(map[string]*ClusterInfo),
	}
	wts.tunnelServer = websocketDialer.New(
		wts.authorizer,
		websocketDialer.DefaultErrorWriter,
		wts.cleanCredential,
	)

	return wts
}

// WsTunnelServerCallback describe the callback for tunnel server
type WsTunnelServerCallback struct {
	tunnelServer *websocketDialer.Server

	clsMutex sync.RWMutex
	cls      map[string]*ClusterInfo
}

// GetTunnelServer return the tunnel server inside struct.
func (wts *WsTunnelServerCallback) GetTunnelServer() *websocketDialer.Server {
	return wts.tunnelServer
}

// GetClusterInfo return the ClusterInfo according to the given clusterID
func (wts *WsTunnelServerCallback) GetClusterInfo(clusterID string) (*ClusterInfo, error) {
	wts.clsMutex.RLock()
	defer wts.clsMutex.RUnlock()

	info, ok := wts.cls[clusterID]
	if !ok {
		return nil, fmt.Errorf("cluster %s info not found", clusterID)
	}

	return info, nil
}

// authorizer should implement the websocketDialer.Authorizer, return clientKey, isAuthed and error
// the request may come from proxy-peer or a managed-cluster for argocd-server
func (wts *WsTunnelServerCallback) authorizer(req *http.Request) (string, bool, error) {
	clusterID := req.Header.Get(BcsArgocdManagedClusterID)
	serverAddr := req.Header.Get(BcsArgocdManagedServerAddress)

	wts.updateCls(&ClusterInfo{
		ClusterID:     clusterID,
		ServerAddress: serverAddr,
	})

	return clusterID, true, nil
}

func (wts *WsTunnelServerCallback) updateCls(info *ClusterInfo) {
	if info == nil {
		return
	}

	wts.clsMutex.Lock()
	defer wts.clsMutex.Unlock()

	wts.cls[info.ClusterID] = info
}

// cleanCredential receive serverKey and do the clean work.
func (wts *WsTunnelServerCallback) cleanCredential(_ string) {
}

// ClusterInfo describe the info of managed-cluster for argocd-server
type ClusterInfo struct {
	ClusterID     string
	ServerAddress string
}
