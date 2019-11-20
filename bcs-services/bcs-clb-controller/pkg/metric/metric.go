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

package metric

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
)

// Resource for a kind of clb metric
type Resource interface {
	Register(container *restful.Container)
}

// ClbMetric clb metrics
type ClbMetric struct {
	container *restful.Container
	server    *http.Server
}

// NewClbMetric create clb metrics
func NewClbMetric(port int) *ClbMetric {
	address := fmt.Sprintf(":%d", port)
	container := restful.NewContainer()
	server := &http.Server{
		Addr:    address,
		Handler: container,
	}
	return &ClbMetric{
		container: container,
		server:    server,
	}
}

// Run start metric server
func (cm *ClbMetric) Run() error {
	return cm.server.ListenAndServe()
}

// Close close clb metric
func (cm *ClbMetric) Close() error {
	return cm.server.Close()
}

// RegisterResource register resource to clb metric
func (cm *ClbMetric) RegisterResource(mr Resource) {
	mr.Register(cm.container)
}
