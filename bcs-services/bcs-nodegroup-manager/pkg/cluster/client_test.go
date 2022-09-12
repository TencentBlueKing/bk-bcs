/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cluster

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster/mock"
	"github.com/stretchr/testify/assert"
)

func TestClient_ListClusterNodes(t *testing.T) {
	opts := &ClientOptions{
		Endpoint: "",
		Token:    "testing",
		Sender:   mock.NewMockRequester(),
	}
	cli := NewClient(opts)
	nodes, err := cli.ListClusterNodes("BCS-K8S-15202")
	assert.Nil(t, err)
	assert.NotNil(t, nodes)
	assert.NotEqual(t, 0, len(nodes))
}

func TestClient_UpdateNodeLabels(t *testing.T) {
	opts := &ClientOptions{
		Endpoint: "",
		Token:    "",
		Sender:   mock.NewMockRequester(),
	}
	cli := NewClient(opts)
	nodes, err := cli.ListClusterNodes("BCS-K8S-15202")
	assert.Nil(t, err)
	assert.NotNil(t, nodes)
	assert.NotEqual(t, 0, len(nodes))
	testNode := nodes[0]
	updateLabels := map[string]string{"testupdate3": "testupdate3"}
	err = cli.UpdateNodeLabels("BCS-K8S-15202", testNode.Name, updateLabels)
	assert.Nil(t, err)
}
