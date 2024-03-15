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
 */

package multicluster

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	clusterresources "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestInBlacklist(t *testing.T) {
	blacklist := []config.ProjectClusterConf{
		{ProjectCode: "a", ClusterIDReg: ".*$"},
		{ProjectCode: "b", ClusterIDReg: "^(c|xxx)$"},
		{ProjectCode: "c", ClusterIDReg: ""},
	}
	config.G.MultiCluster.BlacklistForAPIServerQuery = blacklist

	// same projectCode, wild cluster id
	assert.True(t, inBlackList("a", []*clusterresources.ClusterNamespaces{{ClusterID: "xxx"}}),
		"InBlacklist failed")
	// same projectCode, match clusterID
	assert.True(t, inBlackList("b", []*clusterresources.ClusterNamespaces{{ClusterID: "c"}}),
		"InBlacklist failed")
	// same projectCode, match clusterID
	assert.True(t, inBlackList("b", []*clusterresources.ClusterNamespaces{{ClusterID: "xxx"}}),
		"InBlacklist failed")
	// same projectCode, not match clusterID
	assert.True(t, !inBlackList("b", []*clusterresources.ClusterNamespaces{{ClusterID: "a"}}),
		"InBlacklist failed")
	// different projectCode, not match clusterID
	assert.True(t, !inBlackList("x", []*clusterresources.ClusterNamespaces{{ClusterID: "xxx"}}),
		"InBlacklist failed")
	// same projectCode, empty clusterID
	assert.True(t, !inBlackList("c", []*clusterresources.ClusterNamespaces{{ClusterID: "xxx"}}),
		"InBlacklist failed")
}
