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

package manager

import (
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-consoleproxy/console-proxy/config"
	"github.com/fsouza/go-dockerclient"
	"sync"
)

type manager struct {
	sync.RWMutex
	conf                *config.ConsoleProxyConfig
	dockerClient        *docker.Client
	connectedContainers map[string]bool
}

// NewManager create a Manager object
func NewManager(conf *config.ConsoleProxyConfig) Manager {
	return &manager{
		conf:                conf,
		connectedContainers: make(map[string]bool),
	}
}

// Start create docker client
func (m *manager) Start() error {
	var err error
	m.dockerClient, err = docker.NewClient(m.conf.DockerEndpoint)
	return err
}
