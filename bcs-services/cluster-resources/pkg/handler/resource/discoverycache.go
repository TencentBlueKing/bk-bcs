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
	"context"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// InvalidateDiscoveryCache 清理集群 Discovery 缓存内容，慎用
func (h *Handler) InvalidateDiscoveryCache(
	ctx context.Context, req *clusterRes.InvalidateDiscoveryCacheReq, resp *clusterRes.CommonResp,
) error {
	cli, err := res.NewRedisCacheClient4Conf(ctx, res.NewClusterConfig(req.ClusterID))
	if err != nil {
		return err
	}
	return cli.ClearCache()
}
