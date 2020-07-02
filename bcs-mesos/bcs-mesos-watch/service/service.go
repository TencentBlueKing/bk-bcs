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

package service

import (
	"fmt"
	"sync"

	"github.com/json-iterator/go"

	"github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
)

// HTTPClientConfig is bcs inner service http client config struct.
type HTTPClientConfig struct {
	// URL is http whole url.
	URL string

	// Scheme is http url scheme, http/https.
	Scheme string

	// CAFile is https root certificate authority file path.
	CAFile string

	// CertFile is https certificate file path.
	CertFile string

	// KeyFile is https key file path.
	KeyFile string

	// Password is certificate authority file password.
	Password string
}

// InnerService is bcs inner service for discovery.
type InnerService struct {
	name      string
	mu        sync.RWMutex
	eventChan <-chan *RegisterDiscover.DiscoverEvent
	servers   map[string]*HTTPClientConfig
}

// NewInnerService creates a new serviceName InnerService instance for discovery.
func NewInnerService(serviceName string, eventChan <-chan *RegisterDiscover.DiscoverEvent) *InnerService {
	svc := &InnerService{
		name:      serviceName,
		eventChan: eventChan,
		servers:   make(map[string]*HTTPClientConfig),
	}

	return svc
}

// Watch keeps watching service instance endpoints from ZK.
func (s *InnerService) Watch(cfg *types.CmdConfig) error {
	glog.Infof("start to watch service[%s] from ZK", s.name)

	for data := range s.eventChan {
		glog.Infof("received ZK event, Server: %+v", data.Server)
		if data.Err != nil {
			glog.Errorf("%s service discover failed, %+v", s.name, data.Err)
			continue
		}
		s.update(data.Server, cfg)
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

func (s *InnerService) update(servers []string, cfg *types.CmdConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(servers) == 0 {
		glog.Warnf("get non %s service from ZK, no service instance available", s.name)
		return
	}

	currentServers := make(map[string]string)

	for _, server := range servers {
		serverInfo := commtypes.ServerInfo{}
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
				config.CAFile = cfg.CAFile
				config.CertFile = cfg.CertFile
				config.KeyFile = cfg.KeyFile
				config.Password = cfg.PassWord
			}
			s.servers[address] = &config
		}
	}

	for address := range s.servers {
		if _, exists := currentServers[address]; !exists {
			delete(s.servers, address)
			glog.Infof("delete %s server old address[%s] synced from ZK", s.name, address)
		}
	}
	glog.Infof("update %s service addresses done, final: %+v", s.name, s.servers)
}
