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

package bcs

import (
	"fmt"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"

	"bk-bcs/bcs-common/common/RegisterDiscover"
	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/options"
	"strings"
)

// storage server address maintain on zk(ha)
// so, should watch zk to observe the change

type HTTPClientConfig struct {
	URL      string
	Scheme   string
	CAFile   string
	CertFile string
	KeyFile  string
	Password string
}

type StorageService struct {
	mutex     sync.RWMutex
	eventChan <-chan *RegisterDiscover.DiscoverEvent
	Servers   map[string]*HTTPClientConfig

	customStorageEndpoints []string
}

func GetStorageService(zkHosts string, bcsTLSConfig options.TLS, customStorageEndpoints []string) (*StorageService, error) {

	discovery := RegisterDiscover.NewRegDiscoverEx(zkHosts, 5*time.Second)
	if err := discovery.Start(); nil != err {
		return nil, fmt.Errorf("start get storage zk service failed! Error: %v", err)
	}

	// e.g.
	// zk: 127.0.0.11
	// zknode: bcs/services/endpoints/storage
	path := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_STORAGE)
	eventChan, err := discovery.DiscoverService(path)

	if err != nil {
		return nil, fmt.Errorf("discover storage service fail. Error: %s", err)
	}

	storageService := StorageService{
		customStorageEndpoints: customStorageEndpoints,
		eventChan:              eventChan,
		Servers:                make(map[string]*HTTPClientConfig),
	}

	go storageService.run(bcsTLSConfig)
	return &storageService, nil
}

func (s *StorageService) run(bcsTLSConfig options.TLS) {
	glog.Info("start to watch storage service from zk")
	for data := range s.eventChan {
		glog.Info("receive zk event. from zk got Server: %s", data.Server)
		if data.Err != nil {
			glog.Errorf("storage service discover fail. %s", data.Err.Error())
		}
		s.updateServers(data.Server, bcsTLSConfig)
	}

}

func (s *StorageService) updateServers(servers []string, bcsTLSConfig options.TLS) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(servers) == 0 {
		glog.Warn("Get non storage service from zk, no storage service server available")
		return
	}

	currentServers := make(map[string]string)
	if s.customStorageEndpoints != nil {
		for _, address := range s.customStorageEndpoints {
			scheme := "http"
			if strings.HasPrefix(address, "https") {
				scheme = "https"
			}
			currentServers[address] = ""

			if _, exists := s.Servers[address]; !exists {
				config := HTTPClientConfig{
					URL:    address,
					Scheme: scheme,
				}

				// support https
				if scheme == "https" {
					config.CAFile = bcsTLSConfig.CAFile
					config.CertFile = bcsTLSConfig.CertFile
					config.KeyFile = bcsTLSConfig.KeyFile
					config.Password = bcsTLSConfig.Password
				}
				s.Servers[address] = &config
			}
		}
	} else {
		for _, server := range servers {
			serverInfo := types.ServerInfo{}
			if err := jsoniter.Unmarshal([]byte(server), &serverInfo); err != nil {
				glog.Errorf("json Unmarshal storage info fail. %s", err)
			}

			if len(serverInfo.Scheme) == 0 || len(serverInfo.IP) == 0 || serverInfo.Port == 0 {
				glog.Errorf("got invalid storage server info: %s", server)
			}

			serverAddress := fmt.Sprintf("%s://%s:%d", serverInfo.Scheme, serverInfo.IP, serverInfo.Port)

			currentServers[serverAddress] = ""

			if _, exists := s.Servers[serverAddress]; !exists {
				config := HTTPClientConfig{
					URL:    serverAddress,
					Scheme: serverInfo.Scheme,
				}

				// support https
				if serverInfo.Scheme == "https" {
					config.CAFile = bcsTLSConfig.CAFile
					config.CertFile = bcsTLSConfig.CertFile
					config.KeyFile = bcsTLSConfig.KeyFile
					config.Password = bcsTLSConfig.Password
				}
				s.Servers[serverAddress] = &config
			}
		}
	}

	for serverAddress := range s.Servers {
		if _, exists := currentServers[serverAddress]; !exists {
			delete(s.Servers, serverAddress)
			glog.Infof("delete storage server address: %s from list, synced from zk", serverAddress)
		}
	}

	glog.Infof("after sync with zk, servers at finally: %s", s.Servers)
}
