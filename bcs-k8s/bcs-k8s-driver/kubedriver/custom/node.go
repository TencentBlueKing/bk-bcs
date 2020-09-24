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

package custom

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"

	uuid "github.com/satori/go.uuid"
)

// ServiceNode is node info for bcs services.
type ServiceNode types.ServerInfo

// NewServiceNode create kubedriver service node
func NewServiceNode(info types.ServerInfo) ServiceNode {
	return ServiceNode{
		IP:           info.IP,
		Port:         info.Port,
		ExternalIp:   info.ExternalIp,
		ExternalPort: info.ExternalPort,
		MetricPort:   info.MetricPort,
		HostName:     info.HostName,
		Scheme:       info.Scheme,
		Version:      info.Version,
		Cluster:      info.Cluster,
		Pid:          info.Pid,
	}
}

// PrimaryKey key for indexer
func (n *ServiceNode) PrimaryKey() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}

// Payload content length
func (n *ServiceNode) Payload() []byte {
	result, _ := json.Marshal(n)
	return result
}

// OwnsPayload xxxx
func (n *ServiceNode) OwnsPayload(payload []byte) bool {
	return true
}
