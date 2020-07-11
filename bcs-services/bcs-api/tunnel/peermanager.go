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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
	"golang.org/x/net/context"
)

const (
	defaultPeerToken = "Mx9vWfTZea4MEzc7SlvB8aFl0NhmYQvZzEomOYypDMKkev34Q9kIyh32RjXXCIcn"
)

type peerManager struct {
	sync.Mutex
	ready     bool
	token     string
	urlFormat string
	server    *websocketDialer.Server
	peers     map[string]bool
}

type PeerRDiscover struct {
	rd      *RegisterDiscover.RegDiscover
	rootCxt context.Context
	cancel  context.CancelFunc
}

func StartPeerManager(conf *config.ApiServConfig, dialerServer *websocketDialer.Server) error {
	dialerServer.PeerID = fmt.Sprintf("%s:%d", conf.LocalIp, conf.Port)
	dialerServer.PeerToken = conf.PeerToken
	if dialerServer.PeerToken == "" {
		dialerServer.PeerToken = defaultPeerToken
		blog.Info("use default peer token: [%s]", dialerServer.PeerToken)
	}
	pm := &peerManager{
		token:     dialerServer.PeerToken,
		urlFormat: "wss://%s/bcsapi/v1/websocket/connect",
		server:    dialerServer,
		peers:     map[string]bool{},
	}

	peerRd := &PeerRDiscover{
		rd: RegisterDiscover.NewRegDiscoverEx(conf.RegDiscvSrv, 10*time.Second),
	}
	peerRd.rootCxt, peerRd.cancel = context.WithCancel(context.Background())

	if err := peerRd.rd.Start(); err != nil {
		blog.Error("fail to start register and discover bcs-api peers. err:%s", err.Error())
		return err
	}

	go peerRd.discoveryAndWatchPeer(pm)

	return nil
}

func (p *PeerRDiscover) discoveryAndWatchPeer(pm *peerManager) {
	key := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_APISERVER)
	blog.Infof("start discover service key %s", key)
	event, err := p.rd.DiscoverService(key)
	if err != nil {
		blog.Error("fail to register discover for api. err:%s", err.Error())
		p.cancel()
		os.Exit(1)
	}

	for {
		select {
		case eve := <-event:
			var peerServs []string
			for _, serv := range eve.Server {
				apiServ := new(types.APIServInfo)
				if err := json.Unmarshal([]byte(serv), apiServ); err != nil {
					blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
					continue
				}
				peerServ := fmt.Sprintf("%s:%d", apiServ.IP, apiServ.Port)
				peerServs = append(peerServs, peerServ)
			}

			err := pm.syncPeers(peerServs)
			if err != nil {
				blog.Errorf("failed to discovery and watch peers: %s", err.Error())
			}
		case <-p.rootCxt.Done():
			blog.Warn("zk register path %s and discover done", key)
			return
		}
	}
}

func (p *peerManager) syncPeers(servs []string) error {
	if len(servs) == 0 {
		return errors.New("syncPeers even can't discovery self")
	}

	p.addRemovePeers(servs)

	return nil
}

func (p *peerManager) addRemovePeers(servs []string) {
	p.Lock()
	defer p.Unlock()

	newSet := map[string]bool{}
	ready := false

	for _, serv := range servs {
		if serv == p.server.PeerID {
			ready = true
		} else {
			newSet[serv] = true
		}
	}

	toCreate, toDelete, _ := diff(newSet, p.peers)
	for _, peerServ := range toCreate {
		p.server.AddPeer(fmt.Sprintf(p.urlFormat, peerServ), peerServ, p.token)
	}
	for _, ip := range toDelete {
		p.server.RemovePeer(ip)
	}

	p.peers = newSet
	p.ready = ready
}

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
