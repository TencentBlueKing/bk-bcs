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
	"strings"
	"sync"

	"github.com/json-iterator/go"

	"bk-bcs/bcs-common/common/RegisterDiscover"
	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/options"
)

// HTTPClientConfig is bcs inner service http client config struct.
type HTTPClientConfig struct {
	URL      string
	Scheme   string
	CAFile   string
	CertFile string
	KeyFile  string
	Password string
}

// InnerService is bcs inner service for discovery.
type InnerService struct {
	name            string
	mu              sync.RWMutex
	eventChan       <-chan *RegisterDiscover.DiscoverEvent
	servers         map[string]*HTTPClientConfig
	customEndpoints []string
}

// NewInnerService creates a new serviceName InnerService instance for discovery.
func NewInnerService(serviceName string, eventChan <-chan *RegisterDiscover.DiscoverEvent,
	customEndpoints []string) *InnerService {

	svc := &InnerService{
		name:            serviceName,
		eventChan:       eventChan,
		servers:         make(map[string]*HTTPClientConfig),
		customEndpoints: customEndpoints,
	}

	return svc
}

// Watch keeps watching service instance endpoints from ZK.
func (s *InnerService) Watch(bcsTLSConfig options.TLS) error {
	glog.Infof("start to watch service[%s] from zookeeper", s.name)

	for data := range s.eventChan {
		glog.Infof("received ZK event, Server: %+v", data.Server)
		if data.Err != nil {
			glog.Errorf("%s service discover failed, %+v", s.name, data.Err)
			continue
		}
		s.update(data.Server, bcsTLSConfig)
	}

	return nil
}

// Servers returns current available services instances.
func (s *InnerService) Servers() []*HTTPClientConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfgs := []*HTTPClientConfig{}

	for _, cfg := range s.servers {
		cfgs = append(cfgs, cfg)
	}

	return cfgs
}

func (s *InnerService) update(servers []string, bcsTLSConfig options.TLS) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(servers) == 0 {
		glog.Warnf("get non %s service from ZK, no service instance available", s.name)
		return
	}

	currentServers := make(map[string]string)

	if s.customEndpoints != nil {
		for _, address := range s.customEndpoints {
			scheme := "http"
			if strings.HasPrefix(address, "https") {
				scheme = "https"
			}
			currentServers[address] = ""

			if _, exists := s.servers[address]; !exists {
				config := HTTPClientConfig{
					URL:    address,
					Scheme: scheme,
				}

				// support https.
				if scheme == "https" {
					config.CAFile = bcsTLSConfig.CAFile
					config.CertFile = bcsTLSConfig.CertFile
					config.KeyFile = bcsTLSConfig.KeyFile
					config.Password = bcsTLSConfig.Password
				}
				s.servers[address] = &config
			}
		}
	} else {
		for _, server := range servers {
			serverInfo := types.ServerInfo{}
			if err := jsoniter.Unmarshal([]byte(server), &serverInfo); err != nil {
				glog.Errorf("json unmarshal %s server info failed, %+v", s.name, err)
				continue
			}

			if len(serverInfo.Scheme) == 0 || len(serverInfo.IP) == 0 || serverInfo.Port == 0 {
				glog.Errorf("got invalid %s server info: %s", s.name, server)
				continue
			}

			address := fmt.Sprintf("%s://%s:%d", serverInfo.Scheme, serverInfo.IP, serverInfo.Port)
			currentServers[address] = ""

			if _, exists := s.servers[address]; !exists {
				config := HTTPClientConfig{
					URL:    address,
					Scheme: serverInfo.Scheme,
				}

				// support https.
				if serverInfo.Scheme == "https" {
					config.CAFile = bcsTLSConfig.CAFile
					config.CertFile = bcsTLSConfig.CertFile
					config.KeyFile = bcsTLSConfig.KeyFile
					config.Password = bcsTLSConfig.Password
				}
				s.servers[address] = &config
			}
		}
	}

	for address := range s.servers {
		if _, exists := currentServers[address]; !exists {
			delete(s.servers, address)
			glog.Infof("delete %s server address[%s] synced from ZK from list", s.name, address)
		}
	}
	glog.Infof("update %s service addresses done, final: %+v", s.name, s.servers)
}
