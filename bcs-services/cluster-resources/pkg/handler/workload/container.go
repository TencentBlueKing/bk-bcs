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

package workload

import (
	"context"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/perm"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/resp"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListContainer 获取指定 Pod 容器列表
func (h *Handler) ListContainer(
	ctx context.Context, req *clusterRes.ContainerListReq, resp *clusterRes.CommonListResp,
) (err error) {
	if err := perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildListContainerAPIResp(ctx, req.ClusterID, req.Namespace, req.PodName)
	return err
}

// GetContainer 获取指定容器详情
func (h *Handler) GetContainer(
	ctx context.Context, req *clusterRes.ContainerGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err := perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildGetContainerAPIResp(ctx, req.ClusterID, req.Namespace, req.PodName, req.ContainerName)
	return err
}

// GetContainerEnvInfo 获取指定容器环境变量信息
func (h *Handler) GetContainerEnvInfo(
	ctx context.Context, req *clusterRes.ContainerGetReq, resp *clusterRes.CommonListResp,
) error {
	if err := perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}

	envResp, _, err := cli.NewPodCliByClusterID(ctx, req.ClusterID).ExecCommand(
		req.Namespace, req.PodName, req.ContainerName, []string{"/bin/sh", "-c", "env"},
	)
	if err != nil {
		return err
	}

	// 逐行解析 stdout，生成容器 env 信息
	envs := []map[string]interface{}{}
	for _, info := range strings.Split(envResp, "\n") {
		if len(info) == 0 {
			continue
		}
		key, val := stringx.Partition(info, "=")
		envs = append(envs, map[string]interface{}{
			"name": key, "value": val,
		})
	}
	resp.Data, err = pbstruct.MapSlice2ListValue(envs)
	return err
}
