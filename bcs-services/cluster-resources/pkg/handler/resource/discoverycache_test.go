/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resource

import (
	"testing"

	"github.com/TencentBlueKing/gopkg/stringx"
	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestInvalidateDiscoveryCache(t *testing.T) {
	conf.G.Basic.CacheToken = stringx.Random(8)
	ctx := handler.NewInjectedContext("", "", "")
	req := clusterRes.InvalidateDiscoveryCacheReq{
		ProjectID: envs.TestProjectID, ClusterID: envs.TestClusterID, AuthToken: conf.G.Basic.CacheToken,
	}
	assert.Nil(t, New().InvalidateDiscoveryCache(ctx, &req, &clusterRes.CommonResp{}))
}
