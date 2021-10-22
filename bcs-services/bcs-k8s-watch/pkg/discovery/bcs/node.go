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
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/satori/go.uuid"
)

// ServiceNode is node info for bcs services.
type ServiceNode struct {
	Id   string
	Info types.ServerInfo
}

func NewServiceNode(info types.ServerInfo) ServiceNode {
	id := fmt.Sprintf("%s", uuid.NewV4())
	return ServiceNode{Id: id, Info: info}
}

func (n *ServiceNode) PrimaryKey() string {
	return n.Id
}

func (n *ServiceNode) Payload() []byte {
	result, _ := json.Marshal(n)
	return result
}

func (n *ServiceNode) OwnsPayload(payload []byte) bool {
	var resultNode ServiceNode
	err := json.Unmarshal(payload, &resultNode)
	if err != nil {
		return false
	}

	return resultNode.PrimaryKey() != "" &&
		resultNode.PrimaryKey() == n.PrimaryKey()
}
