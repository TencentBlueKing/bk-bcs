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

package tunnel

import (
	"reflect"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/tunnel"
)

// TestDiff test the diff func
func TestDiff(t *testing.T) {
	desired := map[string]bool{
		"node1": true,
		"node2": true,
		"node3": true,
	}
	actual := map[string]bool{
		"node2": true,
		"node3": true,
		"node4": true,
	}

	toCreate, toDelete, same := diff(desired, actual)
	if !reflect.DeepEqual(toCreate, []string{"node1"}) {
		t.Error("get an unexpected toCreate map")
	}
	if !reflect.DeepEqual(toDelete, []string{"node4"}) {
		t.Error("get an unexpected toDelete map")
	}
	if !reflect.DeepEqual(same, []string{"node2", "node3"}) {
		t.Error("get an unexpected same map")
	}
}

// TestSyncPeers test the syncPeers func
func TestSyncPeers(t *testing.T) {
	tunnelServer := tunnel.NewTunnelServer()
	pm := &peerManager{
		token:     tunnelServer.PeerToken,
		urlFormat: "wss://%s/usermanager/v1/websocket/connect",
		server:    tunnelServer,
		peers:     map[string]bool{},
	}
	err := pm.syncPeers([]string{})
	if err == nil {
		t.Error("register and discovery should always can discovery self")
	}
}
