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

package mongo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

func TestMergePatch(t *testing.T) {
	strategy1 := &storage.NodeGroupMgrStrategy{
		Name:         "testStrategy1",
		Labels:       map[string]string{"test": "test"},
		ResourcePool: "resourcePool1",
		ReservedNodeGroup: &storage.GroupInfo{
			NodeGroupID: "111",
			ClusterID:   "111",
			Weight:      1,
		},
		ElasticNodeGroups: []*storage.GroupInfo{
			{
				NodeGroupID: "222",
				ClusterID:   "222",
				Weight:      1,
			},
			{
				NodeGroupID: "333",
				ClusterID:   "333",
				Weight:      2,
			},
		},
		Strategy: &storage.Strategy{
			Type:              "buffer",
			ScaleUpCoolDown:   0,
			ScaleUpDelay:      0,
			MinScaleUpSize:    0,
			ScaleDownDelay:    0,
			MaxIdleDelay:      0,
			ReservedTimeRange: "",
			Buffer: &storage.BufferStrategy{
				Low:  1,
				High: 2,
			},
		},
		Status: &storage.State{
			Status:      "normal",
			LastStatus:  "",
			Error:       "",
			Message:     "",
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		},
	}
	modifiedStrategy := &storage.NodeGroupMgrStrategy{
		Labels: map[string]string{"testModified": "testModified"},
	}
	mergeByte, err := MergePatch(strategy1, modifiedStrategy, false)
	assert.Nil(t, err)
	mergeStrategy := &storage.NodeGroupMgrStrategy{}
	err = json.Unmarshal(mergeByte, mergeStrategy)
	assert.Nil(t, err)
	assert.NotEqual(t, "", mergeStrategy.ResourcePool)
	mergeByte, err = MergePatch(strategy1, modifiedStrategy, true)
	assert.Nil(t, err)
	mergeStrategy2 := &storage.NodeGroupMgrStrategy{}
	err = json.Unmarshal(mergeByte, mergeStrategy2)
	assert.Nil(t, err)
	assert.Equal(t, "", mergeStrategy2.ResourcePool)
}
