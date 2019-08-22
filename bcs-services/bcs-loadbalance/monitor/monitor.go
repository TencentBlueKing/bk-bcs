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

package monitor

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
)

// Resource for a kind of clb metric
type Resource interface {
	Register(container *restful.Container)
}

// Monitor monitoring data for bcs-loadbalance
type Monitor struct {
	container *restful.Container
	server    *http.Server
}

// NewMonitor create clb monitor
func NewMonitor(addr string, port int) *Monitor {
	address := fmt.Sprintf("%s:%d", addr, port)
	container := restful.NewContainer()
	server := &http.Server{
		Addr:    address,
		Handler: container,
	}
	return &Monitor{
		container: container,
		server:    server,
	}
}

// Run start monitor server
func (cm *Monitor) Run() error {
	return cm.server.ListenAndServe()
}

// Close close monitor
func (cm *Monitor) Close() error {
	return cm.server.Close()
}

// RegisterResource register resource to monitor
func (cm *Monitor) RegisterResource(mr Resource) {
	mr.Register(cm.container)
}
